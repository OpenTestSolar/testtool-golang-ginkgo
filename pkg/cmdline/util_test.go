package cmdline

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveTestCaseLabels(t *testing.T) {
	os.Setenv("TESTSOLAR_TTP_LABELMARK", "label")
	result := removeTestCaseLabels([]string{"case01 [label01]", "case02 [label02]"})
	assert.Equal(t, result, []string{"case01", "case02"})
	os.Unsetenv("TESTSOLAR_TTP_LABELMARK")
	result = removeTestCaseLabels([]string{"case01 [label01]", "case02 [label02]"})
	assert.Equal(t, result, []string{"case01 [label01]", "case02"})
}

func TestGenTestCaseFocusName(t *testing.T) {
	focusName := GenTestCaseFocusName([]string{"[(case01)]", "case02"})
	assert.Equal(t, focusName, "\\[\\(case01\\)\\]|case02")
	err := os.Setenv("TESTSOLAR_TTP_LABELMARK", "a")
	assert.Nil(t, err)
	focusName = GenTestCaseFocusName([]string{"[xxx] [a, b, c]"})
	assert.Equal(t, focusName, "\\[xxx\\]")
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
