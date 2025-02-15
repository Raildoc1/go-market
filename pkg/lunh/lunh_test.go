package lunh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "right 1",
			number:   "236478236",
			expected: true,
		},
		{
			name:     "right 2",
			number:   "7875643830",
			expected: true,
		},
		{
			name:     "right 3",
			number:   "8898209379842213671099999",
			expected: true,
		},
		{
			name:     "wrong 1",
			number:   "263748",
			expected: false,
		},
		{
			name:     "wrong 2",
			number:   "87239784293",
			expected: false,
		},
		{
			name:     "wrong 3",
			number:   "091280387128758123",
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Validate(test.number))
		})
	}
}
