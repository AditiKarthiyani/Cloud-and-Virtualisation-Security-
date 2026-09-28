package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type msg struct {
	Message string `json:"message"`
}

type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, msg{Message: "Hello from backend"})
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	var m msg
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func dbStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"db": "unreachable", "error": err.Error()})
		return
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM todos").Scan(&count); err != nil {
		count = 0
	}
	writeJSON(w, http.StatusOK, map[string]any{"db": "ok", "row_count": count, "time": time.Now().Format(time.RFC3339)})
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listTodos(w, r)
	case http.MethodPost:
		createTodo(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid todo id"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		getTodo(w, r, id)
	case http.MethodPut:
		updateTodo(w, r, id)
	case http.MethodDelete:
		deleteTodo(w, r, id)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func listTodos(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT id, title, done, created_at FROM todos ORDER BY id")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	todos := make([]Todo, 0)
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		todos = append(todos, t)
	}
	writeJSON(w, http.StatusOK, todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.Title) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	var todo Todo
	err := db.QueryRowContext(ctx,
		"INSERT INTO todos (title, done) VALUES ($1, false) RETURNING id, title, done, created_at",
		strings.TrimSpace(input.Title)).Scan(&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, todo)
}

func getTodo(w http.ResponseWriter, r *http.Request, id int) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	var todo Todo
	if err := db.QueryRowContext(ctx, "SELECT id, title, done, created_at FROM todos WHERE id = $1", id).
		Scan(&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func updateTodo(w http.ResponseWriter, r *http.Request, id int) {
	var input struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.Title) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	res, err := db.ExecContext(ctx, "UPDATE todos SET title = $1, done = $2 WHERE id = $3", strings.TrimSpace(input.Title), input.Done, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if count, _ := res.RowsAffected(); count == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
		return
	}
	getTodo(w, r, id)
}

func deleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	res, err := db.ExecContext(ctx, "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if count, _ := res.RowsAffected(); count == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func initDB() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env("DB_USER", "postgres"), env("DB_PASSWORD", "postgres"),
		env("DB_HOST", "localhost"), env("DB_PORT", "5432"), env("DB_NAME", "exercise"))
	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("waiting for db (%d/10): %v", i, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id serial PRIMARY KEY,
		title text NOT NULL,
		done boolean NOT NULL DEFAULT false,
		created_at timestamptz NOT NULL DEFAULT now()
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDB()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/hello", withCORS(helloHandler))
	mux.HandleFunc("/api/echo", withCORS(echoHandler))
	mux.HandleFunc("/api/dbstatus", withCORS(dbStatusHandler))
	mux.HandleFunc("/api/todos", withCORS(todosHandler))
	mux.HandleFunc("/api/todos/", withCORS(todoHandler))

	srv := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	log.Println("starting backend on :8081")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
