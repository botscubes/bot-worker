package model

type BotStatus int

var (
	StatusBotStopped BotStatus
	StatusBotRunning BotStatus = 1
)

type Bot struct {
	Id     int64     `json:"id"`
	UserId int64     `json:"userId,omitempty"`
	Title  *string   `json:"title"`
	Token  *string   `json:"-"`
	Status BotStatus `json:"-"`
}
