package runner

import (
	"os"
	"path/filepath"
	"testing"

	builder "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/builder"

	"github.com/stretchr/testify/assert"
)

func TestRunGinkgoV2Test(t *testing.T) {
	absPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	err = builder.Build(absPath)
	assert.NoError(t, err)
	pkgBin := "../../testdata/demo.test"
	_, err = os.Stat(pkgBin)
	assert.NoError(t, err)
	defer os.Remove("../../testdata/demo.test")
	testResult, err := RunGinkgoV2Test(absPath, "demo.test", "../../testdata/demo_test.go", []string{"Testcase"})
	assert.NoError(t, err)
	defer os.Remove("../../testdata/output.json")
	assert.NotEqual(t, len(testResult), 0)
}
