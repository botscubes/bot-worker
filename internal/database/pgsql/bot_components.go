package pgsql

import (
	"context"
	"strconv"

	"github.com/botscubes/bot-worker/internal/model"
)

func (db *Db) CheckComponentExist(botId int64, compId int64) (bool, error) {
	var c bool
	query := `SELECT EXISTS(SELECT 1 FROM ` + prefixSchema + strconv.FormatInt(botId, 10) + `.component
			WHERE id = $1 AND status = $2) AS "exists";`

	if err := db.Pool.QueryRow(
		context.Background(), query, compId, model.StatusComponentActive,
	).Scan(&c); err != nil {
		return false, err
	}

	return c, nil
}

func (db *Db) Components(botId int64) (*[]*model.Component, error) {
	var data []*model.Component

	query := `SELECT id, data, keyboard, ARRAY(
			SELECT jsonb_build_object('id', id, 'data', data, 'type', type, 'componentId', component_id, 'nextStepId', next_step_id)
				FROM ` + prefixSchema + strconv.FormatInt(botId, 10) + `.command
				WHERE component_id = t.id AND status = $1 ORDER BY id
			), next_step_id, is_main
			FROM ` + prefixSchema + strconv.FormatInt(botId, 10) + `.component t
			WHERE status = $2 ORDER BY id;`

	rows, err := db.Pool.Query(context.Background(), query, model.StatusCommandActive, model.StatusComponentActive)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var r model.Component
		r.Commands = &model.Commands{}
		if err = rows.Scan(&r.Id, &r.Data, &r.Keyboard, r.Commands, &r.NextStepId, &r.IsMain); err != nil {
			return nil, err
		}

		data = append(data, &r)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return &data, nil
}

func (db *Db) Component(botId int64, compID int64) (*model.Component, error) {
	prefix := prefixSchema + strconv.FormatInt(botId, 10)

	query := `SELECT id, data, keyboard, ARRAY(
		SELECT jsonb_build_object('id', id, 'data', data, 'type', type, 'componentId', component_id, 'nextStepId', next_step_id)
		FROM ` + prefix + `.command
		WHERE component_id = t.id AND status = $1 ORDER BY id
	), next_step_id, is_main
	FROM ` + prefix + `.component t
	WHERE status = $2 AND id = $3 ORDER BY id;`

	var r model.Component
	r.Commands = &model.Commands{}
	if err := db.Pool.QueryRow(
		context.Background(), query, model.StatusCommandActive, model.StatusComponentActive, compID,
	).Scan(&r.Id, &r.Data, &r.Keyboard, r.Commands, &r.NextStepId, &r.IsMain); err != nil {
		return nil, err
	}

	return &r, nil
}
