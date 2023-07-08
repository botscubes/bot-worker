package bot

import (
	"strconv"
	"time"

	"github.com/botscubes/bot-worker/internal/config"
	"github.com/botscubes/bot-worker/internal/database/pgsql"
	rdb "github.com/botscubes/bot-worker/internal/database/redis"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"
)

const handlerTimeout = 10 // sec

type BotWorker struct {
	config        *config.ServiceConfig
	log           *zap.SugaredLogger
	redis         *rdb.Rdb
	db            *pgsql.Db
	webhookServer *WebhookServer
	botHandlers   map[int64]*th.BotHandler
}

func NewBotWorker(logger *zap.SugaredLogger, c *config.ServiceConfig, r *rdb.Rdb, db *pgsql.Db, ws *WebhookServer) *BotWorker {
	return &BotWorker{
		config:        c,
		log:           logger,
		redis:         r,
		db:            db,
		webhookServer: ws,
		botHandlers:   make(map[int64]*th.BotHandler),
	}
}

func (bw *BotWorker) RunBot(botId int64, token string) error {
	bot, err := telego.NewBot(token, telego.WithHealthCheck(), telego.WithDefaultDebugLogger())
	if err != nil {
		return err
	}

	bw.log.Infow("Starting bot", "botId", botId)

	updates, err := bot.UpdatesViaWebhook(
		strconv.FormatInt(botId, 10),
		telego.WithWebhookServer(&telego.NoOpWebhookServer{
			RegisterHandlerFunc: bw.webhookServer.RegisterBot,
		}),
	)
	if err != nil {
		return err
	}

	botHandler, err := th.NewBotHandler(bot, updates, th.WithStopTimeout(time.Second*handlerTimeout))
	if err != nil {
		return err
	}

	bw.botHandlers[botId] = botHandler

	// Set middlerwares
	botHandler.Use(th.PanicRecovery)
	botHandler.Use(bw.registerUser(botId))

	// Set handlers
	// Handle command
	botHandler.Handle(bw.commandHandler(botId),
		th.Union(
			th.AnyCommand(),
		))

	// Handle message
	botHandler.Handle(bw.messageHandler(botId),
		th.Union(
			th.AnyMessage(),
			th.AnyEditedMessage(),
		))

	go botHandler.Start()

	return nil
}

func (bw *BotWorker) StopBot(botId int64) {
	bot, ok := bw.botHandlers[botId]
	if ok {
		bot.Stop()
	} else {
		bw.log.Warnw("bot handler not found", "botId", botId)
	}

	bw.webhookServer.RemoveBot(botId)
}
