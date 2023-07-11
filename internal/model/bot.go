package model

type BotStatus int

var (
	StatusBotStopped BotStatus
	StatusBotRunning BotStatus = 1
)

type Bot struct {
	Id    int64   `json:"id"`
	Token *string `json:"-"`
}
