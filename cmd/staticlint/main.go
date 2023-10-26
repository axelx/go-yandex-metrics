// staticlint анализатор кода
// необходимо в папки анализатора создать билд анализатора go build
// старт анализатора по всему проекту ./staticlint ../../...
package main

import (
	"encoding/json"
	"go/ast"
	"os"

	"github.com/Antonboom/errname/pkg/analyzer"
	"github.com/nunnatsa/ginkgolinter"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

// Config — имя файла конфигурации.
const Config = `config.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck []string
}

func main() {
	data, err := os.ReadFile(Config)
	if err != nil {
		panic(err)
	}
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	mychecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		analyzerAnalyzer(),
		ginkgolinterAnalyzer(),
		osExitAnalyzer,
	}
	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}

	for _, v := range staticcheck.Analyzers {
		// добавляем в массив нужные проверки
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}

func analyzerAnalyzer() *analysis.Analyzer {
	return analyzer.New()
}

func ginkgolinterAnalyzer() *analysis.Analyzer {
	return ginkgolinter.Analyzer
}

var osExitAnalyzer = &analysis.Analyzer{
	Name:     "noOsExit",
	Doc:      "Check for direct os.Exit calls in main function",
	Run:      checkOsExit,
	Requires: []*analysis.Analyzer{},
}

func checkOsExit(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				continue
			}

			for _, stmt := range fn.Body.List {
				if callExpr, ok := stmt.(*ast.ExprStmt); ok {
					if call, ok := callExpr.X.(*ast.CallExpr); ok {
						if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := fun.X.(*ast.Ident); ok {
								if ident.Name == "os" && fun.Sel.Name == "Exit" {
									pass.Reportf(call.Pos(), "Avoid direct os.Exit calls in main function")
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, nil
}
