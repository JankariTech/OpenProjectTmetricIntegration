package cmd

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_getEntriesWithoutWorkType(t *testing.T) {
	type args struct {
		timeEntries []TimeEntry
		tmetricUser *TmetricUser
		config      *Config
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
				config: &Config{
					clientIdInTmetric: 123,
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
				config: &Config{
					clientIdInTmetric: 123,
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
				config: &Config{
					clientIdInTmetric: 123,
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
			result := getEntriesWithoutWorkType(tt.args.timeEntries, tt.args.config)
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
			config := Config{
				tmetricToken:        "dummyToken",
				tmetricAPIV3BaseUrl: mockServer.URL + "/",
			}
			user := TmetricUser{
				ActiveAccountId: 1,
			}
			timeEntry := TimeEntry{
				Project: Project{Id: 1},
			}
			result, err := timeEntry.getPossibleWorkTypes(config, user)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("expected result: %v, got: %v", tt.expectedResult, result)
			}
		})
	}
}
