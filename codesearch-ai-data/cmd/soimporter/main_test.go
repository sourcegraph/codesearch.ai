package main

import (
	"context"
	"os"
	"testing"

	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/soimporter"

	"github.com/jackc/pgx/v4"
)

func TestPostsXmlFileImport(t *testing.T) {
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

	err = soimporter.Import(ctx, conn, "./testdata/Posts.xml")
	if err != nil {
		t.Fatal(err)
	}

	var questionsCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM so_questions").Scan(&questionsCount)
	if err != nil {
		t.Fatal(err)
	}

	if questionsCount != 6 {
		t.Fatalf("Expected 6 questions imported, got %d", questionsCount)
	}

	var answersCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM so_answers").Scan(&answersCount)
	if err != nil {
		t.Fatal(err)
	}

	if answersCount != 2 {
		t.Fatalf("Expected 2 answers imported, got %d", answersCount)
	}
}
