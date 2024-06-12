package loader

import (
	"log"
	"os"
	"path/filepath"

	ginkgoTestcase "ginkgo/pkg/testcase"
)

func LoadTestCase(projPath string, selectorPath string) ([]*ginkgoTestcase.TestCase, error) {
	var testcaseList []*ginkgoTestcase.TestCase
	selectorAbsPath := filepath.Join(projPath, selectorPath)
	fi, err := os.Stat(selectorAbsPath)
	if err != nil {
		log.Printf("stat selector abs path: %s failed, err: %s", selectorAbsPath, err.Error())
		return testcaseList, err
	}
	parseMode := os.Getenv("TESTSOLAR_TTP_PARSEMODE")
	log.Printf("Try to load testcases from path %s, parse mode: %s", selectorAbsPath, parseMode)
	if fi.IsDir() {
		if parseMode == "dynamic" {
			loadedTestCases, err := DynamicLoadTestcaseInDir(projPath, selectorAbsPath)
			if err != nil {
				log.Printf("dynamic load testcase from dir failed: %v", err)
				return testcaseList, err
			}
			testcaseList = append(testcaseList, loadedTestCases...)
		} else {
			filepath.Walk(selectorAbsPath, func(path string, fi os.FileInfo, _ error) error {
				loadedTestCases, err := ParseTestCaseInFile(projPath, path)
				if err != nil {
					log.Printf("Static parse testcase within path %s failed, err: %v", path, err)
					return nil
				}
				testcaseList = append(testcaseList, loadedTestCases...)
				return nil
			})
		}
	} else {
		if parseMode == "dynamic" {
			loadedTestCases, err := DynamicLoadTestcaseInFile(projPath, selectorAbsPath)
			if err != nil {
				log.Printf("dynamic load testcase in faile %s failed, err: %v", selectorAbsPath, err)
				return testcaseList, err
			}
			testcaseList = append(testcaseList, loadedTestCases...)
		} else {
			loadedTestCases, err := ParseTestCaseInFile(projPath, selectorAbsPath)
			if err != nil {
				log.Printf("Parse testcase file %s failed, err: %v", selectorAbsPath, err)
				return testcaseList, err
			}
			testcaseList = append(testcaseList, loadedTestCases...)
		}
	}
	return testcaseList, nil
}
