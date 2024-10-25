package loader

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/builder"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestLoadTestCase(t *testing.T) {
	// test static loading testcase in directory
	err := os.Setenv("TESTSOLAR_TTP_PARSEMODE", "static")
	assert.NoError(t, err)
	absPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	testcases, loadErrors := LoadTestCase(absPath, "demo")
	assert.NoError(t, err)
	assert.NotEqual(t, len(testcases), 0)
	assert.Len(t, loadErrors, 0)
	// test static loading testcase in file
	testcases, loadErrors = LoadTestCase(absPath, "demo/demo_test.go")
	assert.NotEqual(t, len(testcases), 0)
	assert.Len(t, loadErrors, 0)
	// test dynamic loading testcase
	err = os.Setenv("TESTSOLAR_TTP_PARSEMODE", "dynamic")
	assert.NoError(t, err)
	absPath, err = filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	err = builder.Build(absPath)
	assert.NoError(t, err)
	defer os.Remove("../../testdata/demo.test")
	defer os.Remove("../../testdata/demo/book.test")
	defer os.Remove("../../testdata/demo/report.json")
	// test dynamic loading testcase in directory
	testcases, loadErrors = LoadTestCase(absPath, "demo")
	assert.NotEqual(t, len(testcases), 0)
	assert.Len(t, loadErrors, 0)
	// test dynamic loading testcase in file
	testcases, loadErrors = LoadTestCase(absPath, "demo/demo_test.go")
	assert.NotEqual(t, len(testcases), 0)
	assert.Len(t, loadErrors, 0)
	// test dynamic loading v1 testcase in directory
	testcases, loadErrors = LoadTestCase(absPath, "demo/v1")
	assert.NotEqual(t, len(testcases), 0)
	assert.Len(t, loadErrors, 0)
	// test dynamic loading empty v1 testcase in directory
	testcases, loadErrors = LoadTestCase(absPath, "demo/v1_empty")
	assert.Len(t, testcases, 0)
	assert.Len(t, loadErrors, 0)
	// test dynamic loading testcase in directory without test binary
	os.Remove("../../testdata/demo.test")
	os.Remove("../../testdata/demo/book.test")
	testcases, loadErrors = LoadTestCase(absPath, "demo")
	assert.NotEqual(t, len(testcases), 0)
	assert.Len(t, loadErrors, 0)
	os.Remove("../../testdata/demo.test")
	os.Remove("../../testdata/demo/book.test")
	// test dynamic loading testcase in directory with ginkgo tool
	BuildTestPackageMock := gomonkey.ApplyFunc(builder.BuildTestPackage, func(projPath string, packagePath string, compress bool) (string, error) {
		return "", nil
	})
	defer BuildTestPackageMock.Reset()
	testcases, loadErrors = LoadTestCase(absPath, "demo")
	if _, err := exec.LookPath("ginkgo"); err != nil {
		assert.Len(t, testcases, 0)
		assert.NotEqual(t, len(loadErrors), 0)
	} else {
		assert.NotEqual(t, len(testcases), 0)
		assert.Len(t, loadErrors, 0)
	}
	defer os.Remove("../../testdata/demo/book/report.json")
	defer os.Remove("../../testdata/report.json")
}
