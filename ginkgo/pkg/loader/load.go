package loader

import (
	"log"
	"os"
	"path/filepath"

	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/testcase"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

func LoadTestCase(projPath string, selectorPath string) ([]*ginkgoTestcase.TestCase, []*sdkModel.LoadError) {
	var testcaseList []*ginkgoTestcase.TestCase
	var loadErrors []*sdkModel.LoadError
	selectorAbsPath := filepath.Join(projPath, selectorPath)
	fi, err := os.Stat(selectorAbsPath)
	if err != nil {
		loadErrors = append(loadErrors, &sdkModel.LoadError{
			Name:    selectorPath,
			Message: err.Error(),
		})
		return nil, loadErrors
	}
	parseMode := os.Getenv("TESTSOLAR_TTP_PARSEMODE")
	log.Printf("Try to load testcases from path %s, parse mode: %s", selectorAbsPath, parseMode)
	if fi.IsDir() {
		if parseMode != "static" {
			loadedTestCases, lErrors := DynamicLoadTestcaseInDir(projPath, selectorAbsPath)
			testcaseList = append(testcaseList, loadedTestCases...)
			loadErrors = append(loadErrors, lErrors...)
		} else {
			err := filepath.Walk(selectorAbsPath, func(path string, fi os.FileInfo, _ error) error {
				loadedTestCases, lErrors := ParseTestCaseInFile(projPath, path)
				testcaseList = append(testcaseList, loadedTestCases...)
				loadErrors = append(loadErrors, lErrors...)
				return nil
			})
			if err != nil {
				log.Printf("Failed to load testcases from %s, err: %s", selectorAbsPath, err)
			}
		}
	} else {
		if parseMode != "static" {
			loadedTestCases, lErrors := DynamicLoadTestcaseInFile(projPath, selectorAbsPath)
			testcaseList = append(testcaseList, loadedTestCases...)
			loadErrors = append(loadErrors, lErrors...)
		} else {
			loadedTestCases, lErrors := ParseTestCaseInFile(projPath, selectorAbsPath)
			testcaseList = append(testcaseList, loadedTestCases...)
			loadErrors = append(loadErrors, lErrors...)
		}
	}
	return testcaseList, loadErrors
}
