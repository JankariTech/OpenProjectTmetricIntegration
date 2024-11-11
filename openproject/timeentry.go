package openproject

import (
	"encoding/json"
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/go-chrono/chrono"
	"github.com/go-resty/resty/v2"
	"net/url"
	"time"
)

type TimeEntry struct {
	Ongoing bool `json:"ongoing"`
	Comment struct {
		Raw string `json:"raw"`
	} `json:"comment"`
	SpentOn string `json:"spentOn"`
	Hours   string `json:"hours"`
	Links   struct {
		WorkPackage struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"workPackage"`
		Activity struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"activity"`
		Project struct {
			Href  string `json:"href,omitempty"`
			Title string `json:"title,omitempty"`
		} `json:"project,omitempty"`
		User struct {
			Href  string `json:"href,omitempty"`
			Title string `json:"title,omitempty"`
		} `json:"user,omitempty"`
	} `json:"_links"`
}

func (timeEntry TimeEntry) Save(config config.Config) error {
	entryJSON, err := json.Marshal(timeEntry)
	if err != nil {
		return fmt.Errorf("error marshalling OpenProject time entry to JSON: %v", err)
	}
	httpClient := resty.New()
	wpURL, _ := url.JoinPath(config.OpenProjectUrl, "/api/v3/time_entries/")
	resp, err := httpClient.R().
		SetBasicAuth("apikey", config.OpenProjectToken).
		SetHeader("Content-Type", "application/hal+json; charset=utf-8").
		SetBody(entryJSON).
		Post(wpURL)
	if err != nil || resp.StatusCode() != 201 {
		return fmt.Errorf(
			"could not save time entry in OpenProject. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}
	return nil
}
func (timeEntry TimeEntry) GetDuration() (time.Duration, error) {
	_, duration, err := chrono.ParseDuration(timeEntry.Hours)
	if err != nil {
		return 0, fmt.Errorf("could not parse duration: %v", err)
	}
	return time.Duration(duration.Nanoseconds()), nil
}

func (timeEntry TimeEntry) GetHumanReadableDuration() (string, error) {
	duration, err := timeEntry.GetDuration()
	if err != nil {
		return "", err
	}
	readableDuration := fmt.Sprintf("%02d:%02d", int(duration.Hours()), int(duration.Minutes())%60)

	return readableDuration, nil
}
