package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/url"
)

type WorkPackage struct {
	Subject string `json:"subject"`
	Id      int    `json:"id"`
}

func getWorkpackage(workPackageId string, config *Config) (WorkPackage, error) {
	openProjectHttpClient := resty.New()

	wpURL, _ := url.JoinPath(config.openProjectUrl, "/api/v3/work_packages/", workPackageId)
	resp, err := openProjectHttpClient.R().
		SetBasicAuth("apikey", config.openProjectToken).
		Get(wpURL)

	if err != nil || resp.StatusCode() != 200 {
		return WorkPackage{}, fmt.Errorf("could not find WP in %v", config.openProjectUrl)
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
