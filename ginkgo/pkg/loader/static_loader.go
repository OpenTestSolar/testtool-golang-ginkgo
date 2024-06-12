package loader

import (
	"fmt"
	ginkgoTestcase "ginkgo/pkg/testcase"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func genTestCaseBySpec(path string, spec *TestCaseSpec, root *ginkgoTestcase.TestCase) ([]*ginkgoTestcase.TestCase, error) {
	var testcases []*ginkgoTestcase.TestCase
	if root == nil {
		root = &ginkgoTestcase.TestCase{
			Path:       path,
			Name:       spec.name,
			Attributes: map[string]string{}, //TODO:
		}
	} else {
		root.Name += "/" + spec.name
	}
	if len(spec.subSpecs) == 0 {
		testcases = append(testcases, root)
		return testcases, nil
	}
	for _, subSpec := range spec.subSpecs {
		newRoot := &ginkgoTestcase.TestCase{
			Path:       root.Path,
			Name:       root.Name,
			Attributes: root.Attributes,
		}
		newTestCases, err := genTestCaseBySpec(path, subSpec, newRoot)
		if err != nil {
			continue
		}
		testcases = append(testcases, newTestCases...)
	}
	return testcases, nil
}

type TestCaseSpec struct {
	kind     string
	name     string
	subSpecs []*TestCaseSpec
}

func parseTestCaseSpec(expr *ast.CallExpr) (*TestCaseSpec, error) {
	testcaseSpec := &TestCaseSpec{}
	name := ""
	exprName := reflect.TypeOf(expr.Fun).String()
	if exprName == "*ast.Ident" {
		name = expr.Fun.(*ast.Ident).Name
	} else if exprName == "*ast.SelectorExpr" {
		if selectorExpr, ok := expr.Fun.(*ast.SelectorExpr); ok {
			name = selectorExpr.Sel.Name
		}
	} else {
		return nil, nil
	}
	if name == "Describe" || name == "Context" || name == "It" {
		testcaseSpec.kind = name
		args := expr.Args
		if len(args) != 2 && len(args) != 3 {
			return nil, fmt.Errorf("Invalid ginkgo spec %s", name)
		}
		switch args[0].(type) {
		case *ast.BasicLit:
			arg1 := args[0].(*ast.BasicLit)
			if arg1.Kind == token.STRING {
				testcaseSpec.name = strings.Replace(arg1.Value, "\"", "", -1)
			}
		case *ast.Ident:
			arg1 := args[0].(*ast.Ident)
			log.Printf("Unsupported testcase description: %v", arg1)
		}

		if name == "It" {
			// leaf node
			return testcaseSpec, nil
		}
		index := 1
		if len(args) == 3 {
			index = 2
		}
		arg2 := args[index].(*ast.FuncLit)
		for _, it := range arg2.Body.List {
			switch it.(type) {
			case *ast.ExprStmt:
				subSpec, err := parseTestCaseSpec(it.(*ast.ExprStmt).X.(*ast.CallExpr))
				if err != nil {
					log.Printf("Parse ginkgo testcase failed: %v", err)
					continue
				}
				if subSpec != nil {
					testcaseSpec.subSpecs = append(testcaseSpec.subSpecs, subSpec)
				}
			case *ast.DeclStmt:
				subSpec, err := parseTestCaseSpec(it.(*ast.DeclStmt).Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0].(*ast.CallExpr))
				if err != nil {
					log.Printf("Parse ginkgo testcase failed: %v", err)
					continue
				}
				if subSpec != nil {
					testcaseSpec.subSpecs = append(testcaseSpec.subSpecs, subSpec)
				}
			}
		}
		return testcaseSpec, nil
	}
	return nil, nil
}

func ParseTestCaseInFile(projPath string, path string) ([]*ginkgoTestcase.TestCase, error) {
	var testcaseList []*ginkgoTestcase.TestCase
	if !strings.HasSuffix(path, "_test.go") {
		return testcaseList, nil
	}
	log.Println("Parse testcase in file", path)
	code, err := os.ReadFile(path)
	if err != nil {
		return testcaseList, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, code, 0)
	if err != nil {
		log.Printf("Invalid go file %s", path)
		return testcaseList, err
	}
	ginkgoVersion := 0
	for _, decl := range file.Decls {
		switch declType := decl.(type) {
		case *ast.GenDecl:
			if declType.Tok == token.IMPORT {
				for _, spec := range declType.Specs {
					importValue := spec.(*ast.ImportSpec).Path.Value
					if strings.Contains(importValue, "github.com/onsi/ginkgo/v2") {
						ginkgoVersion = 2
					} else if strings.Contains(importValue, "github.com/onsi/ginkgo") {
						ginkgoVersion = 1
					}
				}
			} else if declType.Tok == token.VAR {
				for _, spec := range declType.Specs {
					for _, value := range spec.(*ast.ValueSpec).Values {
						switch value := value.(type) {
						case *ast.CallExpr:
							spec, err := parseTestCaseSpec(value)
							if err != nil {
								log.Printf("Parse ginkgo testcase failed: %v", err)
								continue
							}
							if spec != nil {
								testcases, err := genTestCaseBySpec(path[len(projPath)+1:], spec, nil)
								if err != nil {
									log.Printf("Generate testcase failed: %v", err)
									return testcaseList, err
								}
								if ginkgoVersion == 0 {
									log.Printf("ginkgo version was not found, Please check file import")
									return testcaseList, nil
								}
								for _, testcase := range testcases {
									testcase.Attributes["ginkgoVersion"] = strconv.Itoa(ginkgoVersion)
								}
								testcaseList = append(testcaseList, testcases...)
							}
						}
					}
				}
			}
		}
	}
	return testcaseList, nil
}
