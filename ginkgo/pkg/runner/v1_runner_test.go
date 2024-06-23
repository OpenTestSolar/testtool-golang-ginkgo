package runner

import (
	builder "ginkgo/pkg/builder"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunGinkgoV1Test(t *testing.T) {
	absPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	err = builder.Build(absPath)
	assert.NoError(t, err)
	pkgBin := "../../testdata/demo.test"
	_, err = os.Stat(pkgBin)
	assert.NoError(t, err)
	defer os.Remove("../../testdata/demo.test")
	testResult, err := RunGinkgoV1Test(absPath, "demo.test", "../../testdata/demo_test.go", []string{"Testcase"})
	assert.NoError(t, err)
	assert.NotEqual(t, len(testResult), 0)
}
