package bot

import (
	"github.com/mymmrac/telego"

	tu "github.com/mymmrac/telego/telegoutil"
)

type BotIO struct {
	bot              *telego.Bot
	update           *telego.Update
	messageProcessed bool
}

func NewBotIO(bot *telego.Bot, update *telego.Update, messageProcessed bool) *BotIO {
	return &BotIO{
		bot,
		update,
		messageProcessed,
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
	if io.messageProcessed {
		return nil
	}
	text := io.update.Message.Text
	io.messageProcessed = true
	return &text
}
