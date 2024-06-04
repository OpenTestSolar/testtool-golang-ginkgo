package cmdline

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func removeTestCaseLabels(tcNames []string) []string {
	var replacedNames []string
	pattern := `\s\[([^\[\]]*)\]$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Printf("Failed to replace labels, please check if the regular expression is correct: %v", err)
		return tcNames
	}
	for _, name := range tcNames {
		replacedName := re.ReplaceAllString(name, "")
		replacedNames = append(replacedNames, replacedName)
	}
	return replacedNames
}

func GenTestCaseFocusName(tcNames []string) string {
	// 传入的用例名中可能包含用例标签，ginkgo focus参数中只能识别用例名，因此需要去除
	if remove, _ := strconv.ParseBool(os.Getenv("TESTSOLAR_TTP_WITHLABELS")); remove {
		tcNames = removeTestCaseLabels(tcNames)
	}
	name := strings.Join(tcNames, "|")
	name = strings.Replace(name, "/", " ", -1)
	name = strings.Replace(name, "[", "\\[", -1)
	name = strings.Replace(name, "]", "\\]", -1)
	name = strings.Replace(name, "(", "\\(", -1)
	name = strings.Replace(name, ")", "\\)", -1)
	return name
}

func ExtractPackPathFromBinFile(pkgBin, projPath string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(pkgBin, projPath), ".test"), "/")
}
