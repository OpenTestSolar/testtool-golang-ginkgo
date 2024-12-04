package cmdline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenTestCaseFocusName(t *testing.T) {
	focusName := GenTestCaseFocusName([]string{"[(case01)]", "case02"})
	assert.Equal(t, focusName, "\\[\\(case01\\)\\]|case02")
}

func TestExtractPackPathFromBinFile(t *testing.T) {
	testCases := []struct {
		pkgBin   string
		projPath string
		expected string
	}{
		{"/path/to/project/pkg/name.test", "/path/to/project", "pkg/name"},
		{"/path/to/project/pkg/name", "/path/to/project", "pkg/name"},
		{"/path/to/project/pkg/name.test", "/path/to/project/", "pkg/name"},
		{"/path/to/project/pkg/name", "/path/to/project/", "pkg/name"},
	}

	for _, tc := range testCases {
		t.Run(tc.pkgBin, func(t *testing.T) {
			result := ExtractPackPathFromBinFile(tc.pkgBin, tc.projPath)
			assert.Equal(t, tc.expected, result)
		})
	}
}
