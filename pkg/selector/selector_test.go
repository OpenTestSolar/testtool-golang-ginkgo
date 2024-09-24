package selector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExclude(t *testing.T) {
	ts1, err := NewTestSelector("path?name=testname&exclude=true")
	assert.NoError(t, err)
	assert.True(t, ts1.IsExclude())
}

func TestString(t *testing.T) {
	ts1, err := NewTestSelector("path?name=testname&exclude=true")
	assert.NoError(t, err)
	assert.NotEmpty(t, ts1.String())
}

func TestNewTestSelector(t *testing.T) {
	// 测试用例1: 正常的selector字符串
	ts1, err := NewTestSelector("path?name=testname&attr1=value1&attr2=value2")
	assert.NoError(t, err)
	assert.Equal(t, "path", ts1.Path)
	assert.Equal(t, "testname", ts1.Name)
	assert.Equal(t, ts1.Attributes["attr1"], "value1")
	assert.Equal(t, ts1.Attributes["attr2"], "value2")

	// 测试用例2: 不包含查询参数的selector字符串
	ts2, err := NewTestSelector("path")
	assert.NoError(t, err)
	assert.Equal(t, "path", ts2.Path)
	assert.Empty(t, ts2.Name)
	assert.Len(t, ts2.Attributes, 0)

	// 测试用例3: 包含特殊字符的selector字符串
	ts4, err := NewTestSelector("path?name=test%20name&attr1=value%3D1")
	assert.NoError(t, err)
	assert.Equal(t, "path", ts4.Path)
	assert.Equal(t, "test name", ts4.Name)
	assert.Equal(t, "value=1", ts4.Attributes["attr1"])
}
