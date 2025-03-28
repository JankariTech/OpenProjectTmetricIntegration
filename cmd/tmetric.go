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

package cmd

import (
	"errors"
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/JankariTech/OpenProjectTmetricIntegration/openproject"
	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
	"time"
)

func validateOpenProjectWorkPackage(input string) error {
	if len(input) > 0 {
		_, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			return errors.New("Invalid WP")
		}
	}
	return nil
}

func handleEntriesWithoutIssue(timeEntries []tmetric.TimeEntry, tmetricUser tmetric.User, config *config.Config) error {
	entriesWithoutLinkToOpenProject := tmetric.GetEntriesWithoutLinkToOpenProject(config, timeEntries)
	if len(entriesWithoutLinkToOpenProject) > 0 {
		fmt.Println("Some time-entries do not have any workpackages assigned")
	}

	spinner := newSpinner()
	defer spinner.Stop()

	for _, entry := range entriesWithoutLinkToOpenProject {
		prompt := promptui.Prompt{
			Label: fmt.Sprintf(
				"%v => %v %v-%v. Provide a WP number to be assigned to this time-entry (Enter to skip)",
				entry.Project.Name, entry.Note, entry.StartTime, entry.EndTime,
			),
			Validate: validateOpenProjectWorkPackage,
		}

		workpackageFoundOnOpenProject := false
		for !workpackageFoundOnOpenProject {

			workPackageId, err := prompt.Run()

			if err != nil {
				return fmt.Errorf("prompt failed: %v", err)
			}

			if workPackageId == "" {
				fmt.Println("skipping")
				workpackageFoundOnOpenProject = true
				continue
			}
			workPackage, err := openproject.GetWorkpackage(workPackageId, config)

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
				spinner.Start()
				spinner.Prefix = fmt.Sprintf("Updating t-metric entry '%v' ", entry.Note)
				spinner.FinalMSG = "❌\n"
				latestTimeEntry, err := tmetric.CreateDummyTimeEntry(workPackage, tmetricUser, config)
				if err != nil {
					return err
				}
				_ = latestTimeEntry.Delete(*config, tmetricUser)
				entry.Task = latestTimeEntry.Task
				err = entry.Update(*config, tmetricUser)
				if err != nil {
					return err
				}
				spinner.FinalMSG = "✔️\n"
				spinner.Stop()
			}
		}
	}
	return nil
}

func handleEntriesWithoutWorkType(timeEntries []tmetric.TimeEntry, tmetricUser tmetric.User, config *config.Config) error {
	entriesWithoutWorkType := tmetric.GetEntriesWithoutWorkType(timeEntries)
	if len(entriesWithoutWorkType) > 0 {
		fmt.Println("Some time-entries do not have any work type assigned")
	}

	spinner := newSpinner()
	defer spinner.Stop()
	for _, entry := range entriesWithoutWorkType {
		possibleWorkTypes, err := entry.GetPossibleWorkTypes(*config, tmetricUser)

		if err != nil {
			return err
		}

		if len(possibleWorkTypes) == 0 {
			return fmt.Errorf("could not find any work types for project %v", entry.Project.Name)
		}

		var promptItems []string
		for _, workType := range possibleWorkTypes {
			promptItems = append(promptItems, workType.Name)
		}
		promptItems = append(promptItems, "skip")
		promptSelectWorkType := promptui.Select{
			Label: fmt.Sprintf(
				"%v => %v %v-%v. Select work-type",
				entry.Project.Name, entry.Note, entry.StartTime, entry.EndTime,
			),
			Items: promptItems,
		}
		_, workType, err := promptSelectWorkType.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %v", err)
		}
		if workType == "skip" {
			fmt.Println("skipping")
			continue
		}

		promptConfirm := promptui.Prompt{
			Label: fmt.Sprintf(
				"Work-type: %v. Update t-metric entry", workType,
			),
			IsConfirm: true,
		}
		updateTmetricConfirmation, err := promptConfirm.Run()
		if strings.ToLower(updateTmetricConfirmation) == "y" {
			entry.Tags = append(entry.Tags, tmetric.Tag{Name: workType, IsWorkType: true})
			spinner.Start()
			spinner.Prefix = fmt.Sprintf("Updating t-metric entry '%v' ", entry.Note)
			spinner.FinalMSG = "❌\n"

			err = entry.Update(*config, tmetricUser)
			if err != nil {
				return err
			}
			spinner.FinalMSG = "✔️\n"
			spinner.Stop()
		}
	}
	return nil
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
		config := config.NewConfig()

		tmetricUser := tmetric.NewUser()
		timeEntries, err := tmetric.GetAllTimeEntries(config, tmetricUser, startDate, endDate)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		err = handleEntriesWithoutIssue(timeEntries, tmetricUser, config)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		// after and update time entries receive a new id, so we need to fetch them again
		timeEntries, err = tmetric.GetAllTimeEntries(config, tmetricUser, startDate, endDate)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		err = handleEntriesWithoutWorkType(timeEntries, tmetricUser, config)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
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
