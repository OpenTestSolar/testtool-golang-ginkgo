package loader

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ginkgoResult "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/result"

	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/testcase"
	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/util"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
)

// func findTestSuiteDirs(absPath string) ([]string, error) {
// 	var testDirs []string
// 	visited := make(map[string]bool)
// 	log.Printf("find test suite from directory: %s", absPath)
// 	err := filepath.Walk(absPath, func(currentPath string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
// 			dir := filepath.Dir(currentPath) // 获取文件所在的目录路径
// 			if !visited[dir] {
// 				visited[dir] = true
// 				testDirs = append(testDirs, dir)
// 			}
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return testDirs, nil
// }

func findGinkgoVersion(path string) (ginkgoVersion int) {
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
			version := checkGinkgoImportVersion(filePath)
			if version > 0 {
				ginkgoVersion = version
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error searching for test files: %v\n", err)
	}
	return ginkgoVersion
}

func checkGinkgoImportVersion(file string) int {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		return 0
	}

	version := 0
	ast.Inspect(node, func(n ast.Node) bool {
		importSpec, ok := n.(*ast.ImportSpec)
		if ok && importSpec.Path != nil {
			if importSpec.Path.Value == "\"github.com/onsi/ginkgo/v2\"" {
				version = 2
				return false
			} else if importSpec.Path.Value == "\"github.com/onsi/ginkgo\"" {
				version = 1
				return false
			}
		}
		return true
	})

	return version
}

func findBinFile(absSelectorPath string) string {
	testFileName := filepath.Join(absSelectorPath + ".test")
	_, err := os.Stat(testFileName)
	if os.IsNotExist(err) {
		return ""
	}
	absPath, err := filepath.Abs(testFileName)
	if err != nil {
		return ""
	}
	return absPath
}

func ginkgo_v1_load(projPath, pkgBin string, ginkgoVersion int) ([]*ginkgoTestcase.TestCase, error) {
	var caseList []*ginkgoTestcase.TestCase
	cmdline := pkgBin + " --ginkgo.v --ginkgo.dryRun --ginkgo.noColor"
	output, _, err := ginkgoUtil.RunCommandWithOutput(cmdline, projPath)
	if err != nil {
		log.Printf("Ginkgo dry run command exit code: %v", err)
		return nil, err
	}
	testcaseList, err := ginkgoResult.ParseCaseByReg(projPath, output, ginkgoVersion)
	if err != nil {
		log.Printf("find testcase in log error: %v", err)
	}
	caseList = append(caseList, testcaseList...)
	return caseList, nil
}

func ginkgo_v2_load(projPath, path, pkgBin string, ginkgoVersion int) ([]*ginkgoTestcase.TestCase, error) {
	var caseList []*ginkgoTestcase.TestCase
	reportJson := filepath.Join(path, "report.json")
	if exists, err := ginkgoUtil.FileExists(reportJson); err != nil || exists {
		if err != nil {
			log.Printf("Check report.json file exists failed: %v", err)
		}
		err := os.Remove(reportJson)
		if err != nil {
			log.Printf("Remove report.json file failed: %v", err)
		}
	}
	cmdline := pkgBin + fmt.Sprintf(" --ginkgo.v --ginkgo.dry-run --ginkgo.no-color --ginkgo.json-report=%s ", reportJson)
	log.Printf("dry run cmd: %s", cmdline)
	output, _, err := ginkgoUtil.RunCommandWithOutput(cmdline, projPath)
	if err != nil {
		log.Printf("Ginkgo v2 dry run command exit code: %v", err)
		return nil, err
	}
	if exists, err := ginkgoUtil.FileExists(reportJson); err != nil || !exists {
		log.Printf("dry run report json file not exists, try to parse cases by stdout")
		testcaseList, errInfo := ginkgoResult.ParseCaseByReg(projPath, output, ginkgoVersion)
		if errInfo != nil {
			log.Printf("find testcase in log error: %v", errInfo)
		}
		caseList = append(caseList, testcaseList...)
		if caseList != nil {
			return caseList, nil
		}

		if err != nil {
			log.Printf("file report json not exist: %v", err)
			return caseList, err
		}
		log.Printf("report json file not exists, Please check log")
		return caseList, fmt.Errorf("dry run report json file not exist")
	}
	log.Printf("Parse json file %s", reportJson)
	resultParser, err := ginkgoResult.NewResultParser(reportJson, projPath, path, false)
	if err != nil {
		return nil, err
	}
	results, err := resultParser.Parse()
	if err != nil {
		log.Printf("Pasre json file failed")
		return caseList, err
	}
	for _, result := range results {
		name := result.Test.Name
		if strings.Contains(name, "?") {
			casePath := strings.Split(name, "?")
			caseInfo := &ginkgoTestcase.TestCase{
				Path:       casePath[0],
				Name:       casePath[1],
				Attributes: map[string]string{},
			}
			caseInfo.Attributes["ginkgoVersion"] = strconv.Itoa(2)
			caseList = append(caseList, caseInfo)
		}
	}
	return caseList, nil
}

