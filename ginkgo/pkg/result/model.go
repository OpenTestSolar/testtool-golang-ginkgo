package result

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

type TestStepInfo struct {
	Text     string
	Duration int32
}

func newTestStepInfo(rawStr string) (*TestStepInfo, error) {
	var stepInfo TestStepInfo
	if err := json.Unmarshal([]byte(rawStr), &stepInfo); err != nil {
		return nil, fmt.Errorf("Unmarshal step json failed: %s, err: %s", rawStr, err.Error())
	}
	return &stepInfo, nil
}

type LeafNodeLocation struct {
	FileName string
}

type Value struct {
	AsJSON string
}

type ReportEntry struct {
	Name  string
	Time  time.Time
	Value Value
}

func (e *ReportEntry) isValidEntry() bool {
	if e.Name != "By Step" {
		return false
	}
	return true
}

type FailureLocation struct {
	FullStackTrace string
}

type FailureNodeLocation struct {
	FileName   string
	LineNumber int64
}

type Failure struct {
	Message             string
	Location            *FailureLocation
	ForwardedPanic      string
	FailureNodeLocation *FailureNodeLocation
}

func (f *Failure) getMessage() string {
	if f == nil {
		return ""
	}
	return f.Message
}

type Spec struct {
	ContainerHierarchyTexts    []string
	LeafNodeType               string
	LeafNodeText               string
	ContainerHierarchyLabels   [][]string
	LeafNodeLabels             []string
	StartTime                  time.Time
	EndTime                    time.Time
	LeafNodeLocation           *LeafNodeLocation
	RunTime                    time.Duration
	State                      string
	CapturedStdOutErr          string
	CapturedGinkgoWriterOutput string
	ReportEntries              []*ReportEntry
	Failure                    *Failure
	ParallelProcess            int64
}

func (s *Spec) genarateSpecName(splitor string) string {
	var specName string
	if s.ContainerHierarchyTexts == nil {
		// 需要上报beforeSuite和afterSuite的情况
		if s.LeafNodeType != "BeforeSuite" && s.LeafNodeType != "AfterSuite" {
			return ""
		}
		if _, err := strconv.ParseBool(os.Getenv("TESTSOLAR_TTP_SHOWSETUP")); err != nil {
			return ""
		}
		specName = s.LeafNodeType
	} else {
		for _, name := range s.ContainerHierarchyTexts {
			// ignore empty name
			if strings.TrimSpace(name) == "" {
				continue
			}
			if specName != "" && strings.TrimSpace(specName) != "" {
				specName += splitor
			}
			specName += name
		}
		specName += "/" + s.LeafNodeText
		specName = addLabels(specName, s.ContainerHierarchyLabels, s.LeafNodeLabels)
	}
	return specName
}

func (s *Spec) getStepsByOutputLines(output string) []*sdkModel.TestCaseStep {
	outputs := splitByNewline(output)
	var steps []*sdkModel.TestCaseStep
	var lineIndex int
	for _, entry := range s.ReportEntries {
		if !entry.isValidEntry() {
			continue
		}
		stepInfo, err := newTestStepInfo(entry.Value.AsJSON)
		if err != nil {
			log.Printf("Unmarshal step json failed: %s", entry.Value.AsJSON)
			continue
		}
		var logs []*sdkModel.TestCaseLog
		isCurrStep := false
		for lineIndex < len(outputs) {
			if strings.HasPrefix(outputs[lineIndex], "STEP: ") {
				if strings.HasPrefix(outputs[lineIndex], fmt.Sprintf("STEP: %s", stepInfo.Text)) {
					isCurrStep = true
				} else {
					isCurrStep = false
					break
				}
			}
			if isCurrStep && outputs[lineIndex] != "" {
				logs = append(logs, &sdkModel.TestCaseLog{
					Time:    entry.Time, //FIXME
					Level:   sdkModel.LogLevelInfo,
					Content: outputs[lineIndex],
				})
			}
			lineIndex++
		}
		steps = append(steps, &sdkModel.TestCaseStep{
			StartTime: entry.Time,
			EndTime:   entry.Time.Add(time.Second * time.Duration(stepInfo.Duration)),
			Title:     stepInfo.Text,
			Logs:      logs,
		})
	}
	return steps
}

