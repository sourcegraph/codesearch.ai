package socode

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestGetCodeSnippetsFromHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{
			name: "No code",
			html: "",
		},
		{
			name: "Single empty code",
			html: "<code></code>",
		},
		{
			name: "Single code",
			html: "<code>1+1</code>",
		},
		{
			name: "Nested code",
			html: "<code><code>1+1</code></code>",
		},
		{
			name: "Mismatched tags code",
			html: "<code><code>1+1</code>",
		},
		{
			name: "Multiple nested code snippets",
			html: "<code>1<code>2+3</code>4</code>Irrelevant<code>5<code>6<code>7</code></code>8</code>",
		},
		{
			name: "End tag before start tag",
			html: "<code>1</code></code><code><code>2</code>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeSnippets := GetCodeSnippetsFromHTML(tt.html)
			autogold.Equal(t, codeSnippets)
		})
	}
}
