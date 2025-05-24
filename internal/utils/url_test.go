package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUrlValidShouldCorrectlyValidate(t *testing.T) {
	cases := map[string]bool{
		"": false,
		"://github.com/Krzysztofz01/video-lightning-detector":      false,
		"https:///Krzysztofz01/video-lightning-detector":           false,
		"https://github.com/Krzysztofz01/video-lightning-detector": true,
	}

	for c, expected := range cases {
		actual := IsValidUrl(c)
		assert.Equal(t, expected, actual)
	}
}
