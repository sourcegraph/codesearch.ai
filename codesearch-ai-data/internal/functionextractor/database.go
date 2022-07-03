package functionextractor

import (
	"codesearch-ai-data/internal/database"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
)

func repoExists(ctx context.Context, conn *pgx.Conn, repoName string) (bool, error) {
	var repoCount int
	err := conn.QueryRow(ctx, "SELECT COUNT(*) FROM repos WHERE name = $1", repoName).Scan(&repoCount)
	if err != nil {
		return false, err
	}
	return repoCount != 0, nil
}

func insertRepo(ctx context.Context, conn *pgx.Conn, repoName, commitID string) (int, error) {
	var repoID int
	err := conn.QueryRow(ctx, "INSERT INTO repos (name, commit_id) VALUES ($1, $2) RETURNING id", repoName, commitID).Scan(&repoID)
	if err != nil {
		return -1, err
	}
	return repoID, nil
}

const insertExtractedFunctionsBatchSize = 32

const insertExtractedFunctionsQuery = `
INSERT INTO extracted_functions (repo_id, path, docstring, inline_comments, clean_code, clean_code_hash, identifier, start_line, end_line)
VALUES
	%s
ON CONFLICT (clean_code_hash)
DO UPDATE SET
  docstring = TRIM(CONCAT(extracted_functions.docstring, ' ', EXCLUDED.docstring)),
  inline_comments = TRIM(CONCAT(extracted_functions.inline_comments, ' ', EXCLUDED.inline_comments));
`

func deduplicateExtractedFunctions(extractedFunctions []*ExtractedFunction) []*ExtractedFunction {
	hashToDuplicateFunctions := map[string][]*ExtractedFunction{}
	for _, ef := range extractedFunctions {
		hashToDuplicateFunctions[ef.CleanCodeHash] = append(hashToDuplicateFunctions[ef.CleanCodeHash], ef)
	}

	deduplicatedFunctions := make([]*ExtractedFunction, 0, len(hashToDuplicateFunctions))
	for _, duplicateFunctions := range hashToDuplicateFunctions {
		docstrings := make([]string, 0, len(duplicateFunctions))
		inlineComments := make([]string, 0, len(duplicateFunctions))

		for _, duplicateFunction := range duplicateFunctions {
			docstrings = append(docstrings, duplicateFunction.Docstring)
			inlineComments = append(inlineComments, duplicateFunction.InlineComments)
		}

		deduplicatedFunction := duplicateFunctions[0]
		deduplicatedFunction.Docstring = strings.TrimSpace(strings.Join(docstrings, " "))
		deduplicatedFunction.InlineComments = strings.TrimSpace(strings.Join(inlineComments, " "))
		deduplicatedFunctions = append(deduplicatedFunctions, deduplicatedFunction)
	}
	return deduplicatedFunctions
}

func insertExtractedFunctionsFromFile(ctx context.Context, conn *pgx.Conn, repoID int, filePath string, extractedFunctions []*ExtractedFunction) error {
	// Deduplicate extracted functions before inserting them because the ON CONFLICT clause does not work when inserting multiple duplicated values.
	deduplicatedFunctions := deduplicateExtractedFunctions(extractedFunctions)
	length := len(deduplicatedFunctions)
	for i := 0; i < length; i += insertExtractedFunctionsBatchSize {
		end := i + insertExtractedFunctionsBatchSize
		if end > length {
			end = length
		}

		extractedFunctionsBatch := deduplicatedFunctions[i:end]

		insertValuesParameters, valuesArgs := database.PrepareValuesForBulkInsert(extractedFunctionsBatch, 9, func(valueArgs []any, ef *ExtractedFunction) []any {
			return append(valueArgs, repoID, filePath, ef.Docstring, ef.InlineComments, ef.CleanCode, ef.CleanCodeHash, ef.Identifier, ef.StartLine, ef.EndLine)
		})

		_, err := conn.Exec(ctx, fmt.Sprintf(insertExtractedFunctionsQuery, insertValuesParameters), valuesArgs...)
		if err != nil {
			return err
		}
	}

	return nil
}
