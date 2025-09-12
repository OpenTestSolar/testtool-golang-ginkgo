package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	ginkgoResult "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/result"
	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestGenarateCommandLine(t *testing.T) {
	extraArgs := `--procs 3 --timeout 5h --focus "case" --label-filter "( label01||label02)" --timeout 5h `
	jsonFileName := "output.json"
	projPath := "/data/workspace"
	pkgBin := "suite.test"
	tcNames := []string{
		"case01",
		"case02",
	}
	os.Setenv("TESTSOLAR_TTP_FOCUS", "true")
	expected := `ginkgo --v --no-color --trace --json-report "output.json" --output-dir "/data/workspace" --always-emit-ginkgo-writer --procs "3" --timeout "5h" --focus "case" --label-filter "( label01||label02)" suite.test`
	cmdline := genarateCommandLine(extraArgs, jsonFileName, projPath, pkgBin, tcNames, true)
	assert.Equal(t, expected, cmdline, "should return the expected command line")
	extraArgs = "b64://LS1wcm9jcyAzIC0tdGltZW91dCA1aCAtLWZvY3VzICJjYXNlIiAtLWxhYmVsLWZpbHRlciAiKCBsYWJlbDAxfHxsYWJlbDAyKSIgLS10aW1lb3V0IDVoIA=="
	cmdline = genarateCommandLine(extraArgs, jsonFileName, projPath, pkgBin, tcNames, true)
	assert.Equal(t, expected, cmdline, "should return the expected command line")
	extraArgs = `--ginkgo.label-filter "( label01||label02)"`
	expected = "suite.test --ginkgo.v --ginkgo.no-color --ginkgo.trace --ginkgo.json-report=\"output.json\" --ginkgo.always-emit-ginkgo-writer --ginkgo.focus=\"case01$|case02$\" --ginkgo.label-filter \"( label01||label02)\""
	cmdline = genarateCommandLine(extraArgs, jsonFileName, projPath, pkgBin, tcNames, false)
	assert.Equal(t, expected, cmdline, "should return the expected command line")
	extraArgs = ""
	expected = "suite.test --ginkgo.v --ginkgo.no-color --ginkgo.trace --ginkgo.json-report=\"output.json\" --ginkgo.always-emit-ginkgo-writer --ginkgo.focus=\"case01$|case02$\""
	cmdline = genarateCommandLine(extraArgs, jsonFileName, projPath, pkgBin, tcNames, false)
	assert.Equal(t, expected, cmdline, "should return the expected command line")
}

func Test_obtainExpectedExecuteCases(t *testing.T) {
	projPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	packagePath := filepath.Join(projPath, "demo")
	pkgBin := packagePath + ".test"
	_, _, err = ginkgoUtil.RunCommandWithOutput(fmt.Sprintf("go test -c %s -o %s", packagePath, pkgBin), projPath)
	assert.NoError(t, err)
	_, err = os.Stat(pkgBin)
	assert.NoError(t, err)
	defer os.Remove(pkgBin)
	cmdline := fmt.Sprintf("ginkgo --v --no-color --procs 10 --always-emit-ginkgo-writer %s", pkgBin)
	err = os.Setenv("TESTSOLAR_TTP_EXTRAARGS", "1")
	assert.NoError(t, err)
	expectedCases := obtainExpectedExecuteCasesByDryRun(projPath, cmdline, []string{})
	assert.Len(t, expectedCases, 5)
}

func Test_getExpectedCases(t *testing.T) {
	projPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	packagePath := filepath.Join(projPath, "demo")
	pkgBin := packagePath + ".test"
	_, _, err = ginkgoUtil.RunCommandWithOutput(fmt.Sprintf("go test -c %s -o %s", packagePath, pkgBin), projPath)
	assert.NoError(t, err)
	_, err = os.Stat(pkgBin)
	assert.NoError(t, err)
	defer os.Remove(pkgBin)
	cmdline := fmt.Sprintf("ginkgo --v --no-color --procs 10 --always-emit-ginkgo-writer %s", pkgBin)
	err = os.Setenv("TESTSOLAR_TTP_EXTRAARGS", "1")
	assert.NoError(t, err)
	expectedCases := getExpectedCases(cmdline, projPath, "tests/ginkgo/demo/demo_test.go", "tests/ginkgo/demo", []string{})
	assert.Len(t, expectedCases, 5)
}

func Test_convertExpectedCasesToFailedCases(t *testing.T) {
	results := generateFailedCasesWhenSuitePanic("stdout", "stderr", nil, []string{
		"xxx/yyy/zzz.go?xxx",
		"xxx/yy2/zzz2.go?yyy",
	})
	assert.NotEqual(t, len(results), 0)
	spec := &ginkgoResult.Spec{
		StartTime: time.Now(),
		EndTime:   time.Now(),
		RunTime:   100000,
	}
	results = generateFailedCasesWhenSuitePanic("", "", spec, []string{
		"xxx/yyy/zzz.go?xxx",
		"xxx/yy2/zzz2.go?yyy",
	})
	assert.NotEqual(t, len(results), 0)
}

func Test_regenerateDryRunCmd(t *testing.T) {
	dryRunCmd, err := regenerateDryRunCmd(`ginkgo --v --no-color --trace --json-report output.json --output-dir /data/workspace --always-emit-ginkgo-writer --procs "10" --timeout "5h" --label-filter "(label01)" /data/workspace.test`, []string{})
	assert.NoError(t, err)
	assert.Equal(t, dryRunCmd, `ginkgo --dry-run --v --no-color --trace --json-report "output.json" --output-dir "/data/workspace" --timeout "5h" --label-filter "(label01)" /data/workspace.test`)
	dryRunCmd, err = regenerateDryRunCmd(`ginkgo --v --no-color --trace --json-report output.json --output-dir /data/workspace --always-emit-ginkgo-writer --procs "10" --timeout "5h" --label-filter "(label01)" --focus "\[cls\] 环境变量 采集volume$|\[cls\] 环境变量 采集容器路径且包含中文日志$" /data/workspace.test`, []string{"[cls] 环境变量 采集volume", "[cls] 环境变量 采集容器路径且包含中文日志"})
	assert.NoError(t, err)
	assert.Equal(t, dryRunCmd, `ginkgo --dry-run --focus "\[cls\] 环境变量 采集volume$|\[cls\] 环境变量 采集容器路径且包含中文日志$" --v --no-color --trace --json-report "output.json" --output-dir "/data/workspace" --timeout "5h" --label-filter "(label01)" /data/workspace.test`)
}
