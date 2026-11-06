package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/JankariTech/OpenProjectTmetricIntegration/openproject"
	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
	"github.com/spf13/cobra"
)

var openprojectCmd = &cobra.Command{
	Use:   "openproject",
	Short: "Copy time entries from Tmetric to OpenProject",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		config := config.NewConfig()
		tmetricUser := tmetric.NewUser()

		filteredEntries, err := checkTmetricEntries(
			tmetricUser, true, true, config.OpenProjectTmetricClientIds, config)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, tmetricTimeEntry := range filteredEntries {
			err = transferEntryToOpenProject(tmetricTimeEntry, tmetricUser, config)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	},
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

func init() {
	copyCmd.AddCommand(openprojectCmd)

	firstDayOfMonth := time.Now().Format("2006-01-02")
	firstDayOfMonth = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")

	openprojectCmd.Flags().StringVarP(&startDate, "start", "s", firstDayOfMonth, "start date")
	today := time.Now().Format("2006-01-02")
	openprojectCmd.Flags().StringVarP(&endDate, "end", "e", today, "end date")
}
