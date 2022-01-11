package storages

type Config struct {
	Sqlite *SqliteConfig `yaml:"sqlite" mapstructure:"sqlite"`
}

type SqliteConfig struct {
	Path string `yaml:"path" mapstructure:"path"`
}
