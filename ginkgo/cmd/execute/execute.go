package execute

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	ginkgoBuilder "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/builder"
	ginkgoRunner "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/runner"
	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/testcase"
	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/util"

	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	pkgErrors "github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ExecuteOptions struct {
	executePath string
}

// NewExecuteOptions NewBuildOptions new build options with default value
func NewExecuteOptions() *ExecuteOptions {
	return &ExecuteOptions{}
}

// NewCmdExecute NewCmdBuild create a build command
func NewCmdExecute() *cobra.Command {
	o := NewExecuteOptions()
	cmd := cobra.Command{
		Use:   "execute",
		Short: "Execute testcases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunExecute(cmd)
		},
	}
	cmd.Flags().StringVarP(&o.executePath, "path", "p", "", "Path of testcase info")
	_ = cmd.MarkFlagRequired("path")
	return &cmd
}

func groupTestCasesByPathAndName(projPath string, testcases []*ginkgoTestcase.TestCase) (map[string]map[string][]*ginkgoTestcase.TestCase, error) {
	packages := map[string]map[string][]*ginkgoTestcase.TestCase{}
	for _, testcase := range testcases {
		path, name, err := ginkgoUtil.GetPathAndFileName(projPath, testcase.Path)
		if err != nil {
			log.Printf("Get path and file name from %s failed, err: %s", testcase.Path, err.Error())
			return nil, err
		}
		path = strings.TrimSuffix(path, string(os.PathSeparator))
		_, ok := packages[path]
		if !ok {
			packages[path] = map[string][]*ginkgoTestcase.TestCase{}
		}
		_, ok = packages[path][name]
		if !ok {
			packages[path][name] = []*ginkgoTestcase.TestCase{}
		}
		packages[path][name] = append(packages[path][name], testcase)
	}
	return packages, nil
}

func reportTestResults(testResults []*sdkModel.TestResult, reporter api.Reporter) error {
	for _, result := range testResults {
		log.Printf("[PLUGIN]try to report testresult %s", result.Test.Name)
		err := reporter.ReportCaseResult(result)
		if err != nil {
			return pkgErrors.Wrap(err, "failed to report load result")
		}
	}
	return nil
}

func findTestPackagesByPath(path string) ([]string, error) {
	subDirs := []string{}
	foundSubDirs := map[string]bool{}
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return pkgErrors.Wrapf(err, "walk subdir %s failed", p)
		}
		if !d.IsDir() && strings.HasSuffix(p, "_suite_test.go") {
			subDir := filepath.Dir(p)
			if _, ok := foundSubDirs[subDir]; !ok {
				subDirs = append(subDirs, subDir)
				foundSubDirs[subDir] = true
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, pkgErrors.Wrapf(err, "walk dir %s failed", path)
	}
	return subDirs, nil
}

// findCompileBinary 通过用例路径查询对应二进制可执行文件
// 通常情况下预编译生成的二进制文件与用例文件所在目录同级
// 但是部分场景下二进制文件可能处于更上层目录，需要递归查询
func findCompileBinary(path string) string {
	for {
		preCompileFile := path + ".test"
		if _, err := os.Stat(preCompileFile); err != nil {
			path = filepath.Dir(path)
			if path == "." || path == "/" || path == "\\" {
				break
			} else {
				continue
			}
		} else {
			return preCompileFile
		}
	}
	return ""
}

func discoverExecutableTestcases(testcases []*ginkgoTestcase.TestCase) ([]*ginkgoTestcase.TestCase, error) {
	excutableTestcases := []*ginkgoTestcase.TestCase{}
	for _, testcase := range testcases {
		fd, err := os.Stat(testcase.Path)
		if err != nil {
			preCompileFile := findCompileBinary(filepath.Dir(testcase.Path))
			if preCompileFile == "" {
				return nil, pkgErrors.Wrapf(err, "get file info %s failed and there is no precompiled binary file", testcase.Path)
			}
			log.Printf("[PLUGIN]can't find testcase file %s, but find precompiled binary file %s", testcase.Path, preCompileFile)
			testcase.Path = strings.TrimSuffix(preCompileFile, ".test")
			excutableTestcases = append(excutableTestcases, testcase)
			continue
		}
		if !fd.IsDir() {
			excutableTestcases = append(excutableTestcases, testcase)
			continue
		}
		packages, err := findTestPackagesByPath(testcase.Path)
		if err != nil {
			return nil, pkgErrors.Wrapf(err, "find packages in %s failed", testcase.Path)
		}
		if len(packages) == 0 {
			return nil, pkgErrors.Wrapf(err, "failed to found available test packages in dir %s", testcase.Path)
		}
		for _, pack := range packages {
			if pack != testcase.Path {
				log.Printf("[PLUGIN]found test package %s in %s", pack, testcase.Path)
			}
			t := &ginkgoTestcase.TestCase{
				Path:       pack,
				Name:       testcase.Name,
				Attributes: testcase.Attributes,
			}
			excutableTestcases = append(excutableTestcases, t)
		}
	}
	return excutableTestcases, nil
}

