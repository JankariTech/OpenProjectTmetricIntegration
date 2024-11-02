package openproject

import (
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
)

type Activity struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func NewFromWorkType(
	config config.Config, issueId int, workType string,
) (Activity, error) {
	workPackage := WorkPackage{
		Id: issueId,
	}
	activities, err := workPackage.GetAllowedActivities(config)
	if err != nil {
		return Activity{}, err
	}
	workTypeValid := false
	var selectedActivity Activity
	for _, activity := range activities {
		if workType == activity.Name {
			workTypeValid = true
			break
		}
	}
	if !workTypeValid {
		return Activity{}, fmt.Errorf(
			"Work Type '%v' is not a valid activity in OpenProject\n",
			workType,
		)
	}
	return selectedActivity, nil
}
