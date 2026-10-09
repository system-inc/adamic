package css

// src/language-css/utilities/has-string-or-function.js.

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
)

func hasStringOrFunction(groupList []*estree.Node) bool {
	for _, group := range groupList {
		if group.Is("string") ||
			(group.Is("func") &&
				// workaround false-positive func
				!strings.HasSuffix(group.String("value"), "\\")) {
			return true
		}
	}
	return false
}
