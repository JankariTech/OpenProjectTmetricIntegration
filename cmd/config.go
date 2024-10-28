package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	tmetricToken      string
	clientIdInTmetric int
	tmetricAPIBaseUrl string
}

func NewConfig() *Config {
	tmetricToken := viper.GetString("tmetric.token")
	if tmetricToken == "" {
		fmt.Fprintln(os.Stderr, "tmetric.token not set")
		os.Exit(1)
	}
	clientIdInTmetric := viper.GetInt("tmetric.clientId")
	if clientIdInTmetric == 0 {
		fmt.Fprintln(os.Stderr, "tmetric.clientId not set")
		os.Exit(1)
	}
	return &Config{
		tmetricToken:      tmetricToken,
		clientIdInTmetric: clientIdInTmetric,
		tmetricAPIBaseUrl: "https://app.tmetric.com/api/v3/",
	}
}