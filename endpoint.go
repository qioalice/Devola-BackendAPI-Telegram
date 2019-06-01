// Package tgbotapi has functions and types used for interacting with
// the Telegram Bot API.
package tgbotapi

// Telegram endpoints
const (
	// APIEndpoint is the endpoint for all API methods,
	// with formatting for Sprintf.
	APIEndpoint = "https://api.telegram.org/bot"
	// FileEndpoint is the endpoint for downloading a file from Telegram.
	FileEndpoint = "https://api.telegram.org/file/bot"
)

// makeURL is just like fmt.Sprintf but faster than it.
func makeURL(apibase, token, arg string) string {
	return apibase + token + "/" + arg
}
