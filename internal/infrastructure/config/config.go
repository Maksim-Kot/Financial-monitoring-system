package config

import (
	"fms-project/internal/infrastructure/adapter"
)

type InfrastructureConfig struct {
	HttpPort    string
	LogLevel    string
	RedisHost   string
	RedisPort   string
	RedisSecret string
	PostgresUri string
	IsDev       bool
	IsProd      bool
}

type ServicesConfig struct {
	TelegramBotToken  string
	GroqServiceURL    string
	GroqServiceAPIKey string
}

type Config struct {
	Infrastructure InfrastructureConfig
	Services       ServicesConfig
}

func Load() (*Config, error) {
	if err := adapter.InitDotEnv(); err != nil {
		return nil, err
	}

	environment := adapter.Getenv("ENVIRONMENT").Default("development").ToString()

	config := &Config{
		Infrastructure: InfrastructureConfig{
			HttpPort:    adapter.Getenv("PORT").Default("8080").ToString(),
			LogLevel:    adapter.Getenv("LOG_LEVEL").Default("debug").ToString(),
			RedisHost:   adapter.Getenv("REDIS_HOST").Default("localhost").ToString(),
			RedisPort:   adapter.Getenv("REDIS_PORT").Default("6379").ToString(),
			RedisSecret: adapter.Getenv("REDIS_SECRET").Default("").ToString(),
			PostgresUri: adapter.Getenv("POSTGRES_URI").Default("postgres://admin:admin@localhost:5434/fms").ToString(),
			IsDev:       environment == "development",
			IsProd:      environment == "production",
		},
		Services: ServicesConfig{
			TelegramBotToken:  adapter.Getenv("BOT_TOKEN").ToString(),
			GroqServiceURL:    adapter.Getenv("GROQ_SERVICE_URL").ToString(),
			GroqServiceAPIKey: adapter.Getenv("GROQ_SERVICE_API_KEY").ToString(),
		},
	}

	return config, nil
}
