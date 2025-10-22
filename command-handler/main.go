package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
)

type App struct {
	db *sqlx.DB
}

func main() {
	// Logger façon PLIC
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// ✅ Charge automatiquement ton fichier .env (à la racine du projet)
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("⚠️  Aucun fichier .env trouvé, utilisation des variables d'environnement existantes")
	} else {
		log.Info().Msg("✅ Fichier .env chargé")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal().Msg("❌ DATABASE_URL manquant dans le .env")
	}

	// Connexion DB (avec Datadog trace, comme ton main PLIC)
	dbStd, err := sqltrace.Open("pgx", connStr, sqltrace.WithServiceName("plic-db"))
	if err != nil {
		log.Fatal().Err(err).Msg("échec ouverture connexion SQL tracée")
	}
	db := sqlx.NewDb(dbStd, "pgx")

	// Vérification simple
	var version string
	if err := dbStd.QueryRow("select version()").Scan(&version); err != nil {
		log.Fatal().Err(err).Msg("échec lecture version Postgres")
	}
	log.Info().Str("postgres_version", version).Msg("✅ Connexion DB OK")

	dbStd.SetConnMaxLifetime(5 * time.Minute)
	dbStd.SetMaxIdleConns(10)
	dbStd.SetMaxOpenConns(20)

	app := &App{db: db}

	// --- CLI ---
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}
	cmd := os.Args[1]

	switch cmd {
	case "create-match":
		fs := flag.NewFlagSet("create-match", flag.ExitOnError)
		var (
			userID       string
			courtID      string
			sport        string
			participants int
			team         int
			matchID      string
		)
		fs.StringVar(&userID, "user-id", "", "ID du créateur (obligatoire)")
		fs.StringVar(&courtID, "court-id", "", "ID du court (obligatoire)")
		fs.StringVar(&sport, "sport", "basket", "Sport (basket|foot|ping-pong)")
		fs.IntVar(&participants, "participants", 2, "Nombre de participants")
		fs.IntVar(&team, "team", 1, "Équipe du créateur (1 ou 2)")
		fs.StringVar(&matchID, "match-id", "", "UUID du match (optionnel)")
		_ = fs.Parse(os.Args[2:])

		if userID == "" || courtID == "" {
			fs.Usage()
			os.Exit(2)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := RunCreateMatch(ctx, app.db, CreateMatchOptions{
			MatchID:      matchID,
			UserID:       userID,
			CourtID:      courtID,
			Sport:        sport,
			Participants: participants,
			Team:         team,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("❌ create-match a échoué")
		}
		log.Info().Msg("✅ create-match terminé avec succès")

	default:
		log.Error().Str("cmd", cmd).Msg("commande inconnue")
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println(`Usage:
  go run ./command-handler create-match --user-id <uuid> --court-id <uuid> [options]

Options:
  --sport <basket|foot|ping-pong>   Sport (défaut: basket)
  --participants <n>                Nombre de participants (défaut: 2)
  --team <1|2>                      Équipe du créateur (défaut: 1)
  --match-id <uuid>                 Forcer un UUID spécifique (optionnel)`)
}
