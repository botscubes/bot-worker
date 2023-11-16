package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

const (
	MainComponentId = 1
	RedisExpire     = 1 * time.Hour
	ShutdownTimeout = 1 * time.Minute
)

type ServiceConfig struct {
	LoggerType    string `env:"LOGGER_TYPE,required"`
	WebhookPath   string `env:"WEBHOOK_PATH,required"`
	ListenAddress string `env:"LISTEN_ADDRESS,required"`
	Redis         RedisConfig
	Pg            PostgresConfig
	NatsURL       string `env:"NATS_URL,required"`
}

type RedisConfig struct {
	Db   int    `env:"REDIS_DB,required"`
	Pass string `env:"REDIS_PASS,required"`
	Host string `env:"REDIS_HOST,required"`
	Port string `env:"REDIS_PORT,required"`
}

type PostgresConfig struct {
	Db   string `env:"POSTGRES_DB,required"`
	User string `env:"POSTGRES_USER,required"`
	Pass string `env:"POSTGRES_PASSWORD,required"`
	Host string `env:"POSTGRES_HOST,required"`
	Port string `env:"POSTGRES_PORT,required"`
}

func GetConfig() (*ServiceConfig, error) {
	var c ServiceConfig
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, err
	}
	return &c, nil
}
