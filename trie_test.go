package sqlike

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLikeTrie(t *testing.T) {
	tables := []struct {
		expressions []string
		matches     map[string]string
		nomatches   []string
	}{
		// Case of direct match
		// Note: "" doesnt match any expression, not even ""
		{
			[]string{"test", "a", ""},
			map[string]string{
				"test": "test",
				"a":    "a",
			},
			[]string{"", "notest", "b", "tesb"},
		},
		// Case of similar direct matches
		{
			[]string{"test-a-post", "test-b-post", "test-c-post"},
			map[string]string{
				"test-b-post": "test-b-post",
			},
			[]string{"test-d-post"},
		},
		// Case of empty wild card matches
		{
			[]string{"a%b", "a%bb"},
			map[string]string{
				"ab":  "a%b",
				"abb": "a%bb",
			},
			[]string{},
		},
		// Case of similar direct wild card matches
		{
			[]string{"testa-%-posta", "testa-%-postb", "testb-%-postb"},
			map[string]string{
				"testa-abc-posta": "testa-%-posta",
				"testa-abc-postb": "testa-%-postb",
				"testb-abc-postb": "testb-%-postb",
			},
			[]string{"testb-abc-posta", "testa-abc-postc"},
		},
		// Case of partial matches
		{
			[]string{"test-%-post-a", "test-%-post"},
			map[string]string{
				"test-abc-post-a": "test-%-post-a",
				"test-abc-post":   "test-%-post",
			},
			[]string{"test-abc-post-", "test-abc-pos"},
		},
		// Case of prefixes
		{
			[]string{"test-%", "test-%-post", "test-a-%"},
			map[string]string{
				"test-b-post": "test-%-post",
				"test-a-test": "test-a-%",
				"test-b-test": "test-%",
			},
			[]string{"testpost"},
		},
		// Case of suffixes
		{
			[]string{"%-post", "%-a-post", "test-%-post"},
			map[string]string{
				"test-a-post": "test-%-post",
				"pre-a-post":  "%-a-post",
				"pre-b-post":  "%-post",
			},
			[]string{"testpost"},
		},
		// Case of protected sections
		{
			[]string{"test-[%]-post", "test-[%-noclose"},
			map[string]string{
				"test-%-post":     "test-[%]-post",
				"test-[]-noclose": "test-[%-noclose",
			},
			[]string{"test-a-post", "test-[]-post"},
		},
		// Case of similar expressions with protected sections
		{
			[]string{"test-%-post", "test-[%]-post"},
			map[string]string{
				"test-a-post": "test-%-post",
				"test-%-post": "test-[%]-post",
			},
			[]string{"testpost"},
		},
		// Case of protected sections with brackets
		{
			[]string{"test-[[]%[]]-post", "test-[[][]]-test", "pre-[%]%[%]-post"},
			map[string]string{
				"test-[]-post":      "test-[[]%[]]-post",
				"test-[abc]-post":   "test-[[]%[]]-post",
				"test-[%[]]-post":   "test-[[]%[]]-post",
				"test-[]%[]-post":   "test-[[]%[]]-post",
				"test-[[]a[]]-post": "test-[[]%[]]-post",
				"test-[]-test":      "test-[[][]]-test",
				"pre-%%-post":       "pre-[%]%[%]-post",
				"pre-%test%-post":   "pre-[%]%[%]-post",
			},
			[]string{"test-[][]-test", "test-[[][]]-test", "pre-abc-post"},
		},
	}

	for _, table := range tables {
		trie := NewLikeTrie(0)
		expected_meta := map[string]int{}
		for idx, expr := range table.expressions {
			trie.SaveExpression(expr, idx)
			expected_meta[expr] = idx
		}
		for text, match := range table.matches {
			expr, meta, err := trie.FindExpression(text)
			if assert.NoError(t, err, "Error looking for expression") {
				assert.Equal(t, match, expr, "text matched incorrect expression")
				assert.Equal(t, expected_meta[expr], meta, "Incorrect meta stored or fetched")
			}
		}
		for _, text := range table.nomatches {
			expr, meta, err := trie.FindExpression(text)
			assert.Nil(t, meta, "A meta was found for an unmatched text")
			if assert.Error(t, err) {
				assert.Equal(t, ExpressionNotFound, err)
			}
			assert.Equal(t, "", expr, "A matching expression was found")
		}
	}
}
