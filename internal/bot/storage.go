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

func (s *Storage) groupId() int64 {
	return 1
}

func (s *Storage) setContext(userId int64, botId int64, groupId int64, ctx *context.Context) {

}

func (s *Storage) context(userId int64) {

}

func (s *Storage) userStep() {

}

func (s *Storage) setUserStep() {

}

func (s *Storage) addUser(botId int64, from *telego.User) error {
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

	_, err := s.db.AddUser(botId, user)
	if err != nil {
		return err
	}

	if err := s.redis.SetUserStep(botId, config.MainGroupId, from.ID, user.StepId); err != nil {
		return err
	}

	return nil
}

func (s *Storage) checkUserExist(userId int64, botId int64) (bool, error) {
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
