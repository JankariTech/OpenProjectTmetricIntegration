package cmd

type ExternalLink struct {
	Caption string `json:"caption"`
	Link    string `json:"link"`
	IssueId string `json:"issueId"`
}

type Task struct {
	Id           int          `json:"id"`
	Name         string       `json:"name"`
	ExternalLink ExternalLink `json:"externalLink"`
}

type Client struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Client Client `json:"client"`
}

type TimeEntry struct {
	Id        int     `json:"id"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	Task      Task    `json:"task"`
	Project   Project `json:"project"`
	Note      string  `json:"note"`
}
