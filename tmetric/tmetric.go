package tmetric

import (
	"encoding/json"
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"

	"github.com/go-resty/resty/v2"
)

func GetAllTimeEntries(config *config.Config, tmetricUser *User, startDate string, endDate string) ([]TimeEntry, error) {
	httpClient := resty.New()
	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Get(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries?userId=%v&startDate=%v&endDate=%v`,
				config.TmetricAPIV3BaseUrl,
				tmetricUser.ActiveAccountId,
				tmetricUser.Id,
				startDate,
				endDate,
			),
		)
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"cannot read timeentries. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}

	var timeEntries []TimeEntry
	err = json.Unmarshal(resp.Body(), &timeEntries)
	if err != nil {
		return nil, fmt.Errorf("error parsing time entries response: %v\n", err)
	}

	return timeEntries, nil
}

func GetEntriesWithoutWorkType(timeEntries []TimeEntry, config *config.Config) []TimeEntry {
	// get all entries that belong to the client and do not have any work-type set
	var entriesWithoutWorkType []TimeEntry
	for _, entry := range timeEntries {
		if entry.Project.Client.Id == config.ClientIdInTmetric {
			workTypeFound := false
			for _, tag := range entry.Tags {
				if tag.IsWorkType {
					workTypeFound = true
					break
				}
			}
			if !workTypeFound {
				entriesWithoutWorkType = append(entriesWithoutWorkType, entry)
			}
		}
	}
	return entriesWithoutWorkType
}
