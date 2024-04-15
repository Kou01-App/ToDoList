// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// mkasm.go generates assembly trampolines to call library routines from Go.
// This program must be run after mksyscall.go.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func archPtrSize(arch string) int {
	switch arch {
	case "386", "arm":
		return 4
	case "amd64", "arm64", "mips64", "ppc64", "riscv64":
		return 8
	default:
		log.Fatalf("Unknown arch %q", arch)
		return 0
	}
}

func generateASMFile(goos, arch string, inFileNames []string, outFileName string) map[string]bool {
	trampolines := map[string]bool{}
	var orderedTrampolines []string
	for _, inFileName := range inFileNames {
		in, err := os.ReadFile(inFileName)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		for _, line := range strings.Split(string(in), "\n") {
			const prefix = "var "
			const suffix = "_trampoline_addr uintptr"
			if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
				continue
			}
			fn := strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix)
			if !trampolines[fn] {
				orderedTrampolines = append(orderedTrampolines, fn)
				trampolines[fn] = true
			}
		}
	}

	ptrSize := archPtrSize(arch)

	var out bytes.Buffer
	fmt.Fprintf(&out, "// go run mkasm.go %s\n", strings.Join(os.Args[1:], " "))
	fmt.Fprintf(&out, "// Code generated by the command above; DO NOT EDIT.\n")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "#include \"textflag.h\"\n")
	for _, fn := range orderedTrampolines {
		fmt.Fprintf(&out, "\nTEXT %s_trampoline<>(SB),NOSPLIT,$0-0\n", fn)
		if goos == "openbsd" && arch == "ppc64" {
			fmt.Fprintf(&out, "\tCALL\t%s(SB)\n", fn)
			fmt.Fprintf(&out, "\tRET\n")
		} else {
			fmt.Fprintf(&out, "\tJMP\t%s(SB)\n", fn)
		}
		fmt.Fprintf(&out, "GLOBL\t·%s_trampoline_addr(SB), RODATA, $%d\n", fn, ptrSize)
		fmt.Fprintf(&out, "DATA\t·%s_trampoline_addr(SB)/%d, $%s_trampoline<>(SB)\n", fn, ptrSize, fn)
	}

	if err := os.WriteFile(outFileName, out.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write assembly file %q: %v", outFileName, err)
	}

	return trampolines
}

const darwinTestTemplate = `// go run mkasm.go %s
// Code generated by the command above; DO NOT EDIT.

//go:build darwin && go1.12

package unix

// All the _trampoline functions in zsyscall_darwin_%s.s.
var darwinTests = [...]darwinTest{
%s}
`

func writeDarwinTest(trampolines map[string]bool, fileName, arch string) {
	var sortedTrampolines []string
	for fn := range trampolines {
		sortedTrampolines = append(sortedTrampolines, fn)
	}
	sort.Strings(sortedTrampolines)

	var out bytes.Buffer

	const prefix = "libc_"
	for _, fn := range sortedTrampolines {
		fmt.Fprintf(&out, fmt.Sprintf("\t{%q, %s_trampoline_addr},\n", strings.TrimPrefix(fn, prefix), fn))
	}
	lines := out.String()

	out.Reset()
	fmt.Fprintf(&out, darwinTestTemplate, strings.Join(os.Args[1:], " "), arch, lines)

	if err := os.WriteFile(fileName, out.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write test file %q: %v", fileName, err)
	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <goos> <arch>", os.Args[0])
	}
	goos, arch := os.Args[1], os.Args[2]

	syscallFilename := fmt.Sprintf("syscall_%s.go", goos)
	syscallArchFilename := fmt.Sprintf("syscall_%s_%s.go", goos, arch)
	zsyscallArchFilename := fmt.Sprintf("zsyscall_%s_%s.go", goos, arch)
	zsyscallASMFileName := fmt.Sprintf("zsyscall_%s_%s.s", goos, arch)

	inFileNames := []string{
		syscallFilename,
		syscallArchFilename,
		zsyscallArchFilename,
	}

	trampolines := generateASMFile(goos, arch, inFileNames, zsyscallASMFileName)

	if goos == "darwin" {
		writeDarwinTest(trampolines, fmt.Sprintf("darwin_%s_test.go", arch), arch)
	}
}
