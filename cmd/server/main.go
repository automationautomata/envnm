package main

import (
	"context"
	"envmn/config"
	"envmn/internal/bootstrap"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	envFilePath := flag.String("env-file", "", "path to .env")
	flag.Parse()

	var (
		cfg config.StartupConfig
		err error
	)
	if *envFilePath == "" {
		cfg, err = config.Load[config.StartupConfig]()
	} else {
		cfg, err = config.LoadFromEnvFile[config.StartupConfig](*envFilePath)
	}
	if err != nil {
		log.Fatal(err)
	}

	app, err := bootstrap.Build(cfg)
	if err != nil {
		log.Fatal("Failed to build server:", err)
	}

	shutdown := app.Run()
	log.Println("Server started successfully")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdown(ctx)
}
