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
		timeEntries, err := tmetric.GetAllTimeEntries(config, tmetricUser, startDate, endDate)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		var filteredEntries []tmetric.TimeEntry
		for _, entry := range timeEntries {
			hasTransferredTag := false
			for _, tag := range entry.Tags {
				if tag.Name == config.TmetricTagTransferredToOpenProject {
					hasTransferredTag = true
					break
				}
			}
			if !hasTransferredTag {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		if len(tmetric.GetEntriesWithoutWorkType(filteredEntries)) > 0 {
			_, _ = fmt.Fprintln(
				os.Stderr,
				"Some time-entries do not have any work-type assigned, run the 'check tmetric' command to fix it",
			)
			os.Exit(1)
		}

		if len(tmetric.GetEntriesWithoutLinkToOpenProject(filteredEntries)) > 0 {
			_, _ = fmt.Fprintln(
				os.Stderr,
				"Some time-entries are not linked to an OpenProject work-package, run the 'check tmetric' command to fix it",
			)
			os.Exit(1)
		}

		for _, tmetricTimeEntry := range filteredEntries {
			issueId, err := tmetricTimeEntry.GetIssueIdAsInt()
			if err != nil {
				_, _ = fmt.Fprintln(
					os.Stderr, err,
				)
				os.Exit(1)
			}

			workType, err := tmetricTimeEntry.GetWorkType()
			if err != nil {
				_, _ = fmt.Fprintf(
					os.Stderr,
					"Error with time entry '%v' in project '%v'\nError: %v\n",
					tmetricTimeEntry.Note,
					tmetricTimeEntry.Project.Name,
					err,
				)
				os.Exit(1)
			}

			activity, err := openproject.NewFromWorkType(*config, issueId, workType)
			if err != nil {
				_, _ = fmt.Fprintf(
					os.Stderr,
					"Error with time entry '%v' in project '%v'\nError: %v\n",
					tmetricTimeEntry.Note,
					tmetricTimeEntry.Project.Name,
					err,
				)
				os.Exit(1)
			}

			openProjectTimeEntry, err := tmetricTimeEntry.ConvertToOpenProjectTimeEntry(activity)
			if err != nil {
				_, _ = fmt.Fprintf(
					os.Stderr,
					"could not convert time entry '%v' in project '%v' started at '%v' from tmetric to OpenProject\n"+
						"Error: %v\n",
					tmetricTimeEntry.Note, tmetricTimeEntry.Project, tmetricTimeEntry.StartTime, err,
				)
				os.Exit(1)
			}

			err = openProjectTimeEntry.Save(*config)
			if err != nil {
				_, _ = fmt.Fprintf(
					os.Stderr,
					"could not save time entry '%v' for work package '%v' spend on '%v' in OpenProject\n"+
						"Error: %v\n",
					openProjectTimeEntry.Comment.Raw,
					filepath.Base(openProjectTimeEntry.Links.WorkPackage.Href),
					openProjectTimeEntry.SpentOn,
					err,
				)
				os.Exit(1)
			}

			tmetricTimeEntry.TagAsTransferredToOpenProject(*config)
			tmetricUser = tmetric.NewUser()
			err = tmetricTimeEntry.Update(*config, *tmetricUser)
			if err != nil {
				_, _ = fmt.Fprintf(
					os.Stderr,
					"could not tag tmetric entry as being transferred to openproject\n"+
						"Error: %v\n",
					err,
				)
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
