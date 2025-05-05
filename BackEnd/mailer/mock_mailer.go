package mailer

type MockMailer struct {
	SentCounts map[string]int
}

func NewMockMailer() *MockMailer {
	return &MockMailer{
		SentCounts: make(map[string]int),
	}
}

func (m *MockMailer) SendTestMail(to string) error {
	m.SentCounts["test"]++
	return nil
}

func (m *MockMailer) SendPasswordForgotMail(to string, newPassword string) error {
	m.SentCounts["password_forgot"]++
	return nil
}

func (m *MockMailer) GetSentCounts(mail string) int {
	return m.SentCounts[mail]
}
