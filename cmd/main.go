package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/botscubes/bot-service/pkg/logger"
	a "github.com/botscubes/bot-worker/internal/app"
	"github.com/botscubes/bot-worker/internal/config"
	"github.com/botscubes/bot-worker/internal/database/pgsql"
	"github.com/nats-io/nats.go"
)

func main() {
	c, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Get config: %v\n", err)
		return
	}

	log, err := logger.NewLogger(logger.Config{
		Type: c.LoggerType,
	})
	if err != nil {
		fmt.Printf("Create logger: %v\n", err)
		return
	}

	defer func() {
		if err := log.Sync(); err != nil {
			log.Error(err)
		}
	}()

	pgsqlUrl := "postgres://" + c.Pg.User + ":" + c.Pg.Pass + "@" + c.Pg.Host + ":" + c.Pg.Port + "/" + c.Pg.Db

	db, err := pgsql.OpenConnection(pgsqlUrl)
	if err != nil {
		log.Fatalw("Open PostgreSQL connection", "error", err)
	}

	defer db.CloseConnection()

	nc, err := nats.Connect(c.NatsURL, nats.MaxReconnects(-1))
	if err != nil {
		log.Fatalw("NATS connection", "error", err)
	}
	defer nc.Drain() //nolint:errcheck

	app := a.CreateApp(log, c, db, nc)

	done := make(chan struct{}, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Info("Stopping...")

		err = app.Shutdown()
		if err != nil {
			log.Fatalw("Shutdown", "error", err)
		}

		done <- struct{}{}
	}()

	app.Run()

	log.Info("App started")

	<-done
	log.Info("App done")
}