func dynamicLoadTestcase(projPath string, selectorPath string) ([]*ginkgoTestcase.TestCase, error) {
	var caseList []*ginkgoTestcase.TestCase
	absSelectorPath := filepath.Join(projPath, selectorPath)
	// pathList, err := findTestSuiteDirs(absSelectorPath)
	// if err != nil {
	// 	log.Printf("find test suite dirs from %s failed: %v", selectorPath, err)
	// 	return nil, err
	// }
	// log.Printf("find test suite dirs from %s: %v", selectorPath, pathList)
	pkgBin := findBinFile(absSelectorPath)
	if pkgBin == "" {
		log.Printf("package bin file not exist, ignore load testcase")
		return caseList, nil
	}
	ginkgoVersion := findGinkgoVersion(absSelectorPath)
	log.Printf("load testcase by bin file %s under ginkgo %d", pkgBin, ginkgoVersion)
	var err error
	var testcaseList []*ginkgoTestcase.TestCase
	if ginkgoVersion == 2 {
		testcaseList, err = ginkgo_v2_load(projPath, selectorPath, pkgBin, ginkgoVersion)
	} else {
		testcaseList, err = ginkgo_v1_load(projPath, pkgBin, ginkgoVersion)
	}
	if err != nil {
		log.Printf("load testcase by bin file %s failed, err: %v", pkgBin, err)
		return nil, err
	}
	for _, testcase := range testcaseList {
		log.Println(testcase.GetSelector())
	}
	caseList = append(caseList, testcaseList...)
	return caseList, nil
}

func getAvailableSuitePath(projPath, rootPath string) ([]string, error) {
	var packageList []string
	err := filepath.Walk(rootPath, func(path string, fi os.FileInfo, e error) error {
		if e != nil {
			return errors.Wrapf(e, "failed to walk %s", path)
		}
		if !fi.IsDir() && strings.HasSuffix(path, "_suite_test.go") {
			packagePath := filepath.Dir(path)
			if packagePath == projPath {
				packagePath = ""
			} else {
				packagePath = packagePath[len(projPath)+1:]
			}
			if !ginkgoUtil.ElementIsInSlice(packagePath, packageList) {
				packageList = append(packageList, packagePath)
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to walk %s", rootPath)
	}
	return packageList, nil
}

func DynamicLoadTestcaseInDir(projPath string, rootPath string) ([]*ginkgoTestcase.TestCase, []*sdkModel.LoadError) {
	var testcaseList []*ginkgoTestcase.TestCase
	var loadErrors []*sdkModel.LoadError
	packageList, err := getAvailableSuitePath(projPath, rootPath)
	if err != nil {
		log.Printf("get available suite path of %s failed: %v", rootPath, err)
		loadErrors = append(loadErrors, &sdkModel.LoadError{
			Name:    rootPath,
			Message: err.Error(),
		})
		return nil, loadErrors
	}
	log.Printf("Available package list: %v, root path: %s", packageList, rootPath)
	for _, packagePath := range packageList {
		log.Printf("Start dynamic load testcase from: %v", packagePath)
		caseList, err := dynamicLoadTestcase(projPath, packagePath)
		if err != nil {
			log.Printf("dynamic load testcase from %s failed, err: %v", packagePath, err)
			loadErrors = append(loadErrors, &sdkModel.LoadError{
				Name:    packagePath,
				Message: err.Error(),
			})
			continue
		}
		// 如果加载出来的用例实际路径与下发的包路径不一致，则表明该用例为共享用例（用例被其他路径下的用例所引用）
		// 这种情况下为避免sdk由于selector中下发的路径与实际加载出来的用例路径不匹配而将用例过滤，需要将用例的路径更改为包路径
		for _, c := range caseList {
			if c.Path != packagePath && !strings.HasPrefix(c.Path, packagePath) {
				log.Printf("Loaded case [path: %s, name: %s] has different path with package: %s, replace case's path to package path", c.Path, c.Name, packagePath)
				c.Path = packagePath
			}
		}
		testcaseList = append(testcaseList, caseList...)
	}
	return testcaseList, loadErrors
}

func DynamicLoadTestcaseInFile(projPath string, filePath string) ([]*ginkgoTestcase.TestCase, []*sdkModel.LoadError) {
	var loadErrors []*sdkModel.LoadError
	selectorPath, err := filepath.Rel(projPath, filePath)
	if err != nil {
		log.Printf("get path %s rel path %s failed: %v", filePath, projPath, err)
		loadErrors = append(loadErrors, &sdkModel.LoadError{
			Name:    selectorPath,
			Message: err.Error(),
		})
		return nil, loadErrors
	}
	log.Printf("Start dynamic load testcase in file %s", selectorPath)
	// 这里处理文件，扫描文件上一层的用例，然后过滤
	parentDir := filepath.Dir(selectorPath)
	testcaseList, err := dynamicLoadTestcase(projPath, parentDir)
	if err != nil {
		log.Printf("dynamic load testcase in %s failed: %v", selectorPath, err)
		loadErrors = append(loadErrors, &sdkModel.LoadError{
			Name:    selectorPath,
			Message: err.Error(),
		})
		return nil, loadErrors
	}
	return testcaseList, nil
}
