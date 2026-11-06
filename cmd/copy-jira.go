package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
	"github.com/spf13/cobra"
)

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Copy time entries from Tmetric to Jira",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		config := config.NewConfig()
		tmetricUser := tmetric.NewUser()

		filteredEntries, err := checkTmetricEntries(
			tmetricUser, false, true, config.JiraTmetricClientIds, config)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, tmetricTimeEntry := range filteredEntries {
			err = transferEntryToJira(tmetricTimeEntry, tmetricUser, config)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	},
}

func transferEntryToJira(
	tmetricTimeEntry tmetric.TimeEntry, tmetricUser tmetric.User, config *config.Config,
) error {
	spinner := newSpinner()
	defer spinner.Stop()
	spinner.FinalMSG = "❌\n"
	spinner.Prefix = fmt.Sprintf(
		"Transferring data to Jira. Project: '%v', Note: '%v', Start: '%v' ", tmetricTimeEntry.Project.Name, tmetricTimeEntry.Note, tmetricTimeEntry.StartTime,
	)
	spinner.Restart()

	jiraWorklog, err := tmetricTimeEntry.ConvertToJiraWorklog(tmetricUser.TimeZone.Id)
	if err != nil {
		return fmt.Errorf(
			"could not convert time entry '%v' in project '%v' started at '%v' from tmetric to Jira\n"+
				"Error: %v\n",
			tmetricTimeEntry.Note, tmetricTimeEntry.Project, tmetricTimeEntry.StartTime, err,
		)
	}
	err = jiraWorklog.Save(*config)
	if err != nil {
		return fmt.Errorf(
			"could not work log '%v' for issue '%v' started at '%v' in Jira\n"+
				"Error: %v\n",
			jiraWorklog.Comment,
			jiraWorklog.Key,
			jiraWorklog.Started,
			err,
		)
	}

	tmetricTimeEntry.TagAsTransferredToOpenProject(*config)
	tmetricUser = tmetric.NewUser()
	err = tmetricTimeEntry.Update(*config, tmetricUser)
	if err != nil {
		return fmt.Errorf(
			"could not tag tmetric entry as being transferred to project management\n"+
				"Error: %v\n",
			err,
		)
	}
	spinner.FinalMSG = "✔️\n"
	return nil
}

func init() {
	copyCmd.AddCommand(jiraCmd)

	firstDayOfMonth := time.Now().Format("2006-01-02")
	firstDayOfMonth = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")

	jiraCmd.Flags().StringVarP(&startDate, "start", "s", firstDayOfMonth, "start date")
	today := time.Now().Format("2006-01-02")
	jiraCmd.Flags().StringVarP(&endDate, "end", "e", today, "end date")
}
