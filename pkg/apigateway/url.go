package apigw

import "strings"

func SanitizeURL(url string) string {
	return strings.Replace(url, "wss://", "https://", -1)
}
