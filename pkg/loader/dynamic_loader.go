package loader

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	ginkgoBuilder "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/builder"
	ginkgoResult "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/result"
	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/testcase"
	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/util"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
)

func findBinFile(absSelectorPath string) string {
	testFileName := absSelectorPath + ".test"
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

func ginkgo_v1_load(projPath, pkgBin string) ([]*ginkgoTestcase.TestCase, error) {
	var caseList []*ginkgoTestcase.TestCase
	var cmdline string
	workDir := strings.TrimSuffix(pkgBin, ".test")
	if pkgBin != "" {
		cmdline = strings.Join([]string{pkgBin, "--ginkgo.v --ginkgo.dryRun --ginkgo.noColor"}, " ")
	} else {
		if _, err := exec.LookPath("ginkgo"); err != nil {
			return nil, errors.Wrapf(err, "there is no test and ginkgo binary")
		}
		cmdline = "ginkgo --v --dry-run --no-color ."
	}
	log.Printf("dry run cmd: %s in dir: %s", cmdline, workDir)
	stdout, stderr, err := ginkgoUtil.RunCommandWithOutput(cmdline, workDir)
	if err != nil {
		message := fmt.Sprintf("dry run command failed, cmdline: %s, err: %v, stdout: %s, stderr: %s", cmdline, err, stdout, stderr)
		log.Println(message)
		return nil, errors.New(message)
	}
	// 如果加载出来测试套中的用例数为空则直接返回
	if strings.Contains(stdout, "Ran 0 of 0 Specs in 0.000 seconds") {
		log.Printf("no testcases found in %s", pkgBin)
		return []*ginkgoTestcase.TestCase{}, nil
	}
	testcaseList, err := ginkgoResult.ParseCaseByReg(projPath, stdout, 1, "")
	if err != nil {
		message := fmt.Sprintf("find testcase from stdout failed, err: %v, stdout: %s, stderr: %s", err, stdout, stderr)
		log.Print(message)
		return nil, errors.New(message)
	}
	if len(testcaseList) == 0 {
		return nil, fmt.Errorf("failed to find testcases from stdout, stdout: %s, stderr: %s", stdout, stderr)
	}
	caseList = append(caseList, testcaseList...)
	return caseList, nil
}

func ginkgo_v2_load(projPath, path, pkgBin string) ([]*ginkgoTestcase.TestCase, error) {
	var caseList []*ginkgoTestcase.TestCase
	var cmdline string
	var workDir string
	if pkgBin != "" {
		cmdline = strings.Join([]string{pkgBin, "--ginkgo.v --ginkgo.dry-run --ginkgo.no-color --ginkgo.json-report=report.json"}, " ")
		workDir = projPath
	} else {
		if _, err := exec.LookPath("ginkgo"); err != nil {
			return nil, errors.Wrapf(err, "there is no test and ginkgo binary")
		}
		cmdline = "ginkgo --v --dry-run --no-color --json-report=report.json ."
		workDir = filepath.Join(projPath, path)
	}
	log.Printf("dry run cmd: %s\nwork directory: %s", cmdline, workDir)
	stdout, stderr, err := ginkgoUtil.RunCommandWithOutput(cmdline, workDir)
	if err != nil {
		return nil, fmt.Errorf("dry run command %s failed, err: %v", cmdline, err)
	}
	reportJson := filepath.Join(workDir, "report.json")
	if exists, err := ginkgoUtil.FileExists(reportJson); err != nil || !exists {
		log.Printf("dry run report json file not exists, try to parse cases by stdout")
		testcaseList, errInfo := ginkgoResult.ParseCaseByReg(projPath, stdout, 2, path)
		if errInfo != nil {
			return nil, errors.Wrapf(err, "find testcase from stdout failed, stdout: %s, stderr: %s", stdout, stderr)
		}
		if len(testcaseList) == 0 {
			return nil, fmt.Errorf("can't find testcases from stdout: %s, stderr: %s", stdout, stderr)
		}
		caseList = append(caseList, testcaseList...)
		return caseList, nil
	}
	log.Printf("Parse json file %s", reportJson)
	resultParser, err := ginkgoResult.NewResultParser(reportJson, projPath, path, false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse ginkgo dry run output json file %s", reportJson)
	}
	results, err := resultParser.Parse()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse ginkgo dry run output json file %s", reportJson)
	}
	for _, result := range results {
		name := result.Test.Name
		if strings.Contains(name, "?") {
			casePath := strings.Split(name, "?")
			caseInfo := &ginkgoTestcase.TestCase{
				Path:       casePath[0],
				Name:       casePath[1],
				Attributes: result.Test.Attributes,
			}
			caseInfo.Attributes["ginkgoVersion"] = strconv.Itoa(2)
			caseInfo.Attributes["path"] = path
			caseList = append(caseList, caseInfo)
		}
	}
	return caseList, nil
}

