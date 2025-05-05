package mailer

import (
	"PLIC/models"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"time"
)

type MailerInterface interface {
	SendTestMail(to string) error
	SendPasswordForgotMail(to string, newPassword string) error
}

type Mailer struct {
	AlreadySent map[string]bool
	LastSentAt  map[string]time.Time
	Config      *models.MailerConfig
}

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

func (mailer *Mailer) SendPasswordForgotMail(to string, newPassword string) error {
	if mailer.AlreadySent[to] && time.Since(mailer.LastSentAt[to]) < time.Minute {
		log.Println("⛔️ Email déjà envoyé récemment à", to, "→ annulation.")
		return fmt.Errorf("\"⛔️ Email déjà envoyé récemment à\", to, \"→ annulation.\"")
	}

	log.Println("🚀 Envoi de l'email via Mailjet à", to)

	m := gomail.NewMessage()
	m.SetHeader("From", mailer.Config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Récupération de mot de passe")

	textBody := "Votre nouveau mot de passe est : " + newPassword + ""
	htmlBody := fmt.Sprintf(`
	<html>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<div style="max-width: 600px; margin: auto; background: white; padding: 20px; border-radius: 8px;">
				<h2 style="color: #333;">Bonjour,</h2>
				<p style="font-size: 16px;">Voici votre nouveau mot de passe :</p>
				<p style="font-size: 18px; font-weight: bold; background-color: #eee; padding: 10px; border-radius: 4px; text-align: center;">
					%s
				</p>
				<p style="font-size: 14px; color: #555;">Vous pouvez le modifier après vous être connecté.</p>
				<hr style="margin: 20px 0;">
				<small style="color: #888;">Envoyé automatiquement depuis une application Go.</small>
			</div>
		</body>
	</html>
`, newPassword)

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
