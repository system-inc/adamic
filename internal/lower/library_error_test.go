package lower

import (
	"strings"
	"testing"
)

func TestLibraryErrorRefusals(t *testing.T) {
	for _, probe := range []struct{ name, source, reason string }{
		{"stack after capture", `import type { Stats } from 'node:fs'; const e = new Error('no stack'); Error.captureStackTrace(e); console.log(e.stack ?? 'absent');`, "Error.stack: native frames"},
		{"plain target stack", `import type { Stats } from 'node:fs'; const target: { stack?: string } = {}; Error.captureStackTrace(target); console.log(target.stack ?? 'absent');`, "Error.stack: native frames"},
		{"bracket stack", `import type { Stats } from 'node:fs'; const e = new Error('no stack'); Error.captureStackTrace(e); console.log(e['stack'] ?? 'absent');`, "Error.stack: native frames"},
		{"destructured stack", `import type { Stats } from 'node:fs'; const target: { stack?: string } = {}; Error.captureStackTrace(target); const { stack: captured } = target; console.log(captured ?? 'absent');`, "Error.stack: native frames"},
		{"constructor alias", `const constructor = Error;`, "constructor aliases and overloaded calls"},
		{"constructor passed", `function use(constructor: ErrorConstructor): void {} use(Error);`, "constructor aliases and overloaded calls"},
		{"stack limit", `import type { Stats } from 'node:fs'; console.log(String(Error.stackTraceLimit));`, "stackTraceLimit"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q, got %v", probe.reason, err)
			}
		})
	}
}
