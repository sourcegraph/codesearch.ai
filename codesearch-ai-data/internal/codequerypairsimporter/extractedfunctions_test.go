package codequerypairsimporter

import (
	fe "codesearch-ai-data/internal/functionextractor"
	"testing"

	"github.com/hexops/autogold"
)

func TestExtractedFunctionToCodeQueryPair(t *testing.T) {
	tests := []struct {
		name string
		ef   *fe.ExtractedFunction
	}{
		{
			name: "Regular extracted function",
			ef: &fe.ExtractedFunction{
				Docstring:      "This is a docstring",
				InlineComments: "Inline comment A Inline Comment B",
				Identifier:     "FunctionA",
			},
		},
		{
			name: "Extracted function with unicode docstring",
			ef: &fe.ExtractedFunction{
				Docstring:      "Hello, 世 界 with a smiley face 🙂",
				InlineComments: "Inline comment A Inline Comment B",
				Identifier:     "FunctionA",
			},
		},
		{
			name: "Extracted function without docstring",
			ef: &fe.ExtractedFunction{
				InlineComments: "Inline comment A Inline Comment B",
				Identifier:     "FunctionA",
			},
		},
		{
			name: "Extracted function short docstring",
			ef: &fe.ExtractedFunction{
				InlineComments: "Docstring line",
				CleanCode:      "() => 1",
			},
		},
		{
			name: "Extracted function with multi-cased identifier",
			ef: &fe.ExtractedFunction{
				Identifier: "ClassA_FunctionHTML.call",
				CleanCode:  "() => 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cqp := extractedFunctionToCodeQueryPair(tt.ef)
			// Ignore in the autogold snapshot.
			cqp.ExtractedFunctionID = nil
			autogold.Equal(t, cqp)
		})
	}
}
