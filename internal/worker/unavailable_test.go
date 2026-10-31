package worker

import (
	"bytes"

	"github.com/system-inc/adamic/internal/javascript"
)

// workersUnavailable is the first spelling in a file a Worker ships that Workers can't run, or "":
// a Node global, matched as internal/javascript's TestWorkersCalleeWithoutNodeGlobals matches it,
// so ArrayBuffer.isView isn't taken for Node's Buffer, or code built from a string.
func workersUnavailable(source []byte) string {
	if forbidden := javascript.NodeGlobalUse(string(source)); forbidden != "" {
		return forbidden
	}
	for _, forbidden := range []string{"eval(", "new Function("} {
		if bytes.Contains(source, []byte(forbidden)) {
			return forbidden
		}
	}
	return ""
}
