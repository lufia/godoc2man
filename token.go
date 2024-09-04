package main

import (
	"slices"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

var jaTokenizer *tokenizer.Tokenizer

func init() {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}
	jaTokenizer = t
}

func BreakString(s string) string {
	var buf strings.Builder
	tokens := jaTokenizer.Tokenize(s)
	for _, token := range tokens {
		buf.WriteString(token.Surface)
		if canBreakAfter(token.Features()) {
			buf.WriteString(`\:`)
		}
	}
	return buf.String()
}

func canBreakAfter(features []string) bool {
	switch {
	default:
		return false
	case slices.Contains(features, "読点"):
		return true
	case slices.Contains(features, "句点"):
		return true
	case slices.Contains(features, "助詞") && !slices.Contains(features, "連体化"):
		return true
	}
}
