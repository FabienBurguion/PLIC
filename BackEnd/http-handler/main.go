package main

import (
	"PLIC/database"
	"PLIC/httpx"
	"database/sql"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const Port string = "8080"

type Service struct {
	db     database.Database
	server *http.ServeMux
	clock  Clock
}

func (s *Service) InitService() {
	s.db = initDb()
	s.server = http.NewServeMux()
	s.clock = Clock{offset: time.Hour}
}

func initDb() database.Database {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: No .env file found, using environment variables")
		}
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	var version string
	if err := db.QueryRow("select version()").Scan(&version); err != nil {
		panic(err)
	}
	fmt.Printf("version=%s\n", version)
	return database.Database{
		Database: sqlx.NewDb(db, "postgres"),
	}
}

type methodHandlers struct {
	get  func(w http.ResponseWriter, _ *http.Request) error
	post func(w http.ResponseWriter, _ *http.Request) error
}

var handlers = make(map[string]*methodHandlers)

func (s *Service) GET(path string, handlerFunc func(_ http.ResponseWriter, _ *http.Request) error) {
	if handlers[path] == nil {
		handlers[path] = &methodHandlers{}
		s.server.HandleFunc(path, handleRequest)
	}
	handlers[path].get = handlerFunc
}

func (s *Service) POST(path string, handlerFunc func(_ http.ResponseWriter, _ *http.Request) error) {
	if handlers[path] == nil {
		handlers[path] = &methodHandlers{}
		s.server.HandleFunc(path, handleRequest)
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
			_ = httpx.WriteError(w, http.StatusMethodNotAllowed, httpx.MethodNotAllowedError)
		}
	case http.MethodPost:
		if handler.post != nil {
			_ = handler.post(w, r)
		} else {
			_ = httpx.WriteError(w, http.StatusMethodNotAllowed, httpx.MethodNotAllowedError)
		}
	default:
		_ = httpx.WriteError(w, http.StatusMethodNotAllowed, httpx.MethodNotAllowedError)
	}
}

func (s *Service) Start() {
	log.Println("🚀 Serveur démarré sur AWS Lambda")
	lambdaHandler := httpadapter.NewV2(s.server)
	lambda.Start(lambdaHandler.ProxyWithContext)
}

func main() {
	s := &Service{}
	s.InitService()

	s.GET("/", s.GetHelloWorldBasic)
	s.GET("/hello_world", s.GetHelloWorld)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		fmt.Println("🚀 Démarrage sur AWS Lambda...")
		s.Start()
	} else {
		fmt.Println("🌍 Démarrage en local sur le port " + Port + "...")
		log.Fatal(http.ListenAndServe(":"+Port, s.server))
	}
}
