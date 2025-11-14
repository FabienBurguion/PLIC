package mailer

import (
	"PLIC/models"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gopkg.in/gomail.v2"
)

type MailSender interface {
	SendLinkResetPassword(to string, url string) error
	SendWelcomeEmail(userId string, to string, username string) error
	SendMatchResultEmail(matchId string, to string, username string, sport models.Sport, fieldName string, teamScore, oppScore int) error
}

type Mailer struct {
	AlreadySent map[string]bool
	LastSentAt  map[string]time.Time
	Config      *models.MailerConfig
}

func (mailer *Mailer) dialer() *gomail.Dialer {
	d := gomail.NewDialer(mailer.Config.Host, mailer.Config.Port, mailer.Config.Username, mailer.Config.Password)
	d.TLSConfig = &tls.Config{
		ServerName: mailer.Config.Host,
		MinVersion: tls.VersionTLS12,
	}
	return d
}

func (mailer *Mailer) setCommonHeaders(m *gomail.Message, subject, to string) {
	from := mailer.Config.From
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetHeader("Reply-To", "support@tondomaine.tld")
	msgID := fmt.Sprintf("<%s@%s>", uuid.NewString(), "tondomaine.tld")
	m.SetHeader("Message-ID", msgID)
}

func (mailer *Mailer) SendLinkResetPassword(to string, url string) error {
	baseLogger := log.With().
		Str("mail_kind", "reset_password").
		Str("to", to).
		Logger()

	if mailer.AlreadySent[to] && time.Since(mailer.LastSentAt[to]) < 10*time.Second {
		baseLogger.Warn().
			Dur("since_last", time.Since(mailer.LastSentAt[to])).
			Msg("email recently sent → throttled")
		return fmt.Errorf("email recently sent to %s → throttled", to)
	}

	baseLogger.Info().Msg("sending reset password email")

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

	start := time.Now()
	d := mailer.dialer()

	if err := d.DialAndSend(m); err != nil {
		baseLogger.Error().Err(err).Dur("latency", time.Since(start)).Msg("mail send failed")
		return err
	}

	if mailer.AlreadySent == nil {
		mailer.AlreadySent = map[string]bool{}
	}
	if mailer.LastSentAt == nil {
		mailer.LastSentAt = map[string]time.Time{}
	}

	mailer.AlreadySent[to] = true
	mailer.LastSentAt[to] = time.Now()

	baseLogger.Info().Dur("latency", time.Since(start)).Msg("mail sent successfully")
	return nil
}

func (mailer *Mailer) SendWelcomeEmail(userId string, to string, username string) error {
	key := userId + ":welcome"

	baseLogger := log.With().
		Str("mail_kind", "welcome").
		Str("to", to).
		Str("username", username).
		Logger()

	if mailer.AlreadySent[key] && time.Since(mailer.LastSentAt[key]) < time.Minute {
		baseLogger.Warn().
			Dur("since_last", time.Since(mailer.LastSentAt[key])).
			Msg("welcome email recently sent → throttled")
		return fmt.Errorf("welcome email recently sent to %s → throttled", to)
	}

	baseLogger.Info().Msg("sending welcome email")

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
						
						<!-- LOGO / HEADER -->
						<tr>
							<td align="center" style="padding:8px 0 20px 0;">
								<div style="font-size:22px;color:#FF6A00;font-weight:700;letter-spacing:.3px;">
									PLAY THE STREET
								</div>
							</td>
						</tr>

						<!-- CARD -->
						<tr>
							<td style="background:#1A1A1A;border-radius:16px;padding:28px 22px;border:1px solid #2B2B2B;">
								
								<!-- TITRE CENTRÉ -->
								<h1 style="
									margin:0 0 16px 0;
									font-size:24px;
									line-height:30px;
									color:#EDEDED;
									text-align:center;
									font-weight:600;
								">
									Bienvenue, %s 👋
								</h1>

								<!-- Texte intro -->
								<p style="margin:0 0 22px 0;font-size:15px;line-height:22px;color:#CFCFCF;text-align:center;">
									Ravi de te compter parmi nous. À partir d’aujourd’hui, tu peux créer ou rejoindre des matchs,
									suivre tes stats, et découvrir les meilleurs terrains autour de toi.
								</p>

								<!-- LISTE CENTRÉE -->
								<table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 26px auto;">
									<tr>
										<td style="
											background:#2B2B2B;
											border-radius:12px;
											padding:14px 16px;
											border:1px solid #3A3A3A;
											text-align:center;
										">
											<ul style="
												list-style:none;
												padding:0;
												margin:0;
												color:#D9D9D9;
												font-size:14px;
												line-height:22px;
												text-align:center;
											">
												<li>Crée un match en 10&nbsp;secondes</li>
												<li>Invite des joueurs près de chez toi</li>
												<li>Suis tes victoires et ton classement</li>
											</ul>
										</td>
									</tr>
								</table>

								<!-- TEXTE MOTIVANT -->
								<p style="
									margin:12px 0 0 0;
									font-size:14px;
									line-height:22px;
									color:#FF6A00;
									text-align:center;
									font-weight:600;
								">
									🔥 Ne perds pas une seconde : ton prochain match t’attend déjà.
								</p>

							</td>
						</tr>

						<!-- FOOTER -->
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

	start := time.Now()
	d := mailer.dialer()

	if err := d.DialAndSend(m); err != nil {
		baseLogger.Error().Err(err).Dur("latency", time.Since(start)).Msg("mail send failed")
		return err
	}

	if mailer.AlreadySent == nil {
		mailer.AlreadySent = map[string]bool{}
	}
	if mailer.LastSentAt == nil {
		mailer.LastSentAt = map[string]time.Time{}
	}

	mailer.AlreadySent[key] = true
	mailer.LastSentAt[key] = time.Now()

	baseLogger.Info().Dur("latency", time.Since(start)).Msg("mail sent successfully")
	return nil
}

