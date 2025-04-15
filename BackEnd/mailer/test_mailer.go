package mailer

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"time"
)

func (mailer *Mailer) SendTestMail(to string) error {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: No .env file found, using environment variables")
		}
	}

	from := os.Getenv("SMTP_FROM")
	host := os.Getenv("SMTP_HOST")
	port := 587
	fmt.Sscanf(os.Getenv("SMTP_PORT"), "%d", &port)
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	if mailer.AlreadySent[to] && time.Since(mailer.LastSentAt[to]) < 10*time.Second {
		log.Println("⛔️ Email déjà envoyé récemment à", to, "→ annulation.")
		return fmt.Errorf("\"⛔️ Email déjà envoyé récemment à\", to, \"→ annulation.\"")
	}

	log.Println("🚀 Envoi de l'email via Mailjet à", to)

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Hello from Go + Mailjet ✅")
	m.SetBody("text/plain", "Coucou, voici un mail sans tomber en spam 🚀")

	d := gomail.NewDialer(host, port, username, password)

	if err := d.DialAndSend(m); err != nil {
		log.Println("❌ Échec de l'envoi à", to, ":", err)
		return err
	}

	mailer.AlreadySent[to] = true
	mailer.LastSentAt[to] = time.Now()

	log.Println("📤 Email envoyé avec succès à", to)
	return nil
}
