package main

import (
	"codesearch-ai-data/internal/database"
	"context"
	"html/template"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	conn, err := database.ConnectToDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	http.HandleFunc("/inspect/extracted-functions", inspectExtractedFunctionsHandler(ctx, conn))
	http.HandleFunc("/inspect/so-questions", inspectSOQuestionsHandler(ctx, conn))
	http.HandleFunc("/inspect/code-query-pairs", inspectCodeQueryPairsHandler(ctx, conn))

	log.Info("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

type StoredFunction struct {
	ID             int
	Path           string
	Docstring      string
	InlineComments string
	CleanCode      string
	Identifier     string
}

const inspectExtractedFunctionsQuery = "SELECT id, path, docstring, inline_comments, clean_code, identifier FROM extracted_functions"

func inspectExtractedFunctionsHandler(ctx context.Context, conn *pgx.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		after, pageSize := getAfterAndPageSize(r)
		efs, err := database.GetRowsPage(ctx, conn, inspectExtractedFunctionsQuery, "", "", "id", after, pageSize, func(rows pgx.Rows) (*StoredFunction, error) {
			sf := &StoredFunction{}
			err := rows.Scan(
				&sf.ID,
				&sf.Path,
				&sf.Docstring,
				&sf.InlineComments,
				&sf.CleanCode,
				&sf.Identifier,
			)
			if err != nil {
				return nil, err
			}
			return sf, nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("templates/inspect/extracted-functions.template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextAfter := 0
		if len(efs) > 0 {
			nextAfter = efs[len(efs)-1].ID
		}

		data := struct {
			StoredFunctions []*StoredFunction
			After           int
			PageSize        int
		}{
			StoredFunctions: efs,
			After:           nextAfter,
			PageSize:        pageSize,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type StoredSOQuestion struct {
	ID      int
	Title   string
	Tags    string
	Answers []template.HTML
}

const inspectSOQuestionsQuery = `SELECT so_questions.id, so_questions.title, so_questions.tags, array_agg(sa.body order by sa.score desc)::text[]
FROM so_questions
LEFT JOIN so_answers sa on so_questions.id = sa.parent_id`

func inspectSOQuestionsHandler(ctx context.Context, conn *pgx.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		after, pageSize := getAfterAndPageSize(r)
		sqs, err := database.GetRowsPage(ctx, conn, inspectSOQuestionsQuery, "", "so_questions.id", "so_questions.id", after, pageSize, func(rows pgx.Rows) (*StoredSOQuestion, error) {
			sq := &StoredSOQuestion{}
			var answers []*string
			err := rows.Scan(
				&sq.ID,
				&sq.Title,
				&sq.Tags,
				&answers,
			)
			if err != nil {
				return nil, err
			}
			if answers != nil {
				htmlAnswers := []template.HTML{}
				for _, answer := range answers {
					if answer != nil {
						htmlAnswers = append(htmlAnswers, template.HTML(*answer))
					}
				}
				sq.Answers = htmlAnswers
			}
			return sq, nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("templates/inspect/so-questions.template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextAfter := 0
		if len(sqs) > 0 {
			nextAfter = sqs[len(sqs)-1].ID
		}

		data := struct {
			StoredQuestions []*StoredSOQuestion
			After           int
			PageSize        int
		}{
			StoredQuestions: sqs,
			After:           nextAfter,
			PageSize:        pageSize,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type StoredCodeQueryPair struct {
	ID    int
	Code  string
	Query string
}

const inspectCodeQueryPairsQuery = `SELECT id, code, query FROM code_query_pairs`

func inspectCodeQueryPairsHandler(ctx context.Context, conn *pgx.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		after, pageSize := getAfterAndPageSize(r)
		cqps, err := database.GetRowsPage(ctx, conn, inspectCodeQueryPairsQuery, "", "", "id", after, pageSize, func(rows pgx.Rows) (*StoredCodeQueryPair, error) {
			cqp := &StoredCodeQueryPair{}
			err := rows.Scan(
				&cqp.ID,
				&cqp.Code,
				&cqp.Query,
			)
			if err != nil {
				return nil, err
			}
			return cqp, nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("templates/inspect/code-query-pairs.template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextAfter := 0
		if len(cqps) > 0 {
			nextAfter = cqps[len(cqps)-1].ID
		}

		data := struct {
			CodeQueryPairs []*StoredCodeQueryPair
			After          int
			PageSize       int
		}{
			CodeQueryPairs: cqps,
			After:          nextAfter,
			PageSize:       pageSize,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getAfterAndPageSize(r *http.Request) (int, int) {
	query := r.URL.Query()
	var after, pageSize int
	if query.Get("after") != "" {
		after, _ = strconv.Atoi(query.Get("after"))
	}
	if query.Get("pageSize") != "" {
		pageSize, _ = strconv.Atoi(query.Get("pageSize"))
	} else {
		pageSize = 30
	}
	return after, pageSize
}
