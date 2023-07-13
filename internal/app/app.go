package app

import (
	"github.com/botscubes/bot-worker/internal/bot"
	mb "github.com/botscubes/bot-worker/internal/broker"
	ct "github.com/botscubes/bot-worker/internal/components"
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
	mb            mb.Broker
}

func CreateApp(logger *zap.SugaredLogger, c *config.ServiceConfig, db *pgsql.Db, b mb.Broker) *App {
	redis := rdb.NewClient(&c.Redis)
	webhookServer := bot.NewWebhookServer(logger, c)
	worker := bot.NewBotWorker(logger, c, redis, db, webhookServer)
	b.SetWorker(worker)

	return &App{
		log:           logger,
		config:        c,
		redis:         redis,
		db:            db,
		webhookServer: webhookServer,
		worker:        worker,
		mb:            b,
	}
}

func (app *App) Run() {
	app.RegisterComponents()

	go func() {
		if err := app.webhookServer.Start(); err != nil {
			app.log.Fatalw("Start webhook server", "error", err)
		}
	}()

	if err := app.mb.StartBotSub(); err != nil {
		app.log.Fatalw("Broker subscribe: start bot ", "error", err)
	}

	if err := app.mb.StopBotSub(); err != nil {
		app.log.Fatalw("Broker subscribe: stop bot ", "error", err)
	}

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
	for _, b := range *bots {
		if err := app.worker.RunBot(b.Id, *b.Token); err != nil {
			app.log.Errorw("launch bot", "botId", b.Id, "error", err)
		} else {
			n++
		}
	}

	app.log.Infof("Bot launched: %d", n)

	return nil
}

func (app *App) RegisterComponents() {
	app.worker.RegisterComponent("text", ct.Text)
}
