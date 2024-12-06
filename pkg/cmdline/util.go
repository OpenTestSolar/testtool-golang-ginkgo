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
	if ok, err := strconv.ParseBool(os.Getenv("TESTSOLAR_TTP_WITHLABELS")); err == nil && ok {
		tcNames = removeTestCaseLabels(tcNames)
	}
	// ginkgo中focus参数需要输入一个正则表达式，因此需要将用例名中和正则表达式相关的字符进行转义
	var escapedNames []string
	for _, name := range tcNames {
		// 将双引号转义
		escapedNames = append(escapedNames, strings.Replace(regexp.QuoteMeta(name), "\"", "\\\"", -1))
	}
	name := strings.Join(escapedNames, "|")
	return name
}

func ExtractPackPathFromBinFile(pkgBin, projPath string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(pkgBin, projPath), ".test"), "/")
}
