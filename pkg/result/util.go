package result

import (
	"fmt"
	"strings"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

func addLabels(specName string, hierarchyLabels [][]string, nodeLabels []string) string {
	// filter duplicate labels
	inLabels := func(labels []string, label string) bool {
		for _, l := range labels {
			if l == label {
				return true
			}
		}
		return false
	}

	var labels []string
	for _, containerLabels := range hierarchyLabels {
		for _, label := range containerLabels {
			if !inLabels(labels, label) {
				labels = append(labels, label)
			}
		}
	}
	for _, label := range nodeLabels {
		if !inLabels(labels, label) {
			labels = append(labels, label)
		}
	}
	if len(labels) != 0 {
		specName += " " + fmt.Sprintf("[%s]", strings.Join(labels, ", "))
	}
	return specName
}

func splitByNewline(s string) []string {
	return strings.Split(s, "\n")
}

func removeProjectPrefix(filePath string, projectPath string) string {
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[len(projectPath)+1:]
	}
	return filePath
}

func getResultType(result string) sdkModel.ResultType {
	if result == "passed" {
		return sdkModel.ResultTypeSucceed
	} else if result == "failed" || result == "panicked" {
		return sdkModel.ResultTypeFailed
	} else if result == "skipped" {
		return sdkModel.ResultTypeIgnored
	} else {
		return sdkModel.ResultTypeUnknown
	}
}
