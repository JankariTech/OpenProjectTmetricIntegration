package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"os"
)

type TmetricAccount struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type TmetricUser struct {
	Id       int              `json:"id"`
	Name     string           `json:"name"`
	Accounts []TmetricAccount `json:"accounts"`
}

func NewTmetricUser() *TmetricUser {
	config := NewConfig()

	httpClient := resty.New()

	resp, err := httpClient.R().
		SetAuthToken(config.tmetricToken).
		Get(config.tmetricAPIBaseUrl + "user")
	if err != nil || resp.StatusCode() != 200 {
		fmt.Fprintf(os.Stderr, "cannot reach tmetric server\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "status code: %v\n", resp.StatusCode())
		}
		os.Exit(1)
	}
	var user TmetricUser
	err = json.Unmarshal(resp.Body(), &user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(user.Accounts) == 0 {
		fmt.Fprintf(os.Stderr, "could not find accountID in response: %v\n", resp)
		os.Exit(1)
	}
	return &user
}
