package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/blockdaemon/chain_sink/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	ConfigFiles = "BD_CONFIG_FILES"
	PREFIX      = "BD_"
)

func LoadConfig[T any]() (T, error) {

	var cfg T

	files := os.Getenv(ConfigFiles)
	if files == "" {
		return cfg, fmt.Errorf("no config files specified")
	}

	var cfgFile string
	fileSplits := strings.Split(files, " ")
	if len(fileSplits) >= 1 {
		cfgFile = fileSplits[0]
		fileSplits = fileSplits[1:]
	}

	v := viper.GetViper()
	v.SetEnvPrefix(PREFIX)
	v.AutomaticEnv()

	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		return cfg, err
	}

	for _, file := range fileSplits {
		v.SetConfigFile(file)
		if err := v.MergeInConfig(); err != nil {
			return cfg, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	defaults.SetDefaults(&cfg)

	if err := validator.New().Struct(cfg); err != nil {
		return cfg, fmt.Errorf("config validation failed: %w", err)
	}

	logger.Log.Debug("config loaded", zap.Any("config", cfg))

	return cfg, nil
}
