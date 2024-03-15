package bot

import (
	"errors"

	"github.com/botscubes/bot-worker/internal/config"
	rdb "github.com/botscubes/bot-worker/internal/database/redis"
	"github.com/botscubes/bot-worker/internal/model"
	"github.com/mymmrac/telego"
)

// Getting the user step.
// Initially, an attempt is made to retrieve from the cache.
// In case of failure, a request is made to the database.
// The step received from the database is written to the cache
func (bw *BotWorker) getUserStep(botId int64, from *telego.User) (int64, bool) {
	// try get user step from cache
	stepID, err := bw.redis.GetUserStep(botId, from.ID)
	if err == nil {
		return stepID, true
	}

	if !errors.Is(err, rdb.ErrNotFound) {
		bw.log.Errorw("failed get user step from cache", "error", err)
	}

	// user step not found in cache, try get from db
	stepID, err = bw.db.UserStepByTgId(botId, from.ID)
	if err != nil {
		bw.log.Errorw("failed get user step by tg id", "error", err)
		return 0, false
	}

	// write user step to the cache
	if err = bw.redis.SetUserStep(botId, from.ID, stepID); err != nil {
		bw.log.Errorw("failed write user step to the cache", "error", err)
	}

	return stepID, true
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
		return err
	}

	if err := bw.redis.SetUserStep(botId, from.ID, user.StepId); err != nil {
		bw.log.Errorw("failed write user step to the cache", "error", err)
	}

	return nil
}

// Getting the component.
// Initially, an attempt is made to retrieve from the cache.
// In case of failure, it is received from the database.
// Components received from the database are stored in the cache
//func (bw *BotWorker) getComponent(botId int64, stepID int64) (*model.Component, error) {
//	// try get component from cache
//	component, err := bw.redis.GetComponent(botId, stepID)
//	if err == nil {
//		return component, nil
//	}
//
//	if !errors.Is(err, rdb.ErrNotFound) {
//		bw.log.Errorw("failed get component from cache", "error", err)
//	}
//
//	// check bot components exists in cache
//	ex, err := bw.redis.CheckComponentsExist(botId)
//	if err != nil {
//		bw.log.Errorw("failed check components exists in the cache", "error", err)
//	}
//
//	// components not found in the cache
//	if err == nil && ex == 0 {
//		// get all components from db
//		components, err := bw.db.Components(botId)
//		if err != nil {
//			bw.log.Errorw("failed get all components from db", "error", err)
//		} else {
//			// write compoennts to the cache
//			if err := bw.redis.SetComponents(botId, components); err != nil {
//				bw.log.Errorw("failed write components to the cache", "error", err)
//			} else {
//				// get the required component from the cache
//				component, err := bw.redis.GetComponent(botId, stepID)
//				if err == nil {
//					return component, nil
//				}
//
//				if !errors.Is(err, rdb.ErrNotFound) {
//					bw.log.Errorw("failed get component from cache", "error", err)
//				}
//			}
//		}
//	}
//
//	// component not found in cache, try get from db
//	exists, err := bw.db.CheckComponentExist(botId, stepID)
//	if err != nil {
//		bw.log.Errorw("failed check component exists in the db", "error", err)
//		return nil, err
//	}
//
//	if exists {
//		component, err = bw.db.Component(botId, stepID)
//		if err != nil {
//			bw.log.Errorw("failed get component from db", "error", err)
//			return nil, err
//		}
//
//		if err = bw.redis.SetComponent(botId, component); err != nil {
//			bw.log.Errorw("failed write component to the cache", "error", err)
//		}
//
//		return component, nil
//	}
//
//	return nil, ErrNotFound
//}
