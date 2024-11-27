package runner

import (
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	cmdpkg "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/cmdline"
	ginkgoResult "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/result"
	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/util"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

func RunGinkgoV1Test(projPath string, pkgBin string, filepath string, tcNames []string) ([]*sdkModel.TestResult, error) {
	var testResults []*sdkModel.TestResult
	_, filename := path.Split(pkgBin)
	outputXmlFile := path.Join(projPath, fmt.Sprintf("%s_output.xml", filename))
	cmdline := pkgBin + fmt.Sprintf(` --ginkgo.v --ginkgo.noColor --ginkgo.trace --ginkgo.reportFile="%s" --ginkgo.focus="%s" `, outputXmlFile, cmdpkg.GenTestCaseFocusName(tcNames))
	log.Printf("Run cmdline %s", cmdline)
	startTime := time.Now()
	workDir := strings.TrimSuffix(pkgBin, ".test")
	_, stderr, err := ginkgoUtil.RunCommandWithOutput(cmdline, workDir)
	delta := time.Since(startTime)
	log.Printf("Run test command cost %.2fs", delta.Seconds())
	if err != nil {
		log.Printf("Command exit code: %v", err)
	}
	if exists, err := ginkgoUtil.FileExists(outputXmlFile); err != nil || !exists {
		if err != nil {
			return testResults, err
		}
		if stderr == "" {
			stderr = "Output xml file not exist"
		}
		step := &sdkModel.TestCaseStep{
			StartTime: startTime,
			EndTime:   startTime,
			Title:     "Error",
			Logs: []*sdkModel.TestCaseLog{
				{
					Time:    startTime,
					Level:   sdkModel.LogLevelError,
					Content: stderr,
				},
			},
		}
		for _, tcName := range tcNames {
			testResults = append(testResults, &sdkModel.TestResult{
				Test: &sdkModel.TestCase{
					Name: filepath + "?" + tcName,
				},
				ResultType: sdkModel.ResultTypeFailed,
				StartTime:  startTime,
				EndTime:    startTime,
				Message:    ginkgoUtil.ShortenString(stderr, 512),
				Steps:      []*sdkModel.TestCaseStep{step},
			})
		}
		return testResults, nil
	}
	testResults, err = ginkgoResult.ParseXmlResultFile(projPath, outputXmlFile)
	if err != nil {
		return testResults, err
	}
	for _, result := range testResults {
		result.Test.Name = filepath + "?" + result.Test.Name
		result.Test.Attributes["description"] = result.Test.Name
	}
	return testResults, nil
}
