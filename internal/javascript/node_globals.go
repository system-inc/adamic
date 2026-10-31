package javascript

import "regexp"

// A Node global is used where its spelling starts an identifier, so a standard
// global that ends in one, like ArrayBuffer.isView, which every module reads
// through adamicPlaceholderNarrow, is not a use.
var nodeGlobalSpelling = regexp.MustCompile(`(^|[^A-Za-z0-9_$])(node:|process\.|Buffer\.)`)

// NodeGlobalUse is the first Node global in code that Workers lack, or "".
func NodeGlobalUse(code string) string {
	if match := nodeGlobalSpelling.FindStringSubmatch(code); match != nil {
		return match[2]
	}
	return ""
}
