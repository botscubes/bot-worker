package app

import (
	"github.com/botscubes/bot-worker/internal/bot"
	"github.com/botscubes/bot-worker/internal/config"
	"github.com/botscubes/bot-worker/internal/database/pgsql"
	"go.uber.org/zap"

	rdb "github.com/botscubes/bot-worker/internal/database/redis"
)

type App struct {
	log           *zap.SugaredLogger
	config        *config.ServiceConfig
	redis         *rdb.Rdb
	db            *pgsql.Db
	webhookServer *bot.WebhookServer
	worker        *bot.BotWorker
}

func CreateApp(logger *zap.SugaredLogger, c *config.ServiceConfig, db *pgsql.Db) *App {
	redis := rdb.NewClient(&c.Redis)
	webhookServer := bot.NewWebhookServer(logger, c)

	return &App{
		log:           logger,
		config:        c,
		redis:         redis,
		db:            db,
		webhookServer: webhookServer,
		worker:        bot.NewBotWorker(logger, c, redis, db, webhookServer),
	}
}

func (app *App) Run() {
	go func() {
		if err := app.webhookServer.Start(); err != nil {
			app.log.Fatalw("Start webhook server", "error", err)
		}
	}()

	if err := app.launchBots(); err != nil {
		app.log.Fatalw("Launch bots", "error", err)
	}
}

func (app *App) Shutdown() error {
	return app.webhookServer.Shutdown()
}

func (app *App) launchBots() error {
	bots, err := app.db.GetRunningBots()
	if err != nil {
		return err
	}

	n := 0
	for _, bot := range *bots {
		if err := app.worker.RunBot(bot.Id, *bot.Token); err != nil {
			app.log.Errorw("launch bot", "botId", bot.Id, "error", err)
		} else {
			n++
		}
	}

	app.log.Infof("Bot launched: %d", n)

	return nil
}
