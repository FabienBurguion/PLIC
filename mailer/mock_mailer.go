package mailer

import "PLIC/models"

type MockMailer struct {
	SentCounts map[string]int
}

func NewMockMailer() *MockMailer {
	return &MockMailer{
		SentCounts: make(map[string]int),
	}
}

func (m *MockMailer) SendLinkResetPassword(_ string, _ string) error {
	m.SentCounts["link_reset"]++
	return nil
}

func (m *MockMailer) SendWelcomeEmail(_ string, _ string) error {
	m.SentCounts["welcome"]++
	return nil
}

func (m *MockMailer) SendMatchResultEmail(_ string, _ string, _ models.Sport, _ string, _, _ int) error {
	m.SentCounts["result"]++
	return nil
}

func (m *MockMailer) GetSentCounts(mail string) int {
	return m.SentCounts[mail]
}