func (s *Spec) generateDefaultStep(stderr string, stdout string) *sdkModel.TestCaseStep {
	stderrs := splitByNewline(stderr)
	stdouts := splitByNewline(stdout)
	logs := make([]*sdkModel.TestCaseLog, 0)
	if len(stdouts) != 0 {
		for _, line := range stdouts {
			if line != "" {
				logs = append(logs, &sdkModel.TestCaseLog{
					Level:   sdkModel.LogLevelInfo,
					Content: line,
				})
			}
		}
	}
	if len(stderrs) != 0 {
		for _, line := range stderrs {
			if line != "" {
				level := sdkModel.LogLevelInfo
				if strings.HasPrefix(line, "E") {
					level = sdkModel.LogLevelError
				}
				logs = append(logs, &sdkModel.TestCaseLog{
					Level:   level,
					Content: line,
				})
			}
		}
	}
	return &sdkModel.TestCaseStep{
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Title:     "stdout/stderr",
		Logs:      logs,
	}
}

func (s *Spec) isValidResultType() bool {
	if s.State == "" || s.State == "skipped" || s.State == "pending" {
		return false
	}
	return true
}

func (s *Spec) generateFailureStep() *sdkModel.TestCaseStep {
	var step sdkModel.TestCaseStep
	if s.Failure != nil {
		step.ResultType = sdkModel.ResultTypeFailed
		step.Title = s.Failure.Message
		step.StartTime = s.StartTime
		step.EndTime = s.EndTime
		step.Logs = append(step.Logs, &sdkModel.TestCaseLog{
			Time:    s.StartTime,
			Level:   sdkModel.LogLevelError,
			Content: s.Failure.Location.FullStackTrace,
		})
		if s.Failure.ForwardedPanic != "" {
			content := s.Failure.ForwardedPanic
			if s.Failure.FailureNodeLocation != nil {
				content += fmt.Sprintf("\n%s:%d", s.Failure.FailureNodeLocation.FileName, s.Failure.FailureNodeLocation.LineNumber)
			}
			step.Logs = append(step.Logs, &sdkModel.TestCaseLog{
				Time:    s.StartTime,
				Level:   sdkModel.LogLevelError,
				Content: content,
			})
		}
	}
	return &step
}

func (s *Spec) outputTestName(projectPath, packPath, specName string) string {
	casePath := removeProjectPrefix(s.LeafNodeLocation.FileName, projectPath)
	packPath = removeProjectPrefix(packPath, projectPath)
	// 如果解析的用例路径与当前执行用例的包路径不一致，则表明该用例为共享用例，需要在上报结果时将路径替换为包路径
	if !strings.HasPrefix(casePath, packPath) {
		return packPath + "?" + specName
	} else {
		return casePath + "?" + specName
	}
}

func (s *Spec) GenerateSteps() []*sdkModel.TestCaseStep {
	var steps []*sdkModel.TestCaseStep

	// 优先展示失败信息
	if failStep := s.generateFailureStep(); failStep != nil {
		steps = append(steps, failStep)
	}

	if outputSteps := s.getStepsByOutputLines(s.CapturedGinkgoWriterOutput); outputSteps != nil {
		steps = append(steps, outputSteps...)
	}

	// 最后展示默认的标准输出流以及错误输出流
	if defaultStep := s.generateDefaultStep(s.CapturedStdOutErr, s.CapturedGinkgoWriterOutput); defaultStep != nil {
		steps = append(steps, defaultStep)
	}
	return steps
}

type Suite struct {
	SpecReports []*Spec
}

func (s *Suite) getSyncBefSuiteFailedSpec() *Spec {
	for _, spec := range s.SpecReports {
		if spec.LeafNodeType == "SynchronizedBeforeSuite" && spec.ParallelProcess == 1 && spec.State != "passed" {
			return spec
		}
	}
	return nil
}