func dynamicLoadTestcase(projPath string, selectorPath string) ([]*ginkgoTestcase.TestCase, []*sdkModel.LoadError) {
	var caseList []*ginkgoTestcase.TestCase
	absSelectorPath := filepath.Join(projPath, selectorPath)
	pkgBin := findBinFile(absSelectorPath)
	if pkgBin == "" {
		log.Printf("Can't find package bin file %s during loading, try to build it...", pkgBin)
		var err error
		pkgBin, err = ginkgoBuilder.BuildTestPackage(projPath, selectorPath, false)
		if err != nil {
			message := fmt.Sprintf("Build package %s during loading failed, err: %s", selectorPath, err.Error())
			log.Println(message)
			return nil, []*sdkModel.LoadError{
				{
					Name:    selectorPath,
					Message: message,
				},
			}
		}
	}
	ginkgoVersion := ginkgoUtil.FindGinkgoVersion(absSelectorPath)
	log.Printf("load testcase by bin file %s under ginkgo %d", pkgBin, ginkgoVersion)
	var err error
	var testcaseList []*ginkgoTestcase.TestCase
	if ginkgoVersion == 2 {
		testcaseList, err = ginkgo_v2_load(projPath, selectorPath, pkgBin)
	} else {
		testcaseList, err = ginkgo_v1_load(projPath, pkgBin)
	}
	if err != nil {
		message := fmt.Sprintf("load testcase by bin file %s from %s failed, err: %v", pkgBin, selectorPath, err)
		log.Println(message)
		return nil, []*sdkModel.LoadError{
			{
				Name:    selectorPath,
				Message: message,
			},
		}
	}
	log.Printf("load testcase from %s:", selectorPath)
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
				if relPath, err := filepath.Rel(projPath, packagePath); err != nil {
					log.Printf("get rel path failed, basepath %s, targpath: %s, err: %v", projPath, packagePath, err)
					relPath = strings.TrimPrefix(packagePath, projPath)
					relPath = strings.TrimPrefix(relPath, "/")
					packagePath = relPath
				} else {
					packagePath = relPath
				}
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
		caseList, loadError := dynamicLoadTestcase(projPath, packagePath)
		if loadError != nil {
			log.Printf("dynamic load testcase from %s failed, load errors: %v", packagePath, loadError)
			loadErrors = append(loadErrors, loadError...)
			continue
		}
		// 如果加载出来的用例实际路径与下发的包路径不一致，则表明该用例为共享用例（用例被其他路径下的用例所引用）
		// 这种情况下无法确定用例具体对应的文件路径，因此需要将用例文件路径修改为包下的_suite_test.go文件
		for _, c := range caseList {
			if c.Path != packagePath && !strings.HasPrefix(c.Path, packagePath) {
				suiteFileName, err := ginkgoUtil.GetSuiteFileNameInPackage(packagePath)
				if err != nil {
					log.Printf("get suite file name in package %s failed, err: %v", packagePath, err)
					log.Printf("Loaded case [path: %s, name: %s] has different path with package: %s, replace case's path to package path", c.Path, c.Name, packagePath)
					c.Path = packagePath
				} else {
					suitePath := filepath.Join(packagePath, suiteFileName)
					log.Printf("Loaded case [path: %s, name: %s] has different path with package: %s, replace case's path to suite path %s", c.Path, c.Name, packagePath, suitePath)
					c.Path = suitePath
				}
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
	testcaseList, loadErrors := dynamicLoadTestcase(projPath, parentDir)
	for _, c := range testcaseList {
		// 如果加载出来的用例路径不在当前包下，则表明该用例为一个引用用例
		// 低版本ginkgo对于引用用例无法正确解析用例的引用路径，因此需要将用例路径设置为当前下发的加载文件路径
		if !strings.HasPrefix(c.Path, parentDir) {
			c.Path = selectorPath
		}
	}
	if loadErrors != nil {
		log.Printf("dynamic load testcase in %s failed: %v", selectorPath, loadErrors)
		return nil, loadErrors
	}
	return testcaseList, nil
}
