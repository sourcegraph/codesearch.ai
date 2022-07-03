package main

import (
	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/functionextractor"
	"context"
	"flag"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	rand.Seed(0)

	repoName := flag.String("repo-name", "", "Name of the repository to process")
	repoNamesFilePath := flag.String("repo-names-file", "", "Path to the repo names file")
	nWorkers := flag.Int("n-workers", 4, "Number of workers to process the repo names")
	debug := flag.Bool("debug", false, "Enable debug logging")

	flag.Parse()

	if debug != nil && *debug {
		log.SetLevel(log.DebugLevel)
	}

	ctx := context.Background()

	if repoName != nil && *repoName != "" {
		conn, err := database.ConnectToDatabase(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close(ctx)

		err = functionextractor.ProcessRepo(ctx, conn, *repoName)
		if err != nil {
			log.Fatal(err)
		}
	} else if repoNamesFilePath != nil && *repoNamesFilePath != "" {
		processRepoListFile(ctx, *nWorkers, *repoNamesFilePath)
	} else {
		log.Fatal("Provide a valid -repo-name or a valid -repo-list-path command line arguments")
	}
}

func processRepoListFile(ctx context.Context, nWorkers int, repoNamesFilePath string) {
	repoJobs := make(chan string, nWorkers)

	wg := &sync.WaitGroup{}
	for w := 0; w < nWorkers; w++ {
		wg.Add(1)
		go repoWorker(ctx, repoJobs, wg)
	}

	go func() {
		repoNamesFile, err := ioutil.ReadFile(repoNamesFilePath)
		if err != nil {
			log.Fatal(err)
		}
		repoNames := strings.Split(string(repoNamesFile), "\n")
		for _, repoName := range repoNames {
			if len(strings.TrimSpace(repoName)) == 0 {
				continue
			}
			repoJobs <- repoName
		}
		close(repoJobs)
	}()

	wg.Wait()
}

func repoWorker(ctx context.Context, repoJobs <-chan string, wg *sync.WaitGroup) {
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		conn.Close(ctx)
		wg.Done()
	}()

	for repoName := range repoJobs {
		time.Sleep(time.Duration(1+rand.Intn(10)) * time.Second)
		log.Infof("Started processing %s", repoName)
		err := functionextractor.ProcessRepo(ctx, conn, repoName)
		if err != nil {
			log.Error(err)
		}
	}
}
