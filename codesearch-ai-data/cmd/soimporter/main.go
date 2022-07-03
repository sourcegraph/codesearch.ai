package main

import (
	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/soimporter"
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
)

func main() {
	postsXmlPath := flag.String("posts-xml-path", "", "Path to the StackOverflow Posts.xml file")

	flag.Parse()

	if postsXmlPath == nil || *postsXmlPath == "" {
		log.Fatal("Command line argument posts-xml-path is not valid.")
	}

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

	err = soimporter.Import(ctx, conn, *postsXmlPath)
	if err != nil {
		log.Fatal(err)
	}
}
