package app

import (
	"github.com/botscubes/bot-worker/internal/bot"
	"github.com/botscubes/bot-worker/internal/config"
	"github.com/botscubes/bot-worker/internal/database/pgsql"
	"go.uber.org/zap"

	rdb "github.com/botscubes/bot-worker/internal/database/redis"
)

type App struct {
	Log           *zap.SugaredLogger
	Config        *config.ServiceConfig
	Redis         *rdb.Rdb
	Db            *pgsql.Db
	WebhookServer *bot.WebhookServer
	Worker        *bot.BotWorker
}

func CreateApp(logger *zap.SugaredLogger, c *config.ServiceConfig) *App {
	redis := rdb.NewClient(&c.Redis)

	pgsqlUrl := "postgres://" + c.Pg.User + ":" + c.Pg.Pass + "@" + c.Pg.Host + ":" + c.Pg.Port + "/" + c.Pg.Db
	db, err := pgsql.OpenConnection(pgsqlUrl)
	if err != nil {
		logger.Fatalw("Open PostgreSQL connection", "error", err)
	}

	defer db.CloseConnection()

	app := &App{
		Log:           logger,
		Config:        c,
		Redis:         redis,
		Db:            db,
		WebhookServer: bot.NewWebhookServer(logger, c),
		Worker:        bot.NewBotWorker(logger, c, redis, db),
	}

	return app
}

func (app *App) Run() {
	// TODO: try error
	go func() {
		if err := app.WebhookServer.Start(); err != nil {
			app.Log.Fatalw("Start webhook server", "error", err)
		}
	}()
}
