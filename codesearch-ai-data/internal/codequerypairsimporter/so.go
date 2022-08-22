package codequerypairsimporter

import (
	"codesearch-ai-data/internal/database"
	ph "codesearch-ai-data/internal/parsinghelpers"
	"codesearch-ai-data/internal/sitterparsers"
	"codesearch-ai-data/internal/socode"
	tc "codesearch-ai-data/internal/tokencounter"
	"context"
	"errors"
	"math/rand"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4"
	sitter "github.com/smacker/go-tree-sitter"
)

var soTagToLanguage = map[string]string{
	"java":       "java",
	"python":     "python",
	"php":        "php",
	"ruby":       "ruby",
	"javascript": "javascript",
	"go":         "go",
	"django":     "python",
	"jquery":     "javascript",
	"node.js":    "javascript",
	"reactjs":    "javascript",
	"spring":     "java",
	"laravel":    "php",
	"numpy":      "python",
}

// For things like python-3, python-2, ruby-on-rails, etc.
var checkTagPrefix = map[string]bool{
	"java":       false,
	"python":     true,
	"php":        false,
	"ruby":       true,
	"javascript": false,
	"go":         false,
	"django":     false,
	"jquery":     false,
	"node.js":    false,
	"reactjs":    false,
	"spring":     false,
	"laravel":    false,
	"numpy":      false,
}

func getLanguagesFromTags(tags []string) []string {
	languagesMap := map[string]bool{}
	for _, tag := range tags {
		tagLower := strings.ToLower(tag)
		for soTag, language := range soTagToLanguage {
			if checkTagPrefix[soTag] && strings.HasPrefix(tagLower, soTag) {
				languagesMap[language] = true
			} else if !checkTagPrefix[soTag] && tagLower == soTag {
				languagesMap[language] = true
			}
		}
	}

	languages := make([]string, 0, len(languagesMap))
	for language := range languagesMap {
		languages = append(languages, language)
	}
	return languages
}

type SOQuestionWithAnswers struct {
	ID      int
	Title   string
	Tags    string
	Answers []*string
}

func tryParse(ctx context.Context, parser *sitter.Parser, code []byte) (*sitter.Node, error) {
	tree, err := parser.ParseCtx(ctx, nil, code)
	if err != nil {
		return nil, err
	}
	rootNode := tree.RootNode()
	if rootNode.HasError() {
		return nil, errors.New("error encountered while parsing")
	}
	return rootNode, nil
}

func newSOQuestionsPaginator(conn *pgx.Conn, pageSize int) *database.Paginator[SOQuestionWithAnswers] {
	return &database.Paginator[SOQuestionWithAnswers]{
		Conn:          conn,
		AfterID:       0,
		PageSize:      pageSize,
		BaseQuery:     "SELECT so_questions.id, so_questions.title, so_questions.tags, array_agg(sa.body order by sa.score desc)::text[] FROM so_questions LEFT JOIN so_answers sa on so_questions.id = sa.parent_id",
		GroupByColumn: "so_questions.id",
		IDColumn:      "so_questions.id",
		ScanRow: func(rows pgx.Rows) (*SOQuestionWithAnswers, error) {
			q := &SOQuestionWithAnswers{}
			err := rows.Scan(
				&q.ID,
				&q.Title,
				&q.Tags,
				&q.Answers,
			)
			if err != nil {
				return nil, err
			}
			return q, nil
		},
		GetRowID: func(row *SOQuestionWithAnswers) int { return row.ID },
	}
}

func addPHPTagsIfMissing(codeText string) string {
	if !strings.HasPrefix(codeText, "<?php") {
		codeText = "<?php\n" + codeText
	}
	if !strings.HasSuffix(codeText, "?>") {
		codeText = codeText + "\n?>"
	}
	return codeText
}

