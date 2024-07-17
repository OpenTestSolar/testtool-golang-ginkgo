package result

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/testcase"
)

func ParseCaseByReg(proj string, output string, ginkgoVersion int, path string) ([]*ginkgoTestcase.TestCase, error) {
	var caseList []*ginkgoTestcase.TestCase
	regexPattern := `(?s).*?Will run.*?specs(.*?)Ran.*?Specs in.*?seconds`
	re := regexp.MustCompile(regexPattern)
	extractedText := re.FindStringSubmatch(output)
	if len(extractedText) > 1 {
		extractedContent := extractedText[1]
		sections := strings.Split(extractedContent, "------------------------------")
		for _, section := range sections {
			if strings.TrimSpace(section) == "" {
				continue
			} else if strings.Contains(section, "BeforeSuite") {
				continue
			} else if strings.Contains(section, "AfterSuite") {
				continue
			} else if strings.Contains(section, "PASSED") {
				continue
			} else if !strings.Contains(section, proj) {
				continue
			} else if strings.Contains(section, "[PENDING]") {
				continue
			}
			// 按换行符切割 section
			lines := strings.Split(section, "\n")
			var path string
			var nameList []string
			pathRegex := regexp.MustCompile(`.*?(.*?):\d+`)

			for _, line := range lines {
				// 检查路径正则表达式是否匹配
				line = strings.TrimSpace(line)
				pathMatch := pathRegex.FindStringSubmatch(line)
				if len(pathMatch) > 1 {
					// 获取路径并存入 path 变量
					path = pathMatch[1]
				} else if strings.Contains(line, "•") {
					continue
				} else if strings.TrimSpace(line) == "" {
					continue
				} else {
					nameList = append(nameList, line)
				}
			}
			casePath := strings.Split(strings.TrimSpace(path), proj)[1]
			casePath = casePath[1:]
			// TODO: 后续需要将分割符统一切换为空格
			var name string
			if ginkgoVersion == 1 {
				name = strings.Join(nameList, " ")
			} else {
				name = strings.Join(nameList, "/")
			}
			caseInfo := &ginkgoTestcase.TestCase{
				Path:       casePath,
				Name:       strings.TrimSpace(name),
				Attributes: map[string]string{},
			}
			fmt.Printf("find testcase : \n name: %v\n path: %v\n", name, path)
			caseInfo.Attributes["ginkgoVersion"] = strconv.Itoa(ginkgoVersion)
			if path != "" {
				caseInfo.Attributes["path"] = path
			}
			caseList = append(caseList, caseInfo)
		}
	}
	return caseList, nil
}
