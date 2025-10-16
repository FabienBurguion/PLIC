package mailer

type MockMailer struct {
	SentCounts map[string]int
}

func NewMockMailer() *MockMailer {
	return &MockMailer{
		SentCounts: make(map[string]int),
	}
}

func (m *MockMailer) SendTestMail(_ string) error {
	m.SentCounts["test"]++
	return nil
}

func (m *MockMailer) SendLinkResetPassword(_ string, _ string) error {
	m.SentCounts["link_reset"]++
	return nil
}

func (m *MockMailer) GetSentCounts(mail string) int {
	return m.SentCounts[mail]
}
