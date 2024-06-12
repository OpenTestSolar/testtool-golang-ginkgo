package cmdline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmdArgsWithDefauleArgs(t *testing.T) {
	cmdArgs := []*CommandArg{
		{
			Key:   "key1",
			Value: "value1",
		},
	}
	ca := NewCmdArgsWithDefauleArgs(cmdArgs)
	assert.Equal(t, ca.Args, cmdArgs)
}

func TestNewCmdArgsParseByCmdLine(t *testing.T) {
	// 测试正常输入
	cmdline := "ginkgo --arg1 --arg2"
	ca, err := NewCmdArgsParseByCmdLine(cmdline)
	assert.NoError(t, err)
	assert.Equal(t, len(ca.Args), 3)
	// 测试无效输入
	cmdline = ""
	_, err = NewCmdArgsParseByCmdLine(cmdline)
	assert.Error(t, err)
	// 测试错误情况（例如，输入包含无法解析的引号）
	cmdline = "ginkgo \"arg with unclosed quote"
	_, err = NewCmdArgsParseByCmdLine(cmdline)
	assert.Error(t, err)
}

func TestAdd(t *testing.T) {
	cmdline := "ginkgo --arg1 --arg2"
	ca, err := NewCmdArgsParseByCmdLine(cmdline)
	assert.NoError(t, err)
	assert.Equal(t, len(ca.Args), 3)
	ca.Add(&CommandArg{
		Key:   "key1",
		Value: "value1",
	})
	assert.Equal(t, len(ca.Args), 4)
}

func TestExtend(t *testing.T) {
	cmdline := "ginkgo --arg1 --arg2"
	ca, err := NewCmdArgsParseByCmdLine(cmdline)
	assert.NoError(t, err)
	assert.Equal(t, len(ca.Args), 3)
	ca.Extend([]*CommandArg{
		{
			Key:   "key1",
			Value: "value1",
		},
		{
			Key:   "key2",
			Value: "value2",
		},
	})
	assert.Equal(t, len(ca.Args), 5)
}

func TestGetValueByKey(t *testing.T) {
	cmdline := "ginkgo --arg1 --arg2"
	ca, err := NewCmdArgsParseByCmdLine(cmdline)
	assert.NoError(t, err)
	assert.Equal(t, len(ca.Args), 3)
	value := ca.GetValueByKey("arg1")
	assert.Equal(t, value, "")
}

func TestMerge(t *testing.T) {
	cmdline1 := "ginkgo --arg1 --arg2"
	ca1, err := NewCmdArgsParseByCmdLine(cmdline1)
	assert.NoError(t, err)
	assert.Equal(t, len(ca1.Args), 3)
	cmdline2 := "ginkgo --arg1 value1 --arg3 --arg4"
	ca2, err := NewCmdArgsParseByCmdLine(cmdline2)
	assert.NoError(t, err)
	assert.Equal(t, len(ca2.Args), 4)
	ca1.Merge(ca2)
	assert.Equal(t, len(ca1.Args), 5)
}

func TestGenerateCmdLineStr(t *testing.T) {
	cmdline := "ginkgo --arg1 --arg2"
	ca, err := NewCmdArgsParseByCmdLine(cmdline)
	assert.NoError(t, err)
	assert.Equal(t, len(ca.Args), 3)
	asserteStr := ca.GenerateCmdLineStr()
	assert.Equal(t, asserteStr, "ginkgo --arg1 --arg2")
}

func TestNeedFocus(t *testing.T) {
	// need focus
	cmdline := "ginkgo --arg1 --arg2"
	ca, err := NewCmdArgsParseByCmdLine(cmdline)
	assert.NoError(t, err)
	assert.Equal(t, len(ca.Args), 3)
	focus := ca.NeedFocus()
	assert.Equal(t, focus, true)
	// not need focus
	cmdline = "ginkgo --focus xxx --arg2"
	ca, err = NewCmdArgsParseByCmdLine(cmdline)
	assert.NoError(t, err)
	assert.Equal(t, len(ca.Args), 3)
	focus = ca.NeedFocus()
	assert.Equal(t, focus, false)
}

func TestNewCmdArgs(t *testing.T) {
	ca := NewCmdArgs()
	assert.NotNil(t, ca)
}
