package functionextractor

import (
	ph "codesearch-ai-data/internal/parsinghelpers"
	tc "codesearch-ai-data/internal/tokencounter"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

const MAX_FILE_BYTE_SIZE = 1000000 // 1MB
const DEFAULT_MIN_FUNCTION_LINES = 4

type functionExtractor struct {
	parser   *sitter.Parser
	minLines int
}

type ExtractedFunction struct {
	ID             int
	Identifier     string
	Code           string
	CleanCode      string
	CleanCodeHash  string
	InlineComments string
	Docstring      string
	StartLine      int
	EndLine        int
	IsTrain        bool
	TokenCounts    tc.TokenCounter
}

func getSHA1Hash(text string) string {
	hasher := sha1.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func isFunctionRightSize(codeText string, minLines int) bool {
	lines := strings.Split(codeText, "\n")
	return len(lines) >= minLines && len(lines) <= 512
}

func NewExtractedFunction(identifier string, cleanCode string, inlineComments []string, docstring string, node *sitter.Node, code []byte) *ExtractedFunction {
	return &ExtractedFunction{
		Identifier:     identifier,
		Code:           node.Content(code),
		CleanCode:      cleanCode,
		CleanCodeHash:  getSHA1Hash(cleanCode),
		InlineComments: strings.Join(inlineComments, " "),
		Docstring:      ph.GetPrecedingFunctionDocstring(node, code),
		StartLine:      int(node.StartPoint().Row),
		EndLine:        int(node.EndPoint().Row),
		TokenCounts:    tc.CountTokens(node, code),
	}
}

type FunctionExtractor interface {
	Extract(ctx context.Context, code []byte) ([]*ExtractedFunction, error)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getFunctionExtractorForFile(filePath string) FunctionExtractor {
	fileExtension := strings.TrimPrefix(filepath.Ext(filePath), ".")

	switch fileExtension {
	case "rb":
		return NewRubyFunctionExtractor(DEFAULT_MIN_FUNCTION_LINES)
	case "py":
		return NewPythonFunctionExtractor(DEFAULT_MIN_FUNCTION_LINES)
	case "php":
		return NewPhpFunctionExtractor(DEFAULT_MIN_FUNCTION_LINES)
	case "java":
		return NewJavaFunctionExtractor(DEFAULT_MIN_FUNCTION_LINES)
	case "js":
		return NewJavascriptFunctionExtractor(DEFAULT_MIN_FUNCTION_LINES)
	case "go":
		return NewGoFunctionExtractor(DEFAULT_MIN_FUNCTION_LINES)
	}

	return nil
}
