package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
	"github.com/spf13/cobra"
)

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Copy time entries from Tmetric to Jira",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		filteredEntries, err := checkTmetricEntries(
			false, true, cfg.JiraTmetricClientIds)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, tmetricTimeEntry := range filteredEntries {
			err = transferEntryToJira(tmetricTimeEntry)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	},
}

func transferEntryToJira(tmetricTimeEntry tmetric.TimeEntry) error {
	spinner := newSpinner()
	defer spinner.Stop()
	spinner.FinalMSG = "❌\n"
	spinner.Prefix = fmt.Sprintf(
		"Transferring data to Jira. Project: '%v', Note: '%v', Start: '%v' ", tmetricTimeEntry.Project.Name, tmetricTimeEntry.Note, tmetricTimeEntry.StartTime,
	)
	spinner.Restart()

	jiraWorklog, err := tmetricTimeEntry.ConvertToJiraWorklog(tmetricUser)
	if err != nil {
		return fmt.Errorf(
			"could not convert time entry '%v' in project '%v' started at '%v' from tmetric to Jira\n"+
				"Error: %v\n",
			tmetricTimeEntry.Note, tmetricTimeEntry.Project, tmetricTimeEntry.StartTime, err,
		)
	}
	err = jiraWorklog.Save(*cfg)
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

	tmetricTimeEntry.TagAsTransferredToOpenProject(*cfg)
	err = tmetricTimeEntry.Update(*cfg, tmetricUser)
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
	jiraCmd.Flags().StringVarP(
		&userNameFromCmd, "user", "u", "", "name of the user whose entries should be transferred",
	)
}
