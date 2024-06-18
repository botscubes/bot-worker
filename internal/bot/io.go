package bot

import (
	"bytes"

	"github.com/botscubes/bot-components/io"
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

func (io *BotIO) PrintText(text string) error {
	chatID := io.update.Message.Chat.ID
	_, err := io.bot.SendMessage(
		tu.Message(
			tu.ID(chatID),
			text,
		).WithReplyMarkup(tu.ReplyKeyboardRemove()),
	)
	return err
}
func (io *BotIO) ReadText() *string {
	if io.messageProcessed {
		return nil
	}
	text := io.update.Message.Text
	io.messageProcessed = true
	return &text
}

func (io *BotIO) PrintPhoto(file []byte, name string) error {
	chatID := io.update.Message.Chat.ID
	_, err := io.bot.SendPhoto(
		tu.Photo(tu.ID(chatID), tu.File(
			tu.NameReader(
				bytes.NewReader(file),
				name,
			),
		),
		),
	)
	return err
}

func (io *BotIO) PrintButtons(text string, buttons []*io.ButtonData) error {

	chatID := io.update.Message.Chat.ID
	if len(buttons) == 0 {
		err := io.PrintText(text)
		return err
	}
	tbuttons := make([][]telego.KeyboardButton, (len(buttons)-1)/3+1)

	for i, button := range buttons {
		row := i / 3
		tbuttons[row] = append(tbuttons[row], tu.KeyboardButton(button.Text))
	}
	keyboard := tu.Keyboard(
		tbuttons...,
	).WithResizeKeyboard()

	msg := tu.Message(
		tu.ID(chatID),
		text,
	).WithReplyMarkup(keyboard)

	_, err := io.bot.SendMessage(msg)
	return err
}
