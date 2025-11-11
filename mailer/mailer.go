package mailer

import (
	"PLIC/models"
	"fmt"
	"log"
	"time"

	"gopkg.in/gomail.v2"
)

type MailSender interface {
	SendLinkResetPassword(to string, newPassword string) error
	SendWelcomeEmail(to string, username string) error
}

type Mailer struct {
	AlreadySent map[string]bool
	LastSentAt  map[string]time.Time
	Config      *models.MailerConfig
}

func (mailer *Mailer) SendLinkResetPassword(to string, url string) error {
	if mailer.AlreadySent[to] && time.Since(mailer.LastSentAt[to]) < 10*time.Second {
		log.Println("⛔️ Email déjà envoyé récemment à", to, "→ annulation.")
		return fmt.Errorf("⛔️ Email déjà envoyé récemment à %s → annulation", to)
	}

	log.Println("🚀 Envoi de l'email de réinitialisation à", to)

	m := gomail.NewMessage()
	m.SetHeader("From", mailer.Config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Réinitialisation de votre mot de passe")

	textBody := fmt.Sprintf(`Bonjour,

	Vous avez demandé à réinitialiser votre mot de passe.
	
	Veuillez cliquer sur le lien suivant pour définir un nouveau mot de passe (valable 15 minutes) :
	%s
	
	Si vous n'êtes pas à l'origine de cette demande, vous pouvez ignorer cet e-mail.
	
	Cordialement,
	L'équipe Support`, url)

	htmlBody := fmt.Sprintf(`
	<html>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<div style="max-width: 600px; margin: auto; background: white; padding: 20px; border-radius: 8px;">
				<h2 style="color: #333;">Demande de réinitialisation de mot de passe</h2>
				<p style="font-size: 16px;">Vous avez demandé à réinitialiser votre mot de passe.</p>
				<p style="font-size: 16px;">Cliquez sur le lien ci-dessous pour définir un nouveau mot de passe. <strong>Ce lien est valide pendant 15 minutes.</strong></p>
				<p style="text-align: center; margin: 20px 0;">
					<a href="%s" style="display: inline-block; background-color: #007BFF; color: white; padding: 12px 20px; text-decoration: none; border-radius: 5px;">
						Réinitialiser mon mot de passe
					</a>
				</p>
				<p style="font-size: 14px; color: #555;">Si vous n'êtes pas à l'origine de cette demande, vous pouvez ignorer cet e-mail.</p>
				<hr style="margin: 20px 0;">
				<small style="color: #888;">Cet e-mail a été envoyé automatiquement par notre application Go.</small>
			</div>
		</body>
	</html>
	`, url)

	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(mailer.Config.Host, mailer.Config.Port, mailer.Config.Username, mailer.Config.Password)

	if err := d.DialAndSend(m); err != nil {
		log.Println("❌ Échec de l'envoi à", to, ":", err)
		return err
	}

	mailer.AlreadySent[to] = true
	mailer.LastSentAt[to] = time.Now()

	log.Println("📤 Email de réinitialisation envoyé à", to)
	return nil
}

func (mailer *Mailer) SendWelcomeEmail(to string, username string) error {
	key := to + ":welcome"
	if mailer.AlreadySent[key] && time.Since(mailer.LastSentAt[key]) < time.Minute {
		log.Println("⛔️ Email de bienvenue déjà envoyé récemment à", to, "→ annulation.")
		return fmt.Errorf("⛔️ Email de bienvenue déjà envoyé récemment à %s → annulation", to)
	}

	log.Println("🚀 Envoi de l'email de bienvenue à", to)

	m := gomail.NewMessage()
	m.SetHeader("From", mailer.Config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("Bienvenue sur Play The Street, %s 🏀", username))

	textBody := fmt.Sprintf(`Salut %s,

Bienvenue sur Play The Street ! 🙌
Tu peux maintenant :
• Créer ou rejoindre des matchs
• Suivre tes stats & victoires
• Découvrir les terrains près de chez toi

À très vite sur le terrain !
L’équipe Play The Street`, username)

	htmlBody := fmt.Sprintf(`
	<html>
		<body style="margin:0;padding:0;background:#0E0E0E;font-family: Inter, Arial, sans-serif;">
			<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="padding:24px 0;">
				<tr>
					<td align="center">
						<table role="presentation" width="600" cellpadding="0" cellspacing="0" style="width:600px;max-width:94%%;">
							<tr>
								<td align="center" style="padding:8px 0 20px 0;">
									<!-- Si logo embarqué -->
									<!-- <img src="cid:logoPTS" alt="Play The Street" width="120" style="display:block;"> -->
									<!-- Sinon un titre simple -->
									<div style="font-size:22px;color:#FF6A00;font-weight:700;letter-spacing:.3px;">
										PLAY THE STREET
									</div>
								</td>
							</tr>

							<tr>
								<td style="background:#1A1A1A;border-radius:16px;padding:24px 22px;border:1px solid #2B2B2B;">
									<h1 style="margin:0 0 12px 0;font-size:22px;line-height:28px;color:#EDEDED;">
										Bienvenue, %s 👋
									</h1>
									<p style="margin:0 0 16px 0;font-size:15px;line-height:22px;color:#CFCFCF;">
										Ravi de te compter parmi nous. À partir d’aujourd’hui, tu peux <strong style="color:#FFFFFF;">créer</strong> ou <strong style="color:#FFFFFF;">rejoindre</strong> des matchs, suivre tes <strong style="color:#FFFFFF;">stats</strong> et découvrir les <strong style="color:#FFFFFF;">meilleurs terrains</strong> autour de toi.
									</p>

									<table role="presentation" cellpadding="0" cellspacing="0" style="margin:18px 0 6px 0;">
										<tr>
											<td style="background:#2B2B2B;border-radius:12px;padding:14px 16px;border:1px solid #3A3A3A;">
												<ul style="padding-left:18px;margin:0;color:#D9D9D9;font-size:14px;line-height:22px;">
													<li>Crée un match en 10 secondes</li>
													<li>Invite tes amis ou des joueurs proches</li>
													<li>Suis tes victoires et ton classement</li>
												</ul>
											</td>
										</tr>
									</table>

									<div style="text-align:center;margin:22px 0 8px 0;">
										<a href="#" 
											style="display:inline-block;padding:12px 18px;border-radius:999px;background:#FF6A00;color:#0E0E0E;text-decoration:none;font-weight:700;">
											Ouvrir l’app
										</a>
									</div>

									<p style="margin:14px 0 0 0;font-size:12px;color:#9A9A9A;text-align:center;">
										Si le bouton ne fonctionne pas, ouvre directement l’application Play The Street.
									</p>
								</td>
							</tr>

							<tr>
								<td style="padding:18px 6px 0 6px;text-align:center;color:#7A7A7A;font-size:12px;">
									© %d Play The Street • Paris, France
								</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</body>
	</html>
	`, username, time.Now().Year())

	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(mailer.Config.Host, mailer.Config.Port, mailer.Config.Username, mailer.Config.Password)
	if err := d.DialAndSend(m); err != nil {
		log.Println("❌ Échec de l'envoi (welcome) à", to, ":", err)
		return err
	}

	mailer.AlreadySent[key] = true
	mailer.LastSentAt[key] = time.Now()

	log.Println("📤 Email de bienvenue envoyé à", to)
	return nil
}
