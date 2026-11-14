package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// process.recordreplay is a member Node's process doesn't have, which tsc's sys.ts reads (#b7phwab,
// Adamic on #b5w3ycg, Nov 7 2026). The prelude declares it on NodeJS.Process, and these programs read
// it off Node's own global process, as sys.ts does.
func processRecordReplayModules(body string) map[string]string {
	return map[string]string{
		"main.ts": "import 'node:process';\n" + body + "\n",
	}
}

// A plain read is IR undefined on every target, never a slot of the process object.
func TestProcessRecordReplayReadsUndefined(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"console.log(String(!!process.recordreplay));",
		"console.log(typeof process.recordreplay);",
		"const r = process.recordreplay; console.log(typeof r);",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			program, err := lowerModules(t, processRecordReplayModules(body))
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			undefined := false
			check := func(value any) bool {
				if _, ok := value.(ir.Undefined); ok {
					undefined = true
				}
				if property, ok := value.(ir.Property); ok && property.Name == "recordreplay" {
					t.Error("recordreplay read from a slot")
				}
				return true
			}
			walk(program.Main, check)
			for _, function := range program.Functions {
				walk(function.Body, check)
			}
			// typeof may simplify undefined to its string before IR is walked, as browser's does.
			if !undefined && !strings.Contains(body, "typeof process.recordreplay") {
				t.Error("no IR undefined for recordreplay")
			}
		})
	}
}

// Every write, install, reflection and later view stops with recordreplay named, as browser's do.
func TestProcessRecordReplayRoadsStop(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"process.recordreplay = 1;",
		// Language's first attack: a view naming the member, written through.
		"const p: { recordreplay?: unknown } = process; p.recordreplay = 1;",
		// Her second: a copy of process, read.
		"const p = { ...process }; console.log(String(p.recordreplay));",
		"Object.assign(process, { recordreplay: 1 });",
		"Object.defineProperty(process, 'recordreplay', { value: 1 });",
		"Reflect.set(process, 'recordreplay', 1);",
		"process['recordreplay'] = 1;",
		"delete process.recordreplay;",
		"console.log(String('recordreplay' in process));",
		"const k = 'recordreplay'; console.log(String(k in process));",
		"const { recordreplay } = process; console.log(typeof recordreplay);",
		"Object.assign(globalThis.process, { recordreplay: true });",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, processRecordReplayModules(body))
			var notYet *NotYet
			var refused *Refused
			what := ""
			switch {
			case errors.As(err, &notYet):
				what = notYet.What
			case errors.As(err, &refused):
				what = refused.What + "; " + refused.Fix
			}
			// The temporary path contains the subtest's recordreplay spelling; only the payload counts.
			if !strings.Contains(what, "recordreplay") {
				t.Fatalf("want a stop naming recordreplay, got %v", err)
			}
			if strings.HasPrefix(body, "const k") && what != "a run-time key on process, which could name process.browser or process.recordreplay" {
				t.Fatalf("want the run-time key diagnostic, got %v", err)
			}
			t.Log(err)
		})
	}
}
