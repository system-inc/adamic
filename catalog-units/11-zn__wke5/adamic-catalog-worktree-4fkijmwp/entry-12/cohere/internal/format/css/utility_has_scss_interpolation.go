package css

// src/language-css/utilities/has-scss-interpolation.js.

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
)

// hasSCSSInterpolation is not gated on the parser upstream, so it runs for css too.
func hasSCSSInterpolation(groupList []*estree.Node) bool {
	if len(groupList) > 0 {
		for i := len(groupList) - 1; i > 0; i-- {
			// If we find `#{`, return true.
			if groupList[i].Is("word") &&
				groupList[i].String("value") == "{" &&
				groupList[i-1].Is("word") &&
				strings.HasSuffix(groupList[i-1].String("value"), "#") {
				return true
			}
		}
	}
	return false
}
