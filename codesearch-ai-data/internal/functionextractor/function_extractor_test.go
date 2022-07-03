package functionextractor

import (
	"io/ioutil"
	"sort"
	"testing"

	"github.com/hexops/autogold"
)

func TestFunctionExtractors(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		extractor FunctionExtractor
	}{
		{
			name:      "RubyFunctionExtractor",
			path:      "../testdata/test.rb",
			extractor: NewRubyFunctionExtractor(0),
		},
		{
			name:      "GoFunctionExtractor",
			path:      "../testdata/test.go",
			extractor: NewGoFunctionExtractor(0),
		},
		{
			name:      "PythonFunctionExtractor",
			path:      "../testdata/test.py",
			extractor: NewPythonFunctionExtractor(0),
		},
		{
			name:      "JavascriptFunctionExtractor",
			path:      "../testdata/test.js",
			extractor: NewJavascriptFunctionExtractor(0),
		},
		{
			name:      "JavaFunctionExtractor",
			path:      "../testdata/test.java",
			extractor: NewJavaFunctionExtractor(0),
		},
		{
			name:      "PhpFunctionExtractor",
			path:      "../testdata/test.php",
			extractor: NewPhpFunctionExtractor(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCode, err := ioutil.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}

			extractedFunctions, err := tt.extractor.Extract(testCode)
			if err != nil {
				t.Fatal(err)
			}

			sort.SliceStable(extractedFunctions, func(i, j int) bool {
				return extractedFunctions[i].Identifier < extractedFunctions[j].Identifier
			})

			autogold.Equal(t, extractedFunctions)
		})
	}
}
