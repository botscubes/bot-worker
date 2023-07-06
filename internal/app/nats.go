package app

import (
	"github.com/goccy/go-json"
	"github.com/nats-io/nats.go"
)

type ncPayload struct {
	BotId int64  `json:"botId"`
	Token string `json:"token"`
}

var (
	ncCodeOk        = "200"
	ncCodeErrServer = "500"
)

func (app *App) onStartBot(msg *nats.Msg) {
	req := new(ncPayload)
	if err := json.Unmarshal(msg.Data, req); err != nil {
		app.log.Error(err)

		msg.Respond([]byte(ncCodeErrServer)) //nolint:errcheck
		return
	}

	if err := app.worker.RunBot(req.BotId, req.Token); err != nil {
		msg.Respond([]byte(ncCodeErrServer)) //nolint:errcheck
		app.log.Errorw("launch bot", "botId", req.BotId, "error", err)
	} else {
		msg.Respond([]byte(ncCodeOk)) //nolint:errcheck
	}
}

func (app *App) onStopBot(msg *nats.Msg) {
	req := new(ncPayload)
	if err := json.Unmarshal(msg.Data, req); err != nil {
		app.log.Error(err)

		msg.Respond([]byte(ncCodeErrServer)) //nolint:errcheck
		return
	}

	app.worker.StopBot(req.BotId)
	msg.Respond([]byte(ncCodeOk)) //nolint:errcheck
}
