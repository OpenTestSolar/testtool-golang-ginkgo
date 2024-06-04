package util

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func ElementIsInSlice(element string, elements []string) bool {
	for _, item := range elements {
		if element == item {
			return true
		}
	}
	return false
}

// 根据给定路径返回路径对应的目录以及文件名，若路径指向目录则仅返回目录，若路径指向文件则返回文件对应目录以及文件名，如果路径不存在则返回错误
func GetPathAndFileName(projPath, path string) (dir string, file string, err error) {
	if path == "" {
		return "", "", errors.New("path is empty")
	}
	absPath := filepath.Join(projPath, path)
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return "", "", err
	}
	if fileInfo.IsDir() {
		return path, "", nil
	} else {
		relPath, err := filepath.Rel(projPath, filepath.Dir(absPath))
		if err != nil {
			return "", "", err
		}
		return relPath, filepath.Base(absPath), nil
	}
}

// RemoveExistingResultFile 函数用于删除指定路径的文件，如果文件存在的话。
// 参数:
//
//	filePath: 要删除的文件的路径
//
// 返回值:
//
//	如果文件删除成功或文件不存在，则返回 nil。否则返回错误信息。
func RemoveFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			log.Printf("failed to remove existing output json file, err: %s", err.Error())
			return err
		}
	} else if !os.IsNotExist(err) {
		log.Printf("failed to check file exists or not, err: %s", err.Error())
		return err
	}
	return nil
}

// IsJsonFileEmpty 检查指定路径的 JSON 文件是否为空。
// 如果文件为空或者 JSON 对象为空，则返回 true，否则返回 false。
// 如果在读取文件或解析 JSON 时发生错误，返回错误信息。
func IsJsonFileEmpty(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	if len(data) == 0 {
		return true, nil
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return false, err
	}

	switch v.(type) {
	case map[string]interface{}:
		if len(v.(map[string]interface{})) == 0 {
			return true, nil
		}
	case []interface{}:
		if len(v.([]interface{})) == 0 {
			return true, nil
		}
	}

	return false, nil
}

func ShortenString(str string, n int) string {
	if len(str) <= n {
		return str
	} else {
		return str[:n]
	}
}

func GetWorkspace(path string) string {
	var projPath string
	if path != "" {
		projPath = path
	} else {
		projPath = os.Getenv("TESTSOLAR_WORKSPACE")
	}
	return strings.TrimSuffix(projPath, string(os.PathSeparator))
}
