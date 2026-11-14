package lower

import "testing"

// A host member named in a type query, typeof <member> written as a type, is no use of its value: the
// checker counts the name as an expression, but nothing runs it (#7yz6rtp). Adaptation 93 writes tsc's
// getNodeSystem forwarder this way (sys.ts:1500, Parameters<typeof clearTimeout>[0]). Each row lowers,
// one for each rule that judges a host member's use: the timers' value read (nodeTimerValueName),
// library_node.go's use rules (realpath, stdout, environment, platform, stackTraceLimit) and a host
// binding's read (hostBindingRefusal). Mutants: each of those three inTypeQuery clauses dropped, and
// its rows go red.
func TestTypeQueryOfHostMemberIsNoUse(t *testing.T) {
	t.Parallel()
	const use = "\nfunction absent(value: Held | undefined): boolean {\n    return value === undefined;\n}\nconsole.log(String(absent(undefined)));\n"
	for _, row := range []struct{ name, source string }{
		{"clearTimeout", "import 'node:timers';\ntype Held = typeof clearTimeout;"},
		{"setTimeout", "import 'node:timers';\ntype Held = typeof setTimeout;"},
		{"clearTimeout parameter", "import 'node:timers';\ntype Held = Parameters<typeof clearTimeout>[0];"},
		{"realpathSync", "import * as fs from 'node:fs';\ntype Held = typeof fs.realpathSync;"},
		{"realpathSync.native", "import * as fs from 'node:fs';\ntype Held = typeof fs.realpathSync.native;"},
		{"stdout", "import 'node:process';\ntype Held = typeof process.stdout;"},
		{"setBlocking", "import 'node:process';\ntype Held = typeof process.stdout.setBlocking;"},
		{"env", "import 'node:process';\ntype Held = typeof process.env;"},
		{"platform", "import 'node:process';\ntype Held = typeof process.platform;"},
		{"process", "import 'node:process';\ntype Held = typeof process;"},
		{"stackTraceLimit", "import 'node:process';\ntype Held = typeof Error.stackTraceLimit;"},
		{"onProfilerEvent", "declare let onProfilerEvent: ((eventName: string) => void) | undefined;\ntype Held = typeof onProfilerEvent;"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, row.source+use); err != nil {
				t.Fatal(err)
			}
		})
	}
}
