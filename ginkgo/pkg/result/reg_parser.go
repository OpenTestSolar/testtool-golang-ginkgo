package result

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	ginkgoTestcase "github.com/OpenTestSolar/testtool-golang-ginkgo/ginkgo/pkg/testcase"
)

func isValidSection(section, proj string) bool {
	if strings.TrimSpace(section) == "" {
		return false
	} else if strings.Contains(section, "BeforeSuite") {
		return false
	} else if strings.Contains(section, "AfterSuite") {
		return false
	} else if strings.Contains(section, "PASSED") {
		return false
	} else if !strings.Contains(section, proj) {
		return false
	} else if strings.Contains(section, "PENDING") {
		return false
	}
	return true
}

func getValidLines(section string) []string {
	lines := strings.Split(section, "\n")
	var validLines []string
	for _, line := range lines {
		if line != "" {
			validLines = append(validLines, line)
		}
	}
	return validLines
}

func removeExtraSpace(index int, line string) string {
	if index > 0 {
		line = strings.TrimPrefix(line, "  ")
	} else {
		line = strings.TrimSuffix(line, " ")
	}
	return line
}

func ParseCaseByReg(proj string, output string, ginkgoVersion int, packPath string) ([]*ginkgoTestcase.TestCase, error) {
	var caseList []*ginkgoTestcase.TestCase
	regexPattern := `(?s).*?Will run.*?specs(.*?)Ran.*?Specs in.*?seconds`
	re := regexp.MustCompile(regexPattern)
	extractedText := re.FindStringSubmatch(output)
	if len(extractedText) > 1 {
		extractedContent := extractedText[1]
		sections := strings.Split(extractedContent, "------------------------------")
		for _, section := range sections {
			if !isValidSection(section, proj) {
				continue
			}
			lines := getValidLines(section)
			var path string
			var nameList []string
			pathRegex := regexp.MustCompile(`.*?(.*?):\d+`)
			for i, line := range lines {
				line = removeExtraSpace(i, line)
				pathMatch := pathRegex.FindStringSubmatch(line)
				if len(pathMatch) > 1 && strings.Contains(line, proj) {
					path = pathMatch[1]
				} else if strings.Contains(line, "•") {
					continue
				} else {
					nameList = append(nameList, line)
				}
			}
			selectorPath, err := filepath.Rel(proj, path)
			if err != nil {
				fmt.Printf("get rel path failed, err: %v\n", err)
				selectorPath = strings.Split(strings.TrimSpace(path), proj)[1]
				selectorPath = selectorPath[1:]
			}
			// TODO: 后续需要将分割符统一切换为空格
			var name string
			if ginkgoVersion == 1 {
				name = strings.Join(nameList, " ")
			} else {
				name = strings.Join(nameList, "/")
			}
			caseInfo := &ginkgoTestcase.TestCase{
				Path:       selectorPath,
				Name:       name,
				Attributes: map[string]string{},
			}
			fmt.Printf("find testcase : \n name: %v\n path: %v\n", name, path)
			caseInfo.Attributes["ginkgoVersion"] = strconv.Itoa(ginkgoVersion)
			if marshalNameList, err := json.Marshal(nameList); err == nil {
				caseInfo.Attributes["nameList"] = string(marshalNameList)
			}
			if packPath != "" {
				caseInfo.Attributes["path"] = packPath
			}
			caseList = append(caseList, caseInfo)
		}
	}
	return caseList, nil
}
