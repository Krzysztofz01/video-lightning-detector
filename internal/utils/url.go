package utils

import "net/url"

func IsValidUrl(s string) bool {
	u, err := url.Parse(s)

	return err == nil && u.Scheme != "" && u.Host != ""
}
