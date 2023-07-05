package pgsql

import (
	"context"

	"github.com/botscubes/bot-worker/internal/model"
)

func (db *Db) GetBotToken(userId int64, botId int64) (*string, error) {
	var data string
	query := `SELECT token FROM public.bot WHERE id = $1 AND user_id = $2;`
	if err := db.Pool.QueryRow(
		context.Background(), query, botId, userId,
	).Scan(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (db *Db) GetRunningBots() (*[]*model.Bot, error) {
	var data []*model.Bot

	query := `SELECT id, token FROM public.bot WHERE status = $1;`

	rows, err := db.Pool.Query(context.Background(), query, model.StatusBotRunning)
	if err != nil {
		return nil, err
	}

	// WARN: status not used
	for rows.Next() {
		var r model.Bot
		if err = rows.Scan(&r.Id, &r.Token); err != nil {
			return nil, err
		}

		data = append(data, &r)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return &data, nil
}
