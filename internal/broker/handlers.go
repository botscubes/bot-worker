package broker

import (
	"github.com/goccy/go-json"

	"github.com/nats-io/nats.go"
)

var (
	natsCodeOk        = "200"
	natsCodeErrServer = "500"
)

type startBotPayload struct {
	BotId int64  `json:"botId"`
	Token string `json:"token"`
}

func (b *NatsBroker) onStartBot(msg *nats.Msg) {
	req := new(startBotPayload)
	if err := json.Unmarshal(msg.Data, req); err != nil {
		b.doRespond(msg, []byte(natsCodeErrServer))

		b.log.Errorw("failed json unmarshal", "error", err)
		return
	}

	// start bot handler
	if err := b.worker.RunBot(req.BotId, req.Token); err != nil {
		b.doRespond(msg, []byte(natsCodeErrServer))

		b.log.Errorw("failed launch bot", "botId", req.BotId, "error", err)
	} else {
		b.doRespond(msg, []byte(natsCodeOk))
	}
}

type stopBotPayload struct {
	BotId int64 `json:"botId"`
}

func (b *NatsBroker) onStopBot(msg *nats.Msg) {
	req := new(stopBotPayload)
	if err := json.Unmarshal(msg.Data, req); err != nil {
		b.doRespond(msg, []byte(natsCodeErrServer))

		b.log.Errorw("failed json unmarshal", "error", err)
		return
	}

	// stop bot handler
	b.worker.StopBot(req.BotId)

	b.doRespond(msg, []byte(natsCodeOk))
}

func (b *NatsBroker) doRespond(msg *nats.Msg, data []byte) {
	if err := msg.Respond(data); err != nil {
		b.log.Errorw("failed nats send repond", "error", err)
	}
}
