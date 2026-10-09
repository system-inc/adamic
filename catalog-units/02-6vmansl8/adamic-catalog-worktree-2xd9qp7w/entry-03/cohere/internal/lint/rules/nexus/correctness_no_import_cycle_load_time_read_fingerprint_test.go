package nexus

import (
	"crypto/sha256"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// correctnessNoImportCycleLoadTimeReadFingerprintOf is the rule's program fingerprint over a fixture, read
// through the Program the walk hands the rule, viewed under its own reads.
func correctnessNoImportCycleLoadTimeReadFingerprintOf(t *testing.T, files map[string]string, subject string) [sha256.Size]byte {
	t.Helper()
	var fingerprint [sha256.Size]byte
	probe := CorrectnessNoImportCycleLoadTimeRead
	probe.Run = func(ctx rule.Context, options any) rule.Listeners {
		fingerprint = CorrectnessNoImportCycleLoadTimeRead.ProgramFingerprint(ctx.Program, options)
		return nil
	}
	rule_testing.RunTypedFiles(t, probe, files, subject)
	if fingerprint == ([sha256.Size]byte{}) {
		t.Fatal("the probe never ran, so no fingerprint was read")
	}
	return fingerprint
}

// The fingerprint is the shape of the program's import cycles (#n030v3r), proven both ways: an import that
// opens or closes a cycle, an edge added inside one, or a file in one renamed moves it; an edit that touches
// no cycle's runtime imports leaves it, so the findings cache replays across it. What a cycle's files say,
// and which of its edges survive elision, is in each member's import closure, which the type fingerprint
// already holds.
func TestCorrectnessNoImportCycleLoadTimeReadFingerprintsExactlyTheCycles(t *testing.T) {
	t.Parallel()
	const a = "/repository/source/A.ts"
	const b = "/repository/source/B.ts"
	const c = "/repository/source/C.ts"
	const d = "/repository/source/D.ts"
	base := map[string]string{
		a: "import { B } from './B';\nexport const A = () => B;\n",
		b: "import { C } from './C';\nexport const B = () => C;\n",
		c: "import { A } from './A';\nexport const C = () => A;\n",
		d: "import { A } from './A';\nexport const D = A;\n",
	}
	edited := func(changes map[string]string, removed ...string) map[string]string {
		files := map[string]string{}
		for name, contents := range base {
			files[name] = contents
		}
		for name, contents := range changes {
			files[name] = contents
		}
		for _, name := range removed {
			delete(files, name)
		}
		return files
	}
	original := correctnessNoImportCycleLoadTimeReadFingerprintOf(t, base, d)

	moves := map[string]map[string]string{
		"an edge added inside the cycle": edited(map[string]string{
			a: "import { B } from './B';\nimport { C } from './C';\nexport const A = () => B ?? C;\n",
		}),
		"a file closing a new cycle": edited(map[string]string{
			a: "import { B } from './B';\nimport { D } from './D';\nexport const A = () => B ?? D;\n",
		}),
		"the cycle's direction reversed": edited(map[string]string{
			a: "import { C } from './C';\nexport const A = () => C;\n",
			b: "import { A } from './A';\nexport const B = () => A;\n",
			c: "import { B } from './B';\nexport const C = () => B;\n",
		}),
		"the cycle opened": edited(map[string]string{
			c: "export const C = () => 1;\n",
		}),
		"a file in the cycle renamed": edited(map[string]string{
			b:                          "",
			"/repository/source/BB.ts": "import { C } from './C';\nexport const B = () => C;\n",
			a:                          "import { B } from './BB';\nexport const A = () => B;\n",
		}, b),
	}
	for name, files := range moves {
		if correctnessNoImportCycleLoadTimeReadFingerprintOf(t, files, d) == original {
			t.Errorf("%s left the fingerprint unchanged, so a stale finding would replay", name)
		}
	}

	holds := map[string]map[string]string{
		"a comment in a file of the cycle": edited(map[string]string{
			a: "// the first of three\nimport { B } from './B';\nexport const A = () => B;\n",
		}),
		"an import added outside every cycle": edited(map[string]string{
			d:                         "import { A } from './A';\nimport { E } from './E';\nexport const D = A ?? E;\n",
			"/repository/source/E.ts": "export const E = 5;\n",
		}),
		"a file of the cycle importing a file off it": edited(map[string]string{
			a:                         "import { B } from './B';\nimport { E } from './E';\nexport const A = () => B ?? E;\n",
			"/repository/source/E.ts": "export const E = 5;\n",
		}),
		"a type-only import that would close a cycle": edited(map[string]string{
			a: "import { B } from './B';\nimport type { D } from './D';\nexport const A = (): typeof D | undefined => B ? undefined : undefined;\n",
		}),
	}
	for name, files := range holds {
		if correctnessNoImportCycleLoadTimeReadFingerprintOf(t, files, d) != original {
			t.Errorf("%s moved the fingerprint, so files whose verdict cannot change would not replay", name)
		}
	}
}
