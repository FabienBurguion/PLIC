package main

import (
	"PLIC/database"
	"PLIC/mailer"
	"PLIC/models"
	"PLIC/s3_management"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/caarlos0/env/v10"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"
	ddchi "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-chi/chi.v5"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const Port string = "8080"

const DefaultElo = 1000
const KFactor = 32

type Service struct {
	db            database.Database
	server        *chi.Mux
	clock         Clock
	mailer        mailer.MailSender
	s3Service     s3_management.S3Service
	configuration *models.Configuration
	isLambda      bool
}

func loadConfig() (*models.Configuration, error) {
	var cfg models.Configuration
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Service) initService() {
	_ = godotenv.Load()

	appConfig, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	s.configuration = appConfig
	s.isLambda = s.configuration.Lambda.FunctionName != ""

	s.db = s.initDb()

	s.server = chi.NewRouter()

	if s.isLambda {
		s.server.Use(ddchi.Middleware(
			ddchi.WithServiceName("plic-api"),
		))
	}

	s.server.Use(middleware.Logger)

	s.server.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Starting Request ➡️ %s %s", r.Method, r.URL.Path)
			rw := &models.ResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rw, r)
			elapsed := time.Since(start)

			if s.isLambda {
				ddlambda.Metric(
					"plic.http.request_ms",
					float64(elapsed.Milliseconds()),
					"method:"+r.Method,
					"path:"+r.URL.Path,
				)
			}

			log.Printf("Ending Result ➡️ %s %s", r.Method, r.URL.Path)
			log.Printf("Status Code: %d, Duration: %s", rw.StatusCode, elapsed)
		})
	})

	s.server.Use(middleware.Recoverer)
	s.server.Use(middleware.RequestID)
	s.server.Use(middleware.Timeout(10 * time.Second))
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
		log.Println("Failed to load AWS SDK config:", err)
	} else {
		if s.isLambda {
			awstrace.AppendMiddleware(&cfg)
		}
		s3Client := s3.NewFromConfig(cfg)
		s.s3Service = &s3_management.RealS3Service{Client: s3Client}
	}
}

func (s *Service) initDb() database.Database {
	if !s.isLambda {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: No .env file found, using environment variables")
		}
	}

	connStr := s.configuration.Database.ConnectionString
	if connStr == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		panic(err)
	}

	var version string
	if err := db.QueryRow("select version()").Scan(&version); err != nil {
		panic(err)
	}
	fmt.Printf("version=%s\n", version)

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	return database.Database{
		Database: sqlx.NewDb(db, "pgx"),
	}
}

func (s *Service) Start() {
	log.Println("🚀 Serveur démarré sur AWS Lambda (Datadog enabled)")
	lambdaHandler := httpadapter.NewV2(s.server)
	lambda.Start(ddlambda.WrapFunction(lambdaHandler.ProxyWithContext, nil))
}

func main() {
	s := &Service{}
	s.initService()

	// Routes
	// LOGIN
	s.POST("/register", s.Register)
	s.POST("/login", s.Login)
	s.POST("/forgot-password", s.ForgetPassword)
	s.GET("/reset-password/{token}", s.ResetPassword)
	s.POST("/change-password", withAuthentication(s.ChangePassword))

	// TEST
	s.GET("/", withAuthentication(s.GetTime))
	s.GET("/hello_world", s.GetHelloWorld)

	// EMAIL
	s.POST("/email", withAuthentication(s.SendMail))

	// S3
	s.POST("/profile_picture/{id}", withAuthentication(s.UploadProfilePictureToS3))

	// GOOGLE
	s.POST("/place", s.HandleSyncGooglePlaces)

	// COURTS
	s.GET("/court/all", withAuthentication(s.GetAllCourts))
	s.GET("/court/{id}", withAuthentication(s.GetCourtByID))

	// MATCHES
	s.GET("/match/all", withAuthentication(s.GetAllMatches))
	s.GET("/match/{id}", withAuthentication(s.GetMatchByID))
	s.GET("/user/matches/{userId}", withAuthentication(s.GetMatchesByUserID))
	s.GET("/matches/court/{courtId}", withAuthentication(s.GetMatchesByCourtId))
	s.POST("/match", withAuthentication(s.CreateMatch))
	s.POST("/join/match/{id}", withAuthentication(s.JoinMatch))
	s.PATCH("/score/match/{id}", withAuthentication(s.UpdateMatchScore))
	s.DELETE("/match/{id}", withAuthentication(s.DeleteMatch))
	s.PATCH("/match/{id}/start", withAuthentication(s.StartMatch))
	s.PATCH("/match/{id}/finish", withAuthentication(s.FinishMatch))

	// USERS
	s.GET("/users/{id}", withAuthentication(s.GetUserById))
	s.PATCH("/users/{id}", withAuthentication(s.PatchUser))
	s.DELETE("/users/{id}", withAuthentication(s.DeleteUser))

	// RANKINGS
	s.GET("/ranking/court/{id}", withAuthentication(s.GetRankingByCourtId))
	s.GET("/ranking/user/{userId}", withAuthentication(s.GetUserFields))

	if s.isLambda {
		tracer.Start(
			tracer.WithService("plic-api"),
		)
		defer tracer.Stop()

		fmt.Println("🚀 Démarrage sur AWS Lambda...")
		s.Start()
	} else {
		// Pas de Datadog en local
		fmt.Println("🌍 Démarrage en local sur le port " + Port + "...")
		log.Fatal(http.ListenAndServe(":"+Port, s.server))
	}
}
