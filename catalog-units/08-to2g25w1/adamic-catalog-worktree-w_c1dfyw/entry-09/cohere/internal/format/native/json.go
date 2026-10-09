package native

import (
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/javascript"
)

// The JSON printers: Prettier's json and json-stringify parsers, chosen by file name inside
// javascript.FormatJSON, the way language-json's languages choose them.
func init() {
	Register(".json", func(fileName string, text string, options formatoptions.Options) (string, error) {
		return javascript.FormatJSON(fileName, text, options)
	})
	RegisterDoc(".json", javascript.PrintToDoc)
}
