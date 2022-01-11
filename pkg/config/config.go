package config

import (
	"goquotebot/internal/monitoring/logging"
	"goquotebot/internal/monitoring/metrics"
	"goquotebot/pkg/storages"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var configFile string

// Config holds the configuration file
type Config struct {
	Telegram TelegramConfig  `yaml:"telegram" mapstructure:"telegram"`
	Logger   logging.Config  `yaml:"logger" mapstructure:"logger"`
	Metrics  metrics.Config  `yaml:"metrics" mapstructure:"metrics"`
	Storage  storages.Config `yaml:"storage" mapstructure:"storage"`
}

type TelegramConfig struct {
	Token   string `yaml:"token" mapstructure:"token"`
	GroupID string `yaml:"group_id" mapstructure:"group_id"`
}

// RegisterFlags overwrite the configuration with parameter passed with flags
func (cfg *Config) RegisterFlags(flags *pflag.FlagSet) {
	flags.StringVar(&configFile, "config", "config.yaml", "configuration file to use")
}

// RegisterConfigFile loads the new config file is another than default is specified
func (cfg *Config) RegisterConfigFile() error {
	v := viper.New()
	if configFile != "" {
		v.SetConfigFile(configFile)
		err := v.ReadInConfig()
		if err != nil {
			return err
		}
		err = v.Unmarshal(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}
