package discover

import (
	"log"
	"os"

	ginkgoLoader "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/loader"
	ginkgoSelector "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/selector"
	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/testcase"
	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/util"

	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
	pkgErrors "github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type DiscoverOptions struct {
	discoverPath string
}

// NewDiscoverOptions NewBuildOptions new build options with default value
func NewDiscoverOptions() *DiscoverOptions {
	return &DiscoverOptions{}
}

// NewCmdDiscover NewCmdBuild create a build command
func NewCmdDiscover() *cobra.Command {
	o := NewDiscoverOptions()
	cmd := cobra.Command{
		Use:   "discover",
		Short: "Discover testcases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunDiscover(cmd)
		},
	}
	cmd.Flags().StringVarP(&o.discoverPath, "path", "p", "", "Path of testcase info")
	_ = cmd.MarkFlagRequired("path")
	return &cmd
}

func ParseTestSelectors(testSelector []string) []*ginkgoSelector.TestSelector {
	if len(testSelector) == 0 {
		testSelector = []string{"."}
	}
	var targetSelectors []*ginkgoSelector.TestSelector
	for _, selector := range testSelector {
		testSelector, err := ginkgoSelector.NewTestSelector(selector)
		if err != nil {
			log.Printf("Ignore invalid test selector: %s", selector)
			continue
		}
		if !testSelector.IsExclude() {
			targetSelectors = append(targetSelectors, testSelector)
		}
	}
	return targetSelectors
}

func ReportTestcases(testcases []*ginkgoTestcase.TestCase, loadErrors []*sdkModel.LoadError, reporter api.Reporter) error {
	var tests []*sdkModel.TestCase
	for _, testcase := range testcases {
		tests = append(tests, &sdkModel.TestCase{
			Name:       testcase.GetSelector(),
			Attributes: testcase.Attributes,
		})
	}
	err := reporter.ReportLoadResult(&sdkModel.LoadResult{
		Tests:      tests,
		LoadErrors: loadErrors,
	})
	if err != nil {
		return errors.Wrap(err, "failed to report load result")
	}
	return nil
}

func containTestCase(testcases []*ginkgoTestcase.TestCase, testcase *ginkgoTestcase.TestCase) bool {
	for _, t := range testcases {
		if t.Path == testcase.Path && t.Name == testcase.Name {
			return true
		}
	}
	return false
}

func LoadTestcases(projPath string, targetSelectors []*ginkgoSelector.TestSelector) ([]*ginkgoTestcase.TestCase, []*sdkModel.LoadError) {
	var testcases []*ginkgoTestcase.TestCase
	var loadErrors []*sdkModel.LoadError
	loadedSelectorPath := make(map[string]struct{})
	for _, testSelector := range targetSelectors {
		// skip the path that has been loaded
		if _, ok := loadedSelectorPath[testSelector.Path]; ok {
			continue
		}
		loadedSelectorPath[testSelector.Path] = struct{}{}
		loadedTestcases, lErrors := ginkgoLoader.LoadTestCase(projPath, testSelector.Path)
		for _, loadedTestcase := range loadedTestcases {
			if !containTestCase(testcases, loadedTestcase) {
				testcases = append(testcases, loadedTestcase)
			} else {
				log.Printf("[Plugin] repeat selector: %s", loadedTestcase.GetSelector())
			}
		}
		loadErrors = append(loadErrors, lErrors...)
	}
	return testcases, loadErrors
}

func (o *DiscoverOptions) RunDiscover(cmd *cobra.Command) error {
	config, err := ginkgoTestcase.UnmarshalCaseInfo(o.discoverPath)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to unmarshal case info")
	}
	targetSelectors := ParseTestSelectors(config.TestSelectors)
	log.Printf("load testcases from selectors: %s", targetSelectors)
	projPath := ginkgoUtil.GetWorkspace(config.ProjectPath)
	_, err = os.Stat(projPath)
	if err != nil {
		return pkgErrors.Wrapf(err, "stat project path %s failed", projPath)
	}
	testcases, loadErrors := LoadTestcases(projPath, targetSelectors)
	reporter, err := sdkClient.NewReporterClient(config.FileReportPath)
	if err != nil {
		return errors.Wrapf(err, "failed to create reporter")
	}
	err = ReportTestcases(testcases, loadErrors, reporter)
	if err != nil {
		return errors.Wrapf(err, "failed to report testcases")
	}
	return nil
}
