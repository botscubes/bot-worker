package pgsql

import (
	"context"
	"strconv"

	"github.com/botscubes/bot-worker/internal/model"
)

func (db *Db) AddUser(botId int64, m *model.User) (int64, error) {
	var id int64
	prefix := prefixSchema + strconv.FormatInt(botId, 10)

	query := `INSERT INTO ` + prefix + `.user 
	(tg_id, first_name, last_name, username, step_id, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;`

	if err := db.Pool.QueryRow(
		context.Background(), query, m.TgId, m.FirstName, m.LastName, m.Username, m.StepId, m.Status,
	).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (db *Db) CheckUserExistByTgId(botId int64, tgId int64) (bool, error) {
	var c bool
	prefix := prefixSchema + strconv.FormatInt(botId, 10)

	query := `SELECT EXISTS(SELECT 1 FROM ` + prefix + `.user WHERE tg_id = $1 AND status = $2) AS "exists";`
	if err := db.Pool.QueryRow(
		context.Background(), query, tgId, model.StatusUserActive,
	).Scan(&c); err != nil {
		return false, err
	}

	return c, nil
}

// func (db *Db) UserByTgId(userId int64, botId int64) (*model.User, error) {
// 	prefix := prefixSchema + strconv.FormatInt(botId, 10)

// 	query := `SELECT id, tg_id, first_name, last_name, username, status
// 			FROM ` + prefix + `.user WHERE tg_id = $1;`

// 	var r model.User
// 	if err := db.Pool.QueryRow(
// 		context.Background(), query, userId,
// 	).Scan(&r.Id, &r.TgId, &r.FirstName, &r.LastName, &r.Username, &r.Status); err != nil {
// 		return nil, err
// 	}

// 	return &r, nil
// }

func (db *Db) UserStepByTgId(botId int64, userId int64) (int64, error) {
	prefix := prefixSchema + strconv.FormatInt(botId, 10)

	query := `SELECT step_id FROM ` + prefix + `.user WHERE tg_id = $1;`

	var r int64
	if err := db.Pool.QueryRow(
		context.Background(), query, userId,
	).Scan(&r); err != nil {
		return 0, err
	}

	return r, nil
}

func (db *Db) SetUserStepByTgId(botId int64, userId int64, stepID int64) error {
	prefix := prefixSchema + strconv.FormatInt(botId, 10)

	query := `UPDATE ` + prefix + `.user SET step_id = $1 WHERE tg_id = $2;`
	_, err := db.Pool.Exec(context.Background(), query, stepID, userId)
	return err
}
