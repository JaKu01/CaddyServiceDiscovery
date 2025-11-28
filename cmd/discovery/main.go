package main

import (
	"errors"
	"log"
	"log/slog"

	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	"github.com/jaku01/caddyservicediscovery/internal/manager"
	"github.com/spf13/viper"
)

func main() {
	caddyConfig, err := loadConfiguration()
	if err != nil {
		panic(err)
	}
	log.Println(caddyConfig.String())
	slog.Info("Configuration: CaddyAdminUrl", "url", caddyConfig.CaddyAdminUrl)

	providerConnector, err := newServiceDiscoveryProviderConnector()
	if err != nil {
		panic(err)
	}

	caddyConnector := caddy.NewConnector(caddyConfig)
	if err = manager.StartServiceDiscovery(caddyConnector, providerConnector); err != nil {
		panic(err)
	}
}

func loadConfiguration() (caddy.CaddyConfig, error) {
	viper.SetConfigName("configuration")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("CaddyAdminUrl", "http://localhost:2019")
	viper.SetDefault("tls.manual", false)
	viper.SetDefault("tls.certFilePath", "/etc/certs/tls.crt")
	viper.SetDefault("tls.keyFilePath", "/etc/certs/tls.key")

	viper.SetDefault("manualRoutes.routes", []map[string]interface{}{})

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return caddy.CaddyConfig{}, err
		}
		slog.Warn("No configuration file found, using default values")
	} else {
		slog.Info("Configuration file loaded successfully")
	}

	caddyAdminUrl := viper.GetString("CaddyAdminUrl")

	caddyTlsConfig := getCaddyTlsConfig()

	var manualRoutes []caddy.ManualRoute
	if err := viper.UnmarshalKey("manualRoutes.routes", &manualRoutes); err != nil {
		slog.Warn("Failed to unmarshal manual routes, using defaults", "error", err)
		manualRoutes = []caddy.ManualRoute{}
	}

	return caddy.CaddyConfig{
		TLSConfig:     caddyTlsConfig,
		CaddyAdminUrl: caddyAdminUrl,
		ManualRoutes:  manualRoutes,
	}, nil
}

func getCaddyTlsConfig() caddy.TLSConfig {
	var tlsConfig caddy.TLSConfig
	useDefaults := false

	if err := viper.UnmarshalKey("tls", &tlsConfig); err != nil {
		slog.Warn("[TLS-Config] Failed to unmarshal TLS config", "error", err)
		useDefaults = true
	}

	if tlsConfig.CertFilePath == "" {
		slog.Warn("[TLS-Config] No cert file path specified")
		useDefaults = true
	}

	if tlsConfig.KeyFilePath == "" {
		slog.Warn("[TLS-Config] No key file path specified")
		useDefaults = true
	}

	if useDefaults {
		return caddy.TLSConfig{}
	}
	return tlsConfig
}
