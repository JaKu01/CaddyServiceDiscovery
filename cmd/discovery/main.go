package main

import (
	"errors"
	"log/slog"

	"github.com/jaku01/caddyservicediscovery/internal/manager"
	"github.com/spf13/viper"
)

func main() {
	caddyAdminUrl, err := loadConfiguration()
	if err != nil {
		panic(err)
	}
	slog.Info("Configuration: CaddyAdminUrl", "url", caddyAdminUrl)

	conn, err := newProviderConnector()
	if err != nil {
		panic(err)
	}

	if err = manager.StartServiceDiscovery(caddyAdminUrl, conn); err != nil {
		panic(err)
	}
}

func loadConfiguration() (string, error) {
	viper.SetConfigName("configuration")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("CaddyAdminUrl", "http://localhost:2019")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return "", err
		}
		slog.Warn("No configuration file found, using default values")
	} else {
		slog.Info("Configuration file loaded successfully")
	}

	return viper.GetString("CaddyAdminUrl"), nil
}
