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

func (db *Db) Components(botId int64, groupId int64) (map[int64]*model.ComponentData, error) {
	query :=
		`SELECT component_id, type, 
			jsonb_build_object(
				'id', component_id,
				'type', type,
				'path', path,
				'outputs', outputs,
				'data', data
			)
			FROM ` + prefixSchema + strconv.FormatInt(botId, 10) + `.component
			WHERE group_id = $1;`
	var data map[int64]*model.ComponentData = make(map[int64]*model.ComponentData)

	rows, err := db.Pool.Query(context.Background(), query, groupId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var d model.ComponentData
		var id int64
		if err = rows.Scan(&id, &d.Type, &d.Data); err != nil {
			return nil, err
		}

		data[id] = &d
	}

	return data, nil
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
