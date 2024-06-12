// FILEPATH: /data/golang/ginkgo/pkg/runner/util_test.go
package runner

import (
	ginkgoTestcase "ginkgo/pkg/testcase"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGinkgoVersion(t *testing.T) {
	testcases := []*ginkgoTestcase.TestCase{
		{
			Attributes: map[string]string{
				"ginkgoVersion": "1",
			},
		},
		{
			Attributes: map[string]string{
				"ginkgoVersion": "2",
			},
		},
	}

	version := GetGinkgoVersion(testcases)
	assert.Equal(t, "2", version, "应该返回最后一个测试用例的ginkgo版本")
}

func TestCheckGinkgoCli(t *testing.T) {
	if _, err := exec.LookPath("ginkgo"); err != nil {
		assert.False(t, CheckGinkgoCli())
	} else {
		assert.True(t, CheckGinkgoCli())
	}
}
