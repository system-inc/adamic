package markdownblocks

import "testing"

func TestMarkdownWhitespaceLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "whitespace")
}