func executeTestcases(projPath string, packages map[string]map[string][]*ginkgoTestcase.TestCase) ([]*sdkModel.TestResult, error) {
	var testResults []*sdkModel.TestResult
	for path, filesCases := range packages {
		// test one suite each time
		for filename, cases := range filesCases {
			pkgBin := filepath.Join(projPath, path+".test")
			_, err := os.Stat(pkgBin)
			if err != nil {
				log.Printf("Can't find package bin file %s during running, try to build it...", pkgBin)
				_, err := ginkgoBuilder.BuildTestPackage(projPath, path, false)
				if err != nil {
					log.Printf("Build package %s during running failed, err: %s", path, err.Error())
					continue
				}
			}
			tcNames := make([]string, len(cases))
			for i, tc := range cases {
				tcNames[i] = tc.Name
			}
			ginkgoVersion := ginkgoUtil.FindGinkgoVersion(strings.TrimSuffix(pkgBin, ".test"))
			log.Printf("Run test cases: %v by bin file %s", tcNames, pkgBin)
			var results []*sdkModel.TestResult
			if ginkgoVersion == 1 {
				results, err = ginkgoRunner.RunGinkgoV1Test(projPath, pkgBin, filepath.Join(path, filename), tcNames)
			} else {
				results, err = ginkgoRunner.RunGinkgoV2Test(projPath, pkgBin, filepath.Join(path, filename), tcNames)
			}
			if err != nil {
				log.Printf("Run test cases failed, err: %s", err.Error())
				continue
			}
			if len(results) == 0 {
				log.Println("No test results found during executing")
			}
			testResults = append(testResults, results...)
		}
	}
	return testResults, nil
}

func parseTestcases(testSelectors []string) ([]*ginkgoTestcase.TestCase, []*sdkModel.TestResult, error) {
	var testcases []*ginkgoTestcase.TestCase
	var failedResults []*sdkModel.TestResult
	for _, selector := range testSelectors {
		testcase, err := ginkgoTestcase.ParseTestCaseBySelector(selector)
		if err != nil {
			message := fmt.Sprintf("parse testcase [%s] failed, err: %s", selector, err.Error())
			log.Printf("[PLUGIN] %s", message)
			failedResults = append(failedResults, &sdkModel.TestResult{
				Test: &sdkModel.TestCase{
					Name: selector,
				},
				ResultType: sdkModel.ResultTypeFailed,
				Message:    message,
				StartTime:  time.Now(),
				EndTime:    time.Now(),
			})
			continue
		}
		testcases = append(testcases, testcase)
	}
	if len(testcases) == 0 {
		return nil, nil, errors.New("no available testcases")
	}
	return testcases, failedResults, nil
}

func (o *ExecuteOptions) RunExecute(cmd *cobra.Command) error {
	config, err := ginkgoTestcase.UnmarshalCaseInfo(o.executePath)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to unmarshal case info")
	}
	testcases, parseFailedResults, err := parseTestcases(config.TestSelectors)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to parse test selectors")
	}
	// 递归查询包含实际可执行用例的目录
	excutableTestcases, err := discoverExecutableTestcases(testcases)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to discover excutable testcases")
	}
	projPath := ginkgoUtil.GetWorkspace(config.ProjectPath)
	_, err = os.Stat(projPath)
	if err != nil {
		return pkgErrors.Wrapf(err, "stat project path %s failed", projPath)
	}
	packages, err := groupTestCasesByPathAndName(projPath, excutableTestcases)
	if err != nil {
		return pkgErrors.Wrap(err, "failed to group testcases by path and name")
	}
	testResults, err := executeTestcases(projPath, packages)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to execute testcases")
	}
	reporter, err := sdkClient.NewReporterClient(config.FileReportPath)
	if err != nil {
		return pkgErrors.Wrap(err, "failed to create reporter")
	}
	if len(parseFailedResults) != 0 {
		testResults = append(testResults, parseFailedResults...)
	}
	err = reportTestResults(testResults, reporter)
	if err != nil {
		return pkgErrors.Wrap(err, "failed to report test results")
	}
	return nil
}
