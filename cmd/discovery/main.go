package main

import (
	"errors"
	"log"

	"github.com/jaku01/caddyservicediscovery/internal/manager"
	"github.com/spf13/viper"
)

func main() {
	caddyAdminUrl, err := loadConfiguration()
	if err != nil {
		panic(err)
	}
	log.Printf("Configuration: CaddyAdminUrl=%s", caddyAdminUrl)

	err = manager.StartServiceDiscovery(caddyAdminUrl)
	if err != nil {
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
		log.Println("No configuration file found, using default values")
	} else {
		log.Println("Configuration file loaded successfully")
	}

	return viper.GetString("CaddyAdminUrl"), nil
}
