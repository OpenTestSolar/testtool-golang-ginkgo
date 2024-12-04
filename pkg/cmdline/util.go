package cmdline

import (
	"regexp"
	"strings"
)

func GenTestCaseFocusName(tcNames []string) string {
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
