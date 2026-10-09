package native

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/graphql"
	"github.com/system-inc/cohere/internal/format/printing"
)

// The GraphQL printer: Prettier's language-graphql printer over graphql-js's parse, ported in
// internal/format/graphql. Its doc entry is what gql and graphql templates in TypeScript embed.
func init() {
	print := func(_ string, text string, options formatoptions.Options) (string, error) {
		return graphql.Format(text, options)
	}
	printDoc := func(_ string, text string, options formatoptions.Options, _ string, _ string, _ printing.TextToDoc) (doc.Doc, error) {
		return graphql.PrintToDoc(text, options)
	}
	for _, extension := range []string{".graphql", ".gql"} {
		Register(extension, print)
		RegisterDoc(extension, printDoc)
	}
}
