package result

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJsonToObj(t *testing.T) {
	parser, err := NewResultParser("./testdata/report.json", "/data/workspace", "suits/demo", true)
	assert.NoError(t, err)
	results, err := parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	for _, result := range results {
		if !strings.HasPrefix(result.Test.Name, "suits/demo") {
			t.Errorf("incorrect case name: %s", result.Test.Name)
		}
	}

	parser, err = NewResultParser("./testdata/report_with_setup.json", "/data/workspace", "suites/demo", true)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	parser, err = NewResultParser("./testdata/report_with_setup.json", "/data/workspace", "suites/demo", false)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	parser, err = NewResultParser("./testdata/report_with_labels.json", "/data/workspace", "suites/demo", true)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	if results[0].Test.Name != "suites/demo/demo_suite_test.go?HierarchyText01 HierarchyText02/Text [label01, label02, label11, node-label01]" {
		t.Errorf("incorrect case name: %s", results[0].Test.Name)
	}

	parser, err = NewResultParser("./testdata/report_with_failed_setup.json", "/data/workspace", "suites/demo", true)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	parser, err = NewResultParser("./testdata/report_with_panic.json", "/data/workspace", "suites/demo", true)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	panicSuite, err := parser.GetPanicSuite()
	assert.NoError(t, err)
	assert.NotNil(t, panicSuite)
}

func Test_parseCaseByReg(t *testing.T) {
	byteValue, err := os.ReadFile("./testdata/dry_run_output.txt")
	assert.NoError(t, err)
	cases, err := ParseCaseByReg("/data/workspace", string(byteValue), 2, "")
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
}
