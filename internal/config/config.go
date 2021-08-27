package config

import (
	"github.com/ujum/dictap/pkg/config/loader"
)

const (
	envPrefix  = "dictup"
	configType = "yaml"
	prefix     = "config"
)

type Config struct {
	Server     *ServerConfig
	Logger     *LoggerConfig
	Datasource *DatasourceConfig
}

type ServerConfig struct {
	Port int
}

type LoggerConfig struct {
	Level string
}

type DatasourceConfig struct {
	Mongo *MongoDatasourceConfig
}

type MongoDatasourceConfig struct {
	Port   int
	Schema string
	Host   string
}

func New(configDir string) (*Config, error) {
	appConfig := defaultValue()
	settings := &loader.LoadSettings{
		LoadSysEnv:     true,
		ConfigDir:      configDir,
		FileNamePrefix: prefix,
		ConfigType:     configType,
		EnvPrefix:      envPrefix,
	}
	err := loader.Load(appConfig, settings)
	return appConfig, err
}
