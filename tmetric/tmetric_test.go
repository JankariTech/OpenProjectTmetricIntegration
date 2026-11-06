package tmetric

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/stretchr/testify/assert"
)

func Test_GetEntriesWithoutLinkToProjectManagement(t *testing.T) {
	tests := []struct {
		name string
		task Task
	}{
		{
			name: "empty task",
			task: Task{},
		},
		{
			name: "no external link",
			task: Task{
				Id:   123,
				Name: "some WP",
			},
		},
		{
			name: "link is not to openproject nor jira",
			task: Task{
				Id:   345,
				Name: "some WP",
				ExternalLink: ExternalLink{
					Link:    "https://some_host/work_packages/123",
					IssueId: "#123",
				},
			},
		},
		{
			name: "IssueId is empty",
			task: Task{
				Id:   345,
				Name: "some WP",
				ExternalLink: ExternalLink{
					Link:    "https://community.openproject.org/work_packages/123",
					IssueId: "",
				},
			},
		},
		{
			name: "IssueId has wrong format",
			task: Task{
				Id:   345,
				Name: "some WP",
				ExternalLink: ExternalLink{
					Link:    "https://community.openproject.org/work_packages/123",
					IssueId: "!123",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// adds a correct openproject entry at the beginning and a jira entry at the end
			// and makes sure the invalid entry is still in the result
			config := &config.Config{
				OpenProjectTmetricClientIds: []int{123},
				JiraTmetricClientIds:        []int{789},
				JiraUrl:                     "https://jankaritech.atlassian.net",
				TmetricExternalTaskLink:     "https://community.openproject.org/",
			}
			validOpenProjectTimeEntry := TimeEntry{
				Project: Project{
					Client: Client{Id: 123},
					Name:   "Project1",
				},
				Note: "correct entry",
				Task: Task{
					Id:   345,
					Name: "some WP",
					ExternalLink: ExternalLink{
						Link:    "https://community.openproject.org/work_packages/123",
						IssueId: "#123",
					},
				},
			}
			validJiraTimeEntry := TimeEntry{
				Project: Project{
					Client: Client{Id: 789},
					Name:   "Project1",
				},
				Note: "correct entry",
				Task: Task{
					Id:   345,
					Name: "some WP",
					ExternalLink: ExternalLink{
						Link:    "https://jankaritech.atlassian.net/browse/JT-123",
						IssueId: "JT-123",
					},
				},
			}
			invalidTimeEntry := TimeEntry{
				Project: Project{
					Client: Client{Id: 123},
					Name:   "Project1",
				},
				Note: "invalid entry",
				Task: tt.task,
			}
			timeEntriesToCheck := append([]TimeEntry{invalidTimeEntry}, validOpenProjectTimeEntry)
			timeEntriesToCheck = append([]TimeEntry{validJiraTimeEntry}, timeEntriesToCheck...)
			result := GetEntriesWithoutLinkToProjectManagement(config, timeEntriesToCheck, nil)
			if !reflect.DeepEqual(result, []TimeEntry{invalidTimeEntry}) {
				t.Errorf("got %v, want %v", result, []TimeEntry{invalidTimeEntry})
			}
		})
	}
}

func Test_GetEntriesWithoutLinkToProjectManagementLimitToClient(t *testing.T) {
	tests := []struct {
		name          string
		ids           []int
		expectedNotes []string
	}{
		{
			name: "all",
			ids:  nil,
			expectedNotes: []string{
				"first entry first client",
				"second entry first client",
				"first entry second client",
				"second entry second client",
				"first entry third client",
			},
		},
		{
			name:          "one client",
			ids:           []int{1},
			expectedNotes: []string{"first entry first client", "second entry first client"},
		},
		{
			name:          "multiple client",
			ids:           []int{1, 3},
			expectedNotes: []string{"first entry first client", "second entry first client", "first entry third client"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// adds a correct openproject entry at the beginning and a jira entry at the end
			// and makes sure the invalid entry is still in the result
			config := &config.Config{
				OpenProjectTmetricClientIds: []int{1, 3},
				JiraTmetricClientIds:        []int{2},
				JiraUrl:                     "https://jankaritech.atlassian.net",
				TmetricExternalTaskLink:     "https://community.openproject.org/",
			}
			timeEntry1Client1 := TimeEntry{
				Project: Project{
					Client: Client{Id: 1},
					Name:   "Project1",
				},
				Note: "first entry first client",
			}
			timeEntry2Client1 := TimeEntry{
				Project: Project{
					Client: Client{Id: 1},
					Name:   "Project2",
				},
				Note: "second entry first client",
			}
			timeEntry1Client2 := TimeEntry{
				Project: Project{
					Client: Client{Id: 2},
					Name:   "Project1",
				},
				Note: "first entry second client",
			}
			timeEntry2Client2 := TimeEntry{
				Project: Project{
					Client: Client{Id: 2},
					Name:   "Project1",
				},
				Note: "second entry second client",
			}
			timeEntry1Client3 := TimeEntry{
				Project: Project{
					Client: Client{Id: 3},
					Name:   "Project1",
				},
				Note: "first entry first client",
			}

			timeEntriesToCheck := []TimeEntry{timeEntry1Client1, timeEntry2Client1, timeEntry1Client2, timeEntry2Client2, timeEntry1Client3}
			results := GetEntriesWithoutLinkToProjectManagement(config, timeEntriesToCheck, tt.ids)
			assert.Equal(t, len(tt.expectedNotes), len(results))
			for _, result := range results {
				assert.Contains(t, tt.expectedNotes, result.Note)
			}

		})
	}
}

func Test_GetEntriesWithoutWorkType(t *testing.T) {
	tests := []struct {
		name             string
		timeEntries      []TimeEntry
		clientIdsToCheck []int
		expectedResult   []TimeEntry
	}{
		{
			name: "No work type assigned",
			timeEntries: []TimeEntry{
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry1",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags:      []Tag{},
				},
			},
			clientIdsToCheck: []int{123},
			expectedResult: []TimeEntry{
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry1",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags:      []Tag{},
				},
			},
		},
		{
			name: "No work type assigned, but check other client",
			timeEntries: []TimeEntry{
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry1",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags:      []Tag{},
				},
			},
			clientIdsToCheck: []int{456},
			expectedResult:   nil,
		},
		{
			name: "all work types already assigned",
			timeEntries: []TimeEntry{
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry1",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags: []Tag{
						{
							Name:       "something",
							IsWorkType: true,
						},
					},
				},
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry2",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags: []Tag{
						{
							Name:       "something",
							IsWorkType: true,
						},
					},
				},
			},
			clientIdsToCheck: []int{123},
			expectedResult:   nil,
		},
		{
			name: "tags exist but not work type",
			timeEntries: []TimeEntry{
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry1",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags: []Tag{
						{
							Name:       "some tag",
							IsWorkType: false,
						},
						{
							Name:       "something else",
							IsWorkType: false,
						},
					},
				},
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry2",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags: []Tag{
						{
							Name:       "some tag",
							IsWorkType: false,
						},
						{
							Name:       "something else",
							IsWorkType: false,
						},
					},
				},
			},
			clientIdsToCheck: []int{123},
			expectedResult: []TimeEntry{
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry1",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags: []Tag{
						{
							Name:       "some tag",
							IsWorkType: false,
						},
						{
							Name:       "something else",
							IsWorkType: false,
						},
					},
				},
				{
					Project: Project{
						Client: Client{Id: 123},
						Name:   "Project1",
					},
					Note:      "Entry2",
					StartTime: "2023-10-01T09:00:00Z",
					EndTime:   "2023-10-01T10:00:00Z",
					Tags: []Tag{
						{
							Name:       "some tag",
							IsWorkType: false,
						},
						{
							Name:       "something else",
							IsWorkType: false,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEntriesWithoutWorkType(tt.timeEntries, tt.clientIdsToCheck)
			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("got %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func Test_getPossibleWorkTypes(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		expectedResult []Tag
		expectError    bool
	}{
		{
			name:           "successful response with work types",
			mockResponse:   `[{"id":1,"name":"WorkType1","isWorkType":true},{"id":2,"name":"WorkType2","isWorkType":true}]`,
			mockStatusCode: http.StatusOK,
			expectedResult: []Tag{
				{Id: 1, Name: "WorkType1", IsWorkType: true},
				{Id: 2, Name: "WorkType2", IsWorkType: true},
			},
			expectError: false,
		},
		{
			name:           "successful response without work types",
			mockResponse:   `[{"id":1,"name":"Tag1","isWorkType":false}]`,
			mockStatusCode: http.StatusOK,
			expectedResult: nil,
			expectError:    false,
		},
		{
			name:           "successful response without mixed tags",
			mockResponse:   `[{"id":1,"name":"Tag1","isWorkType":false},{"id":2,"name":"Tag2","isWorkType":true}]`,
			mockStatusCode: http.StatusOK,
			expectedResult: []Tag{
				{Id: 2, Name: "Tag2", IsWorkType: true},
			},
			expectError: false,
		},
		{
			name: "error response",

			mockResponse:   `Internal Server Error`,
			mockStatusCode: http.StatusInternalServerError,
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer mockServer.Close()
			config := config.Config{
				TmetricToken:        "dummyToken",
				TmetricAPIV3BaseUrl: mockServer.URL + "/",
			}
			user := User{
				ActiveAccountId: 1,
			}
			timeEntry := TimeEntry{
				Project: Project{Id: 1},
			}
			result, err := timeEntry.GetPossibleWorkTypes(config, user)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("expected result: %v, got: %v", tt.expectedResult, result)
			}
		})
	}
}
