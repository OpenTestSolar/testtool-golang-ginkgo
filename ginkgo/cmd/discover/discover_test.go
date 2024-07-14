package discover

import (
	"path/filepath"
	"testing"

	"github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/loader"
	"github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/selector"
	"github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/testcase"
	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/testcase"

	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkApi "github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewDiscoverOptions(t *testing.T) {
	o := NewDiscoverOptions()
	assert.NotNil(t, o)
}

func TestNewCmdDiscover(t *testing.T) {
	cmd := NewCmdDiscover()
	assert.NotNil(t, cmd)
}

func TestParseTestSelectors(t *testing.T) {
	testSelectors := []string{"path?name=test%20name&attr1=value%3D1"}
	selectors := parseTestSelectors(testSelectors)
	assert.Len(t, selectors, 1)
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

func TestReportTestcases(t *testing.T) {
	NewReporterClientMock := gomonkey.ApplyFunc(sdkClient.NewReporterClient, func() (sdkApi.Reporter, error) {
		return &MockReporterClient{}, nil
	})
	defer NewReporterClientMock.Reset()
	testcases := []*testcase.TestCase{
		{
			Path:       "",
			Name:       "",
			Attributes: map[string]string{},
		},
	}
	loadErrors := []*sdkModel.LoadError{
		{
			Name:    "",
			Message: "",
		},
	}
	err := reportTestcases(testcases, loadErrors, &MockReporterClient{})
	assert.NoError(t, err)
}

func TestLoadTestcases(t *testing.T) {
	LoadTestCaseMock := gomonkey.ApplyFunc(loader.LoadTestCase, func(projPath string, selectorPath string) ([]*testcase.TestCase, error) {
		return []*testcase.TestCase{
			{
				Path:       "path/to/test",
				Name:       "test01",
				Attributes: map[string]string{},
			},
		}, nil
	})
	defer LoadTestCaseMock.Reset()
	testSelectors := []*selector.TestSelector{
		{
			Value:      "",
			Path:       "path/to/test",
			Name:       "test01",
			Attributes: map[string]string{},
		},
		{
			Value:      "",
			Path:       "path/to/test",
			Name:       "test02",
			Attributes: map[string]string{},
		},
	}
	projPath, err := filepath.Abs("../../testdata")
	assert.NoError(t, err)
	testcases, loadErrors := loadTestcases(projPath, testSelectors)
	assert.Len(t, testcases, 1)
	assert.Len(t, loadErrors, 0)
}

func TestRunDiscover(t *testing.T) {
	reportTestcasesMock := gomonkey.ApplyFunc(reportTestcases, func(testcases []*ginkgoTestcase.TestCase, loadErrors []*sdkModel.LoadError, reporter api.Reporter) error {
		return nil
	})
	defer reportTestcasesMock.Reset()
	projPath, err := filepath.Abs("../../testdata")
	assert.NoError(t, err)
	UnmarshalCaseInfoMock := gomonkey.ApplyFunc(ginkgoTestcase.UnmarshalCaseInfo, func(path string) (*sdkModel.EntryParam, error) {
		return &sdkModel.EntryParam{
			TestSelectors: []string{
				"path/to/test?test01",
			},
			ProjectPath:    projPath,
			FileReportPath: projPath,
		}, nil
	})
	defer UnmarshalCaseInfoMock.Reset()
	LoadTestCaseMock := gomonkey.ApplyFunc(loader.LoadTestCase, func(projPath string, selectorPath string) ([]*testcase.TestCase, error) {
		return []*testcase.TestCase{
			{
				Path:       "path/to/test",
				Name:       "test01",
				Attributes: map[string]string{},
			},
		}, nil
	})
	defer LoadTestCaseMock.Reset()
	o := NewDiscoverOptions()
	cmd := NewCmdDiscover()
	err = o.RunDiscover(cmd)
	assert.NoError(t, err)
}
