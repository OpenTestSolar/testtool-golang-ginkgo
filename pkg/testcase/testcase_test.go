package testcase

import (
	"encoding/json"
	"os"
	"testing"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/stretchr/testify/assert"
)

func createMockJsonFile() (string, error) {
	tmpfile, err := os.CreateTemp("", "testcase*.json")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()
	param := &sdkModel.EntryParam{
		Context:        map[string]string{},
		TaskId:         "task id",
		ProjectPath:    "workspace",
		TestSelectors:  []string{"selector01", "selector02"},
		Collectors:     []string{"selector url"},
		FileReportPath: "filepath",
	}
	content, err := json.Marshal(param)
	if err != nil {
		return "", err
	}
	_, err = tmpfile.Write(content)
	if err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
}

func TestUnmarshalCaseInfo(t *testing.T) {
	path, err := createMockJsonFile()
	assert.NoError(t, err)
	defer os.Remove(path)
	config, err := UnmarshalCaseInfo(path)
	assert.NoError(t, err)
	assert.Equal(t, "task id", config.TaskId)
	assert.Equal(t, "workspace", config.ProjectPath)
	assert.Equal(t, map[string]string{}, config.Context)
	assert.Equal(t, []string{"selector01", "selector02"}, config.TestSelectors)
	assert.Equal(t, []string{"selector url"}, config.Collectors)
	assert.Equal(t, "filepath", config.FileReportPath)
}

func TestParseTestCaseBySelector(t *testing.T) {
	// 测试用例1：包含路径和查询参数的selector
	testCase1, err := ParseTestCaseBySelector("path/to/test?name=testName&attr1=value1")
	assert.NoError(t, err)
	assert.Equal(t, testCase1.Path, "path/to/test")
	assert.Equal(t, testCase1.Name, "testName")
	assert.Equal(t, testCase1.Attributes["attr1"], "value1")
	// 测试用例2：包含路径但没有查询参数的selector
	testCase2, err := ParseTestCaseBySelector("path/to/test")
	assert.NoError(t, err)
	assert.Equal(t, testCase2.Path, "path/to/test")
	assert.Empty(t, testCase2.Name)
	assert.Len(t, testCase2.Attributes, 0)
	// 测试用例3：包含查询参数但没有路径的selector
	testCase3, err := ParseTestCaseBySelector("?name=testName&attr1=value1")
	assert.NoError(t, err)
	assert.NotNil(t, testCase3)
	// 测试用例4：包含特殊字符的selector
	testCase4, err := ParseTestCaseBySelector("path/to/test?name=test%20Name&attr1=value%3D1")
	assert.NoError(t, err)
	assert.Equal(t, testCase4.Path, "path/to/test")
	assert.Equal(t, testCase4.Name, "test Name")
	assert.Equal(t, testCase4.Attributes["attr1"], "value=1")
}

func TestGetSelector(t *testing.T) {
	tc := TestCase{
		Path:       "/path/to/test",
		Name:       "TestName",
		Attributes: map[string]string{"key": "value"},
	}
	selector := tc.GetSelector()
	expected := "/path/to/test?TestName"
	assert.Equal(t, selector, expected)
}

func TestGetSelectorEmptyName(t *testing.T) {
	tc := TestCase{
		Path:       "/path/to/test",
		Name:       "",
		Attributes: map[string]string{"key": "value"},
	}
	selector := tc.GetSelector()
	expected := "/path/to/test"
	assert.Equal(t, selector, expected)
}
