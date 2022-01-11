package logging

type Config struct {
	Level       string `yaml:"level" mapstructure:"level"`
	Encoding    string `yaml:"encoding" mapstructure:"encoding"`
	Development bool   `yaml:"development" mapstructure:"development"`
}
