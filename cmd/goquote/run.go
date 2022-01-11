package main

import (
	"goquotebot/internal/monitoring/logging"
	"goquotebot/pkg/config"
	t "goquotebot/pkg/telegram"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func run(cfg *config.Config) error {
	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	logger, err := logging.InitZap(cfg.Logger)
	if err != nil {
		return err
	}

	logger.Info("configuration ", zap.ByteString("config :\n", cfgBytes))

	server, err := t.NewServer(logger, cfg)
	if err != nil {
		return err
	}
	server.Start()
	defer server.Stop()

	return nil
}
