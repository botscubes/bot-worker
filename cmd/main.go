package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/botscubes/bot-service/pkg/logger"
	a "github.com/botscubes/bot-worker/internal/app"
	"github.com/botscubes/bot-worker/internal/config"
)

func main() {
	c, err := config.GetConfig()
	if err != nil {
		fmt.Println("Get config: ", err)
		return
	}

	log, err := logger.NewLogger(logger.Config{
		Type: c.LoggerType,
	})
	if err != nil {
		fmt.Println("Create logger: ", err)
		return
	}

	done := make(chan struct{}, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Info("Stopping...")

		done <- struct{}{}
	}()

	app := a.CreateApp(log, c)
	app.Run()

	log.Info("App started")

	<-done
	log.Info("App done")
}
