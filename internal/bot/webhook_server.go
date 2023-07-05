package bot

import (
	"errors"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/mymmrac/telego"

	"github.com/botscubes/bot-worker/internal/config"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type WebhookServer struct {
	log         *zap.SugaredLogger
	config      *config.ServiceConfig
	server      *fiber.App
	botHandlers map[int64]telego.WebhookHandler
}

func NewWebhookServer(logger *zap.SugaredLogger, c *config.ServiceConfig) *WebhookServer {
	server := fiber.New(fiber.Config{
		AppName:               "Bot Webhook Server",
		DisableStartupMessage: true,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		ErrorHandler:          errorHandler(logger),
	})

	return &WebhookServer{
		log:         logger,
		config:      c,
		server:      server,
		botHandlers: make(map[int64]telego.WebhookHandler),
	}
}

func (w *WebhookServer) RegisterBot(path string, handler telego.WebhookHandler) error {
	botID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		return err
	}

	w.botHandlers[botID] = handler

	return nil
}

func (w *WebhookServer) RemoveBot(botId int64) {
	delete(w.botHandlers, botId)
}

func (w *WebhookServer) Start() error {
	w.server.Post("/webhook/bot/:botID<int>", w.botHandler)
	return w.server.Listen(w.config.ListenAddress)
}

func (w *WebhookServer) botHandler(ctx *fiber.Ctx) error {
	botID, err := strconv.ParseInt(ctx.Params("botID"), 10, 64)
	if err != nil {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	handler, ok := w.botHandlers[botID]
	if !ok {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	if err := handler(ctx.Body()); err != nil {
		w.log.Errorw("webhook bot handler", "bot-id", botID, "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (w *WebhookServer) Shutdown() error {
	return w.server.ShutdownWithTimeout(config.ShutdownTimeout)
}

func errorHandler(log *zap.SugaredLogger) func(ctx *fiber.Ctx, err error) error {
	return func(ctx *fiber.Ctx, err error) error {
		// Status code defaults to 500
		code := fiber.StatusInternalServerError

		// Retrieve the custom status code if it's a *fiber.Error
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
		}

		log.Errorf("Bot Worker panic recovered: %v", err)

		return ctx.Status(code).SendString("Error")
	}
}
