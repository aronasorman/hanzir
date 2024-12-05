package main

import (
	"log"

	"github.com/aronasorman/hanzir/api/internal"
	"github.com/kelseyhightower/envconfig"
	"github.com/samber/do"
	"github.com/sashabaranov/go-openai"
)

type Config struct {
	ServerPort string `envconfig:"PORT" default:"8080"`
	OpenAIKey  string `envconfig:"OPENAI_API_KEY" required:"true"`
}

func NewOpenAIService(i *do.Injector) (*openai.Client, error) {
	cfg := do.MustInvoke[*Config](i)
	return openai.NewClient(cfg.OpenAIKey), nil
}

func InitConfig(i *do.Injector) (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}
	return &cfg, nil
}

func main() {
	injector := do.New()
	do.Provide[*Config](injector, InitConfig)
	do.Provide[*openai.Client](injector, NewOpenAIService)

	runServer(injector)
}

func runServer(i *do.Injector) {
	cfg := do.MustInvoke[*Config](i)
	r := internal.InitRoutes(i)
	log.Fatal(r.Run(":" + cfg.ServerPort))
}
