package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A probe, not for landing: one line per program, its stop or each backend's answer beside Node's.
func TestZZProbeRestHide(t *testing.T) {
	for _, directory := range []string{"internal/oracle/testdata/rest_fixed_view_probe"} {
		entries, err := os.ReadDir(filepath.Join(repository, directory))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			extension := filepath.Ext(entry.Name())
			if extension != ".a" && extension != ".ts" {
				continue
			}
			absolute, _ := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			truth := onNode(t, absolute)
			line := fmt.Sprintf("R %s node=%d %q", entry.Name(), truth.exitCode, strings.TrimSpace(truth.stdout))
			program, err := lowered(t, absolute)
			var notYet *lower.NotYet
			if errors.As(err, &notYet) {
				where := notYet.Where
				if index := strings.LastIndex(where, "/"); index >= 0 {
					where = where[index+1:]
				}
				message := notYet.Error()
				if len(message) > 110 {
					message = message[len(message)-110:]
				}
				fmt.Println(line, "NOTYET", where, message)
				continue
			}
			if err != nil {
				fmt.Println(line, "ERR", strings.ReplaceAll(err.Error(), "\n", " ")[:min(160, len(err.Error()))])
				continue
			}
			sanitized, _ := natively(t, program)
			released := released(t, program)
			javascript := onJavaScriptBackend(t, program)
			fmt.Println(line, "LOWERS", "rel", released.exitCode, disagreement(truth, released) == "", "san", sanitized.exitCode, disagreement(truth, sanitized) == "", strings.ReplaceAll(strings.TrimSpace(sanitized.stderr), "\n", " ")[:min(90, len(strings.TrimSpace(sanitized.stderr)))], "js", disagreement(truth, javascript) == "")
		}
	}
}
