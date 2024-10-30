package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	openProjectUrl        string
	openProjectToken      string
	tmetricToken          string
	clientIdInTmetric     int
	tmetricAPIBaseUrl     string
	tmetricAPIV3BaseUrl   string
	tmetricDummyProjectId int
}

func NewConfig() *Config {
	openProjectUrl := viper.GetString("openproject.url")
	if openProjectUrl == "" {
		fmt.Fprintln(os.Stderr, "openproject.url not set")
		os.Exit(1)
	}
	openProjectToken := viper.GetString("openproject.token")
	if openProjectToken == "" {
		fmt.Fprintln(os.Stderr, "openproject.token not set")
		os.Exit(1)
	}
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
	tmetricDummyProjectId := viper.GetInt("tmetric.dummyProjectId")
	if clientIdInTmetric == 0 {
		fmt.Fprintln(os.Stderr, "tmetric.dummyProjectId not set")
		os.Exit(1)
	}
	return &Config{
		openProjectUrl:        openProjectUrl,
		openProjectToken:      openProjectToken,
		tmetricToken:          tmetricToken,
		clientIdInTmetric:     clientIdInTmetric,
		tmetricAPIBaseUrl:     "https://app.tmetric.com/api/",
		tmetricAPIV3BaseUrl:   "https://app.tmetric.com/api/v3/",
		tmetricDummyProjectId: tmetricDummyProjectId,
	}
}
