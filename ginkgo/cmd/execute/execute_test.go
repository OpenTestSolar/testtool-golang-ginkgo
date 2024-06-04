package execute

import (
	"path/filepath"
	"testing"
	"time"

	"ginkgo/pkg/testcase"

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
	testSelectors := []string{"path?name=test%20name&attr1=value%3D1"}
	testcases, err := parseTestcases(testSelectors)
	assert.NoError(t, err)
	assert.Len(t, testcases, 1)
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
	err := reportTestResults(testResults)
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
