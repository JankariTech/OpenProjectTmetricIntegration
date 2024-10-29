/*
Copyright © 2024 JankariTech <info@jankaritech.com>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var startDate string
var endDate string

func validateOpenProjectWorkPackage(input string) error {
	_, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		return errors.New("Invalid WP")
	}

	return nil
}

func getAllTimeEntries(config *Config, tmetricUser *TmetricUser) ([]TimeEntry, error) {
	httpClient := resty.New()
	resp, err := httpClient.R().
		SetAuthToken(config.tmetricToken).
		Get(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries?userId=%v&startDate=%v&endDate=%v`,
				config.tmetricAPIBaseUrl,
				tmetricUser.Accounts[0].Id,
				tmetricUser.Id,
				startDate,
				endDate,
			),
		)
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"cannot read timeentries. Error: '%v'. HTTP status code:", err, resp.StatusCode(),
		)
	}

	var timeEntries []TimeEntry
	err = json.Unmarshal(resp.Body(), &timeEntries)
	if err != nil {
		return nil, fmt.Errorf("error parsing time entries response: %v\n", err)
	}

	return timeEntries, nil
}

// tmetricCmd represents the tmetric command
var tmetricCmd = &cobra.Command{
	Use:   "tmetric",
	Short: "check the validity of the tmetric data",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return fmt.Errorf("start date is not in the format YYYY-MM-DD")
		}
		_, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return fmt.Errorf("end date is not in the format YYYY-MM-DD")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		config := NewConfig()

		httpClient := resty.New()
		tmetricUser := NewTmetricUser()
		timeEntries, err := getAllTimeEntries(config, tmetricUser)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		var filteredEntries []TimeEntry
		for _, entry := range timeEntries {
			if entry.Project.Client.Id == config.clientIdInTmetric {
				filteredEntries = append(filteredEntries, entry)
			}
		}

		var entriesWithoutIssue []TimeEntry
		for _, entry := range filteredEntries {
			if entry.Task.ExternalLink.IssueId == "" {
				entriesWithoutIssue = append(entriesWithoutIssue, entry)
			}
		}

		if len(entriesWithoutIssue) > 0 {
			fmt.Println("Some time-entries do not have any workpackages assigned")
		}

		openProjectToken := viper.Get("openproject.token").(string)
		openProjectUrl := viper.Get("openproject.url").(string)
		openProjectHttpClient := resty.New()

		for _, entry := range entriesWithoutIssue {
			prompt := promptui.Prompt{
				Label:    fmt.Sprintf("%v %v %v", entry.Note, entry.StartTime, entry.EndTime),
				Validate: validateOpenProjectWorkPackage,
			}

			workpackageFoundOnOpenProject := false
			for !workpackageFoundOnOpenProject {

				workPackageId, err := prompt.Run()

				if err != nil {
					fmt.Printf("Prompt failed %v\n", err)
					return
				}
				wpURL, _ := url.JoinPath(openProjectUrl, "/api/v3/work_packages/", workPackageId)
				resp, err := openProjectHttpClient.R().
					SetBasicAuth("apikey", openProjectToken).
					Get(wpURL)

				if err == nil && resp.StatusCode() == 200 {
					workpackageFoundOnOpenProject = true
				} else {
					fmt.Printf("Could not find WP in %v\n", openProjectUrl)
				}

				var workPackage WorkPackage
				err = json.Unmarshal(resp.Body(), &workPackage)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error parsing work packages response or no work packages found: %v\n", err)
					return
				}

				prompt = promptui.Prompt{
					Label:     fmt.Sprintf("WP: %q. Subject: %q. Updatey", workPackageId, workPackage.Subject),
					IsConfirm: true,
				}
				updateTmetricConfirmation, err := prompt.Run()
				if strings.ToLower(updateTmetricConfirmation) == "y" {
					fmt.Printf("updating t-metric entry '%v'\n", entry.Note)

					resp, err = httpClient.R().
						SetAuthToken(config.tmetricToken).
						SetHeader("Content-Type", "application/json").
						SetBody(fmt.Sprintf(
							`{"task": {"externalLink": { "caption": "%v", "link": "%v", "issueId":"%v"}}}`,
							workPackageId, wpURL, workPackageId,
						)).
						Put(
							fmt.Sprintf(
								`%vaccounts/%v/timeentries/%v`,
								config.tmetricAPIBaseUrl,
								tmetricUser.Accounts[0].Id,
								entry.Id,
							),
						)

					if err != nil || resp.StatusCode() != 200 {
						fmt.Fprintf(os.Stderr, "could not update time entry\n")
						if err != nil {
							fmt.Fprintf(os.Stderr, "error: %v\n", err)
						} else {
							fmt.Fprintf(os.Stderr, "status code: %v\n", resp.StatusCode())
						}
						return
					}
				}

			}
		}
	},
}

func init() {
	checkCmd.AddCommand(tmetricCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tmetricCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tmetricCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	firstDayOfMonth := time.Now().Format("2006-01-02")
	firstDayOfMonth = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")

	tmetricCmd.Flags().StringVarP(&startDate, "start", "s", firstDayOfMonth, "start date")
	today := time.Now().Format("2006-01-02")
	tmetricCmd.Flags().StringVarP(&endDate, "end", "e", today, "end date")
}
