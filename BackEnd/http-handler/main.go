package main

import (
	"PLIC/database"
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

	var version string
	if err := db.QueryRow("select version()").Scan(&version); err != nil {
		panic(err)
	}
	fmt.Printf("version=%s\n", version)
	return database.Database{
		Database: sqlx.NewDb(db, "postgres"),
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

	//LOGIN
	s.POST("/register", s.Register)
	s.POST("/login", s.Login)

	// ENDPOINTS FOR TESTING PURPOSE
	s.GET("/", withAuthentication(s.GetTime))
	s.GET("/hello_world", withAuthentication(s.GetHelloWorld))

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		fmt.Println("🚀 Démarrage sur AWS Lambda...")
		s.Start()
	} else {
		fmt.Println("🌍 Démarrage en local sur le port " + Port + "...")
		log.Fatal(http.ListenAndServe(":"+Port, s.server))
	}
}
