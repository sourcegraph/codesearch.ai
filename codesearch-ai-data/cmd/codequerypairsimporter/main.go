package main

import (
	cqpi "codesearch-ai-data/internal/codequerypairsimporter"
	"codesearch-ai-data/internal/database"
	"context"
	"flag"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

func main() {
	rand.Seed(0)

	importSO := flag.Bool("so", false, "Import SO questions")
	importExtractedFunctions := flag.Bool("extracted-functions", false, "Import extracted functions")
	soTrainTestRatio := flag.Float64("so-train-test-ratio", 0.95, "SO train test ratio")

	flag.Parse()

	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal("Unable to connect to database", err)
	}
	defer func() {
		err := conn.Close(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	if *importSO {
		log.Info("Importing StackOverflow code query pairs")
		err = cqpi.ImportSOCodeQueryPairs(ctx, conn, *soTrainTestRatio)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *importExtractedFunctions {
		log.Info("Importing extracted functions code query pairs")
		err = cqpi.ImportExtractedFunctionsCodeQueryPairs(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
	}
}
