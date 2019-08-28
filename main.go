package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/duo-labs/webauthn.io/config"
	log "github.com/duo-labs/webauthn.io/logger"
	"github.com/duo-labs/webauthn.io/models"
	"github.com/duo-labs/webauthn.io/server"
	_ "github.com/lib/pq"
)

func main() {
	config, err := config.LoadConfig("config.json")
	if err != nil {
		log.Info("error loading config.json")
		log.Fatal(err)
	}

	err = models.Setup(config)
	if err != nil {
		log.Info("error models.Setup(config) call")
		log.Fatal(err)
	}

	err = log.Setup(config)
	if err != nil {
		log.Info("error log.Setup(config) call")
		log.Fatal(err)
	}

	server, err := server.NewServer(config)
	if err != nil {
		log.Info("error server.NewServer")
		log.Fatal(err)
	}
	go server.Start()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)

	<-c
	log.Info("Shutting down...")
	server.Shutdown()
}
