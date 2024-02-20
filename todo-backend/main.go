package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Todo struct {
	ID        int    `json:"id"`
	Task      string `json:"task"`
	Completed bool   `json:"completed"`
}

var todos []Todo

func main() {
	http.HandleFunc("/todos", handleTodos)
	http.HandleFunc("/todos/", updateTodoCompletion)
	log.Fatal(http.ListenAndServe(":8000", addCorsHeaders(http.DefaultServeMux)))
}

func addCorsHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin with the specified methods.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			// Preflight request, respond with success status
			return
		}

		h.ServeHTTP(w, r)
	})
}

func handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTodos(w, r)
	case http.MethodPost:
		addTodo(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
	}
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	// Connect to MySQL database
	db, err := sql.Open("mysql", "root:admin&1898@tcp(localhost:3306)/todo_db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query todos from database
	rows, err := db.Query("SELECT id, task, completed FROM todos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create slice to store todos
	var todos []Todo

	// Iterate through rows and populate todos slice
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Task, &todo.Completed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode todos slice as JSON and send in response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Connect to MySQL database
	db, err := sql.Open("mysql", "root:admin&1898@tcp(localhost:3306)/todo_db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert todo into database
	_, err = db.Exec("INSERT INTO todos (task, completed) VALUES (?, ?)", todo.Task, todo.Completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the added todo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func updateTodoCompletion(w http.ResponseWriter, r *http.Request) {
	// Parse todo ID from URL path
	todoID := r.URL.Path[len("/todos/"):]
	id, err := strconv.Atoi(todoID)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	// Decode JSON request body containing updated todo
	var updatedTodo Todo
	if err := json.NewDecoder(r.Body).Decode(&updatedTodo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Connect to MySQL database
	db, err := sql.Open("mysql", "root:admin&1898@tcp(localhost:3306)/todo_db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Update completion status of todo in the database
	_, err = db.Exec("UPDATE todos SET Completed = ? WHERE ID = ?", updatedTodo.Completed, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the updated todo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTodo)
}
