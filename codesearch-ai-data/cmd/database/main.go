package main

import (
	"codesearch-ai-data/internal/database"
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
)

func main() {
	initializeSchema := flag.Bool("init", false, "Initialize database schema")
	resetSchema := flag.Bool("reset", false, "Reset database schema")

	flag.Parse()

	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := conn.Close(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	if initializeSchema != nil && *initializeSchema {
		err = database.InitializeDatabaseSchema(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
	} else if resetSchema != nil && *resetSchema {
		err = database.ResetDatabaseSchema(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
	}
}
