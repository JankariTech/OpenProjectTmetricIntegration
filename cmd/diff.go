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
	"github.com/spf13/cobra"
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

		for currentDay := start; !currentDay.After(end); currentDay = currentDay.AddDate(0, 0, 1) {
			row := tableRow{}
			row.Date = currentDay.Format("2006-01-02")
			sumDurationTmetric := 0
			for _, entry := range tmetricTimeEntries {
				entryDate, _ := time.Parse("2006-01-02", entry.StartTime[:10])
				if entryDate.Equal(currentDay) {
					row.TmetricEntry += fmt.Sprintf("%v => %v\n", entry.Project.Name, entry.Note)
					duration, _ := entry.GetDuration()
					sumDurationTmetric += int(duration.Minutes())
					humanReadableDuration, _ := entry.GetHumanReadableDuration()

					row.TmetricDuration += fmt.Sprintf("%v\n", humanReadableDuration)
				}
			}
			sumDurationOpenProject := 0
			for _, entry := range openProjectTimeEntries {
				entryDate, _ := time.Parse("2006-01-02", entry.SpentOn)
				if entryDate.Equal(currentDay) {
					row.OpenProjectEntry += fmt.Sprintf(
						"%v (#%v) => %v\n",
						entry.Links.WorkPackage.Title,
						path.Base(entry.Links.WorkPackage.Href),
						entry.Comment.Raw,
					)
					duration, _ := entry.GetDuration()
					sumDurationOpenProject += int(duration.Minutes())
					humanReadableDuration, _ := entry.GetHumanReadableDuration()
					row.OpenProjectDuration += fmt.Sprintf("%v\n", humanReadableDuration)
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
