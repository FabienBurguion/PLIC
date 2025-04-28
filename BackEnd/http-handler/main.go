package main

import (
	"PLIC/database"
	"PLIC/mailer"
	"PLIC/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/gin-gonic/gin"
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
	db       database.Database
	server   *http.ServeMux
	clock    Clock
	mailer   *mailer.Mailer
	s3Client *s3.Client
}

func (s *Service) InitService() {
	s.db = initDb()
	s.server = http.NewServeMux()
	s.clock = Clock{offset: time.Hour}
	s.mailer = &mailer.Mailer{
		LastSentAt:  make(map[string]time.Time),
		AlreadySent: make(map[string]bool),
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println("Failed to load SDK config:", err)
	} else {
		s.s3Client = s3.NewFromConfig(cfg)
	}
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

// DOCS ONLY: fake router for swaggo
func docsRouter() {
	r := gin.New()
	s := &Service{}
	s.InitService()
	r.POST("/register", func(c *gin.Context) {
		_ = s.Register(c.Writer, c.Request, models.AuthInfo{})
	})
	r.POST("/login", func(c *gin.Context) {
		_ = s.Login(c.Writer, c.Request, models.AuthInfo{})
	})

	r.GET("/", func(c *gin.Context) {
		_ = s.GetTime(c.Writer, c.Request, models.AuthInfo{})
	})
	r.GET("/hello_world", func(c *gin.Context) {
		_ = s.GetHelloWorld(c.Writer, c.Request, models.AuthInfo{})
	})

	r.POST("/email", func(c *gin.Context) {
		_ = s.SendMail(c.Writer, c.Request, models.AuthInfo{})
	})

	r.POST("/image", func(c *gin.Context) {
		_ = s.UploadImageToS3(c.Writer, c.Request, models.AuthInfo{})
	})

	r.GET("/image", func(c *gin.Context) {
		_ = s.GetS3Image(c.Writer, c.Request, models.AuthInfo{})
	})
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

	// ENDPOINTS FOR EMAIL
	s.POST("/email", s.SendMail)

	// ENDPOINTS FOR S3
	s.POST("/image", s.UploadImageToS3)
	s.GET("/image", s.GetS3Image)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		fmt.Println("🚀 Démarrage sur AWS Lambda...")
		s.Start()
	} else {
		fmt.Println("🌍 Démarrage en local sur le port " + Port + "...")
		log.Fatal(http.ListenAndServe(":"+Port, s.server))
	}
}
