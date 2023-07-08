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
		app.doNatsRespond(msg, []byte(ncCodeErrServer))

		app.log.Errorw("failed json unmarshal", "error", err)
		return
	}

	// start bot handler
	if err := app.worker.RunBot(req.BotId, req.Token); err != nil {
		app.doNatsRespond(msg, []byte(ncCodeErrServer))

		app.log.Errorw("failed launch bot", "botId", req.BotId, "error", err)
	} else {
		app.doNatsRespond(msg, []byte(ncCodeOk))
	}
}

func (app *App) onStopBot(msg *nats.Msg) {
	req := new(ncPayload)
	if err := json.Unmarshal(msg.Data, req); err != nil {
		app.doNatsRespond(msg, []byte(ncCodeErrServer))

		app.log.Errorw("failed json unmarshal", "error", err)
		return
	}

	// stop bot handler
	app.worker.StopBot(req.BotId)

	app.doNatsRespond(msg, []byte(ncCodeOk))
}

func (app *App) doNatsRespond(msg *nats.Msg, data []byte) {
	if err := msg.Respond(data); err != nil {
		app.log.Errorw("failed nats send repond", "error", err)
	}
}
