package execute

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/testcase"

	sdkApi "github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewExecuteOptions(t *testing.T) {
	o := NewExecuteOptions()
	assert.NotNil(t, o)
}

func TestNewCmdExecute(t *testing.T) {
	cmd := NewCmdExecute()
	assert.NotNil(t, cmd)
}

func TestParseTestcases(t *testing.T) {
	testSelectors := []string{"path?name=test%20name&attr1=value%3D1", "path?name=test%20name&attr1=value%"}
	testcases, parseFailedResults, err := parseTestcases(testSelectors)
	assert.NoError(t, err)
	assert.Len(t, testcases, 1)
	assert.Len(t, parseFailedResults, 1)
}

type MockReporterClient struct{}

func (m *MockReporterClient) ReportLoadResult(loadResult *sdkModel.LoadResult) error {
	return nil
}
func (m *MockReporterClient) ReportCaseResult(caseResult *sdkModel.TestResult) error {
	return nil
}
func (m *MockReporterClient) Close() error {
	return nil
}

func TestReportTestResults(t *testing.T) {
	NewReporterClientMock := gomonkey.ApplyFunc(sdkClient.NewReporterClient, func() (sdkApi.Reporter, error) {
		return &MockReporterClient{}, nil
	})
	defer NewReporterClientMock.Reset()
	testResults := []*sdkModel.TestResult{
		{
			Test:       &sdkModel.TestCase{},
			StartTime:  time.Now(),
			ResultType: sdkModel.ResultTypeSucceed,
			Message:    "",
			EndTime:    time.Now(),
			Steps:      []*sdkModel.TestCaseStep{},
		},
	}
	err := reportTestResults(testResults, &MockReporterClient{})
	if err != nil {
		t.Errorf("reportTestResults failed: %v", err)
	}
}

func TestGroupTestCasesByPathAndName(t *testing.T) {
	projPath := "../../testdata"
	testcases := []*testcase.TestCase{
		{
			Path:       "demo",
			Name:       "Testcase cont demo test",
			Attributes: map[string]string{},
		},
		{
			Path:       "demo",
			Name:       "Testcase cont demo test2",
			Attributes: map[string]string{},
		},
		{
			Path:       "demo",
			Name:       "Testcase [cont3] [demo test3]",
			Attributes: map[string]string{},
		},
	}
	result, err := groupTestCasesByPathAndName(projPath, testcases)
	assert.NoError(t, err)
	assert.Equal(t, result, map[string]map[string][]*testcase.TestCase{
		"demo": {
			"": {
				{
					Path:       "demo",
					Name:       "Testcase cont demo test",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "Testcase cont demo test2",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "Testcase [cont3] [demo test3]",
					Attributes: map[string]string{},
				},
			},
		},
	})
}

func TestExecuteTestcases(t *testing.T) {
	projPath, err := filepath.Abs("../../testdata")
	assert.NoError(t, err)
	packages := map[string]map[string][]*testcase.TestCase{
		"demo": {
			"": {
				{
					Path:       "demo",
					Name:       "Testcase cont demo test",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "Testcase cont demo test2",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "Testcase [cont3] [demo test3]",
					Attributes: map[string]string{},
				},
			},
		},
	}
	results, err := executeTestcases(projPath, packages)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
}

func Test_discoverExecutableTestcases(t *testing.T) {
	projPath, err := filepath.Abs("../../testdata")
	assert.NoError(t, err)
	curWd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(projPath)
	assert.NoError(t, err)
	defer os.Chdir(curWd) //nolint:all
	// 验证可以基于指定目录路径找到路径下对应的所有包含测试用例的子目录
	testcases := []*testcase.TestCase{
		{
			Path: "demo",
			Name: "",
		},
	}
	execTestcases, err := discoverExecutableTestcases(testcases)
	assert.NoError(t, err)
	assert.Len(t, execTestcases, 4)
	// 验证如果传入的是文件路径则直接返回
	testcases = []*testcase.TestCase{
		{
			Path: "demo/demo_test.go",
			Name: "",
		},
		{
			Path: "demo/book/book_test.go",
			Name: "",
		},
	}
	execTestcases, err = discoverExecutableTestcases(testcases)
	assert.NoError(t, err)
	assert.Len(t, execTestcases, 2)
	// 验证如果传入的已经是子目录则不会返回额外用例
	testcases = []*testcase.TestCase{
		{
			Path: "demo/book",
			Name: "",
		},
	}
	execTestcases, err = discoverExecutableTestcases(testcases)
	assert.NoError(t, err)
	assert.Len(t, execTestcases, 1)
}

func Test_findCompileBinary(t *testing.T) {
	result := findCompileBinary("path")
	assert.Equal(t, "", result)
	result = findCompileBinary("/path/to/file")
	assert.Equal(t, "", result)
	result = findCompileBinary("path/to/file")
	assert.Equal(t, "", result)
	result = findCompileBinary("path\\to\\file")
	assert.Equal(t, "", result)
	result = findCompileBinary("")
	assert.Equal(t, "", result)
	result = findCompileBinary("/")
	assert.Equal(t, "", result)
	result = findCompileBinary("//")
	assert.Equal(t, "", result)
	result = findCompileBinary("\\")
	assert.Equal(t, "", result)
	// 测试存在二进制文件场景
	projPath, err := filepath.Abs("../../testdata")
	assert.NoError(t, err)
	binFile := filepath.Join(projPath, "demo01.test")
	_, err = os.Create(binFile)
	assert.NoError(t, err)
	result = findCompileBinary(filepath.Join(projPath, "demo01"))
	assert.Equal(t, binFile, result)
	result = findCompileBinary(filepath.Join(projPath, "demo01", "test"))
	assert.Equal(t, binFile, result)
	// 测试存在不与目录同名的二进制文件时
	err = os.Remove(binFile)
	assert.NoError(t, err)
	otherBinFile := filepath.Join(projPath, "demo01_xxx.test")
	_, err = os.Create(otherBinFile)
	assert.NoError(t, err)
	result = findCompileBinary(filepath.Join(projPath, "demo01"))
	assert.NotEqual(t, binFile, result)
	defer os.Remove(binFile)
}
