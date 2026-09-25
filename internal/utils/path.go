package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Extract the base filename from the path and remove all extension suffixes. Function will return an error
// if the resulting path is empty or the input path has a lot (more than 16) suffixes.
func GetFilenameWithoutExtensions(path string) (string, error) {
	var (
		p     string
		limit int = 16
	)

	for {
		if len(path) == 0 || limit <= 0 {
			return "", fmt.Errorf("utils: can not extract the filename without extensions")
		}

		if p = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)); p == path {
			return p, nil
		} else {
			path = p
			limit -= 1
		}
	}
}
