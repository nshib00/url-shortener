package main

import (
	"flag"
	"fmt"
	"go-url-shortener/internal/config"
	"log"
	"log/slog"
	"os"
)

func setupLogger(envType config.EnvType) *slog.Logger {
	switch envType {
	case config.EnvLocal:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case config.EnvProd:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return nil
}

func main() {
	cfgTypeArg := flag.String("config", "local", "Путь до файла с конфигами")
	flag.Parse()

	envType, err := config.ParseEnvType(*cfgTypeArg)
	if err != nil {
		log.Fatal("main: wrong env type")
	}
	cfg := config.MustLoad(envType)
	log := setupLogger(envType)

	log.Info(fmt.Sprintf("main: starting app [%s]", slog.String("env", cfg.Env)))
	log.Debug("main: debug mode on")
}
