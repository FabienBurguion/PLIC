package mailer

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"time"
)

func (mailer *Mailer) SendTestMail(to string) error {
	if mailer.AlreadySent[to] && time.Since(mailer.LastSentAt[to]) < time.Minute {
		log.Println("⛔️ Email déjà envoyé récemment à", to, "→ annulation.")
		return fmt.Errorf("\"⛔️ Email déjà envoyé récemment à\", to, \"→ annulation.\"")
	}

	log.Println("🚀 Envoi de l'email via Mailjet à", to)

	m := gomail.NewMessage()
	m.SetHeader("From", mailer.Config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Hello from Go + Mailjet ✅")

	textBody := "Coucou, voici un mail sans tomber en spam 🚀"
	htmlBody := `
	<html>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<div style="max-width: 600px; margin: auto; background: white; padding: 20px; border-radius: 8px;">
				<h2 style="color: #333;">Salut 👋</h2>
				<p style="font-size: 16px;">Voici un <strong>email stylisé en HTML</strong> envoyé avec Go et Mailjet.</p>
				<p>🚀 Profite bien de ta journée !</p>
				<hr style="margin: 20px 0;">
				<small style="color: #888;">Envoyé automatiquement depuis une app Go.</small>
			</div>
		</body>
	</html>
`

	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(mailer.Config.Host, mailer.Config.Port, mailer.Config.Username, mailer.Config.Password)

	if err := d.DialAndSend(m); err != nil {
		log.Println("❌ Échec de l'envoi à", to, ":", err)
		return err
	}

	mailer.AlreadySent[to] = true
	mailer.LastSentAt[to] = time.Now()

	log.Println("📤 Email envoyé avec succès à", to)
	return nil
}
