package mailer

import "time"

type Mailer struct {
	AlreadySent map[string]bool
	LastSentAt  map[string]time.Time
}
