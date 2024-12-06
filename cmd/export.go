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
	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var arbitraryString []string
var tmplFile string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export data using a template e.g. to generate an invoice",
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
			"ArbitraryString": func(i int) string {
				return arbitraryString[i]
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
	rootCmd.AddCommand(exportCmd)

	firstDayOfMonth := time.Now().Format("2006-01-02")
	firstDayOfMonth = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")

	exportCmd.Flags().StringVarP(&startDate, "start", "s", firstDayOfMonth, "start date")
	today := time.Now().Format("2006-01-02")
	exportCmd.Flags().StringVarP(&endDate, "end", "e", today, "end date")
	exportCmd.Flags().StringArrayVarP(
		&arbitraryString,
		"arbitraryString",
		"a",
		nil,
		"any string that should be placed on the export, e.g. the invoice number",
	)
	exportCmd.MarkFlagRequired("arbitraryString")
	exportCmd.Flags().StringVarP(&tmplFile, "template", "t", today, "the template file")
	exportCmd.MarkFlagRequired("template")
}
