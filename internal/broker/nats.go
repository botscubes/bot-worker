package broker

import (
	"github.com/botscubes/bot-worker/internal/bot"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsBroker struct {
	nc     *nats.Conn
	log    *zap.SugaredLogger
	worker *bot.BotWorker
}

func NewNatsBroker(natsURL string, logger *zap.SugaredLogger) (*NatsBroker, error) {
	nc, err := nats.Connect(natsURL, nats.MaxReconnects(-1))
	if err != nil {
		return nil, err
	}

	return &NatsBroker{
		nc:  nc,
		log: logger,
	}, nil
}

func (b *NatsBroker) CloseConnection() {
	b.nc.Drain() //nolint:errcheck
}

func (b *NatsBroker) SetWorker(w *bot.BotWorker) {
	b.worker = w
}

func (b *NatsBroker) StartBotSub() error {
	if _, err := b.nc.Subscribe("worker.start", b.onStartBot); err != nil {
		return err
	}

	return nil
}

func (b *NatsBroker) StopBotSub() error {
	if _, err := b.nc.Subscribe("worker.stop", b.onStopBot); err != nil {
		return err
	}

	return nil
}
