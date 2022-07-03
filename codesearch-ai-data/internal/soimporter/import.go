package soimporter

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"

	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"codesearch-ai-data/internal/database"

	"github.com/jackc/pgx/v4"
)

const TIMESTAMP_LAYOUT = "2006-01-02T15:04:05.000"
const MAX_LINE_LENGTH = 1024 * 1024
const BATCH_SIZE = 1024

func Import(ctx context.Context, conn *pgx.Conn, postsXmlPath string) error {
	if _, err := os.Stat(postsXmlPath); errors.Is(err, os.ErrNotExist) {
		return err
	}

	file, err := os.Open(postsXmlPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, MAX_LINE_LENGTH)
	scanner.Buffer(buf, MAX_LINE_LENGTH)

	questionsBuffer := make([]*SOQuestion, 0, BATCH_SIZE)
	answersBuffer := make([]*SOAnswer, 0, BATCH_SIZE)

	rowNumber := 0
	for scanner.Scan() {
		line := scanner.Text()

		rowNumber++
		if rowNumber%1_000_000 == 0 {
			log.Infof("Processed row number %d", rowNumber)
		}

		if !strings.HasPrefix(line, "  <row") {
			continue
		}

		// Unmarshal the XML row and escape HTML encoded fields
		var row SOPostRow
		err := xml.Unmarshal([]byte(line), &row)
		if err != nil {
			log.Fatal(err)
		}
		row.Body = html.UnescapeString(row.Body)
		if row.Tags != nil {
			unescapedTags := html.UnescapeString(*row.Tags)
			row.Tags = &unescapedTags
		}

		// Skip questions and answers with a negative score
		if row.Score < 0 {
			continue
		}

		if row.PostTypeID == 1 {
			// Skip questions with no answers
			if intOrZero(row.AnswerCount) == 0 {
				continue
			}

			if len(questionsBuffer) == BATCH_SIZE {
				err = importQuestions(ctx, conn, questionsBuffer)
				if err != nil {
					return err
				}
				questionsBuffer = questionsBuffer[:0]
			}

			questionsBuffer = append(questionsBuffer, &SOQuestion{
				ID:               row.ID,
				Title:            stringOrEmpty(row.Title),
				Tags:             stringOrEmpty(row.Tags),
				Score:            row.Score,
				AcceptedAnswerID: row.AcceptedAnswerID,
				CreationDate:     row.CreationDate,
				LastEditDate:     row.LastEditDate,
			})
		} else if row.PostTypeID == 2 {
			// Sanity check, skip answer if it doesn't have a parent question
			if row.ParentID == nil {
				continue
			}

			if len(answersBuffer) == BATCH_SIZE {
				err = importAnswers(ctx, conn, answersBuffer)
				if err != nil {
					return err
				}
				answersBuffer = answersBuffer[:0]
			}

			answersBuffer = append(answersBuffer, &SOAnswer{
				ID:           row.ID,
				Body:         row.Body,
				Score:        row.Score,
				ParentID:     *row.ParentID,
				CreationDate: row.CreationDate,
				LastEditDate: row.LastEditDate,
			})
		}
	}

	err = importQuestions(ctx, conn, questionsBuffer)
	if err != nil {
		return err
	}

	err = importAnswers(ctx, conn, answersBuffer)
	if err != nil {
		return err
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func importQuestions(ctx context.Context, conn *pgx.Conn, questions []*SOQuestion) error {
	if len(questions) == 0 {
		return nil
	}

	insertValuesParameters, valuesArgs := database.PrepareValuesForBulkInsert(questions, 7, func(valueArgs []any, question *SOQuestion) []any {
		return append(valueArgs, question.ID, question.Title, question.Tags, question.Score, question.AcceptedAnswerID, question.CreationDate, question.LastEditDate)
	})

	_, err := conn.Exec(
		ctx,
		fmt.Sprintf("INSERT INTO so_questions (id, title, tags, score, accepted_answer_id, creation_date, last_edit_date) VALUES %s", insertValuesParameters),
		valuesArgs...,
	)
	return err
}

func importAnswers(ctx context.Context, conn *pgx.Conn, answers []*SOAnswer) error {
	if len(answers) == 0 {
		return nil
	}

	insertValuesParameters, valuesArgs := database.PrepareValuesForBulkInsert(answers, 6, func(valueArgs []any, answer *SOAnswer) []any {
		return append(valueArgs, answer.ID, answer.Body, answer.Score, answer.ParentID, answer.CreationDate, answer.LastEditDate)
	})

	_, err := conn.Exec(
		ctx,
		fmt.Sprintf("INSERT INTO so_answers (id, body, score, parent_id, creation_date, last_edit_date) VALUES %s", insertValuesParameters),
		valuesArgs...,
	)
	return err
}

func intOrZero(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
