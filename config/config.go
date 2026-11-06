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
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	OpenProjectUrl                     string
	OpenProjectToken                   string
	OpenProjectTeam                    string
	OpenProjectTmetricClientIds        []int
	TmetricToken                       string
	ClientIdInTmetric                  int
	TmetricAPIBaseUrl                  string
	TmetricAPIV3BaseUrl                string
	TmetricDummyProjectId              int
	TmetricTagTransferredToOpenProject string
	TmetricExternalTaskLink            string
	JiraUrl                            string
	JiraToken                          string
	JiraTmetricClientIds               []int
	JiraUser                           string
}

func NewConfig() *Config {
	config := &Config{}
	errorMessage := ""
	openproject := viper.Sub("openproject")
	if openproject != nil {
		config.OpenProjectUrl = openproject.GetString("url")
		if config.OpenProjectUrl == "" {
			errorMessage = "openproject.url not set"
		}
		config.OpenProjectToken = openproject.GetString("token")
		if config.OpenProjectToken == "" {
			errorMessage = "openproject.token not set"
		}
		config.OpenProjectTeam = openproject.GetString("team")
		config.OpenProjectTmetricClientIds = openproject.GetIntSlice("tmetricClientIds")
		if len(config.OpenProjectTmetricClientIds) == 0 {
			errorMessage = "openproject.tmetricClientIds not set"
		}
	}

	jira := viper.Sub("jira")
	if jira != nil {
		config.JiraUrl = jira.GetString("url")
		if config.JiraUrl == "" {
			errorMessage = "jira.url not set"
		}
		config.JiraUser = jira.GetString("user")
		if config.JiraUser == "" {
			errorMessage = "jira.user not set"
		}
		config.JiraToken = jira.GetString("token")
		if config.JiraToken == "" {
			errorMessage = "jira.token not set"
		}
		config.JiraTmetricClientIds = jira.GetIntSlice("tmetricClientIds")
		if len(config.JiraTmetricClientIds) == 0 {
			errorMessage = "jira.tmetricClientIds not set"
		}
	}

	if jira == nil && openproject == nil {
		errorMessage = "neither jira nor openproject settings are set"
	}

	config.TmetricToken = viper.GetString("tmetric.token")
	if config.TmetricToken == "" {
		errorMessage = "tmetric.token not set"
	}

	config.ClientIdInTmetric = viper.GetInt("tmetric.clientId")
	if config.ClientIdInTmetric != 0 {
		errorMessage =
			"tmetric.clientId is obsolete! Use `openproject.tmetricClientIds` and `jira.tmetricClientIds`"
	}

	config.TmetricDummyProjectId = viper.GetInt("tmetric.dummyProjectId")
	if config.TmetricDummyProjectId == 0 {
		errorMessage = "tmetric.dummyProjectId not set"
	}

	if errorMessage != "" {
		_, _ = fmt.Fprintln(os.Stderr, errorMessage)
		os.Exit(1)
	}

	config.TmetricAPIBaseUrl = "https://app.tmetric.com/api/"
	config.TmetricAPIV3BaseUrl = "https://app.tmetric.com/api/v3/"
	config.TmetricTagTransferredToOpenProject = "transferred-to-openproject"
	// this value has always to be "https://community.openproject.org"
	// otherwise tmetric does not recognize the integration and does not allow to create the external task
	config.TmetricExternalTaskLink = "https://community.openproject.org/"

	return config
}
