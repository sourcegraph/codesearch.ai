package database

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v4"
)

const SCHEMA_UP = `
CREATE TABLE so_questions (
    id bigint NOT NULL PRIMARY KEY,
    title text NOT NULL,
    tags text NOT NULL,
    score integer NOT NULL,
    accepted_answer_id integer,
    creation_date text NOT NULL,
    last_edit_date text NOT NULL
);

CREATE TABLE so_answers (
    id bigint NOT NULL PRIMARY KEY,
    body text NOT NULL,
    score integer NOT NULL,
    parent_id integer NOT NULL,
    creation_date text NOT NULL,
    last_edit_date text NOT NULL
);

CREATE INDEX so_answers_parent_id_idx ON so_answers USING btree (parent_id);

CREATE TABLE repos (
    id bigserial NOT NULL PRIMARY KEY,
    commit_id text NOT NULL,
    name text NOT NULL UNIQUE,
    is_train bool NOT NULL DEFAULT false
);

CREATE TABLE extracted_functions (
    id bigserial NOT NULL PRIMARY KEY,
    path text NOT NULL,
    docstring text NOT NULL,
    inline_comments text NOT NULL,
    clean_code text NOT NULL,
    clean_code_hash text NOT NULL UNIQUE,
	token_counts jsonb NOT NULL,
    identifier text NOT NULL,
    start_line integer NOT NULL,
    end_line integer NOT NULL,
    repo_id integer NOT NULL,

    CONSTRAINT extracted_functions_repo_fk FOREIGN KEY (repo_id) REFERENCES repos (id) ON DELETE CASCADE
);

CREATE INDEX extracted_functions_repo_id_idx ON extracted_functions USING btree (repo_id);

CREATE TABLE code_query_pairs (
    id bigserial NOT NULL PRIMARY KEY,
    code text NOT NULL,
	code_hash text NOT NULL UNIQUE,
    query text NOT NULL,
    is_train bool NOT NULL DEFAULT false,
	token_counts jsonb,
    so_question_id integer,
    extracted_function_id integer,

    CONSTRAINT code_query_pairs_so_question_id_fk FOREIGN KEY (so_question_id) REFERENCES so_questions (id) ON DELETE SET NULL,

    CONSTRAINT code_query_pairs_extracted_function_id_fk FOREIGN KEY (extracted_function_id) REFERENCES extracted_functions (id) ON DELETE SET NULL
);

CREATE INDEX code_query_pairs_so_question_id_idx ON code_query_pairs USING btree (so_question_id);

CREATE INDEX code_query_pairs_extracted_function_id_idx ON code_query_pairs USING btree (extracted_function_id);
`

const SCHEMA_DOWN = `
DROP TABLE code_query_pairs;
DROP TABLE so_questions;
DROP TABLE so_answers;
DROP TABLE extracted_functions;
DROP TABLE repos;
`

func InitializeDatabaseSchema(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, SCHEMA_UP)
	return err
}

func ResetDatabaseSchema(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, SCHEMA_DOWN)
	return err
}

func ConnectToDatabase(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, os.Getenv("CODESEARCH_AI_DATA_DATABASE_URL"))
}

func PrepareValuesForBulkInsert[T any](s []*T, argsPerValue int, appendValueArgs func(valueArgs []any, value *T) []any) (string, []any) {
	insertValuesParameters := make([]string, 0, len(s))
	insertValueParametersBuffer := make([]string, argsPerValue)
	flatValuesArgs := make([]any, 0, len(s)*argsPerValue)

	for idx, value := range s {
		offset := idx * argsPerValue
		for i := 0; i < argsPerValue; i++ {
			insertValueParametersBuffer[i] = fmt.Sprintf("$%d", offset+i+1)
		}
		insertValuesParameters = append(insertValuesParameters, fmt.Sprintf("(%s)", strings.Join(insertValueParametersBuffer, ",")))
		flatValuesArgs = appendValueArgs(flatValuesArgs, value)
	}

	return strings.Join(insertValuesParameters, ","), flatValuesArgs
}

func GetRowsPage[T any](ctx context.Context, conn *pgx.Conn, baseQuery string, baseCondition string, groupByColumn string, idColumn string, afterID int, pageSize int, scanRow func(rows pgx.Rows) (*T, error)) ([]*T, error) {
	conditionClause := fmt.Sprintf("WHERE %s > $1", idColumn)
	if baseCondition != "" {
		conditionClause += fmt.Sprintf(" AND (%s)", baseCondition)
	}
	groupByClause := ""
	if groupByColumn != "" {
		groupByClause = fmt.Sprintf("GROUP BY %s", groupByColumn)
	}
	orderByClause := fmt.Sprintf("ORDER BY %s ASC", idColumn)
	query := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", baseQuery, conditionClause, groupByClause, orderByClause, "LIMIT $2")
	rows, err := conn.Query(ctx, query, afterID, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return ScanRows(ctx, rows, scanRow)
}

func ScanRows[T any](ctx context.Context, rows pgx.Rows, scanRow func(rows pgx.Rows) (*T, error)) ([]*T, error) {
	scannedRows := []*T{}
	for rows.Next() {
		row, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		scannedRows = append(scannedRows, row)
	}
	return scannedRows, nil
}

type Paginator[T any] struct {
	Conn          *pgx.Conn
	AfterID       int
	PageSize      int
	BaseQuery     string
	BaseCondition string
	GroupByColumn string
	IDColumn      string
	ScanRow       func(rows pgx.Rows) (*T, error)
	GetRowID      func(row *T) int
	err           error
}

func (p *Paginator[T]) Next(ctx context.Context) []*T {
	rows, err := GetRowsPage(ctx, p.Conn, p.BaseQuery, p.BaseCondition, p.GroupByColumn, p.IDColumn, p.AfterID, p.PageSize, p.ScanRow)
	if err != nil {
		p.err = err
		return nil
	}
	if len(rows) > 0 {
		p.AfterID = p.GetRowID(rows[len(rows)-1])
	}
	return rows
}

func (p *Paginator[T]) Error() error {
	return p.err
}
