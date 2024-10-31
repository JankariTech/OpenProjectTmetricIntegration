package openproject

import (
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/tmetric"
)

type TimeEntry struct {
	Ongoing bool `json:"ongoing"`
	Comment struct {
		Raw string `json:"raw"`
	} `json:"comment"`
	SpentOn string `json:"spentOn"`
	Hours   string `json:"hours"`
	Links   struct {
		WorkPackage struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"workPackage"`
		Activity struct {
			Href string `json:"href"`
		} `json:"activity"`
	} `json:"_links"`
}

func NewFromTmetricTimeEntry(
	tmetricTimeEntry tmetric.TimeEntry, activityId int,
) (TimeEntry, error) {
	opTimeEnty := TimeEntry{
		Ongoing: false,
	}
	opTimeEnty.Comment.Raw = tmetricTimeEntry.Note
	issueId, err := tmetricTimeEntry.GetIssueIdAsInt()
	if err != nil {
		return TimeEntry{}, err
	}
	opTimeEnty.Links.WorkPackage.Href = fmt.Sprintf("/api/v3/work_packages/%d", issueId)
	opTimeEnty.Links.Activity.Href = fmt.Sprintf("/api/v3/time_entries/activities/%d", activityId)
}
