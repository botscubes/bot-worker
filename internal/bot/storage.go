package bot

import (
	"github.com/botscubes/bot-components/context"
	rdb "github.com/botscubes/bot-worker/internal/database/redis"
	"github.com/botscubes/bot-worker/internal/model"

	"github.com/botscubes/bot-worker/internal/database/pgsql"

	"github.com/botscubes/bot-worker/internal/config"
	"github.com/mymmrac/telego"

	"go.uber.org/zap"
)

type Storage struct {
	redis *rdb.Rdb
	db    *pgsql.Db
	log   *zap.SugaredLogger
}

func newStorage(redis *rdb.Rdb, db *pgsql.Db, log *zap.SugaredLogger) *Storage {
	return &Storage{
		redis,
		db,
		log,
	}

}

func (s *Storage) components(botId int64, groupId int64) (map[int64]*model.ComponentData, error) {
	components, err := s.redis.GetComponents(botId, groupId)
	if err != nil {

		return nil, err
	}
	if components != nil {
		s.log.Debug("get components from cache")
		return *components, nil
	}
	dbComps, err := s.db.Components(botId, groupId)
	if err != nil {
		return nil, err
	}
	err = s.redis.SetComponents(botId, groupId, dbComps)
	if err != nil {
		return nil, err
	}

	s.log.Debug("get components from db")
	return dbComps, nil

}

func (s *Storage) clearComponentCache(botId int64) error {
	return s.redis.DeleteComponents(botId, s.groupId())
}

func (s *Storage) groupId() int64 {
	return config.MainGroupId
}

func (s *Storage) setContext(botId int64, groupId int64, userId int64, ctx *context.Context) error {
	err := s.db.SetUserContextByTgId(botId, userId, ctx)
	if err != nil {
		return err
	}
	err = s.redis.SetUserContext(botId, groupId, userId, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) context(botId int64, groupId int64, userId int64) (*context.Context, error) {
	ctx, err := s.redis.GetUserContext(botId, groupId, userId)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx, err = s.db.UserContextByTgId(botId, userId)
		if err != nil {
			return nil, err
		}
		if ctx != nil {
			err = s.redis.SetUserContext(botId, groupId, userId, ctx)
			if err != nil {
				return nil, err
			}
		}
		s.log.Debugw("get user context from db", "botId", botId, "userId", userId)
		return ctx, nil
	}

	s.log.Debugw("get user context from redis", "botId", botId, "userId", userId)
	return ctx, nil
}

func (s *Storage) userStep(botId int64, groupId int64, userId int64) (int64, error) {
	step, err := s.redis.GetUserStep(botId, groupId, userId)
	if err != nil {
		return 0, err
	}
	if step == nil {
		step, err := s.db.UserStepByTgId(botId, userId)
		if err != nil {
			return 0, err
		}
		err = s.redis.SetUserStep(botId, groupId, userId, step)
		if err != nil {
			return 0, err
		}

		s.log.Debugw("get user step from db", "botId", botId, "userId", userId)
		return step, nil
	}

	s.log.Debugw("get user step from redis", "botId", botId, "userId", userId)
	return *step, nil
}

func (s *Storage) setUserStep(botId int64, groupId int64, userId int64, stepId int64) error {
	err := s.db.SetUserStepByTgId(botId, userId, stepId)
	if err != nil {
		return err
	}
	err = s.redis.SetUserStep(botId, groupId, userId, stepId)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) addUser(botId int64, from *telego.User) error {

	ctx := context.NewContext()

	user := &model.User{
		TgId:      from.ID,
		FirstName: &from.FirstName,
		LastName:  &from.LastName,
		Username:  &from.Username,
		StepID: model.StepID{
			StepId: config.MainComponentId,
		},
		Context: ctx,
		Status:  model.StatusUserActive,
	}

	_, err := s.db.AddUser(botId, user)
	if err != nil {
		return err
	}

	if err := s.redis.SetUserStep(botId, config.MainGroupId, from.ID, user.StepId); err != nil {
		return err
	}

	return nil
}

func (s *Storage) checkUserExist(botId int64, userId int64) (bool, error) {
	ex, err := s.redis.CheckUserExist(botId, userId)
	if err != nil {
		return false, err
	}

	if ex == 0 {

		exist, err := s.db.CheckUserExistByTgId(botId, userId)
		if err != nil {
			return false, err
		}
		s.log.Debugw("get user existence from db", "user", userId)
		return exist, nil

	}

	s.log.Debugw("get user existence from redis", "user", userId)
	return true, nil
}
