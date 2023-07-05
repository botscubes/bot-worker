package redis

import (
	"context"
	"errors"
	"strconv"

	"github.com/botscubes/bot-worker/internal/config"
	"github.com/botscubes/bot-worker/internal/model"
	"github.com/redis/go-redis/v9"
)

func (rdb *Rdb) SetComponent(botId int64, comp *model.Component) error {
	ctx := context.Background()

	key := "bot" + strconv.FormatInt(botId, 10) + ":component"

	if err := rdb.HSet(ctx, key, strconv.FormatInt(comp.Id, 10), comp).Err(); err != nil {
		return err
	}

	return rdb.Expire(ctx, key, config.RedisExpire).Err()
}

func (rdb *Rdb) SetComponents(botId int64, comps *[]*model.Component) error {
	for _, v := range *comps {
		if err := rdb.SetComponent(botId, v); err != nil {
			return err
		}
	}

	return nil
}

func (rdb *Rdb) GetComponent(botId int64, compId int64) (*model.Component, error) {
	ctx := context.Background()

	key := "bot" + strconv.FormatInt(botId, 10) + ":component"

	component := &model.Component{}

	result, err := rdb.HGet(ctx, key, strconv.FormatInt(compId, 10)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if result == "" {
		return nil, ErrNotFound
	}

	if err := component.UnmarshalBinary([]byte(result)); err != nil {
		return nil, err
	}

	return component, nil
}

func (rdb *Rdb) CheckComponentsExist(botId int64) (int64, error) {
	ctx := context.Background()
	key := "bot" + strconv.FormatInt(botId, 10) + ":component"
	return rdb.Exists(ctx, key).Result()
}
