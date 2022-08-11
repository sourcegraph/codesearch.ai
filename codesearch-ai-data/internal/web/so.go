package web

import (
	"codesearch-ai-data/internal/database"
	"codesearch-ai-data/internal/socode"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

const TIMESTAMP_LAYOUT = "2006-01-02T15:04:05.000"

type SOQuestionWithAnswers struct {
	ID           int         `json:"id"`
	Title        string      `json:"title"`
	Tags         string      `json:"tags"`
	CreationDate string      `json:"creationDate"`
	Score        int         `json:"score"`
	Answers      []*SOAnswer `json:"answers"`
	URL          string      `json:"url"`
}

type SOAnswer struct {
	ID           int    `json:"id"`
	Body         string `json:"body"`
	Score        int    `json:"score"`
	CreationDate string `json:"creation_date"`
}

const soQuestionsWithAnswersQuery = `SELECT so_questions.id, so_questions.title, so_questions.tags, so_questions.score, so_questions.creation_date, json_agg(sa order by sa.score desc)
FROM so_questions
LEFT JOIN so_answers sa on so_questions.id = sa.parent_id
WHERE so_questions.id = ANY($1)
GROUP BY so_questions.id`

func GetSOQuestionsWithAnswersByID(ctx context.Context, conn *pgx.Conn, ids []int) ([]*SOQuestionWithAnswers, error) {
	rows, err := conn.Query(ctx, soQuestionsWithAnswersQuery, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	qs, err := database.ScanRows(ctx, rows, func(rows pgx.Rows) (*SOQuestionWithAnswers, error) {
		sq := &SOQuestionWithAnswers{}
		err := rows.Scan(
			&sq.ID,
			&sq.Title,
			&sq.Tags,
			&sq.Score,
			&sq.CreationDate,
			&sq.Answers,
		)
		if err != nil {
			return nil, err
		}
		sq.URL = fmt.Sprintf("https://stackoverflow.com/questions/%d", sq.ID)
		timestamp, err := time.Parse(TIMESTAMP_LAYOUT, sq.CreationDate)
		// Ignore if we can't parse the timestamp
		if err == nil {
			sq.CreationDate = timestamp.Format("Jan 02, 2006")
		}

		answers := sq.Answers[:0]
		for _, answer := range sq.Answers {
			if answer == nil {
				continue
			}
			timestamp, err := time.Parse(TIMESTAMP_LAYOUT, answer.CreationDate)
			// Ignore if we can't parse the timestamp
			if err == nil {
				answer.CreationDate = timestamp.Format("Jan 02, 2006")
			}
			answer.Body = socode.EscapeCodeSnippetsInHTML(answer.Body)
			answers = append(answers, answer)
		}
		sq.Answers = answers
		return sq, nil
	})
	if err != nil {
		return nil, err
	}

	idToQuestion := map[int]*SOQuestionWithAnswers{}
	for _, q := range qs {
		idToQuestion[q.ID] = q
	}

	orderedQuestions := make([]*SOQuestionWithAnswers, 0, len(ids))
	for _, id := range ids {
		orderedQuestions = append(orderedQuestions, idToQuestion[id])
	}
	return orderedQuestions, nil
}
