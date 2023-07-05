package redis

import (
	"context"
	"errors"
	"strconv"

	"github.com/botscubes/bot-worker/internal/config"
	"github.com/redis/go-redis/v9"
)

func (rdb *Rdb) SetUserStep(botId int64, userID int64, stepID int64) error {
	ctx := context.Background()

	key := "bot" + strconv.FormatInt(botId, 10) + ":user:" + strconv.FormatInt(userID, 10)

	if err := rdb.HSet(ctx, key, "step", stepID).Err(); err != nil {
		return err
	}

	return rdb.Expire(ctx, key, config.RedisExpire).Err()
}

func (rdb *Rdb) GetUserStep(botId int64, userID int64) (int64, error) {
	ctx := context.Background()

	key := "bot" + strconv.FormatInt(botId, 10) + ":user:" + strconv.FormatInt(userID, 10)

	var stepID int64
	err := rdb.HGet(ctx, key, "step").Scan(&stepID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}

	if errors.Is(err, redis.Nil) {
		return 0, ErrNotFound
	}

	return stepID, nil
}

func (rdb *Rdb) CheckUserExist(botId int64, userID int64) (int64, error) {
	ctx := context.Background()
	key := "bot" + strconv.FormatInt(botId, 10) + ":user:" + strconv.FormatInt(userID, 10)
	return rdb.Exists(ctx, key).Result()
}
