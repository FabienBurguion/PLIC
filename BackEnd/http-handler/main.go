package main

import (
	"PLIC/database"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

type Service struct {
	db database.Database
}

func (s *Service) InitService() {
	s.db = initDb()
}

func initDb() database.Database {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var version string
	if err := db.QueryRow("select version()").Scan(&version); err != nil {
		panic(err)
	}
	fmt.Printf("version=%s\n", version)
	return database.Database{
		Database: sqlx.NewDb(db, "postgres"),
	}
}

var mux = http.NewServeMux()

type methodHandlers struct {
	get  func(w http.ResponseWriter, _ *http.Request) error
	post func(w http.ResponseWriter, _ *http.Request) error
}

var handlers = make(map[string]*methodHandlers)

func (s *Service) GET(path string, handlerFunc func(w http.ResponseWriter, _ *http.Request) error) {
	if handlers[path] == nil {
		handlers[path] = &methodHandlers{}
		mux.HandleFunc(path, handleRequest)
	}
	handlers[path].get = handlerFunc
}

func (s *Service) POST(path string, handlerFunc func(w http.ResponseWriter, _ *http.Request) error) {
	if handlers[path] == nil {
		handlers[path] = &methodHandlers{}
		mux.HandleFunc(path, handleRequest)
	}
	handlers[path].post = handlerFunc
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	handler := handlers[r.URL.Path]
	if handler == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if handler.get != nil {
			_ = handler.get(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	case http.MethodPost:
		if handler.post != nil {
			_ = handler.post(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func Start(port string) {
	_ = http.ListenAndServe(port, mux)
}

func main() {
	s := &Service{}
	s.InitService()

	s.GET("/hello_world", s.GetHelloWorld)

	fmt.Println("Server running on port 8080...")
	Start(":8080")
}
