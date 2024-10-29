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
				config.tmetricAPIV3BaseUrl,
				tmetricUser.ActiveAccountId,
				tmetricUser.Id,
				startDate,
				endDate,
			),
		)
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"cannot read timeentries. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}

	var timeEntries []TimeEntry
	err = json.Unmarshal(resp.Body(), &timeEntries)
	if err != nil {
		return nil, fmt.Errorf("error parsing time entries response: %v\n", err)
	}

	return timeEntries, nil
}

/*
*
this is the only way to create an external task in tmetric.
This task is needed to have an issueId of OpenProject assigned to a time entry.
*/
func createDummyTimeEntry(
	workPackage WorkPackage, tmetricUser *TmetricUser, config *Config, openProjectUrl string,
) (*TimeEntry, error) {
	dummyTimeEntry := newDummyTimeEntry(workPackage, openProjectUrl, config.tmetricDummyProjectId)
	dummyTimerString, _ := json.Marshal(dummyTimeEntry)
	httpClient := resty.New()
	resp, err := httpClient.R().
		SetAuthToken(config.tmetricToken).
		SetHeader("Content-Type", "application/json").
		SetBody(dummyTimerString).
		Post(fmt.Sprintf(
			`%vaccounts/%v/timer/issue`,
			config.tmetricAPIBaseUrl,
			tmetricUser.ActiveAccountId,
		))
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"could not create dummy time entry. Error : '%v'. HTTP-Status-Code: %v",
			err, resp.StatusCode(),
		)
	}

	resp, err = httpClient.R().
		SetAuthToken(config.tmetricToken).
		SetQueryString("userId=" + strconv.Itoa(tmetricUser.Id)).
		Get(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries/latest`,
				config.tmetricAPIV3BaseUrl,
				tmetricUser.ActiveAccountId,
			),
		)

	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"could not find latest time entry. Error : '%v'. HTTP-Status-Code: %v",
			err, resp.StatusCode(),
		)
	}
	latestTimeEntry := TimeEntry{}
	err = json.Unmarshal(resp.Body(), &latestTimeEntry)

	if err != nil || latestTimeEntry.Note != "to-delete-only-created-to-create-an-external-task" {
		return nil, fmt.Errorf(
			"could not find dummy time entry",
		)
	}
	return &latestTimeEntry, nil
}

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

		tmetricUser := NewTmetricUser()
		timeEntries, err := getAllTimeEntries(config, tmetricUser)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		// get all entries that belong to the client and do not have an external link
		var entriesWithoutIssue []TimeEntry
		for _, entry := range timeEntries {
			if entry.Project.Client.Id == config.clientIdInTmetric && entry.Task.ExternalLink.IssueId == "" {
				entriesWithoutIssue = append(entriesWithoutIssue, entry)
			}
		}

		if len(entriesWithoutIssue) > 0 {
			fmt.Println("Some time-entries do not have any workpackages assigned")
		}

		openProjectUrl := viper.Get("openproject.url").(string)

		for _, entry := range entriesWithoutIssue {
			prompt := promptui.Prompt{
				Label:    fmt.Sprintf("%v %v %v. Provide a WP number to be assigned to this time-entry", entry.Note, entry.StartTime, entry.EndTime),
				Validate: validateOpenProjectWorkPackage,
			}

			workpackageFoundOnOpenProject := false
			for !workpackageFoundOnOpenProject {

				workPackageId, err := prompt.Run()

				if err != nil {
					fmt.Printf("Prompt failed %v\n", err)
					return
				}
				workPackage, err := getWorkpackage(workPackageId)

				if err == nil {
					workpackageFoundOnOpenProject = true
				} else {
					fmt.Printf("%v\n", err)
					continue
				}

				prompt = promptui.Prompt{
					Label: fmt.Sprintf(
						"WP: %v. Subject: %v. Update t-metric entry?", workPackage.Id, workPackage.Subject,
					),
					IsConfirm: true,
				}
				updateTmetricConfirmation, err := prompt.Run()
				if strings.ToLower(updateTmetricConfirmation) == "y" {
					fmt.Printf("updating t-metric entry '%v'\n", entry.Note)
					latestTimeEntry, err := createDummyTimeEntry(workPackage, tmetricUser, config, openProjectUrl)
					if err != nil {
						fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
						return
					}
					_ = latestTimeEntry.delete(*config, *tmetricUser)
					entry.Task = latestTimeEntry.Task
					entry.update(*config, *tmetricUser)
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
