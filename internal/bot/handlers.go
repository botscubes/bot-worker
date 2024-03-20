package bot

import (
	"strings"

	"github.com/botscubes/bot-components/context"
	"github.com/botscubes/bot-components/exec"
	"github.com/botscubes/bot-worker/internal/config"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// Handles incoming Message & EditedMessage from Telegram.
func (bw *BotWorker) messageHandler(botId int64) th.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		bw.log.Infow("handle user action", "botId", botId, "user", update.Message.From)
		//	chatID := update.Message.Chat.ID
		//	bot.SendMessage(
		//		tu.Message(
		//			tu.ID(chatID),
		//			update.Message.Text,
		//		),
		//	)
		var groupId int64 = config.MainGroupId
		components, err := bw.storage.components(botId, groupId)
		if err != nil {
			bw.log.Errorw("failed get components", "error", err)
			return
		}
		userId := update.Message.From.ID
		ctx, err := bw.storage.context(botId, groupId, userId)
		if err != nil {
			bw.log.Errorw("failed get context", "error", err)
			return
		}
		step, err := bw.storage.userStep(botId, groupId, userId)
		if err != nil {
			bw.log.Errorw("failed get user step", "error", err)
			return
		}
		if step == 0 {
			bw.log.Info("")
			return
		}
		e := exec.NewExecutor(ctx, NewBotIO(bot, &update))

		for {
			componentData, ok := components[step]
			if !ok {
				break
			}
			bw.log.Info(string(componentData.Data))
			component, err := componentData.Component()
			if err != nil {
				bw.log.Errorw("failed get component", "error", err)

				break
			}
			st, err := e.Execute(component)
			if err != nil {
				var val any = err.Error()
				ctx.SetValue("error", &val)
			}
			if st == nil {
				break
			}
			step = *st

		}
	}
}

// Handles incoming Message with command (eg. /start) from Telegram.
// So far only /start command !!!
func (bw *BotWorker) commandHandler(botId int64) th.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		message := update.Message

		if !commandEqual(message.Text, "start") {
			return
		}
		bw.log.Infow("user start bot", "user", update.Message.From)

		stepId := int64(config.MainComponentId)
		groupId := int64(config.MainGroupId)
		userId := update.Message.From.ID
		bw.storage.setUserStep(botId, groupId, userId, stepId)
		bw.storage.redis.SetUserContext(botId, groupId, userId, context.NewContext())

	}
}

// Determining the next step in the bot structure
//
//	func (bw *BotWorker) findComponent(botId int64, stepID int64, message *telego.Message) (bool, *model.Component, int64) {
//		var origComponent *model.Component
//		var component *model.Component
//		origStepID := stepID
//
//		// for cycle detect
//		stepsPassed := make(map[int64]struct{})
//		isFound := false
//
//		for {
//			// This part of the loop (before the "isFound" condition) is used to automatically
//			// skip the starting component and undefined components.
//			// Also, the next component is selected here by the id found in the second part of the cycle.
//
//			// check cycle
//			if _, ok := stepsPassed[stepID]; ok {
//				if origStepID == stepID {
//					component = origComponent
//					stepID = origStepID
//					break
//				}
//
//				bw.log.Debugw("find component for execute", "message", "cycle detected")
//				return false, nil, 0
//			}
//
//			stepsPassed[stepID] = struct{}{}
//
//			// get component
//			var err error
//			component, err = bw.getComponent(botId, stepID)
//			if err != nil {
//				if errors.Is(err, ErrNotFound) {
//					stepID = config.MainComponentId
//					continue
//				}
//
//				return false, nil, 0
//			}
//
//			if origComponent == nil {
//				origComponent = component
//			}
//
//			// check main component
//			if component.IsMain {
//				if component.NextStepId == nil || *component.NextStepId == stepID {
//					bw.log.Debugw("find component for execute", "message", "error referring to the next component in the main component")
//					return false, nil, 0
//				}
//
//				stepID = *component.NextStepId
//				isFound = true
//				continue
//			}
//
//			if isFound {
//				// next component was found successfully
//				break
//			}
//
//			// In this part there is a search for the ID of the next component.
//			// In case of successful identification of the ID, an additional check occurs in the first part of the cycle.
//
//			isFound = true
//
//			if component.NextStepId != nil {
//				stepID = *component.NextStepId
//				continue
//			}
//
//			command := findComponentCommand(&message.Text, component.Commands)
//			if command != nil && command.NextStepId != nil {
//				stepID = *command.NextStepId
//				continue
//			}
//
//			// next component not found, will be executed initial (current) component
//			component = origComponent
//			stepID = origStepID
//			break
//		}
//
//		return true, component, stepID
//	}
//
//	func (bw *BotWorker) execute(bot *telego.Bot, message *telego.Message, component *model.Component) error {
//		action, ok := bw.components[*component.Data.Type]
//		if ok {
//			return action(bot, message, component)
//		}
//
//		return errors.New("unknown component type")
//	}
//
// // Determine commnad by !message text!
//
//	func findComponentCommand(mes *string, commands *model.Commands) *model.Command {
//		// work with command type - text
//		for _, command := range *commands {
//			// The comparison is not case sensitive
//			if strings.EqualFold(*command.Data, *mes) {
//				return command
//			}
//		}
//
//		return nil
//	}
func commandEqual(messageText string, command string) bool {
	matches := th.CommandRegexp.FindStringSubmatch(messageText)
	if len(matches) != th.CommandMatchGroupsLen {
		return false
	}

	return strings.EqualFold(matches[1], command)
}
