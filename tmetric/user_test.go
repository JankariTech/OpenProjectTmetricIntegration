package tmetric

import (
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFindUserByName(t *testing.T) {

	validResponseMultipleUsers := `[
  {
    "accountMemberId": 777,
    "userProfile": {
      "userProfileId": 1111,
      "activeAccountId": 8888,
      "userName": "Peter Pan"
    },
    "accountMemberScope": {
      "groupMembership": [
        {
          "name": "my group",
          "id": 5652
        }
      ]
    }
  },
  {
    "accountMemberId": 222,
    "userProfile": {
      "userProfileId": 456,
      "activeAccountId": 8888,
      "userName": "Peter Fan"
    },
    "accountMemberScope": {
      "groupMembership": [
        {
          "name": "my group",
          "id": 5652
        }
      ]
    }
  }
]`
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
				Id:              1111,
				Name:            "Peter Pan",
				ActiveAccountId: 8888,
			},
		},
		{
			name:           "no user found",
			mockResponse:   validResponseMultipleUsers,
			mockStatusCode: http.StatusOK,
			search:         "does not exist",
			want:           User{},
			wantErr:        true,
			wantErrMessage: "cannot find a user in tmetric with a name matching 'does not exist'",
		},
		{
			name:           "wrongly formatted response",
			mockResponse:   `{[],]}`,
			mockStatusCode: http.StatusOK,
			search:         "user",
			want:           User{},
			wantErr:        true,
			wantErrMessage: "error parsing users response",
		},
		{
			name:           "no permission",
			mockResponse:   `{"statusCode": 403,"message": "Restricted workspace access."}`,
			mockStatusCode: http.StatusForbidden,
			search:         "user",
			want:           User{},
			wantErr:        true,
			wantErrMessage: "cannot get members for tmetric account '4567'. Error: '<nil>'. HTTP status code: 403",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestPath string
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestPath = r.URL.Path
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer mockServer.Close()
			config := config.Config{
				TmetricToken:      "dummyToken",
				TmetricAPIBaseUrl: mockServer.URL + "/",
			}

			tmetricUserMe := User{
				Id:              1234,
				Name:            "admin user",
				ActiveAccountId: 4567,
			}
			got, err := FindUserByName(&config, tmetricUserMe, tt.search)
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
			assert.Equal(t, "/accounts/4567/members", requestPath)

		})
	}
}
