package main

import (
	"PLIC/database"
	"PLIC/mailer"
	"PLIC/models"
	"PLIC/s3_management"
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/caarlos0/env/v10"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

const Port string = "8080"

type Service struct {
	db            database.Database
	server        *chi.Mux
	clock         Clock
	mailer        mailer.MailerInterface
	s3Service     s3_management.S3Service
	configuration *models.Configuration
}

func LoadConfig() (*models.Configuration, error) {
	var cfg models.Configuration
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Service) InitService() {
	_ = godotenv.Load()
	appConfig, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	s.configuration = appConfig

	s.db = s.initDb()
	s.server = chi.NewRouter()
	s.server.Use(middleware.Logger)
	s.server.Use(middleware.Recoverer)
	s.server.Use(middleware.RequestID)
	s.server.Use(middleware.Timeout(5 * time.Second))
	s.server.Use(middleware.Heartbeat("/ping"))

	parisLocation, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}

	s.clock = Clock{location: parisLocation}

	s.mailer = &mailer.Mailer{
		LastSentAt:  make(map[string]time.Time),
		AlreadySent: make(map[string]bool),
		Config:      &appConfig.Mailer,
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println("Failed to load SDK config:", err)
	} else {
		s3Client := s3.NewFromConfig(cfg)
		s.s3Service = &s3_management.RealS3Service{Client: s3Client}
	}
}

func (s *Service) initDb() database.Database {
	if s.configuration.Lambda.FunctionName == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: No .env file found, using environment variables")
		}
	}

	connStr := s.configuration.Database.ConnectionString
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
	s.POST("/register", withAuthentication(s.Register))
	s.POST("/login", withAuthentication(s.Login))
	s.POST("/forgot-password", withAuthentication(s.ForgetPassword))
	s.GET("/reset-password/{token}", withAuthentication(s.ResetPassword))
	s.POST("/change-password", withAuthentication(s.ChangePassword))

	// ENDPOINTS FOR TESTING PURPOSE
	s.GET("/", withAuthentication(s.GetTime))
	s.GET("/hello_world", withAuthentication(s.GetHelloWorld))

	// ENDPOINTS FOR EMAIL - TESTING
	s.POST("/email", withAuthentication(s.SendMail))

	// ENDPOINTS FOR S3
	s.POST("/image", withAuthentication(s.UploadImageToS3))
	s.GET("/image", withAuthentication(s.GetS3Image))
	s.POST("/profile_picture/{id}", withAuthentication(s.UploadProfilePictureToS3))

	// GOOGLE
	s.POST("/place", withAuthentication(s.HandleSyncGooglePlaces))

	// ENDPOINTS FOR TERRAINS
	s.GET("/court/all", withAuthentication(s.GetAllTerrains))

	// ENDPOINTS FOR MATCHES
	s.GET("/match/all", withAuthentication(s.GetAllMatches))
	s.GET("/match/{id}", withAuthentication(s.GetMatchByID))
	s.GET("/user/matches/{userId}", withAuthentication(s.GetMatchesByUserID))
	s.POST("/match", withAuthentication(s.CreateMatch))
	s.POST("/join/match/{id}", withAuthentication(s.JoinMatch))
	s.DELETE("/match/{id}", withAuthentication(s.DeleteMatch))

	//ENDPOINTS FOR USERS
	s.GET("/users/{id}", withAuthentication(s.GetUserById))
	s.PATCH("/users/{id}", withAuthentication(s.PatchUser))
	s.DELETE("/users/{id}", withAuthentication(s.DeleteUser))

	if s.configuration.Lambda.FunctionName != "" {
		fmt.Println("🚀 Démarrage sur AWS Lambda...")
		s.Start()
	} else {
		fmt.Println("🌍 Démarrage en local sur le port " + Port + "...")
		log.Fatal(http.ListenAndServe(":"+Port, s.server))
	}
}
