package utils

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGetFilenameWithoutExtensionShouldCorrectlyExtractFilename(t *testing.T) {
	cases := map[string]string{
		"foo":               "foo",
		"foo.bar":           "foo",
		"foo.bar.xyz":       "foo",
		"foo/foo.bar":       "foo",
		"foo/foo":           "foo",
		"/foo/foo.bar":      "foo",
		"/foo/foo.bar.xyz":  "foo",
		"./foo/foo.bar.xyz": "foo",
	}

	for c, expected := range cases {
		actual, err := GetFilenameWithoutExtensions(c)

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	}
}

func TestGetFilenameWithoutExtensionShouldFailOnExtensionOnlyFilenames(t *testing.T) {
	cases := []string{
		"",
		".foo",
		".foo.bar",
		".foo.bar.xyz",
		"..foo",
		"..foo.bar",
	}

	for _, c := range cases {
		_, err := GetFilenameWithoutExtensions(c)

		assert.NotNil(t, err)
	}

}

func TestGetFilenameWithoutExtensionShouldFailOnLongExtensionChains(t *testing.T) {
	const (
		filename       = "foo"
		extension      = ".bar"
		extensionCount = 32
	)

	fullFilenameBuilder := new(strings.Builder)
	fullFilenameBuilder.WriteString(filename)

	for i := 0; i < extensionCount; i += 1 {
		fullFilenameBuilder.WriteString(extension)
	}

	_, err := GetFilenameWithoutExtensions(fullFilenameBuilder.String())
	assert.NotNil(t, err)
}
