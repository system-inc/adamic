package native_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// records_define_hidden.a is Object.defineProperty's hidden record key, as tsc's commandLineParser hides
// options.configFile, with the attributes Node gives it: held by graphCompilerWitness to source Node, the
// JavaScript backend, native under ASan and UBSan, and the leak check.
func recordDefineHiddenSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "records_define_hidden.a"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestRecordDefineHiddenAgainstNode(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, recordDefineHiddenSource(t), nil)
}

var (
	recordWriteCheck  = regexp.MustCompile(`if \(!(adamic_record_set\(.*\))\) adamic_record_refused\(.*, false\);`)
	recordDeleteCheck = regexp.MustCompile(`if \(!\w+\) adamic_record_refused\(.*, true\);`)
)

// Mutant: the write and delete checks dropped from the C, so a write or a delete of a hidden key is refused
// without a throw, while the define's own check, which is IR, stays. Node's output must catch it.
func TestRecordDefineHiddenWriteChecksMutant(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, recordDefineHiddenSource(t), func(code string) string {
		return recordDeleteCheck.ReplaceAllString(recordWriteCheck.ReplaceAllString(code, "(void)$1;"), "")
	})
}

// A record write or delete is checked only when its key may be hidden: one known only as it runs, or a literal
// a define names. A literal no define names emits the bare call, and without a define nothing is checked.
func TestRecordWriteChecksFollowHiddenKeys(t *testing.T) {
	t.Parallel()
	const writes = `const r: Record<string, number> = {};
const key = ['a', 'b'].join('');
r.plain = 1;
r[key] = 2;
delete r[key];
delete r.plain;
r.other ??= 3;
`
	const define = `Object.defineProperty(r, 'hidden', { enumerable: false, writable: false, value: 5 });
try {
    r.hidden = 6;
}
catch(error) {
    if(error instanceof Error) console.log(error.message);
}
`
	for _, probe := range []struct {
		name, source    string
		checks, defines int
	}{
		{"no define", writes + "console.log(JSON.stringify(r));\n", 0, 0},
		// r[key] and delete r[key] are dynamic, and r.hidden names the hidden key; r.plain, delete r.plain and
		// r.other ??= name none.
		{"one define", writes + define + "console.log(JSON.stringify(r));\n", 3, 1},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(probe.source), 0600); err != nil {
				t.Fatal(err)
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			if got := strings.Count(code, "adamic_record_refused("); got != probe.checks {
				t.Errorf("%d checked writes and deletes, want %d", got, probe.checks)
			}
			if got := strings.Count(code, "adamic_record_define_hidden("); got != probe.defines {
				t.Errorf("%d defines, want %d", got, probe.defines)
			}
		})
	}
}
