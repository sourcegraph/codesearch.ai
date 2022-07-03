package codequerypairsimporter

import (
	"codesearch-ai-data/internal/database"
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func importCodeQueryPairs(ctx context.Context, conn *pgx.Conn, pairs []*CodeQueryPair) error {
	if len(pairs) == 0 {
		return nil
	}

	codes := map[string]bool{}
	deduplicatedPairs := []*CodeQueryPair{}
	for _, pair := range pairs {
		if codes[pair.CodeHash] {
			continue
		}
		deduplicatedPairs = append(deduplicatedPairs, pair)
		codes[pair.CodeHash] = true
	}

	insertValuesParameters, valuesArgs := database.PrepareValuesForBulkInsert(deduplicatedPairs, 6, func(valueArgs []any, cqp *CodeQueryPair) []any {
		return append(valueArgs, cqp.Code, cqp.CodeHash, cqp.Query, cqp.IsTrain, cqp.SOQuestionID, cqp.ExtractedFunctionID)
	})

	_, err := conn.Exec(
		ctx,
		fmt.Sprintf("INSERT INTO code_query_pairs (code, code_hash, query, is_train, so_question_id, extracted_function_id) VALUES %s ON CONFLICT (code_hash) DO NOTHING", insertValuesParameters),
		valuesArgs...,
	)
	return err
}
