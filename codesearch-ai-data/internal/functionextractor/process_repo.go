package functionextractor

import (
	"codesearch-ai-data/internal/githelpers"
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

func hasFileTooManyColumns(fileText string) bool {
	lines := strings.Split(fileText, "\n")
	for _, line := range lines {
		if len(line) > 1024 {
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
		return fmt.Errorf("repo %s already exists", repoName)
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

	return processRepoPath(ctx, conn, repoName, repoPath)
}

func ProcessRepoPath(ctx context.Context, conn *pgx.Conn, repoName string, repoPath string) error {
	exists, err := repoExists(ctx, conn, repoName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("repo %s already exists", repoName)
	}
	return processRepoPath(ctx, conn, repoName, repoPath)
}

func processRepoPath(ctx context.Context, conn *pgx.Conn, repoName string, repoPath string) error {
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

		extractedFunctions, err := functionExtractor.Extract(ctx, code)
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
