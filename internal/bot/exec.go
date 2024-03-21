package bot

import (
	"github.com/botscubes/bot-components/context"
	"github.com/botscubes/bot-components/exec"
	"github.com/botscubes/bot-components/io"
)

func (bw *BotWorker) execute(botId int64, groupId int64, userId int64, io io.IO, step int64, ctx *context.Context) error {
	components, err := bw.storage.components(botId, groupId)
	if err != nil {
		return err
	}

	e := exec.NewExecutor(ctx, io)
	for {
		componentData, ok := components[step]
		if !ok {
			break
		}
		bw.log.Info(string(componentData.Data))
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