func getCodeAnswers(ctx context.Context, answers []*string, languages []string) ([]string, tc.TokenCounter, error) {
	codeAnswers := map[string]struct{}{}
	codeAnswersDeduplicated := []string{}
	tokenCounts := tc.TokenCounter{}
	for _, answer := range answers {
		if answer == nil {
			continue
		}

		for _, codeSnippet := range socode.GetCodeSnippetsFromHTML(*answer) {
			codeLines := []string{}
			for _, line := range strings.Split(codeSnippet, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "..") {
					continue
				}
				codeLines = append(codeLines, line)
			}
			joinedCodeLines := strings.Join(codeLines, "\n")

			for _, language := range languages {
				codeText := joinedCodeLines
				if language == "php" {
					codeText = addPHPTagsIfMissing(codeText)
				}

				code := []byte(codeText)
				parser := sitterparsers.GetParserForLanguage(language)
				rootNode, err := tryParse(ctx, parser, code)
				if err != nil {
					continue
				}

				var skipNodeFn ph.SkipNodeFn = nil
				if language == "python" {
					docstringNodes, err := ph.GetPythonDocstringNodes(rootNode)
					if err != nil {
						continue
					}
					skipNodeFn = ph.SkipPythonDocstringNodesFn(docstringNodes)
				} else if language == "php" {
					skipNodeFn = func(node *sitter.Node) bool {
						nodeType := node.Type()
						return nodeType == "php_tag" || nodeType == "?>"
					}
				}

				filteredNodes, _ := ph.StripComments(rootNode, skipNodeFn)
				prettyFormattedCode := ph.PrettyFormatNodes(filteredNodes, code)

				if len(prettyFormattedCode) >= 10 {
					_, ok := codeAnswers[prettyFormattedCode]
					if !ok {
						codeAnswersDeduplicated = append(codeAnswersDeduplicated, prettyFormattedCode)
						tokenCounts.Extend(tc.CountTokens(rootNode, code))
					}
					codeAnswers[prettyFormattedCode] = struct{}{}
				}

				// Found a language to parse the code, we can exit early.
				break
			}
		}
	}
	return codeAnswersDeduplicated, tokenCounts, nil
}

func questionToCodeQueryPair(ctx context.Context, conn *pgx.Conn, question *SOQuestionWithAnswers, isTrain bool) (*CodeQueryPair, error) {
	title := strings.TrimSpace(removeNonAsciiChars(question.Title))
	if len(title) == 0 {
		return nil, nil
	}

	tags := strings.Split(strings.TrimPrefix(strings.TrimSuffix(question.Tags, ">"), "<"), "><")
	languages := getLanguagesFromTags(tags)
	if len(languages) == 0 {
		return nil, nil
	}

	codeAnswers, tokenCounter, err := getCodeAnswers(ctx, question.Answers, languages)
	if err != nil || len(codeAnswers) == 0 {
		return nil, err
	}

	codeAnswer := strings.Join(codeAnswers, "\n")
	return newCodeQueryPair(codeAnswer, title, isTrain, tokenCounter, &question.ID, nil), nil
}

func ImportSOCodeQueryPairs(ctx context.Context, conn *pgx.Conn, trainTestSplitRatio float64) error {
	page := 1
	questionsPaginator := newSOQuestionsPaginator(conn, 100_000)
	questionsPage := questionsPaginator.Next(ctx)
	pairsBuffer := make([]*CodeQueryPair, 0, BATCH_SIZE)
	processedRows := 0
	for len(questionsPage) > 0 {
		log.Infof("Processing page %d, len %d", page, len(questionsPage))
		for _, question := range questionsPage {
			cqp, err := questionToCodeQueryPair(ctx, conn, question, rand.Float64() < trainTestSplitRatio)
			if cqp == nil || err != nil {
				continue
			}
			pairsBuffer = append(pairsBuffer, cqp)

			if len(pairsBuffer) == BATCH_SIZE {
				err := importCodeQueryPairs(ctx, conn, pairsBuffer)
				if err != nil {
					return err
				}
				pairsBuffer = pairsBuffer[:0]
			}
		}

		processedRows += len(questionsPage)
		log.Infof("Processed page %d, total processed rows %d", page, processedRows)
		questionsPage = questionsPaginator.Next(ctx)
		page += 1
	}

	if len(pairsBuffer) > 0 {
		err := importCodeQueryPairs(ctx, conn, pairsBuffer)
		if err != nil {
			return err
		}
	}

	return questionsPaginator.Error()
}
