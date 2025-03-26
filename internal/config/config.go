package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env       string `yaml:"env" env:"ENV" env-default:"local" env-required:"true"`
	StoragePath   string `yaml:"storage_path" env-required:"true"`
	HTTPServer HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Adress     string `yaml:"address" env-default:"localhost:8080"`
	Timeout    time.Duration    `yaml:"timeout" env-default:"4"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60"`
}

func MustLoad() *Config {
	err := godotenv.Load("../../.env")		
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return &cfg


}

