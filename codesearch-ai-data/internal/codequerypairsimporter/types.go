package codequerypairsimporter

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"unicode"
)

const BATCH_SIZE = 10_000

type CodeQueryPair struct {
	ID                  int    `json:"id"`
	Code                string `json:"code"`
	CodeHash            string `json:"-"`
	Query               string `json:"query"`
	IsTrain             bool   `json:"-"`
	SOQuestionID        *int   `json:"soQuestionId"`
	ExtractedFunctionID *int   `json:"extractedFunctionId"`
}

func getSHA1Hash(text string) string {
	hasher := sha1.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func removeNonAsciiChars(text string) string {
	return strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return -1
		}
		return r
	}, text)
}

func newCodeQueryPair(code string, query string, isTrain bool, soQuestionID *int, extractedFunctionID *int) *CodeQueryPair {
	return &CodeQueryPair{
		Code:                code,
		CodeHash:            getSHA1Hash(code),
		Query:               query,
		IsTrain:             isTrain,
		SOQuestionID:        soQuestionID,
		ExtractedFunctionID: extractedFunctionID,
	}
}
