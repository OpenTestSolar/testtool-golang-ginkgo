package execute

import (
	"errors"
	"fmt"
	ginkgoBuilder "ginkgo/pkg/builder"
	ginkgoRunner "ginkgo/pkg/runner"
	ginkgoTestcase "ginkgo/pkg/testcase"
	ginkgoUtil "ginkgo/pkg/util"
	"log"
	"os"
	"path/filepath"
	"strings"

	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
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

func reportTestResults(testResults []*sdkModel.TestResult) error {
	reporter, err := sdkClient.NewReporterClient()
	if err != nil {
		fmt.Printf("Failed to create reporter: %v\n", err)
		return err
	}
	for _, result := range testResults {
		err = reporter.ReportCaseResult(result)
		if err != nil {
			fmt.Printf("Failed to report load result: %v\n", err)
			return err
		}
	}
	err = reporter.Close()
	if err != nil {
		fmt.Printf("Failed to close report: %v\n", err)
		return err
	}
	return nil
}

func executeTestcases(projPath string, packages map[string]map[string][]*ginkgoTestcase.TestCase) ([]*sdkModel.TestResult, error) {
	var testResults []*sdkModel.TestResult
	for path, filesCases := range packages {
		pkgBin := filepath.Join(projPath, path+".test")
		_, err := os.Stat(pkgBin)
		if err != nil {
			log.Printf("Can't find package bin file %s during running, try to build it...", pkgBin)
			_, err := ginkgoBuilder.BuildTestPackage(projPath, path, false)
			if err != nil {
				return nil, fmt.Errorf("Build package %s during running failed, err: %s", path, err.Error())
			}
		}
		// test one suite each time
		for filename, cases := range filesCases {
			tcNames := make([]string, len(cases))
			for i, tc := range cases {
				tcNames[i] = tc.Name
			}
			log.Printf("Run test cases: %v by bin file %s", tcNames, pkgBin)
			var results []*sdkModel.TestResult
			if ginkgoVersion := ginkgoRunner.GetGinkgoVersion(cases); ginkgoVersion == "1" {
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

func parseTestcases(testSelectors []string) ([]*ginkgoTestcase.TestCase, error) {
	var testcases []*ginkgoTestcase.TestCase
	for _, selector := range testSelectors {
		testcase, err := ginkgoTestcase.ParseTestCaseBySelector(selector)
		if err != nil {
			log.Printf("parse testcase by selector [%s] failed, err: %s", selector, err.Error())
			continue
		}
		testcases = append(testcases, testcase)
	}
	if len(testcases) == 0 {
		return nil, errors.New("no available testcases")
	}
	return testcases, nil
}

func (o *ExecuteOptions) RunExecute(cmd *cobra.Command) error {
	// load case info from yaml file
	config, err := ginkgoTestcase.UnmarshalCaseInfo(o.executePath)
	if err != nil {
		return err
	}
	// parse testcases
	testcases, err := parseTestcases(config.TestSelectors)
	if err != nil {
		return err
	}
	// get workspace
	projPath := ginkgoUtil.GetWorkspace(config.ProjectPath)
	_, err = os.Stat(projPath)
	if err != nil {
		return fmt.Errorf("stat project path %s failed, err: %s", projPath, err.Error())
	}
	// get testcases grouped by path and name
	packages, err := groupTestCasesByPathAndName(projPath, testcases)
	if err != nil {
		return err
	}
	// run testcases
	testResults, err := executeTestcases(projPath, packages)
	if err != nil {
		return err
	}
	// report test results
	return reportTestResults(testResults)
}
