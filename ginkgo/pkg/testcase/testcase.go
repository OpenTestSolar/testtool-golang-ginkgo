package testcase

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

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

func ParseTestCaseBySelector(selector string) (*TestCase, error) {
	u, err := url.Parse(selector)
	if err != nil {
		return nil, err
	}
	path := u.Path
	rawQuery := u.RawQuery
	query, _ := url.ParseQuery(rawQuery)
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
