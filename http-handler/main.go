package main

import (
	"PLIC/database"
	"PLIC/mailer"
	"PLIC/models"
	"PLIC/s3_management"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/caarlos0/env/v10"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"
	ddchi "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-chi/chi.v5"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
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

// ----------------------
// CONFIGURATION
// ----------------------

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
		log.Fatal().Err(err).Msg("failed to load configuration")
	}
	s.configuration = appConfig

	s.isLambda = s.configuration.Lambda.FunctionName != ""

	if s.isLambda {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		})
	}

	s.db = s.initDb()
	s.server = chi.NewRouter()

	if s.isLambda {
		s.server.Use(ddchi.Middleware(ddchi.WithServiceName("plic-api")))
	}

	s.server.Use(middleware.Logger)

	s.server.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &models.ResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}

			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("➡️ Incoming request")

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

			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rw.StatusCode).
				Dur("duration", elapsed).
				Msg("✅ Request completed")
		})
	})

	s.server.Use(middleware.Recoverer)
	s.server.Use(middleware.RequestID)
	s.server.Use(middleware.Heartbeat("/ping"))

	parisLocation, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load Paris timezone")
	}
	s.clock = Clock{location: parisLocation}

	s.mailer = &mailer.Mailer{
		LastSentAt:  make(map[string]time.Time),
		AlreadySent: make(map[string]bool),
		Config:      &appConfig.Mailer,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Error().Err(err).Msg("failed to load AWS SDK config")
	} else {
		if s.isLambda {
			awstrace.AppendMiddleware(&cfg)
		}
		s3Client := s3.NewFromConfig(cfg)
		s.s3Service = &s3_management.RealS3Service{Client: s3Client}
		log.Info().Msg("S3 service initialized successfully")
	}
}

// ----------------------
// DATABASE
// ----------------------

func (s *Service) initDb() database.Database {
	if !s.isLambda {
		if err := godotenv.Load(); err != nil {
			log.Warn().Msg("no .env file found, using environment variables")
		}
	}

	connStr := s.configuration.Database.ConnectionString
	if connStr == "" {
		log.Fatal().Msg("DATABASE_URL environment variable is not set")
	}

	db, err := sqltrace.Open("pgx", connStr, sqltrace.WithServiceName("plic-db"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open traced SQL connection")
	}

	var version string
	if err := db.QueryRow("select version()").Scan(&version); err != nil {
		log.Fatal().Err(err).Msg("failed to query DB version")
	}
	log.Info().Str("postgres_version", version).Msg("database connection established")

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	return database.Database{
		Database: sqlx.NewDb(db, "pgx"),
	}
}

// ----------------------
// START
// ----------------------

func (s *Service) Start() {
	log.Info().Msg("🚀 Starting server on AWS Lambda (Datadog enabled)")
	lambdaHandler := httpadapter.NewV2(s.server)
	lambda.Start(ddlambda.WrapFunction(lambdaHandler.ProxyWithContext, nil))
}

// ----------------------
// MAIN
// ----------------------

func main() {
	s := &Service{}
	s.initService()

	s.POST("/register", s.Register)
	s.POST("/login", s.Login)
	s.POST("/forgot-password", s.ForgetPassword)
	s.GET("/reset-password/{token}", s.ResetPassword)
	s.POST("/change-password", withAuthentication(s.ChangePassword))

	s.GET("/", withAuthentication(s.GetTime))
	s.GET("/hello_world", s.GetHelloWorld)

	s.POST("/profile_picture", withAuthentication(s.UploadProfilePictureToS3))

	s.POST("/place", s.HandleSyncGooglePlaces)

	s.GET("/court/all", withAuthentication(s.GetAllCourts))
	s.GET("/court/{id}", withAuthentication(s.GetCourtByID))

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

	s.GET("/users/{id}", withAuthentication(s.GetUserById))
	s.PATCH("/users/{id}", withAuthentication(s.PatchUser))
	s.DELETE("/users/{id}", withAuthentication(s.DeleteUser))

	s.GET("/ranking/court/{id}", withAuthentication(s.GetRankingByCourtId))
	s.GET("/ranking/user/{userId}", withAuthentication(s.GetUserFields))

	if s.isLambda {
		tracer.Start(tracer.WithService("plic-api"))
		defer tracer.Stop()

		log.Info().Msg("🚀 Running in AWS Lambda mode...")
		s.Start()
	} else {
		log.Info().Str("port", Port).Msg("🌍 Running locally...")
		if err := http.ListenAndServe(":"+Port, s.server); err != nil {
			log.Fatal().Err(err).Msg("server failed")
		}
	}
}
