package broker

import "github.com/botscubes/bot-worker/internal/bot"

type Broker interface {
	StartBotSub() error
	StopBotSub() error
	CloseConnection()
	SetWorker(w *bot.BotWorker)
}
