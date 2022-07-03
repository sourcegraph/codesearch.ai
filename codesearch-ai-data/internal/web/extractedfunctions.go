package web

import (
	"codesearch-ai-data/internal/database"
	"context"
	"fmt"
	"html/template"

	"github.com/jackc/pgx/v4"
)

type HighlightedExtractedFunction struct {
	ID              int
	RepositoryName  string
	CommitID        string
	FilePath        string
	StartLine       int
	EndLine         int
	HighlightedHTML template.HTML
	URL             string
}

const extractedFunctionsWithRepoQuery = `SELECT extracted_functions.id, r.name, r.commit_id, extracted_functions.path, extracted_functions.start_line, extracted_functions.end_line
FROM extracted_functions
LEFT JOIN repos r ON r.id = extracted_functions.repo_id
WHERE extracted_functions.id = ANY ($1)`

func GetExtractedFunctionsByID(ctx context.Context, conn *pgx.Conn, ids []int) ([]*HighlightedExtractedFunction, error) {
	rows, err := conn.Query(ctx, extractedFunctionsWithRepoQuery, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fns, err := database.ScanRows(ctx, rows, func(rows pgx.Rows) (*HighlightedExtractedFunction, error) {
		hef := &HighlightedExtractedFunction{}
		err := rows.Scan(
			&hef.ID,
			&hef.RepositoryName,
			&hef.CommitID,
			&hef.FilePath,
			&hef.StartLine,
			&hef.EndLine,
		)
		if err != nil {
			return nil, err
		}
		hef.URL = fmt.Sprintf("https://sourcegraph.com/%s@%s/-/blob/%s?L%d-%d", hef.RepositoryName, hef.CommitID, hef.FilePath, hef.StartLine+1, hef.EndLine+1)
		return hef, nil
	})
	if err != nil {
		return nil, err
	}

	idToFunction := map[int]*HighlightedExtractedFunction{}
	for _, f := range fns {
		idToFunction[f.ID] = f
	}

	orderedFunctions := make([]*HighlightedExtractedFunction, 0, len(ids))
	for _, id := range ids {
		orderedFunctions = append(orderedFunctions, idToFunction[id])
	}
	return orderedFunctions, nil
}
