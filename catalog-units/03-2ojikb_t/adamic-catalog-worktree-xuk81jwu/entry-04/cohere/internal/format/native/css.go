package native

import (
	"github.com/system-inc/cohere/internal/format/css"
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/printing"
)

// The CSS printer: Prettier's language-css printer over postcss and its three sub-parsers, ported in
// internal/format/css. The scss parser is registered for embedding only: upstream formats css and
// styled templates in TypeScript as scss (language-js/embed/css.js). .scss and .less files are not
// formatted.
func init() {
	RegisterDoc(".scss", func(_ string, text string, options formatoptions.Options, _ string, _ string, _ printing.TextToDoc) (doc.Doc, error) {
		return css.PrintToDocSCSS(text, options)
	})
	Register(".css", func(_ string, text string, options formatoptions.Options) (string, error) {
		return css.Format(text, options)
	})
	RegisterDoc(".css", func(_ string, text string, options formatoptions.Options, _ string, _ string, _ printing.TextToDoc) (doc.Doc, error) {
		return css.PrintToDoc(text, options)
	})
}
