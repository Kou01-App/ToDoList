import axios from "axios"
import { useState, useEffect } from "react"
import "./ToDoList.css"

function ToDoList() {
	const [completedTodos, setCompletedTodos] = useState([])
	const [todos, setTodos] = useState([])
	const [todo, setTodo] = useState("")

	function handlePush() {
		if (todo !== "") {
			axios
				.post("http://localhost:8000/api/tasks", {
					name: todo,
				})
				.then(() => {
					handleGetTodos()
				})
			setTodo("")
		}
	}
	function handleComplete(index) {
		console.log("取得したidid:", todos[index].ID);
		axios.put(`http://localhost:8000/api/tasks/` + todos[index].ID).then(() => {
			handleGetTodos()
		})
	}
	function handleDelete(index) {
		axios
			.delete(`http://localhost:8000/api/tasks/` + completedTodos[index].ID)
			.then(() => {
				handleGetTodos()
			})
	}
	function handleGetTodos() {
		axios.get("http://localhost:8000/api/tasks").then((res) => {
			let tmpCompletedTodos = []
			let tmpTodos = []
			for (let i = 0; i < res.data.length; i++) {
				if (res.data[i].Finished) {
					tmpCompletedTodos = tmpCompletedTodos.concat([
						{
							ID: res.data[i].ID,
							name: res.data[i].Name,
						},
					])
				} else {
					console.log("取得したname:", res.data[i].Name);
					tmpTodos = tmpTodos.concat([
						{
							ID: res.data[i].ID,
							name: res.data[i].Name,
						},
					])
				}
			}
			setCompletedTodos(tmpCompletedTodos)
			setTodos(tmpTodos)
			console.log("取得したcom:", completedTodos);
			console.log("取得したto:", todos);
		})
	}
	useEffect(() => {
		handleGetTodos()
	}, []) // eslint-disable-line react-hooks/exhaustive-deps

	return (
		<div className="todoList">
			<h1 className="title">ToDoリスト</h1>
			<input value={todo} onChange={(e) => setTodo(e.target.value)}/>
			<button onClick={handlePush} className="button">
				登録
			</button>
			<h2>タスク一覧</h2>
			<h3>完</h3>
			<ul className="ul">
				{completedTodos.map((completedTodo, index) => (
					<li key={completedTodo.id} className="li">
						{completedTodo.name}
						<button onClick={() => handleDelete(index)} className="button">
							タスクを消す
						</button>
					</li>
				))}
			</ul>

			<h3>未完</h3>
			<ul className="ul">
				{todos.map((todo, index) => (
					<li key={todo.id} className="li">
						{todo.name}
						<button onClick={() => handleComplete(index)} className="button">
							完了する
						</button>
					</li>
				))}
			</ul>
		</div>
	)
}

export default ToDoList