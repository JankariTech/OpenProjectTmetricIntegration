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
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/JankariTech/OpenProjectTmetricIntegration/openproject"
	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

func checkTmetricEntries(tmetricUser tmetric.User, config *config.Config) ([]tmetric.TimeEntry, error) {
	spinner := newSpinner()
	defer spinner.Stop()
	spinner.Prefix = "Checking time entries in Tmetric... "
	spinner.FinalMSG = "❌\n"
	timeEntries, err := tmetric.GetAllTimeEntries(config, tmetricUser, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(tmetric.GetEntriesWithoutWorkType(timeEntries)) > 0 {
		return nil, fmt.Errorf(
			"some time-entries do not have any work-type assigned, run the 'check tmetric' command to fix it",
		)
	}

	if len(tmetric.GetEntriesWithoutLinkToOpenProject(config, timeEntries)) > 0 {
		return nil, fmt.Errorf(
			"some time-entries are not linked to an OpenProject work-package, run the 'check tmetric' command to fix it",
		)
	}

	filteredEntries := tmetric.GetEntriesNotTransferredToOpenProject(
		timeEntries, config.TmetricTagTransferredToOpenProject,
	)

	spinner.FinalMSG = "✔️\n"
	return filteredEntries, err
}

func transferEntryToOpenProject(
	tmetricTimeEntry tmetric.TimeEntry, tmetricUser tmetric.User, config *config.Config,
) error {
	spinner := newSpinner()
	defer spinner.Stop()
	spinner.FinalMSG = "❌\n"
	spinner.Prefix = fmt.Sprintf(
		"Transferring data to OpenProject. Project: '%v', Note: '%v', Start: '%v' ", tmetricTimeEntry.Project.Name, tmetricTimeEntry.Note, tmetricTimeEntry.StartTime,
	)
	spinner.Restart()
	issueId, err := tmetricTimeEntry.GetIssueIdAsInt()
	if err != nil {
		return err
	}

	workType, err := tmetricTimeEntry.GetWorkType()
	if err != nil {
		return fmt.Errorf(
			"Error with time entry '%v' in project '%v'\nError: %v\n",
			tmetricTimeEntry.Note,
			tmetricTimeEntry.Project.Name,
			err,
		)
	}

	activity, err := openproject.NewActivityFromWorkType(*config, issueId, workType)
	if err != nil {
		return fmt.Errorf(
			"Error with time entry '%v' in project '%v'\nError: %v\n",
			tmetricTimeEntry.Note,
			tmetricTimeEntry.Project.Name,
			err,
		)
	}

	openProjectTimeEntry, err := tmetricTimeEntry.ConvertToOpenProjectTimeEntry(activity)
	if err != nil {
		return fmt.Errorf(
			"could not convert time entry '%v' in project '%v' started at '%v' from tmetric to OpenProject\n"+
				"Error: %v\n",
			tmetricTimeEntry.Note, tmetricTimeEntry.Project, tmetricTimeEntry.StartTime, err,
		)
	}

	err = openProjectTimeEntry.Save(*config)
	if err != nil {
		return fmt.Errorf(
			"could not save time entry '%v' for work package '%v' spend on '%v' in OpenProject\n"+
				"Error: %v\n",
			openProjectTimeEntry.Comment.Raw,
			filepath.Base(openProjectTimeEntry.Links.WorkPackage.Href),
			openProjectTimeEntry.SpentOn,
			err,
		)
	}

	tmetricTimeEntry.TagAsTransferredToOpenProject(*config)
	tmetricUser = tmetric.NewUser()
	err = tmetricTimeEntry.Update(*config, tmetricUser)
	if err != nil {
		return fmt.Errorf(
			"could not tag tmetric entry as being transferred to openproject\n"+
				"Error: %v\n",
			err,
		)
	}
	spinner.FinalMSG = "✔️\n"
	return nil
}

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy time entries from Tmetric to OpenProject",
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

		filteredEntries, err := checkTmetricEntries(tmetricUser, config)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		for _, tmetricTimeEntry := range filteredEntries {
			err = transferEntryToOpenProject(tmetricTimeEntry, tmetricUser, config)
			if err != nil {
				_, _ = fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	firstDayOfMonth := time.Now().Format("2006-01-02")
	firstDayOfMonth = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")

	copyCmd.Flags().StringVarP(&startDate, "start", "s", firstDayOfMonth, "start date")
	today := time.Now().Format("2006-01-02")
	copyCmd.Flags().StringVarP(&endDate, "end", "e", today, "end date")
}
