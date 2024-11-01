package openproject

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
