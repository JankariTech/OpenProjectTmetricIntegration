package tmetric

import (
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_GetEntriesWithoutLinkToOpenProject(t *testing.T) {
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
			name: "link is not to openproject",
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
			// adds a correct entry at the beginning and the end and makes sure the invalid
			// entry is still in the result
			config := &config.Config{
				ClientIdInTmetric:       123,
				TmetricExternalTaskLink: "https://community.openproject.org/",
			}
			validTimeEntry := TimeEntry{
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
			invalidTimeEntry := TimeEntry{
				Project: Project{
					Client: Client{Id: 123},
					Name:   "Project1",
				},
				Note: "invalid entry",
				Task: tt.task,
			}
			timeEntriesToCheck := append([]TimeEntry{invalidTimeEntry}, validTimeEntry)
			timeEntriesToCheck = append([]TimeEntry{validTimeEntry}, timeEntriesToCheck...)
			result := GetEntriesWithoutLinkToOpenProject(config, timeEntriesToCheck)
			if !reflect.DeepEqual(result, []TimeEntry{invalidTimeEntry}) {
				t.Errorf("got %v, want %v", result, []TimeEntry{invalidTimeEntry})
			}
		})
	}
}

func Test_GetEntriesWithoutWorkType(t *testing.T) {
	type args struct {
		timeEntries []TimeEntry
		tmetricUser *User
		config      *config.Config
	}
	tests := []struct {
		name           string
		args           args
		expectedResult []TimeEntry
	}{
		{
			name: "No work type assigned",
			args: args{
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
				config: &config.Config{
					ClientIdInTmetric: 123,
				},
			},
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
			name: "all work types already assigned",
			args: args{
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
				config: &config.Config{
					ClientIdInTmetric: 123,
				},
			},
			expectedResult: nil,
		},
		{
			name: "tags exist but not work type",
			args: args{
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
				config: &config.Config{
					ClientIdInTmetric: 123,
				},
			},
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
			result := GetEntriesWithoutWorkType(tt.args.timeEntries)
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
