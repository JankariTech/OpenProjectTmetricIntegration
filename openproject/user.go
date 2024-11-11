package openproject

import (
	"encoding/json"
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"net/url"
)

type User struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

// FindUserByName searches for users that match the given search and returns the first match.
func FindUserByName(config *config.Config, search string) (User, error) {
	httpClient := resty.New()
	openProjectUrl, _ := url.JoinPath(config.OpenProjectUrl, "/api/v3/principals")
	resp, err := httpClient.R().
		SetBasicAuth("apikey", config.OpenProjectToken).
		SetHeader("Content-Type", "application/json").
		SetQueryParam("sortBy", "[[\"name\",\"desc\"]]").
		SetQueryParam("filters", fmt.Sprintf(
			`[{"status":{"operator":"!","values":["3"]}},{"type":{"operator":"=","values":["User"]}},{"name":{"operator":"~","values":["%v"]}}]`,
			search),
		).
		Get(openProjectUrl)
	if err != nil || resp.StatusCode() != 200 {
		return User{}, fmt.Errorf(
			"cannot lookup users in OpenProject. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}
	var users []User
	usersJSON := gjson.GetBytes(resp.Body(), "_embedded.elements")
	err = json.Unmarshal([]byte(usersJSON.String()), &users)
	if err != nil {
		return User{}, fmt.Errorf(
			"error parsing user search response from OpenProject: %v", err,
		)
	}
	if len(users) > 0 {
		return users[0], nil
	}
	return User{}, fmt.Errorf(
		"cannot find a user in OpenProject with a name matching '%v'", search,
	)
}
