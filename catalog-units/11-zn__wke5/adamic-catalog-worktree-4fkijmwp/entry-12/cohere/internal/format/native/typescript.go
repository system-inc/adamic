package native

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/javascript"
	"github.com/system-inc/cohere/internal/types/sourcename"
)

// The TypeScript and TSX printer: Prettier's language-js printer, ported in internal/format/javascript,
// over the ESTree tree internal/format/estree converts from typescript-go's AST.
//
// An Adamic `.a` file is TypeScript under its own name, and prints exactly as the same text named `.ts`
// (#6mhafvb): estree parses it as TS, and the printer's one name-dependent choice, `<T,>`'s comma, asks
// sourcename.TreatedAs. Which `.a` files reach the printer is the program's to say, not this table's: the
// format walk holds them back until the program claims them (formatfiles.Enumeration.Adamic).
func init() {
	print := func(fileName string, text string, options formatoptions.Options) (string, error) {
		return javascript.Format(fileName, text, options, TextToDoc(options, "typescript", nil))
	}
	printParsed := func(fileName string, text string, parsed *ast.SourceFile, options formatoptions.Options) (string, error) {
		return javascript.FormatParsed(fileName, text, parsed, options, TextToDoc(options, "typescript", nil))
	}
	Register(".ts", print)
	Register(".tsx", print)
	Register(sourcename.AdamicExtension, print)
	RegisterParsed(".ts", printParsed)
	RegisterParsed(".tsx", printParsed)
	RegisterParsed(sourcename.AdamicExtension, printParsed)
	RegisterDoc(".ts", javascript.PrintToDoc)
	RegisterDoc(".tsx", javascript.PrintToDoc)
	RegisterDoc(sourcename.AdamicExtension, javascript.PrintToDoc)
}
