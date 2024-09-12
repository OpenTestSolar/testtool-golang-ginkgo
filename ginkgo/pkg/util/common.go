package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math/rand"
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

// GetPathAndFileName 根据给定路径返回路径对应的目录以及文件名，若路径指向目录则仅返回目录，若路径指向文件则返回文件对应目录以及文件名，如果路径不存在则返回错误
func GetPathAndFileName(projPath, path string) (dir string, file string, err error) {
	if path == "" {
		return "", "", errors.New("path is empty")
	}
	absPath := filepath.Join(projPath, path)
	if !strings.HasSuffix(absPath, ".go") {
		return path, "", nil
	} else {
		relPath, err := filepath.Rel(projPath, filepath.Dir(absPath))
		if err != nil {
			return "", "", err
		}
		return relPath, filepath.Base(absPath), nil
	}
}

// RemoveFile 函数用于删除指定路径的文件，如果文件存在的话。
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

	switch v := v.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			return true, nil
		}
	case []interface{}:
		if len(v) == 0 {
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func FindGinkgoVersion(path string) (ginkgoVersion int) {
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
			version := checkGinkgoImportVersion(filePath)
			if version > 0 {
				ginkgoVersion = version
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error searching for test files: %v\n", err)
	}
	return ginkgoVersion
}

func checkGinkgoImportVersion(file string) int {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		return 0
	}

	version := 0
	ast.Inspect(node, func(n ast.Node) bool {
		importSpec, ok := n.(*ast.ImportSpec)
		if ok && importSpec.Path != nil {
			if importSpec.Path.Value == "\"github.com/onsi/ginkgo/v2\"" {
				version = 2
				return false
			} else if importSpec.Path.Value == "\"github.com/onsi/ginkgo\"" {
				version = 1
				return false
			}
		}
		return true
	})

	return version
}

func FindFilesWithSuffix(path, suffix string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), suffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ListFiles(dirPath string) error {
	// 使用 ioutil.ReadDir 读取目录内容
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// 遍历并打印目录内容
	for _, file := range files {
		log.Println(file.Name())
	}

	return nil
}
