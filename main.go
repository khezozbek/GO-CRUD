package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	// Connect to the PostgreSQL database
	connectionString := "postgres://postgres:password@localhost/todo_app?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	// Create the 'todos' table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			description TEXT
		);
	`)
	if err != nil {
		fmt.Println("Error creating table:", err)
		os.Exit(1)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		todo := r.FormValue("todo")
		if todo != "" {
			_, err := db.Exec("INSERT INTO todos (description) VALUES ($1)", todo)
			if err != nil {
				http.Error(w, "Error adding todo", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := db.Query("SELECT description FROM todos")
	if err != nil {
		http.Error(w, "Error querying todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []string
	for rows.Next() {
		var todo string
		err := rows.Scan(&todo)
		if err != nil {
			http.Error(w, "Error scanning todos", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, struct{ Todos []string }{todos})
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		todo := r.FormValue("todo")
		tmpl := template.Must(template.ParseFiles("templates/todo_detail.html"))
		tmpl.Execute(w, todo)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		todo := r.FormValue("todo")
		_, err := db.Exec("DELETE FROM todos WHERE description = $1", todo)
		if err != nil {
			http.Error(w, "Error deleting todo", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/todo", todoHandler)
	http.HandleFunc("/delete", deleteHandler)

	http.ListenAndServe(":8080", nil)
}
