package selector

import "net/url"

type TestSelector struct {
	Value      string
	Path       string
	Name       string
	Attributes map[string]string
}

func NewTestSelector(selector string) (*TestSelector, error) {
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

	testSelector := &TestSelector{
		Value:      selector,
		Path:       path,
		Name:       name,
		Attributes: attributes,
	}
	return testSelector, nil
}

func (ts *TestSelector) IsExclude() bool {
	exclude, ok := ts.Attributes["exclude"]
	return ok && exclude == "true"
}

func (ts *TestSelector) String() string {
	strSelector := ts.Path
	if ts.Name != "" {
		strSelector += "?" + ts.Name
	}
	return strSelector
}
