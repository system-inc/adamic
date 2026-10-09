package css

// src/language-css/utilities/get-function-arguments-text.js.

import "github.com/system-inc/cohere/internal/format/estree"

// getFunctionArgumentsText is the text between a function's parentheses, trimmed. sourceIndex is in
// bytes, as the values parser reports it, so the slice is of the Go string.
func getFunctionArgumentsText(node *estree.Node, walk valueRootWalk) string {
	return estree.TrimJavaScript(
		sliceJavaScript(
			getValueRoot(walk).String("text"),
			intProperty(node.Child("group").Child("open"), "sourceIndex")+1,
			intProperty(node.Child("group").Child("close"), "sourceIndex"),
		),
	)
}
