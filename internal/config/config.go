package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  HTTPServer
	Auth        Auth
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8319"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Auth struct {
	SecretKey string `env:"JWT_SECRET" env-required:"true"`
}

func MustLoad(envType EnvType) *Config {
	if envType == EnvUnknown {
		log.Fatal("config: config path is not set")
	}
	configPath := envType.String()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config: config file %s does not exist", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("config: file read error: %v", err)
	}
	return &cfg
}
