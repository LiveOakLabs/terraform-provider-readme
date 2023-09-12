package testdata

import "encoding/json"

const (
	// testURL is a dummy URL the provider is configured with and the mock HTTP
	// service responds to.
	testURL = "http://testing/api/v1"
	// testToken is a dummy token the provider is configured with and used
	// throughout tests.
	testToken = "hunter2"
	// providerConfig is a shared configuration that sets a mock url and token.
	// The URL points to our gock mock server.
	providerConfig = (`
		provider "readme" {
			api_token = "` + testToken + `"
			api_url   = "` + testURL + `"
		}
	`)
)

func ToJSON(data interface{}) string {
	bytes, _ := json.Marshal(data)

	return string(bytes)
}
