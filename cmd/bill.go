/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
	"github.com/spf13/cobra"
	"math"
	"os"
	"strings"
	"text/template"
	"time"
)

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
		//
		tmetricUser := tmetric.NewUser()
		tmetricTimeEntries, err := tmetric.GetAllTimeEntries(config, tmetricUser, startDate, endDate)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		funcMap := template.FuncMap{
			"DetailedReport": func(clientName string, tagName string, groupName string) tmetric.Report {
				report, _ := tmetric.GetDetailedReport(clientName, tagName, groupName, startDate, endDate)
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
			"Increment": func(i int) int {
				return i + 1
			},
			"Round": func(f float64) float64 {
				return math.Round(f*100) / 100
			},
			"FormatFloat": func(f float64) string {
				s := fmt.Sprintf("%.2f", f)
				return strings.Replace(s, ".", ",", -1)
			},
			"Add": func(a float64, b float64) float64 {
				return a + b
			},
			"Multiply": func(a float64, b float64) float64 {
				return a * b
			},
			"ServiceDate": func() string {
				startTime, _ := time.Parse("2006-01-02", startDate)
				return startTime.Format("01/2006")
			},
		}
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		var tmplFile = "openproject.tmpl"

		tmpl, err := template.New(tmplFile).Funcs(funcMap).ParseFiles(tmplFile)
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(os.Stdout, tmetricTimeEntries)
		if err != nil {
			panic(err)
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
}
