package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/botscubes/bot-worker/internal/config"
	"github.com/botscubes/bot-worker/internal/model"
	"github.com/redis/go-redis/v9"
)

//func (rdb *Rdb) SetComponent(botId int64, comp *model.Component) error {
//	ctx := context.Background()
//
//	key := "bot" + strconv.FormatInt(botId, 10) + ":component"
//
//	if err := rdb.HSet(ctx, key, strconv.FormatInt(comp.Id, 10), comp).Err(); err != nil {
//		return err
//	}
//
//	return rdb.Expire(ctx, key, config.RedisExpire).Err()
//}

func (rdb *Rdb) SetComponents(botId int64, groupId int64, comps map[int64]*model.ComponentData) error {
	ctx := context.Background()
	key := "bot" + strconv.FormatInt(botId, 10) + ":group" + strconv.FormatInt(groupId, 10)
	data, err := json.Marshal(comps)
	if err != nil {
		return nil
	}
	if err := rdb.HSet(ctx, key, "components", data).Err(); err != nil {
		return err
	}
	return rdb.Expire(ctx, key, config.RedisExpire).Err()
}

func (rdb *Rdb) GetComponents(botId int64, groupId int64) (*map[int64]*model.ComponentData, error) {
	ctx := context.Background()

	key := "bot" + strconv.FormatInt(botId, 10) + ":group" + strconv.FormatInt(groupId, 10)

	components := new(map[int64]*model.ComponentData)

	result, err := rdb.HGet(ctx, key, "components").Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(result, components); err != nil {
		return nil, err
	}

	return components, nil
}

//func (rdb *Rdb) CheckComponentsExist(botId int64) (int64, error) {
//	ctx := context.Background()
//	key := "bot" + strconv.FormatInt(botId, 10) + ":component"
//	return rdb.Exists(ctx, key).Result()
//}
