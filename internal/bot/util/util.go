package util

import (
	"github.com/botscubes/bot-worker/internal/model"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func Keyboard(commands *model.Commands, _ *model.Keyboard) *telego.ReplyKeyboardMarkup {
	// markup (keyboard): unused - for future

	rows := [][]telego.KeyboardButton{}
	for _, v := range *commands {
		row := []telego.KeyboardButton{
			tu.KeyboardButton(*v.Data),
		}
		rows = append(rows, row)
	}

	return tu.Keyboard(rows...)
}
