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
		stepID, ok := bw.getUserStep(botId, message.From)
		if !ok {
			return
		}

		// find next component for execute
		component, nextStepId, err := bw.findComponent(botId, stepID, message)
		if err != nil {
			bw.log.Errorw("failed find next component for execute", "error", err)
			return
		}

		// update stepId
		if nextStepId != stepID {
			if err := bw.redis.SetUserStep(botId, message.From.ID, nextStepId); err != nil {
				bw.log.Errorw("failed redis set user step", "error", err)

				// upd stepID in db
				if err := bw.db.SetUserStepByTgId(botId, message.From.ID, stepID); err != nil {
					bw.log.Errorw("failed update user step by tg id (db)", "error", err)
				}
			} else {
				// Async upd stepID in db
				go func() {
					if err := bw.db.SetUserStepByTgId(botId, message.From.ID, stepID); err != nil {
						bw.log.Errorw("failed update user step by tg id (db)", "error", err)
					}
				}()
			}
		}

		if err := bw.execute(bot, message, component); err != nil {
			bw.log.Errorw("message handler: failed execute method", "error", err)
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
		component, nextStepId, err := bw.findComponent(botId, stepID, message)
		if err != nil {
			bw.log.Errorw("failed find next component for execute", "error", err)
			return
		}

		// update stepId
		if nextStepId != stepID {
			if err := bw.redis.SetUserStep(botId, message.From.ID, nextStepId); err != nil {
				bw.log.Errorw("failed redis set user step", "error", err)

				// upd stepID in db
				if err := bw.db.SetUserStepByTgId(botId, message.From.ID, stepID); err != nil {
					bw.log.Errorw("failed update user step by tg id (db)", "error", err)
				}
			} else {
				// Async upd stepID in db
				go func() {
					if err := bw.db.SetUserStepByTgId(botId, message.From.ID, stepID); err != nil {
						bw.log.Errorw("failed update user step by tg id (db)", "error", err)
					}
				}()
			}
		}

		if err := bw.execute(bot, message, component); err != nil {
			bw.log.Errorw("command handler: failed exec method", "error", err)
		}
	}
}

// Determining the next step in the bot structure
func (bw *BotWorker) findComponent(botId int64, stepID int64, message *telego.Message) (*model.Component, int64, error) {
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

		// check cycle
		if _, ok := stepsPassed[stepID]; ok {
			if origStepID == stepID {
				break
			}
			return nil, 0, errors.New("cycle detected")
		}

		stepsPassed[stepID] = struct{}{}

		var err error
		component, err = bw.getComponent(botId, stepID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				stepID = config.MainComponentId
				continue
			}

			return nil, 0, errors.New("failed get component")
		}

		if origComponent == nil {
			origComponent = component
		}

		// check main component
		if component.IsMain {
			if component.NextStepId == nil || *component.NextStepId == stepID {
				return nil, 0, errors.New("error referring to the next component in the main component")
			}

			stepID = *component.NextStepId
			isFound = true
			continue
		}

		if isFound {
			// next component was found successfully
			break
		}

		// In this part there is a search for the ID of the next component.
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

	return component, stepID, nil
}

func (bw *BotWorker) execute(bot *telego.Bot, message *telego.Message, component *model.Component) error {
	action, ok := bw.components[*component.Data.Type]
	if ok {
		return action(bot, message, component)
	}

	return errors.New("Unknown component type")
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
