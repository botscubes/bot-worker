package bot

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func (bw *BotWorker) registerUser(botId int64) th.Middleware {
	return func(bot *telego.Bot, update telego.Update, next th.Handler) {
		var user *telego.User

		// Get user ID from the message
		if update.Message != nil && update.Message.From != nil {
			user = update.Message.From
		}

		// Get user ID from the callback query
		if update.CallbackQuery != nil {
			user = &update.CallbackQuery.From
		}

		ex, err := bw.storage.checkUserExist(botId, user.ID)
		if err != nil {
			bw.log.Errorw("failed check user exists", "error", err)
		}
		if !ex {
			bw.log.Debugw("register user", "botId,", botId, "user", user.ID)
			if err = bw.storage.addUser(botId, user); err != nil {
				bw.log.Errorw("failed register user", "error", err)
				return

			}
		}

		next(bot, update)
	}
}
