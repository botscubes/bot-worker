package bot

import (
	"errors"

	"github.com/botscubes/bot-worker/internal/config"
	rdb "github.com/botscubes/bot-worker/internal/database/redis"
	"github.com/botscubes/bot-worker/internal/model"
	"github.com/mymmrac/telego"
)

func (bw *BotWorker) getUserStep(botId int64, from *telego.User) (int64, error) {
	// try get userStep from cache
	stepID, err := bw.redis.GetUserStep(botId, from.ID)
	if err == nil {
		return stepID, nil
	}

	if !errors.Is(err, rdb.ErrNotFound) {
		bw.log.Error(err)
	}

	// userStep not found in cache, try get from db
	stepID, err = bw.db.UserStepByTgId(botId, from.ID)
	if err != nil {
		bw.log.Error(err)
		return 0, err
	}

	if err = bw.redis.SetUserStep(botId, from.ID, stepID); err != nil {
		bw.log.Error(err)
	}

	return stepID, nil
}

func (bw *BotWorker) addUser(botId int64, from *telego.User) error {
	user := &model.User{
		TgId:      from.ID,
		FirstName: &from.FirstName,
		LastName:  &from.LastName,
		Username:  &from.Username,
		StepID: model.StepID{
			StepId: config.MainComponentId,
		},
		Status: model.StatusUserActive,
	}

	_, err := bw.db.AddUser(botId, user)
	if err != nil {
		bw.log.Error(err)
		return err
	}

	if err := bw.redis.SetUserStep(botId, from.ID, user.StepId); err != nil {
		bw.log.Error(err)
	}

	return nil
}

func (bw *BotWorker) getComponent(botId int64, stepID int64) (*model.Component, error) {
	// try get component from cache
	component, err := bw.redis.GetComponent(botId, stepID)
	if err == nil {
		return component, nil
	}

	if !errors.Is(err, rdb.ErrNotFound) {
		bw.log.Error(err)
	}

	// check bot components exists in cache
	ex, err := bw.redis.CheckComponentsExist(botId)
	if err != nil {
		bw.log.Error(err)
	}

	// components not found in cache
	if err == nil && ex == 0 {
		// get all components from db
		components, err := bw.db.ComponentsForBot(botId)
		if err != nil {
			bw.log.Error(err)
		}

		if err := bw.redis.SetComponents(botId, components); err != nil {
			bw.log.Error(err)
		}

		component, err := bw.redis.GetComponent(botId, stepID)
		if err == nil {
			return component, nil
		}

		if !errors.Is(err, rdb.ErrNotFound) {
			bw.log.Error(err)
		}
	}

	// component not found in cache, try get from db
	exist, err := bw.db.CheckComponentExist(botId, stepID)
	if err != nil {
		bw.log.Error(err)
		return nil, err
	}

	if exist {
		component, err = bw.db.ComponentForBot(botId, stepID)
		if err != nil {
			bw.log.Error(err)
			return nil, err
		}

		if err = bw.redis.SetComponent(botId, component); err != nil {
			bw.log.Error(err)
		}

		return component, nil
	}

	return nil, ErrNotFound
}

func (bw *BotWorker) setUserStep(botId int64, userId int64, stepID int64) {
	if err := bw.db.SetUserStepByTgId(botId, userId, stepID); err != nil {
		bw.log.Error(err)
	}
}
