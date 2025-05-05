package mailer

import (
	"PLIC/models"
	"time"
)

type Mailer struct {
	AlreadySent map[string]bool
	LastSentAt  map[string]time.Time
	Config      *models.MailerConfig
}
