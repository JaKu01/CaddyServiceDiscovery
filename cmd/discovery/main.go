package main

import (
	"errors"
	"log/slog"

	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	"github.com/jaku01/caddyservicediscovery/internal/manager"
	"github.com/spf13/viper"
)

func main() {
	caddyAdminUrl, tlsConfig, err := loadConfiguration()
	if err != nil {
		panic(err)
	}
	slog.Info("Configuration: CaddyAdminUrl", "url", caddyAdminUrl)

	conn, err := newProviderConnector()
	if err != nil {
		panic(err)
	}

	caddyConnector := caddy.NewConnector(caddyAdminUrl, tlsConfig)
	if err = manager.StartServiceDiscovery(caddyConnector, conn); err != nil {
		panic(err)
	}
}

func loadConfiguration() (string, caddy.TLSConfig, error) {
	viper.SetConfigName("configuration")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("CaddyAdminUrl", "http://localhost:2019")
	viper.SetDefault("tls.manual", false)
	viper.SetDefault("tls.certFilePath", "/etc/certs/tls.crt")
	viper.SetDefault("tls.keyFilePath", "/etc/certs/tls.key")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return "", caddy.TLSConfig{}, err
		}
		slog.Warn("No configuration file found, using default values")
	} else {
		slog.Info("Configuration file loaded successfully")
	}

	caddyAdminUrl := viper.GetString("CaddyAdminUrl")

	var tlsCfg caddy.TLSConfig
	if err := viper.UnmarshalKey("tls", &tlsCfg); err != nil {
		return caddyAdminUrl, caddy.TLSConfig{}, err
	}

	return caddyAdminUrl, tlsCfg, nil
}
