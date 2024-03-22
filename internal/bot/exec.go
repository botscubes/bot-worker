package bot

import (
	"errors"

	"github.com/botscubes/bot-components/context"
	"github.com/botscubes/bot-components/exec"
	"github.com/botscubes/bot-components/io"
	"github.com/botscubes/bot-worker/internal/config"
)

func (bw *BotWorker) execute(botId int64, groupId int64, userId int64, io io.IO, step int64, ctx *context.Context) error {
	components, err := bw.storage.components(botId, groupId)
	if err != nil {
		return err
	}
	const MAX_VISIT = config.MaxLoopInExecution
	visitedComponents := make(map[int64]int64)
	e := exec.NewExecutor(ctx, io)
	for {
		_, ok := visitedComponents[step]
		if ok {
			visitedComponents[step]++
		} else {
			visitedComponents[step] = 1
		}
		if visitedComponents[step] > MAX_VISIT {

			return errors.New("Loop too long")
		}

		componentData, ok := components[step]
		if !ok {
			break
		}

		bw.log.Debug(string(componentData.Data))
		component, err := componentData.Component()
		if err != nil {
			return err
		}
		st, err := e.Execute(component)
		if err != nil {
			var val any = err.Error()
			ctx.SetValue("error", &val)
		}
		if st == nil {
			step = 0
			break
		}
		if step == *st {
			break
		}

		step = *st
	}

	err = bw.storage.setUserStep(botId, groupId, userId, step)
	if err != nil {
		return err
	}
	err = bw.storage.setContext(botId, groupId, userId, ctx)
	if err != nil {
		return err
	}
	return nil
}
