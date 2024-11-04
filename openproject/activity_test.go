package openproject

import (
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewActivityFromWorkType(t *testing.T) {
	type args struct {
		issueId  int
		workType string
	}

	validResponseAllowedActivities := `{
									  "_type": "Form",
									  "_embedded": {
										"schema": {
										  "activity": {
											"type": "TimeEntriesActivity",
											"name": "Activity",
											"_embedded": {
											  "allowedValues": [
												{
												  "id": 1,
												  "name": "Management"
												},
												{
												  "id": 2,
												  "name": "Specification"
												},
												{
												  "id": 3,
												  "name": "Development"
												}
											  ]
											}
										  }
										},
										"validationErrors": {
										}
									  }
									}`
	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		args           args
		want           Activity
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:           "Valid work type",
			mockResponse:   validResponseAllowedActivities,
			mockStatusCode: http.StatusOK,
			args: args{
				issueId:  1234,
				workType: "Development",
			},
			want: Activity{
				Id:   3,
				Name: "Development",
			},
		},
		{
			name:           "work type has no equivalent activity",
			mockResponse:   validResponseAllowedActivities,
			mockStatusCode: http.StatusOK,
			args: args{
				issueId:  1234,
				workType: "My WorkType",
			},
			want:           Activity{},
			wantErr:        true,
			wantErrMessage: "Work Type 'My WorkType' is not a valid activity in OpenProject\n",
		},
		{
			name: "Forbidden to access WorkPackage",
			mockResponse: `{
				"_type" : "Error",
				"errorIdentifier" : "urn:openproject-org:api:v3:errors:MissingPermission",
				"message" : "You are not authorized to access this resource."
			}`,
			mockStatusCode: http.StatusForbidden,
			args: args{
				issueId:  1234,
				workType: "My WorkType",
			},
			want:           Activity{},
			wantErr:        true,
			wantErrMessage: "could not fetch allowed activities for work package '1234' from",
		},
		{
			name: "not existing work package",
			mockResponse: `{
				"_embedded": {
					"validationErrors": {
						"workPackage": {
							"message": "Work package is invalid."
						}
					}
				}
			}`,
			mockStatusCode: http.StatusOK,
			args: args{
				issueId:  123567,
				workType: "My WorkType",
			},
			want:           Activity{},
			wantErr:        true,
			wantErrMessage: "work package '123567' not found. Error: Work package is invalid.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody string
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bodyBytes, _ := io.ReadAll(r.Body)
				requestBody = string(bodyBytes)
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer mockServer.Close()
			tmetricConfig := config.Config{
				OpenProjectToken: "dummyToken",
				OpenProjectUrl:   mockServer.URL + "/",
			}

			got, err := NewActivityFromWorkType(tmetricConfig, tt.args.issueId, tt.args.workType)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.ErrorContains(t, err, tt.wantErrMessage, "errorMessage = %v, wantErrMessage %v", err, tt.wantErrMessage)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
			wantRequestBody := fmt.Sprintf(`{"_links":{"workPackage":{"href":"/api/v3/work_packages/%d"}}}`, tt.args.issueId)
			if requestBody != wantRequestBody {
				t.Errorf("Request body = %v, want %v", requestBody, wantRequestBody)
			}
		})
	}
}
