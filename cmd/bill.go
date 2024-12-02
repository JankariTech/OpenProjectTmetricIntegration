/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/JankariTech/OpenProjectTmetricIntegration/openproject"
	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var billNumber string
var tmplFile string

// billCmd represents the bill command
var billCmd = &cobra.Command{
	Use:   "bill",
	Short: "A brief description of your command",
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

		funcMap := template.FuncMap{
			"BillNumber": func() string {
				return billNumber
			},
			"DetailedReport": func(clientName string, tagName string, groupName string) tmetric.Report {
				report, _ := tmetric.GetDetailedReport(config, tmetricUser, clientName, tagName, groupName, startDate, endDate)
				return report
			},
			"AllWorkTypes": func() []tmetric.Tag {
				workTypes, _ := tmetric.GetAllWorkTypes(config, tmetricUser)
				return workTypes
			},
			"AllTeams": func() []tmetric.Team {
				teams, _ := tmetric.GetAllTeams(config, tmetricUser)
				return teams
			},
			"formatFloat": func(f float64) string {
				s := fmt.Sprintf("%.2f", f)
				return strings.Replace(s, ".", ",", -1)
			},
			"ServiceDate": func() string {
				startTime, _ := time.Parse("2006-01-02", startDate)
				return startTime.Format("01/2006")
			},
			"AllTimeEntriesFromOpenProject": func(user string) []openproject.TimeEntry {
				var openProjectUser openproject.User
				openProjectUser, err := openproject.FindUserByName(config, user)
				if err != nil {
					_, _ = fmt.Fprint(os.Stderr, err)
					os.Exit(1)
				}
				openProjectTimeEntries, err := openproject.GetAllTimeEntries(config, openProjectUser, startDate, endDate)
				if err != nil {
					_, _ = fmt.Fprint(os.Stderr, err)
					os.Exit(1)
				}
				return openProjectTimeEntries
			},
		}
		// add all the functions from sprig
		for i, f := range sprig.FuncMap() {
			funcMap[i] = f
		}

		tmpl, err := template.New(filepath.Base(tmplFile)).Funcs(funcMap).ParseFiles(tmplFile)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, fmt.Errorf("could not parse template file '%v': %v", tmplFile, err))
			os.Exit(1)
		}
		err = tmpl.Execute(os.Stdout, nil)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, fmt.Errorf("could not execute template: %v", err))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(billCmd)

	firstDayOfMonth := time.Now().Format("2006-01-02")
	firstDayOfMonth = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")

	billCmd.Flags().StringVarP(&startDate, "start", "s", firstDayOfMonth, "start date")
	today := time.Now().Format("2006-01-02")
	billCmd.Flags().StringVarP(&endDate, "end", "e", today, "end date")
	billCmd.Flags().StringVarP(&billNumber, "billNumber", "b", today, "the bill number")
	billCmd.MarkFlagRequired("billNumber")
	billCmd.Flags().StringVarP(&tmplFile, "template", "t", today, "the template file")
	billCmd.MarkFlagRequired("template")
}
