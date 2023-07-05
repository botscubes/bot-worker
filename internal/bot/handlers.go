package bot

import (
	"errors"
	"strings"

	"github.com/botscubes/bot-worker/internal/config"
	"github.com/botscubes/bot-worker/internal/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// Handles incoming Message & EditedMessage from Telegram.
func (bw *BotWorker) messageHandler(botId int64) th.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		var message *telego.Message
		if update.Message != nil {
			message = update.Message
		} else {
			message = update.EditedMessage
		}

		// Get user stepID
		stepID, err := bw.getUserStep(botId, message.From)
		if err != nil {
			return
		}

		// find next component for execute
		ok, component, nextStepId := bw.findComponent(botId, stepID, message)
		if !ok {
			return
		}

		if nextStepId != stepID {
			if err := bw.redis.SetUserStep(botId, message.From.ID, nextStepId); err != nil {
				bw.log.Error(err)
			}
			// Async upd stepID in db
			go bw.setUserStep(botId, message.From.ID, nextStepId)
		}

		if err := bw.execMethod(bot, message, component); err != nil {
			bw.log.Error(err)
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

		stepID := int64(config.MainComponentId)

		// find next component for execute
		ok, component, nextStepId := bw.findComponent(botId, stepID, message)
		if !ok {
			return
		}

		if nextStepId != stepID {
			if err := bw.redis.SetUserStep(botId, message.From.ID, nextStepId); err != nil {
				bw.log.Error(err)
			}
			// Async upd stepID in db
			go bw.setUserStep(botId, message.From.ID, nextStepId)
		}

		if err := bw.execMethod(bot, message, component); err != nil {
			bw.log.Error(err)
		}
	}
}

// Determining the next step in the bot structure
func (bw *BotWorker) findComponent(botId int64, stepID int64, message *telego.Message) (bool, *model.Component, int64) {
	var origComponent *model.Component
	var component *model.Component
	origStepID := stepID

	// for cycle detect
	stepsPassed := make(map[int64]struct{})
	isFound := false

	for {
		// This part of the loop (before the "isFound" condition) is used to automatically
		// skip the starting component and undefined components.
		// Also, the next component is selected here by the id found in the second part of the cycle.

		if _, ok := stepsPassed[stepID]; ok {
			if origStepID == stepID {
				break
			}

			// TODO: return errors
			// bw.log.Warnf("cycle detected: bot #%d", botId)
			return false, nil, 0
		}

		stepsPassed[stepID] = struct{}{}

		var err error
		component, err = bw.getComponent(botId, stepID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				stepID = config.MainComponentId
				continue
			}

			return false, nil, 0
		}

		if origComponent == nil {
			origComponent = component
		}

		// check main component
		if component.IsMain {
			if component.NextStepId == nil || *component.NextStepId == stepID {
				// TODO: return errors
				// bw.log.Warnf("error referring to the next component in the main component: bot #%d", botId)
				return false, nil, 0
			}

			stepID = *component.NextStepId
			isFound = true
			continue
		}

		if isFound {
			// next component was found successfully
			break
		}

		// In this part, the id of the next component is determined.
		// In case of successful identification of the ID, an additional check occurs in the first part of the cycle.

		isFound = true

		if component.NextStepId != nil {
			stepID = *component.NextStepId
			continue
		}

		command := findComponentCommand(&message.Text, component.Commands)
		if command != nil && command.NextStepId != nil {
			stepID = *command.NextStepId
			continue
		}

		// next component not found, will be executed initial (current) component
		component = origComponent
		stepID = origStepID
		break
	}

	return true, component, stepID
}

func (bw *BotWorker) execMethod(bot *telego.Bot, message *telego.Message, component *model.Component) error {
	switch *component.Data.Type {
	case "text":
		if err := sendMessage(bot, message, component); err != nil {
			return err
		}
	default:
		bw.log.Warn("Unknown type method: ", *component.Data.Type)
	}

	return nil
}

// Determine commnad by !message text!
func findComponentCommand(mes *string, commands *model.Commands) *model.Command {
	// work with command type - text
	for _, command := range *commands {
		// The comparison is not case sensitive
		if strings.EqualFold(*command.Data, *mes) {
			return command
		}
	}

	return nil
}

func commandEqual(messageText string, command string) bool {
	matches := th.CommandRegexp.FindStringSubmatch(messageText)
	if len(matches) != th.CommandMatchGroupsLen {
		return false
	}

	return strings.EqualFold(matches[1], command)
}
