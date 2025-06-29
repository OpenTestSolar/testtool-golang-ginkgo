package result

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

type Parser interface {
	Parse() ([]*sdkModel.TestResult, error)
}

type ResultParser struct {
	suites   []*Suite
	projPath string
	packPath string
	run      bool
	filePath string
}

func NewResultParser(jsonFile string, projectPath string, packPath string, filePath string, run bool) (*ResultParser, error) {
	byteValue, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("read result file failed, err: %s", err.Error())
	}
	fmt.Printf("output json file: \n%s\n", byteValue)
	var suites []*Suite
	err = json.Unmarshal(byteValue, &suites)
	if err != nil {
		return nil, fmt.Errorf("unmarshal result failed, err: %s", err.Error())
	}
	return &ResultParser{
		suites:   suites,
		projPath: projectPath,
		packPath: packPath,
		run:      run,
		filePath: filePath,
	}, nil
}

func (p *ResultParser) getSuite() *Suite {
	if p.suites != nil {
		return p.suites[0]
	}
	return nil
}

func (p *ResultParser) validate() error {
	if len(p.suites) != 1 {
		log.Printf("uncorrect suite num")
		return fmt.Errorf("uncorrect suite num: %d", len(p.suites))
	}
	return nil
}

func (p *ResultParser) GetPanicSuite() (*Spec, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	suite := p.getSuite()
	if suite == nil {
		fmt.Print("no valid suite in results")
		return nil, fmt.Errorf("no valid suite in results")
	}
	return suite.getBefSuiteFailedSpec(), nil
}

func (p *ResultParser) Parse() ([]*sdkModel.TestResult, error) {
	/*
		解析用例执行结果
		前置工作:
		1. 验证是否有且仅有一个测试套;
		2. 获取测试套;
		3. 获取用例分隔符.
		如果测试套的SynchronizedBeforeSuite失败则直接解析返回，否则解析用例:
		1. 判断用例结果是否有效;
		2. 生成用例名;
		3. 生成用例步骤信息;
		4. 添加至结果集中.
	*/
	var testResults []*sdkModel.TestResult
	if err := p.validate(); err != nil {
		return []*sdkModel.TestResult{}, nil
	}
	suite := p.getSuite()
	if suite == nil {
		fmt.Print("no valid suite in results")
		return []*sdkModel.TestResult{}, nil
	}
	for _, spec := range suite.SpecReports {
		if !spec.isValidResultType() {
			continue
		}
		// 若用例为BeforeSuite｜AfterSuite｜ReportAfterSuite，则需要在执行阶段(p.run==false)且状态不为passed时才上报
		if spec.IsSetUpStage() && (!p.run || spec.State == "passed") {
			continue
		}
		containerName, leafName := spec.getContainerAndLeafName()
		if containerName == "" && leafName == "" {
			continue
		}
		specName := strings.Join([]string{containerName, leafName}, " ")
		var nameList string
		if marshalNameList, err := json.Marshal([]string{containerName, leafName}); err == nil {
			nameList = string(marshalNameList)
		}
		var labelList string
		var owner string
		var description string
		if labels := getLabels(spec.ContainerHierarchyLabels, spec.LeafNodeLabels); len(labels) > 0 {
			if marshalLabelList, err := json.Marshal(labels); err == nil {
				labelList = string(marshalLabelList)
			}
			for _, label := range labels {
				if strings.HasPrefix(label, "owner:") {
					owner = strings.TrimSpace(strings.TrimPrefix(label, "owner:"))
				}
				if strings.HasPrefix(label, "description:") {
					description = strings.TrimSpace(strings.TrimPrefix(label, "description:"))
				}
			}
		}
		steps := spec.GenerateSteps()
		var name string
		// 如果已经传入文件路径则直接使用文件路径作为上报用例结果的路径
		if p.filePath != "" && strings.HasSuffix(p.filePath, ".go") {
			name = p.filePath + "?" + specName
		} else {
			name = spec.outputTestName(p.projPath, p.packPath, specName)
		}
		testResults = append(testResults, &sdkModel.TestResult{
			Test: &sdkModel.TestCase{
				Name: name,
				Attributes: map[string]string{
					"nameList":    nameList,
					"label":       labelList,
					"tags":        labelList,
					"owner":       owner,
					"description": description,
					"name":        specName,
				},
			},
			StartTime:  spec.StartTime,
			EndTime:    spec.EndTime,
			ResultType: getResultType(spec.State),
			Message:    spec.Failure.getMessage(),
			Steps:      steps,
		})
	}
	return testResults, nil
}
