package functionextractor

import (
	"codesearch-ai-data/internal/database"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
)

func countExtractedFunctions(ctx context.Context, conn *pgx.Conn) (int, error) {
	var extractedFunctionsCount int
	err := conn.QueryRow(ctx, "SELECT COUNT(*) FROM extracted_functions").Scan(&extractedFunctionsCount)
	if err != nil {
		return -1, err
	}
	return extractedFunctionsCount, nil
}

func getExtractedFunctionDocstringAndInlineComments(ctx context.Context, conn *pgx.Conn, cleanCodeHash string) (string, string, error) {
	row := conn.QueryRow(ctx, "SELECT docstring, inline_comments FROM extracted_functions WHERE clean_code_hash = $1", cleanCodeHash)
	var docstring, inlineComments string
	err := row.Scan(
		&docstring,
		&inlineComments,
	)
	return docstring, inlineComments, err
}

func TestInsertingDuplicateExtractedFunctions(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("CODESEARCH_AI_DATA_TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal("Unable to connect to database", err)
	}

	err = database.InitializeDatabaseSchema(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := database.ResetDatabaseSchema(ctx, conn)
		if err != nil {
			t.Fatal(err)
		}
		err = conn.Close(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	var repoID int
	repoID, err = insertRepo(ctx, conn, "Test", "commit")
	if err != nil {
		t.Fatal(err)
	}

	extractedFunctionsBatchWithDuplicates := []*ExtractedFunction{
		{CleanCode: "a", CleanCodeHash: "a", Docstring: "", InlineComments: ""},
		{CleanCode: "a", CleanCodeHash: "a", Docstring: "d1", InlineComments: "il1"},
		{CleanCode: "b", CleanCodeHash: "b", Docstring: "", InlineComments: ""},
		{CleanCode: "a", CleanCodeHash: "a", Docstring: "d2", InlineComments: "il2"},
	}

	err = insertExtractedFunctionsFromFile(ctx, conn, repoID, "/path", extractedFunctionsBatchWithDuplicates)
	if err != nil {
		t.Fatal(err)
	}

	var extractedFunctionsCount int
	extractedFunctionsCount, err = countExtractedFunctions(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}

	if extractedFunctionsCount != 2 {
		t.Fatalf("Expected 2 extracted functions, got %d", extractedFunctionsCount)
	}

	var docstring, inlineComments string
	docstring, inlineComments, err = getExtractedFunctionDocstringAndInlineComments(ctx, conn, "a")
	if err != nil {
		t.Fatal(err)
	}

	if docstring != "d1 d2" || inlineComments != "il1 il2" {
		t.Fatalf("Expected docstring to be `d1 d2`, got `%s`. Expected inline comments to be `il1 il2`, got `%s`", docstring, inlineComments)
	}

	extractedFunctionsBatchWithDatabaseDuplicates := []*ExtractedFunction{
		{CleanCode: "a", CleanCodeHash: "a", Docstring: "d3", InlineComments: "il3"},
	}

	err = insertExtractedFunctionsFromFile(ctx, conn, repoID, "/path", extractedFunctionsBatchWithDatabaseDuplicates)
	if err != nil {
		t.Fatal(err)
	}

	docstring, inlineComments, err = getExtractedFunctionDocstringAndInlineComments(ctx, conn, "a")
	if err != nil {
		t.Fatal(err)
	}

	if docstring != "d1 d2 d3" || inlineComments != "il1 il2 il3" {
		t.Fatalf("Expected docstring to be `d1 d2 d3`, got `%s`. Expected inline comments to be `il1 il2 il3`, got `%s`", docstring, inlineComments)
	}
}

func TestInsertingLargeExtractedFunctionsBatch(t *testing.T) {
	// TODO: Refactor in subtests with above test
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("CODESEARCH_AI_DATA_TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal("Unable to connect to database", err)
	}

	err = database.InitializeDatabaseSchema(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := database.ResetDatabaseSchema(ctx, conn)
		if err != nil {
			t.Fatal(err)
		}
		err = conn.Close(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	var repoID int
	repoID, err = insertRepo(ctx, conn, "Test", "commit")
	if err != nil {
		t.Fatal(err)
	}

	nExtractedFunctions := 100
	extractedFunctions := make([]*ExtractedFunction, 0, nExtractedFunctions)
	for i := 0; i < nExtractedFunctions; i++ {
		extractedFunctions = append(extractedFunctions, &ExtractedFunction{CleanCodeHash: fmt.Sprintf("%d", i)})
	}

	err = insertExtractedFunctionsFromFile(ctx, conn, repoID, "/path", extractedFunctions)
	if err != nil {
		t.Fatal(err)
	}

	var extractedFunctionsCount int
	extractedFunctionsCount, err = countExtractedFunctions(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}

	if extractedFunctionsCount != nExtractedFunctions {
		t.Fatalf("Expected %d extracted functions, got %d", nExtractedFunctions, extractedFunctionsCount)
	}
}
