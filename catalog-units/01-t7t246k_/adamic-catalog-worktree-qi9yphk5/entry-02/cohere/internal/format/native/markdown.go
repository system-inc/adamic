package native

import (
	"sync"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/markdown"
)

// embedDocs hold the slabs markdown formats clean their embeds' docs into (#v6ksqg3): front matter's
// stripped doc sits inside the markdown doc until markdown.Format has laid it out, so the slab is the
// markdown file's, taken here and released once Format returns. A pool, because files are formatted on
// several goroutines at once and each Get is that caller's alone. Under cohere_poison a released part
// reads as visible text, so a doc read after its file was done prints wrong.
var embedDocs = sync.Pool{New: func() any {
	return &arena.Slab[doc.Doc]{Poison: doc.Text("released")}
}}

// The markdown printer: Prettier's language-markdown printer, ported in internal/format/markdown, over
// the tree its own micromark port builds.
func init() {
	// A fenced code block or front matter is formatted by the native printer for its language, through
	// TextToDoc. markdown's embed passes upstream's parser name, or for ts, typescript and tsx the file
	// name upstream overrides the filepath with, because the trailing comma of type parameters depends
	// on it.
	Register(".md", func(_ string, text string, options formatoptions.Options) (string, error) {
		docs := embedDocs.Get().(*arena.Slab[doc.Doc])
		defer func() {
			docs.Reset()
			embedDocs.Put(docs)
		}()
		return markdown.Format(text, options, TextToDoc(options, "markdown", docs))
	})
}
