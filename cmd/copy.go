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
	"time"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
	"github.com/spf13/cobra"
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy time entries from Tmetric to OpenProject or Jira",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
}

func checkTmetricEntries(tmetricUser tmetric.User, checkWorktype bool, checkLink bool, clientIdsToCheck []int, config *config.Config) ([]tmetric.TimeEntry, error) {
	spinner := newSpinner()
	defer spinner.Stop()
	spinner.Prefix = "Checking time entries in Tmetric... "
	spinner.FinalMSG = "❌\n"
	timeEntries, err := tmetric.GetAllTimeEntries(config, tmetricUser, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if checkWorktype && len(tmetric.GetEntriesWithoutWorkType(timeEntries, clientIdsToCheck)) > 0 {
		return nil, fmt.Errorf(
			"some time-entries do not have any work-type assigned, run the 'check tmetric' command to fix it",
		)
	}

	if checkLink && len(tmetric.GetEntriesWithoutLinkToProjectManagement(config, timeEntries, clientIdsToCheck)) > 0 {
		return nil, fmt.Errorf(
			"some time-entries are not linked to a work-package or issue, run the 'check tmetric' command to fix it",
		)
	}

	filteredEntries := tmetric.GetEntriesNotTransferredToProjectManagement(
		timeEntries, config.TmetricTagTransferredToOpenProject, clientIdsToCheck,
	)

	spinner.FinalMSG = "✔️\n"
	return filteredEntries, err
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
