package functionextractor

import (
	"codesearch-ai-data/internal/githelpers"
	ph "codesearch-ai-data/internal/parsinghelpers"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	sitter "github.com/smacker/go-tree-sitter"
)

const MAX_FILE_BYTE_SIZE = 1000000 // 1MB
const DEFAULT_MIN_FUNCTION_LINES = 4

// TODO: Clean up things like:
// {@link TaskExecutorCustomizer TaskExecutorCustomizers} {@link ThreadPoolTaskExecutor} `something`

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
}

func getSHA1Hash(text string) string {
	hasher := sha1.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
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
	}
}

func isFunctionRightSize(codeText string, minLines int) bool {
	lines := strings.Split(codeText, "\n")
	return len(lines) >= minLines && len(lines) <= 512
}

func hasFileTooManyColumns(fileText string) bool {
	lines := strings.Split(fileText, "\n")
	for _, line := range lines {
		if len(line) > 1024 {
			return true
		}
	}
	return false
}

type FunctionExtractor interface {
	Extract(code []byte) ([]*ExtractedFunction, error)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func repoURLExists(repoURL string) bool {
	resp, err := http.Get(repoURL)
	if err != nil {
		return false
	}
	return resp.StatusCode == 200
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

func ProcessRepo(ctx context.Context, conn *pgx.Conn, repoName string) error {
	repoURL := fmt.Sprintf("https://%s", repoName)

	if !repoURLExists(repoURL) {
		return fmt.Errorf("repo URL %s does not exist", repoURL)
	}

	exists, err := repoExists(ctx, conn, repoName)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("repo %s already exists", repoURL)
	}

	repoPath, err := ioutil.TempDir("", "cloned-repo")
	if err != nil {
		return err
	}

	err = githelpers.CloneRepoWithTimeout(repoURL, repoPath, 300)
	if err != nil {
		log.Debugf("Error cloning repo %s: %s", repoName, err)
		return err
	}
	// Clean up cloned repo.
	defer func() { os.RemoveAll(repoPath) }()

	commitID, err := githelpers.GetRepoCommitID(repoPath)
	if err != nil {
		return err
	}

	repoID, err := insertRepo(ctx, conn, repoName, commitID)
	if err != nil {
		return err
	}

	err = filepath.Walk(repoPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			// Skip .git directory
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.Contains(path, ".pb.") || strings.HasSuffix(path, ".min.js") || info.Size() > MAX_FILE_BYTE_SIZE {
			return nil
		}

		relativePath := strings.TrimPrefix(strings.TrimPrefix(path, repoPath), "/")
		functionExtractor := getFunctionExtractorForFile(relativePath)
		if functionExtractor == nil {
			return nil
		}

		code, err := ioutil.ReadFile(path)
		if err != nil {
			log.Debugf("Error reading file %s/%s: %s", repoName, relativePath, err)
			// In case of a read error, skip file.
			return nil
		}

		if hasFileTooManyColumns(string(code)) {
			return nil
		}

		extractedFunctions, err := functionExtractor.Extract(code)
		if err != nil {
			log.Debugf("Error extracting functions %s/%s: %s", repoName, relativePath, err)
			// In case of a parse error, skip file.
			return nil
		}

		err = insertExtractedFunctionsFromFile(ctx, conn, repoID, relativePath, extractedFunctions)
		if err != nil {
			log.Debugf("Error inserting functions %s/%s: %s", repoName, relativePath, err)
			return nil
		}

		return nil
	})

	return err
}
