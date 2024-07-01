// Пакет описывает анализатор (multichecker) для проверок основного проекта.
package main

import (
	"go/ast"

	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
)

// MainAnalyzer анализатор пакета main
var MainAnalyzer = &analysis.Analyzer{
	Name: "mainChecker",
	Doc:  "check for os.Exit in pkg main",
	Run:  runMainAnalyzer,
}

// runMainAnalyzer отслеживает некорректные ситуации в пакете main:
//
// Запрещающено использовать прямой вызов os.Exit в функции main пакета main
func runMainAnalyzer(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {

			switch x := node.(type) {
			case *ast.File:
				if x.Name.Name != "main" {
					return false
				}
			case *ast.FuncDecl:
				if x.Name.Name != "main" {
					return false
				}
			case *ast.Ident:
				if x.Name == "Exit" {
					pass.Reportf(x.Pos(), "it is forbidden to use exit in the main package in the main function")
				}
				if x.Name != "os" {
					return false
				}
			}
			return true
		})
	}
	return nil, nil
}

// checks содержит слайс анализаторов для запуска.
type checks struct {
	analyzers []*analysis.Analyzer
}

// addAnalysisPasses добавляет в слайс стандартные статические анализаторы пакета golang.org/x/tools/go/analysis/passes.
func (c *checks) addAnalysisPasses() {
	c.analyzers = append(c.analyzers, printf.Analyzer)
	c.analyzers = append(c.analyzers, shadow.Analyzer)
	c.analyzers = append(c.analyzers, structtag.Analyzer)
}

// addStaticcheck добавляет анализаторы класса 'SA' пакета "staticcheck.io".
func (c *checks) addStaticcheck() {
	for _, v := range staticcheck.Analyzers {
		c.analyzers = append(c.analyzers, v.Analyzer)
	}
}

// addQuickFix добавляет анализаторы класса 'QF' пакета "staticcheck.io".
func (c *checks) addQuickFix() {
	for _, v := range quickfix.Analyzers {
		c.analyzers = append(c.analyzers, v.Analyzer)
	}
}

// addPublicAnalyzers добавляет публичные анализаторы
//
// errcheck проверяет чтобы все ошибки были получены и проверены
//
// Не пойму от куда брать дольше публичных анализаторов нужна помощь
func (c *checks) addPublicAnalyzers() {
	c.analyzers = append(c.analyzers, errcheck.Analyzer)
}

// addMyAnalyzers добавляет самописные анализаторы
func (c *checks) addMyAnalyzers() {
	c.analyzers = append(c.analyzers, MainAnalyzer)
}

/*
Запуск multichecker, который состоит из:
 1. Стандартных статических анализаторов пакета "golang.org/x/tools/go/analysis/passes";
 2. Всех анализаторов класса 'SA' пакета "staticcheck.io";
 3. Не менее одного анализатора остальных классов ('QF') пакета "staticcheck.io";
 4. Двух или более любых публичных анализаторов.
*/
func main() {
	c := checks{
		analyzers: make([]*analysis.Analyzer, 0),
	}
	c.addAnalysisPasses()
	c.addStaticcheck()
	c.addQuickFix()
	c.addPublicAnalyzers()
	c.addMyAnalyzers()

	multichecker.Main(
		c.analyzers...,
	)
}
