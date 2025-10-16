package models

type MailerConfig struct {
	From     string `env:"SMTP_FROM"`
	Host     string `env:"SMTP_HOST"`
	Port     int    `env:"SMTP_PORT" envDefault:"587"`
	Username string `env:"SMTP_USERNAME"`
	Password string `env:"SMTP_PASSWORD"`
}

type LambdaConfig struct {
	FunctionName string `env:"AWS_LAMBDA_FUNCTION_NAME"`
}

type DatabaseConfig struct {
	ConnectionString string `env:"DATABASE_URL"`
}

type GoogleConfig struct {
	ApiKey string `env:"GOOGLE_APIKEY"`
}

type Configuration struct {
	Mailer   MailerConfig
	Lambda   LambdaConfig
	Database DatabaseConfig
	Google   GoogleConfig
}
