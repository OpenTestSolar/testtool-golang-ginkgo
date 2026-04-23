package result

import (
	"os"
	"strings"
	"testing"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/stretchr/testify/assert"
)

func TestParseJsonToObj(t *testing.T) {
	parser, err := NewResultParser("./testdata/report.json", "/data/workspace", "suits/demo", "", true)
	assert.NoError(t, err)
	results, err := parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	for _, result := range results {
		if !strings.HasPrefix(result.Test.Name, "test/test_test.go") {
			t.Errorf("incorrect case name: %s", result.Test.Name)
		}
	}
	assert.Equal(t, results[0].Test.Attributes["owner"], "tom")
	assert.Equal(t, results[0].Test.Attributes["description"], "demo test")

	parser, err = NewResultParser("./testdata/report_with_setup.json", "/data/workspace", "suites/demo", "", true)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	parser, err = NewResultParser("./testdata/report_with_setup.json", "/data/workspace", "suites/demo", "", false)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	parser, err = NewResultParser("./testdata/report_with_labels.json", "/data/workspace", "suites/demo", "", true)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	if results[0].Test.Name != "suites/demo/demo_suite_test.go?HierarchyText01 HierarchyText02   Text [label01, label02, label11, node-label01]" {
		t.Errorf("incorrect case name: %s", results[0].Test.Name)
	}

	parser, err = NewResultParser("./testdata/report_with_failed_setup.json", "/data/workspace", "suites/demo", "", true)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	parser, err = NewResultParser("./testdata/report_with_panic.json", "/data/workspace", "suites/demo", "", true)
	assert.NoError(t, err)
	results, err = parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	panicSuite, err := parser.GetPanicSuite()
	assert.NoError(t, err)
	assert.NotNil(t, panicSuite)
}

func Test_parseCaseByReg(t *testing.T) {
	byteValue, err := os.ReadFile("./testdata/dry_run_output.txt")
	assert.NoError(t, err)
	cases, err := ParseCaseByReg("/data/workspace", string(byteValue), 2, "")
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
}

func TestGetStepsByOutputLines_WithSpecEvents(t *testing.T) {
	parser, err := NewResultParser(
		"./testdata/report_with_output_step.json",
		"/data/workspace", "testcase/media", "", true,
	)
	assert.NoError(t, err)

	results, err := parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	steps := results[0].Steps
	// 过滤出由 getStepsByOutputFromSpecEvents 生成的步骤（Title 包含 "步骤"）
	var outputSteps []*sdkModel.TestCaseStep
	for _, step := range steps {
		if strings.Contains(step.Title, "步骤") {
			outputSteps = append(outputSteps, step)
		}
	}
	assert.Len(t, outputSteps, 4, "should parse 4 steps from output")

	// 步骤1：创建对话获取 chat_id
	assert.Contains(t, outputSteps[0].Title, "步骤1：创建对话获取 chat_id")
	assert.NotEmpty(t, outputSteps[0].Logs, "step1 should have logs")
	assert.Contains(t, outputSteps[0].Logs[0].Content, "步骤1")

	// 步骤2：调用 ASR start 接口
	assert.Contains(t, outputSteps[1].Title, "步骤2：调用 ASR start 接口")
	assert.NotEmpty(t, outputSteps[1].Logs, "step2 should have logs")
	assert.Contains(t, outputSteps[1].Logs[0].Content, "步骤2")

	// 步骤3：验证 ASR 返回码
	assert.Contains(t, outputSteps[2].Title, "步骤3：验证 ASR 返回码")
	assert.NotEmpty(t, outputSteps[2].Logs, "step3 should have logs")

	// 步骤4：校验 asrText 相似度
	assert.Contains(t, outputSteps[3].Title, "步骤4：校验 asrText 相似度")
	assert.NotEmpty(t, outputSteps[3].Logs, "step4 should have logs")
	// 验证步骤4的日志中包含相似度校验相关信息
	var hasSimilarityLog bool
	for _, log := range outputSteps[3].Logs {
		if strings.Contains(log.Content, "similarity") {
			hasSimilarityLog = true
			break
		}
	}
	assert.True(t, hasSimilarityLog, "step4 logs should contain similarity info")
}
