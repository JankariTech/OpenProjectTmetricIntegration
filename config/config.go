/*
Copyright © 2024 JankariTech Pvt. Ltd. info@jankaritech.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	OpenProjectUrl                     string
	OpenProjectToken                   string
	TmetricToken                       string
	ClientIdInTmetric                  int
	TmetricAPIBaseUrl                  string
	TmetricAPIV3BaseUrl                string
	TmetricDummyProjectId              int
	TmetricTagTransferredToOpenProject string
	TmetricExternalTaskLink            string
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
		OpenProjectUrl:                     openProjectUrl,
		OpenProjectToken:                   openProjectToken,
		TmetricToken:                       tmetricToken,
		ClientIdInTmetric:                  clientIdInTmetric,
		TmetricAPIBaseUrl:                  "https://app.tmetric.com/api/",
		TmetricAPIV3BaseUrl:                "https://app.tmetric.com/api/v3/",
		TmetricDummyProjectId:              tmetricDummyProjectId,
		TmetricTagTransferredToOpenProject: "transferred-to-openproject",
		// this value has always to be "https://community.openproject.org"
		// otherwise tmetric does not recognize the integration and does not allow to create the external task
		TmetricExternalTaskLink:            "https://community.openproject.org/",
	}
}
