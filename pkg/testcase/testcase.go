package testcase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	ginkgoUtil "github.com/OpenTestSolar/testtool-golang-ginkgo/pkg/util"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

type TestCase struct {
	Path       string
	Name       string
	Attributes map[string]string
}

func (tc *TestCase) GetSelector() string {
	strSelector := tc.Path
	if tc.Name != "" {
		strSelector += "?" + tc.Name
	}
	return strSelector
}

func parseAttrSliceValue(rawString string) ([]string, error) {
	var result []string
	err := json.Unmarshal([]byte(rawString), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (tc *TestCase) MatchAttr(attr map[string]string) bool {
	if len(attr) == 0 {
		// 若不指定属性，则直接返回true
		return true
	}
	for key, value := range attr {
		var valueList []string
		if strings.Contains(value, ",") {
			pars := strings.Split(value, ",")
			for _, v := range pars {
				valueList = append(valueList, strings.TrimSpace(v))
			}
		} else {
			valueList = []string{strings.TrimSpace(value)}
		}
		// 若指定属性不存在，则直接返回false
		if _, ok := tc.Attributes[key]; !ok {
			return false
		}
		// 若属性值为空则直接返回false
		if tc.Attributes[key] == "" {
			return false
		}
		strListValue, err := parseAttrSliceValue(tc.Attributes[key])
		if err != nil {
			log.Printf("[Plugin] parse attr value %s failed, err: %v", tc.Attributes[key], err)
			strListValue = []string{tc.Attributes[key]}
		}
		for _, v := range valueList {
			// 若任何一个期望属性与用例属性相匹配则返回true
			if strListValue != nil {
				if ginkgoUtil.ElementIsInSlice(v, strListValue) {
					return true
				}
			}
		}
	}
	// 执行到这一步表示下发的期望属性存在于用例用例中，但是具体的属性值与用例属性值没有匹配上，所以返回false
	return false
}

func ParseTestCaseBySelector(selector string) (*TestCase, error) {
	selector = strings.Replace(selector, "+", "%2B", -1)
	u, err := url.Parse(selector)
	if err != nil {
		return nil, err
	}
	path := u.Path
	rawQuery := u.RawQuery
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return nil, err
	}
	name := ""
	attributes := map[string]string{}
	for k, v := range query {
		if k == "name" {
			if len(v) == 1 {
				name = v[0]
			}
		} else if len(v) == 1 && v[0] == "" {
			if len(query) == 1 {
				name = k
			}
		} else {
			if len(v) >= 1 {
				attributes[k] = v[0]
			}
		}
	}
	if name == "" && strings.Contains(rawQuery, "&") {
		log.Printf("[Plugin] case name contain `&`, selector: %s", selector)
		name = rawQuery
	}
	testCase := &TestCase{
		Path:       path,
		Name:       name,
		Attributes: attributes,
	}
	return testCase, nil
}

func UnmarshalCaseInfo(path string) (*sdkModel.EntryParam, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read case info failed, err: %v", err)
	}
	var config sdkModel.EntryParam
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal case info into model failed, err: %v", err)
	}
	return &config, nil
}
