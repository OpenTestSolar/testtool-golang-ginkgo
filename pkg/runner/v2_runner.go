package runner

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	cmdpkg "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/cmdline"

	ginkgoResult "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/result"
	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/testcase"
	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/util"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

func genarateCommandLine(extraArgs, jsonFileName, projPath, pkgBin string, tcNames []string, hasClient bool) string {
	if hasClient {
		defaultCmdLine := fmt.Sprintf("ginkgo --v --no-color --trace --json-report %s --output-dir %s --always-emit-ginkgo-writer", jsonFileName, projPath)
		cmdArgs, err := cmdpkg.NewCmdArgsParseByCmdLine(defaultCmdLine)
		if err != nil {
			log.Printf("Parse cmdline [%s] error: %v", defaultCmdLine, err)
			return ""
		}
		if extraArgs != "" {
			if strings.HasPrefix(extraArgs, "b64://") {
				decodedValue, err := base64.StdEncoding.DecodeString(extraArgs[6:])
				if err != nil {
					log.Printf("Decode %s failed, err: %s", extraArgs, err.Error())
				} else {
					extraArgs = string(decodedValue)
				}
			}
			extraCmdArgs, err := cmdpkg.NewCmdArgsParseByCmdLine(extraArgs)
			if err != nil {
				log.Printf("Parse extra cmd args [%s] error: %v", extraArgs, err)
			} else {
				cmdArgs.Merge(extraCmdArgs)
			}
		}
		// 通过环境变量控制是否需要以`--focus`的形式下发用例执行
		// 部分场景下ginkgo用例名中存在特殊字符，拼接到命令行中会导致报错，因此需要避免使用focus参数
		if cmdArgs.NeedFocus() {
			cmdArgs.AddIfNotExists([]*cmdpkg.CommandArg{{Key: "--focus", Value: fmt.Sprintf("\"%s\"", cmdpkg.GenTestCaseFocusName(tcNames, false))}})
		}
		cmdArgs.Add(&cmdpkg.CommandArg{Key: "", Value: pkgBin})
		cmdline := cmdArgs.GenerateCmdLineStr()
		return cmdline
	} else {
		return pkgBin + fmt.Sprintf(` --ginkgo.v --ginkgo.no-color --ginkgo.trace --ginkgo.json-report="%s" --ginkgo.always-emit-ginkgo-writer --ginkgo.focus="%s"`, jsonFileName, cmdpkg.GenTestCaseFocusName(tcNames, false))
	}
}

// RegenerateDryRunCmd 根据输入的ginkgo执行命令重新生成对应的dry run命令
// 输入参数如下：
// - cmd: 需要处理的原始命令行参数字符串
// 返回值：
//   - newCmd: 生成的新命令字符合成后的字符串表示形式
//     注意：当传入不合法或者不符合要求的命令时将会抛出错误
//
// 示例
// - 输入: cmd = "ginkgo --v --no-color --procs 10 --always-emit-ginkgo-writer -p /path/to/your/file.test"
// - 输出: newCmd = "ginkgo --dry-run --v --no-color /path/to/your/file.test"
func regenerateDryRunCmd(cmd string) (string, error) {
	dryRunCmdExcludedKeys := []string{"ginkgo", "--procs", "--always-emit-ginkgo-writer"}

	originalCmdArgs, err := cmdpkg.NewCmdArgsParseByCmdLine(cmd)
	if err != nil {
		return "", err
	}
	dryRunCmdArgs := cmdpkg.NewCmdArgs()
	dryRunCmdArgs.Add(&cmdpkg.CommandArg{
		Key:   "ginkgo",
		Value: "",
	})
	dryRunCmdArgs.Add(&cmdpkg.CommandArg{
		Key:   "--dry-run",
		Value: "",
	})
	for _, arg := range originalCmdArgs.Args {
		if !ginkgoUtil.ElementIsInSlice(arg.Key, dryRunCmdExcludedKeys) {
			dryRunCmdArgs.Add(arg)
		}
	}
	dryRunCmdLine := dryRunCmdArgs.GenerateCmdLineStr()
	return dryRunCmdLine, nil
}

// obtainExpectedExecuteCasesByDryRun 根据给定的项目路径和命令行字符串，获取预期执行的测试用例列表。
// 参数 projPath 是项目的路径。
// 参数 cmdline 是命令行字符串。
func obtainExpectedExecuteCasesByDryRun(projPath, cmdline string) []*ginkgoTestcase.TestCase {
	log.Printf("regenerate dry run cmd by run cmd: [%s]", cmdline)
	dryRunCmd, err := regenerateDryRunCmd(cmdline)
	if err != nil {
		log.Printf("regenerate dry run cmd by run cmd [%s] failed, err: %s", cmdline, err.Error())
		return nil
	}
	log.Printf("execute regenerated dry run cmdline: [%s]", dryRunCmd)
	output, _, err := ginkgoUtil.RunCommandWithOutput(dryRunCmd, projPath)
	if err != nil {
		log.Printf("execute regenerated dry run cmd err: %v", err)
		return nil
	}
	testcaseList, err := ginkgoResult.ParseCaseByReg(projPath, output, 2, "")
	if err != nil {
		log.Printf("find testcase in log error: %v", err)
		return nil
	}
	var filteredCases []*ginkgoTestcase.TestCase
	for _, c := range testcaseList {
		if !strings.HasPrefix(c.Name, "P [PENDING]") {
			filteredCases = append(filteredCases, c)
		}
	}
	return filteredCases
}

