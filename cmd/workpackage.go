package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"net/url"
)

type WorkPackage struct {
	Subject string `json:"subject"`
	Id      int    `json:"id"`
}

func getWorkpackage(workPackageId string) (WorkPackage, error) {
	openProjectToken := viper.Get("openproject.token").(string)
	openProjectUrl := viper.Get("openproject.url").(string)
	openProjectHttpClient := resty.New()

	wpURL, _ := url.JoinPath(openProjectUrl, "/api/v3/work_packages/", workPackageId)
	resp, err := openProjectHttpClient.R().
		SetBasicAuth("apikey", openProjectToken).
		Get(wpURL)

	if err != nil || resp.StatusCode() != 200 {
		return WorkPackage{}, fmt.Errorf("could not find WP in %v", openProjectUrl)
	}

	var workPackage WorkPackage
	err = json.Unmarshal(resp.Body(), &workPackage)
	if err != nil {
		return WorkPackage{}, fmt.Errorf(
			"error parsing work packages response or no work packages found: %v", err,
		)
	}
	return workPackage, err
}
