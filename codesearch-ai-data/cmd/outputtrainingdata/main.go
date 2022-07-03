package main

import (
	cqpi "codesearch-ai-data/internal/codequerypairsimporter"
	"codesearch-ai-data/internal/database"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4"
)

type codeQueryPairsOptions struct {
	IsTrain                *bool
	SOOnly                 bool
	ExtractedFunctionsOnly bool
}

func (o *codeQueryPairsOptions) Condition() string {
	conds := []string{}
	if o.IsTrain != nil {
		if *o.IsTrain {
			conds = append(conds, "is_train")
		} else {
			conds = append(conds, "not is_train")
		}
	}
	if o.SOOnly {
		conds = append(conds, "so_question_id is not null")
	} else if o.ExtractedFunctionsOnly {
		conds = append(conds, "extracted_function_id is not null")
	}

	if len(conds) == 0 {
		return "1=1"
	}
	return strings.Join(conds, " AND ")
}

func newCodeQueryPairsPaginator(conn *pgx.Conn, pageSize int, options *codeQueryPairsOptions) *database.Paginator[cqpi.CodeQueryPair] {
	return &database.Paginator[cqpi.CodeQueryPair]{
		Conn:          conn,
		AfterID:       0,
		PageSize:      pageSize,
		BaseQuery:     "SELECT id, code, query, so_question_id, extracted_function_id FROM code_query_pairs",
		BaseCondition: options.Condition(),
		IDColumn:      "id",
		ScanRow: func(rows pgx.Rows) (*cqpi.CodeQueryPair, error) {
			cqp := &cqpi.CodeQueryPair{}
			err := rows.Scan(
				&cqp.ID,
				&cqp.Code,
				&cqp.Query,
				&cqp.SOQuestionID,
				&cqp.ExtractedFunctionID,
			)
			if err != nil {
				return nil, err
			}
			return cqp, nil
		},
		GetRowID: func(row *cqpi.CodeQueryPair) int { return row.ID },
	}
}

func outputCodeQueryPairsToFile(ctx context.Context, conn *pgx.Conn, options *codeQueryPairsOptions, outputPath string) error {
	fo, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	zero := 0
	paginator := newCodeQueryPairsPaginator(conn, 100_000, options)
	page := paginator.Next(ctx)
	newline := []byte("\n")
	for len(page) > 0 {
		for _, cqp := range page {
			// Replace nils with zero, to have consistent data types in JSON.
			if cqp.ExtractedFunctionID == nil {
				cqp.ExtractedFunctionID = &zero
			}
			if cqp.SOQuestionID == nil {
				cqp.SOQuestionID = &zero
			}
			b, err := json.Marshal(cqp)
			if err != nil {
				return err
			}
			fo.Write(b)
			fo.Write(newline)
		}
		page = paginator.Next(ctx)
	}
	return nil
}

func main() {
	outputTrain := flag.Bool("train", false, "Output train")
	outputTest := flag.Bool("test", false, "Output test")
	outputSO := flag.Bool("so", false, "Output SO questions")
	outputExtractedFunctions := flag.Bool("extracted-functions", false, "Output extracted functions")
	outputDirectory := flag.String("output-directory", "/tmp", "Output directory for the training files")

	flag.Parse()

	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	t := true
	f := false
	if *outputTrain {
		log.Info("Outputting train.jsonl file")
		err = outputCodeQueryPairsToFile(ctx, conn, &codeQueryPairsOptions{IsTrain: &t}, path.Join(*outputDirectory, "train.jsonl"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if *outputTest {
		log.Info("Outputting test.jsonl file")
		err = outputCodeQueryPairsToFile(ctx, conn, &codeQueryPairsOptions{IsTrain: &f}, path.Join(*outputDirectory, "test.jsonl"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if *outputSO {
		log.Info("Outputting so.jsonl file")
		err = outputCodeQueryPairsToFile(ctx, conn, &codeQueryPairsOptions{SOOnly: true}, path.Join(*outputDirectory, "so.train.jsonl"))
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Outputting so.train.jsonl file")
		err = outputCodeQueryPairsToFile(ctx, conn, &codeQueryPairsOptions{SOOnly: true, IsTrain: &t}, path.Join(*outputDirectory, "so.train.jsonl"))
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Outputting so.test.jsonl file")
		err = outputCodeQueryPairsToFile(ctx, conn, &codeQueryPairsOptions{SOOnly: true, IsTrain: &f}, path.Join(*outputDirectory, "so.test.jsonl"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if *outputExtractedFunctions {
		log.Info("Outputting extracted-functions.jsonl file")
		err = outputCodeQueryPairsToFile(ctx, conn, &codeQueryPairsOptions{ExtractedFunctionsOnly: true}, path.Join(*outputDirectory, "extracted-functions.jsonl"))
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Outputting extracted-functions.train.jsonl file")
		err = outputCodeQueryPairsToFile(ctx, conn, &codeQueryPairsOptions{ExtractedFunctionsOnly: true, IsTrain: &t}, path.Join(*outputDirectory, "extracted-functions.train.jsonl"))
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Outputting extracted-functions.test.jsonl file")
		err = outputCodeQueryPairsToFile(ctx, conn, &codeQueryPairsOptions{ExtractedFunctionsOnly: true, IsTrain: &f}, path.Join(*outputDirectory, "extracted-functions.test.jsonl"))
		if err != nil {
			log.Fatal(err)
		}
	}
}
