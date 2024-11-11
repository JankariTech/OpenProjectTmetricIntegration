package openproject

import (
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestFindUserByName(t *testing.T) {

	validResponseMultipleUsers := `{
  "_embedded": {
    "elements": [
      {
        "id": 1234,
        "name": "Peter Pan",
        "_type": "User"
      },
      {
        "id": 567,
        "name": "Peter Fan",
        "_type": "User"
      }
    ]
  }
}`
	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		search         string
		want           User
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:           "Valid multiple users are returned",
			mockResponse:   validResponseMultipleUsers,
			mockStatusCode: http.StatusOK,
			search:         "Peter",
			want: User{
				Id:   1234,
				Name: "Peter Pan",
			},
		},
		{
			name:           "no user found",
			mockResponse:   `{"_embedded": {"elements": []}}`,
			mockStatusCode: http.StatusOK,
			search:         "does not exist",
			want:           User{},
			wantErr:        true,
			wantErrMessage: "cannot find a user in OpenProject with a name matching 'does not exist'",
		},
		{
			name:           "wrongly formatted response",
			mockResponse:   `{"_embedded": {"some-entry-but-no-elements": []}}`,
			mockStatusCode: http.StatusOK,
			search:         "user",
			want:           User{},
			wantErr:        true,
			wantErrMessage: "error parsing user search response from OpenProject: unexpected end of JSON input",
		},
		{
			name:           "no permission",
			mockResponse:   `{"_type":"Error","errorIdentifier":"urn:openproject-org:api:v3:errors:Unauthenticated","message":"You did not provide the correct credentials."}`,
			mockStatusCode: http.StatusUnauthorized,
			search:         "user",
			want:           User{},
			wantErr:        true,
			wantErrMessage: "cannot lookup users in OpenProject. Error: '<nil>'. HTTP status code: 401",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestQuery url.Values
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestQuery = r.URL.Query()
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer mockServer.Close()
			config := config.Config{
				OpenProjectToken: "dummyToken",
				OpenProjectUrl:   mockServer.URL + "/",
			}

			got, err := FindUserByName(&config, tt.search)
			if (err != nil) != tt.wantErr {
				t.Errorf("error: '%v', wantErr: '%v'", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.ErrorContains(
					t,
					err,
					tt.wantErrMessage,
					"errorMessage: '%v', wantErrMessage: '%v'",
					err,
					tt.wantErrMessage,
				)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
			wantQuery := url.Values{
				"filters": {fmt.Sprintf(`[{"status":{"operator":"!","values":["3"]}},{"type":{"operator":"=","values":["User"]}},{"name":{"operator":"~","values":["%v"]}}]`, tt.search)},
			}
			if !reflect.DeepEqual(requestQuery.Get("filters"), wantQuery.Get("filters")) {
				t.Errorf("Request query:\n\t %v\n want:\n\t %v", requestQuery, wantQuery)
			}
		})
	}
}