func sportMeta(s models.Sport) (label, emoji string) {
	switch s {
	case models.Basket:
		return "Basket", "🏀"
	case models.Foot:
		return "Football", "⚽️"
	case models.PingPong:
		return "Ping-pong", "🏓"
	default:
		return string(s), "🎯"
	}
}

func (mailer *Mailer) SendMatchResultEmail(matchId string, to string, username string, sport models.Sport, fieldName string, teamScore, oppScore int) error {
	baseLogger := log.With().
		Str("mail_kind", "match_result").
		Str("to", to).
		Str("username", username).
		Str("sport", string(sport)).
		Str("field", fieldName).
		Int("team_score", teamScore).
		Int("opp_score", oppScore).
		Logger()

	key := matchId + ":match_result"
	if mailer.AlreadySent[key] && time.Since(mailer.LastSentAt[key]) < 10*time.Second {
		baseLogger.Warn().Dur("since_last", time.Since(mailer.LastSentAt[key])).Msg("match result email throttled")
		return fmt.Errorf("match result email recently sent to %s → throttled", to)
	}

	label, emoji := sportMeta(sport)
	resultWord := "Match nul"
	resultBadgeBg := "#3F3F46"
	if teamScore > oppScore {
		resultWord = "Victoire"
		resultBadgeBg = "#22C55E"
	} else if teamScore < oppScore {
		resultWord = "Défaite"
		resultBadgeBg = "#EF4444"
	}

	subject := fmt.Sprintf("%s • %s à %s — %s %d–%d", label, emoji, fieldName, resultWord, teamScore, oppScore)

	baseLogger.Info().Msg("sending match result email")

	m := gomail.NewMessage()
	m.SetHeader("From", mailer.Config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	textBody := fmt.Sprintf(`Salut %s,

Résultat de ton match de %s à %s :
%s — %d-%d

À bientôt sur le terrain !
Play The Street`,
		username, label, fieldName, resultWord, teamScore, oppScore)

	htmlBody := fmt.Sprintf(`
<html>
	<body style="margin:0;padding:0;background:#0E0E0E;font-family: Inter, Arial, sans-serif;">
		<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="padding:24px 0;">
			<tr>
				<td align="center">
					<table role="presentation" width="600" cellpadding="0" cellspacing="0" style="width:600px;max-width:94%%;">
						<tr>
							<td align="center" style="padding:8px 0 20px 0;">
								<div style="font-size:22px;color:#FF6A00;font-weight:700;letter-spacing:.3px;">
									PLAY THE STREET
								</div>
							</td>
						</tr>

						<tr>
							<td style="background:#1A1A1A;border-radius:16px;padding:28px 22px;border:1px solid #2B2B2B;">
								
								<h1 style="margin:0 0 14px 0;font-size:22px;line-height:28px;color:#EDEDED;text-align:center;font-weight:600;">
									%s — %s
								</h1>
								<p style="margin:0 0 6px 0;font-size:14px;line-height:20px;color:#BDBDBD;text-align:center;">
									%s • %s
								</p>

								<div style="text-align:center;margin:18px 0 8px 0;">
									<span style="display:inline-block;padding:8px 14px;border-radius:999px;background:%s;color:#0E0E0E;font-weight:800;font-size:13px;">
										%s
									</span>
								</div>

								<div style="text-align:center;margin:14px 0 20px 0;">
									<span style="display:inline-block;font-size:34px;line-height:38px;color:#FFFFFF;font-weight:800;">
										%d&nbsp;–&nbsp;%d
									</span>
								</div>

								<p style="margin:6px 0 0 0;font-size:14px;line-height:22px;color:#FF6A00;text-align:center;font-weight:600;">
									%s, %s ! Continue sur ta lancée.
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
`, label, emoji, fieldName, username, resultBadgeBg, resultWord, teamScore, oppScore,
		func() string {
			if teamScore > oppScore {
				return "Belle victoire"
			}
			if teamScore == oppScore {
				return "Beau match"
			}
			return "Ce n’est que partie remise"
		}(), username, time.Now().Year())

	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	start := time.Now()
	d := mailer.dialer()
	if err := d.DialAndSend(m); err != nil {
		baseLogger.Error().Err(err).Dur("latency", time.Since(start)).Msg("mail send failed")
		return err
	}

	if mailer.AlreadySent == nil {
		mailer.AlreadySent = map[string]bool{}
	}
	if mailer.LastSentAt == nil {
		mailer.LastSentAt = map[string]time.Time{}
	}
	mailer.AlreadySent[key] = true
	mailer.LastSentAt[key] = time.Now()

	baseLogger.Info().Dur("latency", time.Since(start)).Msg("mail sent successfully")
	return nil
}
