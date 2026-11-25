package oracle

import (
	"strings"
	"testing"
)

func TestNodeArrayTupleProbe(t *testing.T) {
	for _, name := range []string{"implied-textrange.ts", "implied-extra-field.ts", "implied-writable.ts", "implied-string-pos.ts"} {
		path := nodeArrayTuplePath(t, name)
		truth := onNode(t, path)
		_, err := lowered(t, path)
		message := "lowers"
		if err != nil {
			message = strings.ReplaceAll(err.Error(), "\n", " | ")
		}
		t.Logf("PROBE %s node=%d %q lower=%s", name, truth.exitCode, truth.stdout, message)
	}
}
