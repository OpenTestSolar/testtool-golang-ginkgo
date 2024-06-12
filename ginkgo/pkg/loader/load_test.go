package loader

import (
	"ginkgo/pkg/builder"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTestCase(t *testing.T) {
	// test static loading testcase in directory
	absPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	testcases, err := LoadTestCase(absPath, "demo")
	assert.NoError(t, err)
	assert.NotEqual(t, len(testcases), 0)
	// test static loading testcase in file
	testcases, err = LoadTestCase(absPath, "demo/demo_test.go")
	assert.NoError(t, err)
	assert.NotEqual(t, len(testcases), 0)
	// test dynamic loading testcase in directory
	err = os.Setenv("TESTSOLAR_TTP_PARSEMODE", "dynamic")
	assert.NoError(t, err)
	absPath, err = filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	err = builder.Build(absPath)
	assert.NoError(t, err)
	defer os.Remove("../../testdata/demo.test")
	defer os.Remove("../../testdata/demo/report.json")
	testcases, err = LoadTestCase(absPath, "demo")
	assert.NoError(t, err)
	assert.NotEqual(t, len(testcases), 0)
	// test dynamic loading testcase in file
	testcases, err = LoadTestCase(absPath, "demo/demo_test.go")
	assert.NoError(t, err)
	assert.NotEqual(t, len(testcases), 0)
}
