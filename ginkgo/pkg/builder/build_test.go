package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_build(t *testing.T) {
	// test build
	absPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	err = Build(absPath)
	assert.NoError(t, err)
	pkgBin := filepath.Join(absPath, "demo.test")
	_, err = os.Stat(pkgBin)
	assert.NoError(t, err)
	err = os.Remove("../../testdata/demo.test")
	assert.NoError(t, err)
	// test build with env
	err = os.Setenv("TESTSOlAR_TTP_CONCURRENTBUILD", "true")
	assert.NoError(t, err)
	err = os.Setenv("TESTSOLAR_TTP_COMPRESSBINARY", "true")
	assert.NoError(t, err)
	err = Build(absPath)
	assert.NoError(t, err)
	pkgBin = filepath.Join(absPath, "demo.test")
	_, err = os.Stat(pkgBin)
	assert.NoError(t, err)
	defer os.Remove(pkgBin)
}
