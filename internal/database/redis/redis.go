package redis

import (
	"github.com/botscubes/bot-worker/internal/config"
	"github.com/redis/go-redis/v9"
)

type Rdb struct {
	*redis.Client
}

func NewClient(c *config.RedisConfig) *Rdb {
	return &Rdb{
		redis.NewClient(&redis.Options{
			Addr:     c.Host + ":" + c.Port,
			Password: c.Pass,
			DB:       c.Db,
		}),
	}
}
