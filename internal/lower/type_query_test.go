package lower

import "testing"

// A host member named in a type query, typeof <member> written as a type, is no use of its value: the
// checker counts the name as an expression, but nothing runs it (#7yz6rtp). Adaptation 93 writes tsc's
// getNodeSystem forwarder this way (sys.ts:1500, Parameters<typeof clearTimeout>[0]). Each row lowers,
// one for each rule that judges a host member's use: the timers' value read (nodeTimerValueName, which a
// namespace member reaches through callableNamespaceUse), library_node.go's use rules (realpath, stdout,
// environment, platform, stackTraceLimit) and a host binding's read (hostBindingRefusal).
// process.stdout's setBlocking isn't on its declared type, so it has no row. Mutants: each of those
// three inTypeQuery clauses dropped, and its rows go red.
func TestTypeQueryOfHostMemberIsNoUse(t *testing.T) {
	t.Parallel()
	// The alias is exported and nothing holds a value of it, so a row stops only where its name is judged.
	const use = "\nconsole.log('typed');\n"
	for _, row := range []struct{ name, source string }{
		{"clearTimeout", "import 'node:timers';\nexport type Held = typeof clearTimeout;"},
		{"setTimeout", "import 'node:timers';\nexport type Held = typeof setTimeout;"},
		{"clearTimeout through a namespace", "import * as timers from 'node:timers';\nexport type Held = typeof timers.clearTimeout;"},
		{"clearTimeout parameter", "import 'node:timers';\nexport type Held = Parameters<typeof clearTimeout>[0];"},
		{"realpathSync", "import * as fs from 'node:fs';\nexport type Held = typeof fs.realpathSync;"},
		{"realpathSync.native", "import * as fs from 'node:fs';\nexport type Held = typeof fs.realpathSync.native;"},
		{"stdout", "import 'node:process';\nexport type Held = typeof process.stdout;"},
		{"env", "import 'node:process';\nexport type Held = typeof process.env;"},
		{"platform", "import 'node:process';\nexport type Held = typeof process.platform;"},
		{"process", "import 'node:process';\nexport type Held = typeof process;"},
		{"stackTraceLimit", "import 'node:process';\nexport type Held = typeof Error.stackTraceLimit;"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, row.source+use); err != nil {
				t.Fatal(err)
			}
		})
	}
	// A host binding is declared ambiently in its own module, tsc's performance.ts (profilerModules).
	t.Run("onProfilerEvent", func(t *testing.T) {
		t.Parallel()
		if _, err := lowerModules(t, profilerModules("\ttype Held = typeof onProfilerEvent;\n\tconsole.log(name);\n")); err != nil {
			t.Fatal(err)
		}
	})
}
