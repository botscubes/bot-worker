package redis

import (
	"context"
	"errors"
	"strconv"

	bctx "github.com/botscubes/bot-components/context"
	"github.com/botscubes/bot-worker/internal/config"
	"github.com/redis/go-redis/v9"
)

func (rdb *Rdb) SetUserStep(botId int64, groupId int64, userID int64, stepID int64) error {
	ctx := context.Background()

	key := "user:" + strconv.FormatInt(userID, 10) +
		":bot:" + strconv.FormatInt(botId, 10)

	field := "group:" + strconv.FormatInt(groupId, 10) + ":step"
	if err := rdb.HSet(ctx, key, field, stepID).Err(); err != nil {
		return err
	}

	return rdb.Expire(ctx, key, config.RedisExpire).Err()
}

func (rdb *Rdb) GetUserStep(botId int64, groupId int64, userID int64) (*int64, error) {
	ctx := context.Background()

	key := "user:" + strconv.FormatInt(userID, 10) +
		":bot:" + strconv.FormatInt(botId, 10)

	field := "group:" + strconv.FormatInt(groupId, 10) + ":step"

	var stepID int64
	err := rdb.HGet(ctx, key, field).Scan(&stepID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	return &stepID, nil
}
func (rdb *Rdb) SetUserContext(botId int64, groupId int64, userId int64, bctx *bctx.Context) error {
	ctx := context.Background()

	key := "user:" + strconv.FormatInt(userId, 10) +
		":bot:" + strconv.FormatInt(botId, 10)

	field := "group:" + strconv.FormatInt(groupId, 10) + ":context"
	bytes, err := bctx.ToJSON()
	if err != nil {
		return err
	}
	if err := rdb.HSet(ctx, key, field, bytes).Err(); err != nil {
		return err
	}

	return rdb.Expire(ctx, key, config.RedisExpire).Err()
}

func (rdb *Rdb) GetUserContext(botId int64, groupId int64, userId int64) (*bctx.Context, error) {
	ctx := context.Background()

	key := "user:" + strconv.FormatInt(userId, 10) +
		":bot:" + strconv.FormatInt(botId, 10)

	field := "group:" + strconv.FormatInt(groupId, 10) + ":context"

	bytes, err := rdb.HGet(ctx, key, field).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	return bctx.NewContextFromJSON(bytes)
}

func (rdb *Rdb) CheckUserExist(botId int64, userID int64) (int64, error) {
	ctx := context.Background()
	key := "user:" + strconv.FormatInt(userID, 10) +
		":bot:" + strconv.FormatInt(botId, 10)
	return rdb.Exists(ctx, key).Result()
}
