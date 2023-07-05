package bot

import (
	bu "github.com/botscubes/bot-worker/internal/bot/util"
	"github.com/botscubes/bot-worker/internal/model"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func sendMessage(bot *telego.Bot, message *telego.Message, component *model.Component) error {
	mes := tu.Message(
		tu.ID(message.Chat.ID),
		*(*component.Data.Content)[0].Text,
	)

	if len(*component.Commands) > 0 {
		mes.WithReplyMarkup(bu.Keyboard(component.Commands, component.Keyboard).WithResizeKeyboard())
	}

	_, err := bot.SendMessage(mes)
	return err
}
