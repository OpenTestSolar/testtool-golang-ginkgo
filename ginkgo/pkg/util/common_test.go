package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkspace(t *testing.T) {
	workspace := GetWorkspace("/data")
	assert.NotEmpty(t, workspace)
	err := os.Setenv("TESTSOLAR_WORKSPACE", "/data")
	assert.NoError(t, err)
	workspace = GetWorkspace("")
	assert.NotEmpty(t, workspace)
}

func TestShortenString(t *testing.T) {
	str := ShortenString("abcdefghijklmnopqrstuvwxyz", 10)
	assert.Equal(t, str, "abcdefghij")
	str = ShortenString("abcdefghijklmnopqrstuvwxyz", 20)
	assert.Equal(t, str, "abcdefghijklmnopqrst")
}

func TestIsJsonFileEmpty(t *testing.T) {
	// 创建一个临时目录来存放测试文件
	tempDir := t.TempDir()
	emptyJsonFile := filepath.Join(tempDir, "empty.json")
	_ = os.WriteFile(emptyJsonFile, []byte("{}"), 0644)

	// 创建一个非空的JSON对象文件
	nonEmptyObjectFile := filepath.Join(tempDir, "nonempty_object.json")
	_ = os.WriteFile(nonEmptyObjectFile, []byte(`{"key": "value"}`), 0644)

	// 创建一个非空的JSON数组文件
	nonEmptyArrayFile := filepath.Join(tempDir, "nonempty_array.json")
	_ = os.WriteFile(nonEmptyArrayFile, []byte(`[1, 2, 3]`), 0644)

	// 创建一个非JSON文件
	nonJsonFile := filepath.Join(tempDir, "nonjson.txt")
	_ = os.WriteFile(nonJsonFile, []byte("This is not JSON."), 0644)

	// 测试用例
	testCases := []struct {
		path        string
		expected    bool
		expectError bool
	}{
		{emptyJsonFile, true, false},
		{nonEmptyObjectFile, false, false},
		{nonEmptyArrayFile, false, false},
		{nonJsonFile, false, true},
		{"nonexistent.json", false, true},
	}

	for _, tc := range testCases {
		isEmpty, err := IsJsonFileEmpty(tc.path)
		if tc.expectError {
			require.Error(t, err, "Expected an error for path: %s", tc.path)
		} else {
			require.NoError(t, err, "Unexpected error for path: %s", tc.path)
			assert.Equal(t, tc.expected, isEmpty, "Unexpected result for path: %s", tc.path)
		}
	}
}

func TestRemoveFile(t *testing.T) {
	// 创建一个临时文件用于测试
	tempFile, err := os.CreateTemp("", "temp_file_for_test")
	require.NoError(t, err, "创建临时文件时出错")
	tempFilePath := tempFile.Name()
	defer os.Remove(tempFilePath) // 确保测试结束后删除临时文件
	tempFile.Close()
	// 测试文件存在的情况
	err = RemoveFile(tempFilePath)
	assert.NoError(t, err, "删除存在的文件时出错")
	_, err = os.Stat(tempFilePath)
	assert.True(t, os.IsNotExist(err), "文件应该已经被删除")
	// 测试文件不存在的情况
	err = RemoveFile(tempFilePath)
	assert.NoError(t, err, "删除不存在的文件时出错")
	// 测试文件路径为空的情况
	err = RemoveFile("")
	assert.NoError(t, err, "删除空路径文件时返回为空")
	// 测试无效路径的情况
	err = RemoveFile(filepath.Join("invalid", "path"))
	assert.NoError(t, err, "删除无效路径的文件时返回为空")
}

// 测试GetPathAndFileName函数
func TestGetPathAndFileName(t *testing.T) {
	testdata, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	// 测试空路径
	_, _, err = GetPathAndFileName(testdata, "")
	assert.Equal(t, err.Error(), "path is empty")
	// 测试目录路径
	dir, file, err := GetPathAndFileName(testdata, "demo")
	assert.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.Empty(t, file)
	// 测试文件路径
	dir, file, err = GetPathAndFileName(filepath.Join(testdata, "demo"), "demo_test.go")
	assert.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.Equal(t, file, "demo_test.go")
}

func TestElementIsInSlice(t *testing.T) {
	testCases := []struct {
		element  string
		elements []string
		want     bool
	}{
		{"a", []string{"a", "b", "c"}, true},
		{"d", []string{"a", "b", "c"}, false},
		{"", []string{"a", "b", "c"}, false},
		{"a", []string{}, false},
		{"", []string{}, false},
	}
	for _, tc := range testCases {
		got := ElementIsInSlice(tc.element, tc.elements)
		assert.Equal(t, got, tc.want)
	}
}

func TestFileExistsWhenFileExists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	// 测试用例1：文件存在
	exists, err := FileExists(tmpFile.Name())
	assert.NoError(t, err)
	assert.True(t, exists)
	nonExistentPath := "/path/to/non/existent/file"
	// 测试用例2：文件不存在
	exists, err = FileExists(nonExistentPath)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestGenRandomString(t *testing.T) {
	got := GenRandomString(10)
	assert.Equal(t, 10, len(got))
}
