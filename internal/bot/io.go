package bot

import (
	"github.com/mymmrac/telego"

	tu "github.com/mymmrac/telego/telegoutil"
)

type BotIO struct {
	bot    *telego.Bot
	update *telego.Update
}

func NewBotIO(bot *telego.Bot, update *telego.Update) *BotIO {
	return &BotIO{
		bot,
		update,
	}
}

func (io *BotIO) OutputText(text string) {
	chatID := io.update.Message.Chat.ID
	io.bot.SendMessage(
		tu.Message(
			tu.ID(chatID),
			text,
		),
	)
}
func (io *BotIO) InputText() *string {

	return nil
}
