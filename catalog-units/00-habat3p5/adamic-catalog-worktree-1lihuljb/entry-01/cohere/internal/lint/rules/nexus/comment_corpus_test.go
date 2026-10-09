package nexus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
)

// The corpus is text scanning with no dependency on the comment scanner, which is why it stays here
// while `comments.go` moved to `internal/utilities/comments/`. `shouting_mask_gate_test.go` needs real
// comment text to mask; it does not need the parser's view of where comments are.

// realCommentCorpus collects comment lines from the tree, so a gate over comments is measured
// against the comments that exist rather than the ones somebody thought to write.
func realCommentCorpus(t *testing.T) []string {
	t.Helper()

	var comments []string
	root := corpus.Structure.Path(t, "source")
	fileCount := 0
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || fileCount >= 400 {
			return nil
		}
		if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") {
			return nil
		}
		fileCount++
		data, readError := os.ReadFile(path)
		if readError != nil {
			return nil
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			isCommentish := strings.HasPrefix(trimmed, "//") ||
				strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*")
			if isCommentish {
				comments = append(comments, trimmed)
			}
		}
		return nil
	})

	if len(comments) < 1000 {
		t.Fatalf("the structure corpus gave only %d comments, too few to be a corpus", len(comments))
	}
	return comments
}
