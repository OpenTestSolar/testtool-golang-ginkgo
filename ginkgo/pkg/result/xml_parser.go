package result

import (
	"os"
	"strconv"
	"strings"
	"time"

	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/util"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/antchfx/xmlquery"
)

func ParseXmlResultFile(projPath string, path string) ([]*sdkModel.TestResult, error) {
	var testResults []*sdkModel.TestResult
	buff, err := os.ReadFile(path)
	if err != nil {
		return testResults, err
	}
	doc, err := xmlquery.Parse(strings.NewReader(string(buff)))
	if err != nil {
		return testResults, err
	}
	endTime := time.Now()
	root := xmlquery.FindOne(doc, "//testsuite")
	testcases := root.SelectElements("/testcase")
	for _, testcase := range testcases {
		var name string
		var duration time.Duration
		var message string

		for _, attr := range testcase.Attr {
			if attr.Name.Local == "name" {
				name = attr.Value
			} else if attr.Name.Local == "time" {
				v, err := strconv.ParseFloat(attr.Value, 64)
				if err != nil {
					return testResults, err
				}
				duration = time.Duration(-1000000000 * v)
			}

		}
		startTime := endTime.Add(time.Nanosecond * duration)
		var steps []*sdkModel.TestCaseStep
		var step *sdkModel.TestCaseStep

		resultType := sdkModel.ResultTypeSucceed
		skipped := testcase.SelectElement("/skipped")
		if skipped != nil {
			resultType = sdkModel.ResultTypeIgnored
		} else {
			stdouts := testcase.SelectElements("/system-out")
			for _, stdout := range stdouts {
				text := stdout.InnerText()
				if strings.HasPrefix(text, "STEP:") {
					// add new step
					for _, line := range strings.Split(text, "\n") {
						if strings.HasPrefix(line, "STEP:") {
							step = &sdkModel.TestCaseStep{
								Title:      line[6:],
								StartTime:  startTime,
								EndTime:    endTime,
								ResultType: sdkModel.ResultTypeSucceed,
							}
							steps = append(steps, step)
						} else {
							step.Logs = append(step.Logs, &sdkModel.TestCaseLog{
								Level:   sdkModel.LogLevelInfo,
								Content: line,
							})
						}
					}
				} else {
					if len(steps) == 0 {
						// add default step
						step = &sdkModel.TestCaseStep{
							Title:      "Default",
							StartTime:  startTime,
							EndTime:    endTime,
							ResultType: sdkModel.ResultTypeSucceed,
						}
						steps = append(steps, step)
					}
					step.Logs = append(step.Logs, &sdkModel.TestCaseLog{
						Level:   sdkModel.LogLevelInfo,
						Content: text,
					})
				}

			}
			failures := testcase.SelectElements("/failure")
			for _, failure := range failures {
				if len(steps) == 0 {
					// add default step
					step = &sdkModel.TestCaseStep{
						Title:      "Default",
						StartTime:  startTime,
						EndTime:    endTime,
						ResultType: sdkModel.ResultTypeSucceed,
					}
					steps = append(steps, step)
				}
				step.Logs = append(step.Logs, &sdkModel.TestCaseLog{
					Level:   sdkModel.LogLevelError,
					Content: failure.InnerText(),
				})
				step.ResultType = sdkModel.ResultTypeFailed
				resultType = sdkModel.ResultTypeFailed
				message = ginkgoUtil.ShortenString(failure.InnerText(), 512)
			}
		}
		if len(steps) == 0 {
			// add default step
			step = &sdkModel.TestCaseStep{
				Title:      "Default",
				StartTime:  startTime,
				EndTime:    endTime,
				ResultType: resultType,
			}
			steps = append(steps, step)
		}
		testResult := &sdkModel.TestResult{
			Test: &sdkModel.TestCase{
				Name:       name,
				Attributes: map[string]string{}, //TODO:
			},
			StartTime:  startTime,
			EndTime:    endTime,
			ResultType: resultType,
			Steps:      steps,
			Message:    message,
		}
		testResults = append(testResults, testResult)
	}

	return testResults, nil
}
