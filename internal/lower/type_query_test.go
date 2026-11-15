package lower

import "testing"

// A host member named in a type query, typeof <member> written as a type, is no use of its value: the
// checker counts the name as an expression, but nothing runs it (#7yz6rtp). Adaptation 93 writes tsc's
// getNodeSystem forwarder this way (sys.ts:1500, Parameters<typeof clearTimeout>[0]). Each row lowers,
// one for each rule that judges a host member's use: the timers' value read (nodeTimerValueName),
// library_node.go's use rules (realpath, stdout, environment, platform, stackTraceLimit) and a host
// binding's read (hostBindingRefusal). process.stdout's setBlocking isn't on its declared type, so it
// has no row. Mutants: each of those three inTypeQuery clauses dropped, and its rows go red.
func TestTypeQueryOfHostMemberIsNoUse(t *testing.T) {
	t.Parallel()
	// The alias is exported and nothing holds a value of it, so a row stops only where its name is judged.
	const use = "\nconsole.log('typed');\n"
	for _, row := range []struct{ name, source, extension string }{
		{"clearTimeout", "import 'node:timers';\nexport type Held = typeof clearTimeout;", ".a"},
		{"setTimeout", "import 'node:timers';\nexport type Held = typeof setTimeout;", ".a"},
		{"clearTimeout parameter", "import 'node:timers';\nexport type Held = Parameters<typeof clearTimeout>[0];", ".a"},
		{"realpathSync", "import * as fs from 'node:fs';\nexport type Held = typeof fs.realpathSync;", ".a"},
		{"realpathSync.native", "import * as fs from 'node:fs';\nexport type Held = typeof fs.realpathSync.native;", ".a"},
		{"stdout", "import 'node:process';\nexport type Held = typeof process.stdout;", ".a"},
		{"env", "import 'node:process';\nexport type Held = typeof process.env;", ".a"},
		{"platform", "import 'node:process';\nexport type Held = typeof process.platform;", ".a"},
		{"process", "import 'node:process';\nexport type Held = typeof process;", ".a"},
		{"stackTraceLimit", "import 'node:process';\nexport type Held = typeof Error.stackTraceLimit;", ".a"},
		// A host binding is declared ambiently, which only .ts admits (host_bindings/shadow/performance.ts).
		{"onProfilerEvent", "declare let onProfilerEvent: ((eventName: string) => void) | undefined;\nexport type Held = typeof onProfilerEvent;", ".ts"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, row.source+use, row.extension); err != nil {
				t.Fatal(err)
			}
		})
	}
}