// generateFailedCasesWhenSuitePanic 函数用于在测试套件发生 panic 时生成失败的测试用例。
// 它接收标准输出、标准错误输出、测试规范和预期的测试用例列表作为参数。
func generateFailedCasesWhenSuitePanic(stdout, stderr string, spec *ginkgoResult.Spec, expectedCases []string) []*sdkModel.TestResult {
	failedResults := make([]*sdkModel.TestResult, 0, len(expectedCases))
	var steps []*sdkModel.TestCaseStep
	var startTime time.Time
	var endTime time.Time
	if spec != nil {
		steps, startTime, endTime = spec.GenerateSteps(), spec.StartTime, spec.EndTime
	} else {
		failureLogs := []*sdkModel.TestCaseLog{{
			Level:   sdkModel.LogLevelError,
			Content: fmt.Sprintf("stdout: %s\nstderr: %s\n", stdout, stderr),
		}}
		steps = []*sdkModel.TestCaseStep{{
			Title:      "testsuite panic",
			Logs:       failureLogs,
			ResultType: sdkModel.ResultTypeFailed,
		}}
		startTime, endTime = time.Now(), time.Now()
	}
	for _, c := range expectedCases {
		failedResults = append(failedResults, &sdkModel.TestResult{
			Test: &sdkModel.TestCase{
				Name:       c,
				Attributes: map[string]string{},
			},
			StartTime:  startTime,
			EndTime:    endTime,
			ResultType: sdkModel.ResultTypeFailed,
			Message:    "suite panic",
			Steps:      steps,
		})
	}
	return failedResults
}

// getExpectedCases 函数用于获取预期的测试用例列表。
// 参数：
// cmdline: 命令行参数
// projPath: 项目路径
// filepath: 文件路径
// packPath: 包路径
// tcNames: 测试用例名称列表
func getExpectedCases(cmdline, projPath, filepath, packPath string, tcNames []string) []string {
	var expectedCases []*ginkgoTestcase.TestCase
	var finalCases []string
	if os.Getenv("TESTSOLAR_TTP_EXTRAARGS") != "" {
		expectedCases = obtainExpectedExecuteCasesByDryRun(projPath, cmdline)
	}
	if len(expectedCases) == 0 {
		for _, name := range tcNames {
			expectedCases = append(expectedCases, &ginkgoTestcase.TestCase{Path: filepath, Name: name})
		}
	}
	for _, c := range expectedCases {
		var fullName string
		if !strings.HasPrefix(c.Path, packPath) {
			fullName = packPath + "?" + c.Name
		} else {
			fullName = c.Path + "?" + c.Name
		}
		finalCases = append(finalCases, fullName)
	}
	return finalCases
}

func RunGinkgoV2Test(projPath, pkgBin, filepath string, tcNames []string) ([]*sdkModel.TestResult, error) {
	outputJsonFile := fmt.Sprintf("output-%s.json", ginkgoUtil.GenRandomString(8))
	err := ginkgoUtil.RemoveFile(outputJsonFile)
	if err != nil {
		log.Printf("failed to remove output json file, err: %s", err.Error())
	}
	cmdline := genarateCommandLine(os.Getenv("TESTSOLAR_TTP_EXTRAARGS"), outputJsonFile, projPath, pkgBin, tcNames, CheckGinkgoCli())
	log.Printf("Run cmdline %s", cmdline)
	stdout, stderr, err := ginkgoUtil.RunCommandWithOutput(cmdline, projPath)
	if err != nil {
		log.Printf("Command excute failed, stdout: %s, stderr %s, err: %v", stdout, stderr, err)
	}
	packPath := cmdpkg.ExtractPackPathFromBinFile(pkgBin, projPath)
	if empty, err := ginkgoUtil.IsJsonFileEmpty(outputJsonFile); err != nil || empty {
		// 如果输出结果文件为空，说明测试套执行失败，需要将本次期望执行的用例置为失败并上报
		expectedCases := getExpectedCases(cmdline, projPath, filepath, packPath, tcNames)
		return generateFailedCasesWhenSuitePanic(stdout, stderr, nil, expectedCases), nil
	}
	log.Printf("Parse json file %s", outputJsonFile)
	resultParser, err := ginkgoResult.NewResultParser(outputJsonFile, projPath, packPath, true)
	if err != nil {
		log.Printf("instantiate result parser failed, err: %s", err.Error())
		return nil, err
	}
	suite, err := resultParser.GetPanicSuite()
	if err != nil {
		log.Printf("get panic suite failed, err: %s", err.Error())
		return nil, err
	}
	if suite != nil {
		// 如果存在状态为panic的测试套，说明测试套执行失败，需要将本次期望执行的用例置为失败并上报
		expectedCases := getExpectedCases(cmdline, projPath, filepath, packPath, tcNames)
		return generateFailedCasesWhenSuitePanic("", "", suite, expectedCases), nil
	}
	return resultParser.Parse()
}
