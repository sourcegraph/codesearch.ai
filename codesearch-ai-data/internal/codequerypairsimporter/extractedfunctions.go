package codequerypairsimporter

import (
	"codesearch-ai-data/internal/database"
	fe "codesearch-ai-data/internal/functionextractor"
	"context"
	"regexp"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/jackc/pgx/v4"
)

func newExtractedFunctionsPaginator(conn *pgx.Conn, pageSize int) *database.Paginator[fe.ExtractedFunction] {
	return &database.Paginator[fe.ExtractedFunction]{
		Conn:      conn,
		AfterID:   0,
		PageSize:  pageSize,
		BaseQuery: "SELECT extracted_functions.id, docstring, inline_comments, clean_code, identifier, is_train FROM extracted_functions JOIN repos r on r.id = extracted_functions.repo_id",
		IDColumn:  "extracted_functions.id",
		ScanRow: func(rows pgx.Rows) (*fe.ExtractedFunction, error) {
			ef := &fe.ExtractedFunction{}
			err := rows.Scan(
				&ef.ID,
				&ef.Docstring,
				&ef.InlineComments,
				&ef.CleanCode,
				&ef.Identifier,
				&ef.IsTrain,
			)
			if err != nil {
				return nil, err
			}
			return ef, nil
		},
		GetRowID: func(row *fe.ExtractedFunction) int { return row.ID },
	}
}

var DOT_UNDERSCORE_REGEX = regexp.MustCompile("[._]")
var IDENTIFIER_TOKEN_REGEX = regexp.MustCompile("[_a-zA-Z][_a-zA-Z0-9]*")

func identifierToDocstring(identifier string) string {
	if len(identifier) == 0 {
		return ""
	}

	parts := DOT_UNDERSCORE_REGEX.Split(identifier, -1)
	validIdentifierParts := []string{}
	for _, part := range parts {
		camelCaseParts := camelcase.Split(part)
		for _, camelCasePart := range camelCaseParts {
			camelCasePartTrimmed := strings.TrimSpace(camelCasePart)
			if len(camelCasePartTrimmed) == 0 {
				continue
			}
			if IDENTIFIER_TOKEN_REGEX.MatchString(camelCasePart) {
				validIdentifierParts = append(validIdentifierParts, camelCasePartTrimmed)
			}
		}
	}

	return strings.Join(validIdentifierParts, " ")
}

func extractedFunctionToCodeQueryPair(ef *fe.ExtractedFunction) *CodeQueryPair {
	docstring := removeNonAsciiChars(ef.Docstring)
	if len(docstring) == 0 {
		docstring = removeNonAsciiChars(identifierToDocstring(ef.Identifier) + " " + ef.InlineComments)
	}

	if strings.Count(docstring, " ") < 3 {
		docstring = ""
	}

	return newCodeQueryPair(
		ef.CleanCode,
		strings.TrimSpace(docstring),
		ef.IsTrain,
		nil,
		&ef.ID,
	)
}

func ImportExtractedFunctionsCodeQueryPairs(ctx context.Context, conn *pgx.Conn) error {
	extractedFunctionsPaginator := newExtractedFunctionsPaginator(conn, 100_000)
	extractedFunctionsPage := extractedFunctionsPaginator.Next(ctx)

	pairsBuffer := make([]*CodeQueryPair, 0, BATCH_SIZE)
	for len(extractedFunctionsPage) > 0 {
		for _, ef := range extractedFunctionsPage {
			pairsBuffer = append(pairsBuffer, extractedFunctionToCodeQueryPair(ef))

			if len(pairsBuffer) == BATCH_SIZE {
				err := importCodeQueryPairs(ctx, conn, pairsBuffer)
				if err != nil {
					return err
				}
				pairsBuffer = pairsBuffer[:0]
			}
		}

		extractedFunctionsPage = extractedFunctionsPaginator.Next(ctx)
	}

	if len(pairsBuffer) > 0 {
		err := importCodeQueryPairs(ctx, conn, pairsBuffer)
		if err != nil {
			return err
		}
	}

	return extractedFunctionsPaginator.Error()
}
