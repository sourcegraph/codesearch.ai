package sitterparsers

import (
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
)

func GetRubyParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(ruby.GetLanguage())
	return parser
}

func GetGoParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	return parser
}

func GetPythonParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	return parser
}

func GetJavascriptParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	return parser
}

func GetJavaParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())
	return parser
}

func GetPhpParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())
	return parser
}

func GetParserForLanguage(language string) *sitter.Parser {
	switch language {
	case "ruby":
		return GetRubyParser()
	case "python":
		return GetPythonParser()
	case "php":
		return GetPhpParser()
	case "java":
		return GetJavaParser()
	case "javascript":
		return GetJavascriptParser()
	case "go":
		return GetGoParser()
	}

	panic("unknown language: " + language)
}
