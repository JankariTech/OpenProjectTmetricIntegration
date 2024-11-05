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
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

type tableRow struct {
	Date                string
	TmetricEntry        string
	TmetricDuration     string
	OpenProjectEntry    string
	OpenProjectDuration string
	DiffInTime          string
}

var widthOfFixedColumns = 45 // rough size of all columns that have a fixed width

// tries to find out the width of the terminal and returns 80 if it fails
func getTerminalWidth() int {
	if term.IsTerminal(0) {
		terminalWidth, _, err := term.GetSize(0)
		if err != nil {
			return 80
		}
		return terminalWidth
	} else {
		return 80
	}
}

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "show the difference between the entries in tmetric and OpenProject",
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
		tmetricTimeEntries, err := tmetric.GetAllTimeEntries(config, tmetricUser, startDate, endDate)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		sort.Slice(tmetricTimeEntries, func(i, j int) bool {
			return tmetricTimeEntries[i].Note < tmetricTimeEntries[j].Note
		})

		openProjectTimeEntries, err := openproject.GetAllTimeEntries(config, startDate, endDate)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		sort.Slice(openProjectTimeEntries, func(i, j int) bool {
			return openProjectTimeEntries[i].Comment.Raw < openProjectTimeEntries[j].Comment.Raw
		})

		start, _ := time.Parse("2006-01-02", startDate)
		end, _ := time.Parse("2006-01-02", endDate)
		outputTable := table.NewWriter()
		outputTable.SetOutputMirror(os.Stdout)

		outputTable.AppendHeader(
			table.Row{"date", "tmetric entry", "tm\ndur", "OpenProject entry", "OP\ndur", "time\ndiff"},
		)
		widthContentColumns := int((getTerminalWidth() - widthOfFixedColumns) / 2)
		outputTable.SetColumnConfigs([]table.ColumnConfig{
			{Number: 2, WidthMax: widthContentColumns},
			{Number: 4, WidthMax: widthContentColumns},
		})

		for currentDay := start; !currentDay.After(end); currentDay = currentDay.AddDate(0, 0, 1) {
			row := tableRow{}
			row.Date = currentDay.Format("2006-01-02")
			sumDurationTmetric := 0
			for _, entry := range tmetricTimeEntries {
				entryDate, _ := time.Parse("2006-01-02", entry.StartTime[:10])
				if entryDate.Equal(currentDay) {
					workType, _ := entry.GetWorkType()
					description := fmt.Sprintf("Description: %v", entry.Note)
					project := fmt.Sprintf("- Project: %v", entry.Project.Name)
					wpId := fmt.Sprintf("- WP ID: %v", entry.Task.ExternalLink.IssueId)
					wp := fmt.Sprintf("- WP: %v", entry.Task.Name)
					workType = fmt.Sprintf("- Work Type: %v", workType)
					row.TmetricEntry += fmt.Sprintf(
						"%v\n%v\n%v\n%v\n%v\n\n",
						text.Snip(description, widthContentColumns, "~"),
						text.Snip(project, widthContentColumns, "~"),
						text.Snip(wpId, widthContentColumns, "~"),
						text.Snip(wp, widthContentColumns, "~"),
						text.Snip(workType, widthContentColumns, "~"),
					)
					duration, _ := entry.GetDuration()
					sumDurationTmetric += int(duration.Minutes())
					humanReadableDuration, _ := entry.GetHumanReadableDuration()

					row.TmetricDuration += fmt.Sprintf("%v\n\n\n\n\n\n", humanReadableDuration)
				}
			}
			sumDurationOpenProject := 0
			for _, entry := range openProjectTimeEntries {
				entryDate, _ := time.Parse("2006-01-02", entry.SpentOn)
				if entryDate.Equal(currentDay) {
					comment := fmt.Sprintf("Comment: %v", entry.Comment.Raw)
					project := fmt.Sprintf("- Project: %v", entry.Links.Project.Title)
					wpId := fmt.Sprintf("- WP ID: #%v", path.Base(entry.Links.WorkPackage.Href))
					wp := fmt.Sprintf("- WP: %v", entry.Links.WorkPackage.Title)
					activity := fmt.Sprintf("- Activity: %v", entry.Links.Activity.Title)
					row.OpenProjectEntry += fmt.Sprintf(
						"%v\n%v\n%v\n%v\n%v\n\n",
						text.Snip(comment, widthContentColumns, "~"),
						text.Snip(project, widthContentColumns, "~"),
						text.Snip(wpId, widthContentColumns, "~"),
						text.Snip(wp, widthContentColumns, "~"),
						text.Snip(activity, widthContentColumns, "~"),
					)
					duration, _ := entry.GetDuration()
					sumDurationOpenProject += int(duration.Minutes())
					humanReadableDuration, _ := entry.GetHumanReadableDuration()
					row.OpenProjectDuration += fmt.Sprintf("%v\n\n\n\n\n\n", humanReadableDuration)
				}
			}
			if sumDurationTmetric > sumDurationOpenProject {
				row.DiffInTime = strconv.Itoa(sumDurationTmetric - sumDurationOpenProject)
			} else {
				row.DiffInTime = strconv.Itoa(sumDurationOpenProject - sumDurationTmetric)
			}

			outputTable.AppendRow(table.Row{
				row.Date,
				strings.Trim(row.TmetricEntry, "\n"),
				strings.Trim(row.TmetricDuration, "\n"),
				strings.Trim(row.OpenProjectEntry, "\n"),
				strings.Trim(row.OpenProjectDuration, "\n"),
				row.DiffInTime,
			})
			outputTable.AppendSeparator()
		}
		outputTable.Render()
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	firstDayOfMonth := time.Now().Format("2006-01-02")
	firstDayOfMonth = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")

	diffCmd.Flags().StringVarP(&startDate, "start", "s", firstDayOfMonth, "start date")
	today := time.Now().Format("2006-01-02")
	diffCmd.Flags().StringVarP(&endDate, "end", "e", today, "end date")
}
