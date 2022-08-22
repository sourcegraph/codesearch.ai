package tokencounter

import (
	sp "codesearch-ai-data/internal/sitterparsers"
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/hexops/autogold"
	sitter "github.com/smacker/go-tree-sitter"
)

func tokenCounterToString(tc TokenCounter) string {
	keys := []string{}
	for k, _ := range tc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	output := []string{}
	for _, k := range keys {
		output = append(output, fmt.Sprintf("[%s]: %d", k, tc[k]))
	}
	return strings.Join(output, "\n")
}

func TestTokenCounter(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		parser *sitter.Parser
	}{
		{
			name:   "PythonTokenCounter",
			path:   "./testdata/test.py",
			parser: sp.GetPythonParser(),
		},
		{
			name:   "JavaTokenCounter",
			path:   "./testdata/test.java",
			parser: sp.GetJavaParser(),
		},
		{
			name:   "PhpTokenCounter",
			path:   "./testdata/test.php",
			parser: sp.GetPhpParser(),
		},
		{
			name:   "RubyTokenCounter",
			path:   "./testdata/test.rb",
			parser: sp.GetRubyParser(),
		},
		{
			name:   "GoTokenCounter",
			path:   "./testdata/test.go",
			parser: sp.GetGoParser(),
		},
		{
			name:   "JavascriptTokenCounter",
			path:   "./testdata/test.js",
			parser: sp.GetJavascriptParser(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testCode, err := ioutil.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}

			tree, err := tt.parser.ParseCtx(ctx, nil, testCode)
			if err != nil {
				t.Fatal(err)
			}
			rootNode := tree.RootNode()
			if rootNode.HasError() {
				t.Fatal("error encountered while parsing")
			}

			tokenCounter := CountTokens(rootNode, testCode)
			autogold.Equal(t, tokenCounterToString(tokenCounter))
		})
	}
}
