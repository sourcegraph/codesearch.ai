package main

import (
	"codesearch-ai-data/internal/database"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4"
)

type extractedFunctionsPerRepoCount struct {
	RepoID int
	Count  int
}

func newExtractedFunctionsPerRepoCountPaginator(conn *pgx.Conn, pageSize int) *database.Paginator[extractedFunctionsPerRepoCount] {
	return &database.Paginator[extractedFunctionsPerRepoCount]{
		Conn:          conn,
		AfterID:       0,
		PageSize:      pageSize,
		BaseQuery:     "SELECT repo_id, COUNT(*) FROM extracted_functions",
		GroupByColumn: "repo_id",
		IDColumn:      "repo_id",
		ScanRow: func(rows pgx.Rows) (*extractedFunctionsPerRepoCount, error) {
			efc := &extractedFunctionsPerRepoCount{}
			err := rows.Scan(
				&efc.RepoID,
				&efc.Count,
			)
			if err != nil {
				return nil, err
			}
			return efc, nil
		},
		GetRowID: func(row *extractedFunctionsPerRepoCount) int { return row.RepoID },
	}
}

func markTrainRepos(ctx context.Context, conn *pgx.Conn, repos []*extractedFunctionsPerRepoCount) error {
	length := len(repos)
	batchSize := 1024
	for i := 0; i < length; i += batchSize {
		end := i + batchSize
		if end > length {
			end = length
		}

		updateValuesParameters := []string{}
		valuesArgs := []interface{}{}
		for idx, repo := range repos[i:end] {
			valuesArgs = append(valuesArgs, repo.RepoID)
			updateValuesParameters = append(updateValuesParameters, fmt.Sprintf("$%d", idx+1))
		}

		_, err := conn.Exec(ctx, fmt.Sprintf("UPDATE repos SET is_train = true WHERE id IN (%s)", strings.Join(updateValuesParameters, ",")), valuesArgs...)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	rand.Seed(0)

	trainTestRatio := flag.Float64("train-test-ratio", 0.95, "Train test ratio")

	flag.Parse()

	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	efcs := []*extractedFunctionsPerRepoCount{}
	paginator := newExtractedFunctionsPerRepoCountPaginator(conn, 1024)
	page := paginator.Next(ctx)
	for len(page) > 0 {
		efcs = append(efcs, page...)
		page = paginator.Next(ctx)
	}
	err = paginator.Error()
	if err != nil {
		log.Fatal(err)
	}

	nExtractedFunctions := 0
	for _, efc := range efcs {
		nExtractedFunctions += efc.Count
	}
	nTrainSamples := int(*trainTestRatio * float64(nExtractedFunctions))

	rand.Shuffle(len(efcs), func(i, j int) {
		efcs[i], efcs[j] = efcs[j], efcs[i]
	})

	countsPrefixSum := make([]int, len(efcs))
	countsPrefixSum[0] = efcs[0].Count
	for i := 1; i < len(countsPrefixSum); i++ {
		countsPrefixSum[i] = countsPrefixSum[i-1] + efcs[i].Count
	}

	trainIdx := -1
	for idx, count := range countsPrefixSum {
		if count >= nTrainSamples {
			trainIdx = idx
			break
		}
	}

	if trainIdx == -1 {
		panic("trainIdx == -1")
	}

	trainSum := 0
	testSum := 0
	for idx, efc := range efcs {
		if idx < trainIdx {
			trainSum += efc.Count
		} else {
			testSum += efc.Count
		}
	}

	log.Infof("Num repos: %d, train idx: %d, train samples: %d, test samples: %d", len(efcs), trainIdx, trainSum, testSum)

	err = markTrainRepos(ctx, conn, efcs[:trainIdx])
	if err != nil {
		log.Fatal(err)
	}
}
