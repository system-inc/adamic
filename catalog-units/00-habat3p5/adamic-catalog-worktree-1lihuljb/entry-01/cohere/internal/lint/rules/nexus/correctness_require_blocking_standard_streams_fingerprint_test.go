package nexus

import (
	"crypto/sha256"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// correctnessRequireBlockingStandardStreamsFingerprintOf is the rule's program fingerprint over a fixture,
// read through the Program the walk hands the rule, viewed under its own reads.
func correctnessRequireBlockingStandardStreamsFingerprintOf(t *testing.T, files map[string]string, subject string) [sha256.Size]byte {
	t.Helper()
	var fingerprint [sha256.Size]byte
	probe := CorrectnessRequireBlockingStandardStreams
	probe.Run = func(ctx rule.Context, options any) rule.Listeners {
		fingerprint = CorrectnessRequireBlockingStandardStreams.ProgramFingerprint(ctx.Program, options)
		return nil
	}
	rule_testing.RunTypedFiles(t, probe, files, subject)
	if fingerprint == ([sha256.Size]byte{}) {
		t.Fatal("the probe never ran, so no fingerprint was read")
	}
	return fingerprint
}

// The fingerprint is the set of imported files (#5yfbdse), proven both ways: an edit that makes a file
// imported or unimported moves it, and one that does neither leaves it, so the findings cache replays across
// it. The rest of what a verdict reads, what an entry's imports reach and say, is its import closure, which
// the type fingerprint holds.
func TestCorrectnessRequireBlockingStandardStreamsFingerprintsExactlyTheImportedFiles(t *testing.T) {
	t.Parallel()
	const entry = "/repository/source/Entry.ts"
	const worker = "/repository/source/Worker.ts"
	const helper = "/repository/source/Helper.ts"
	const other = "/repository/source/Other.ts"
	base := map[string]string{
		correctnessRequireBlockingStandardStreamsNexusFile: correctnessRequireBlockingStandardStreamsNexus,
		entry:  "import { help } from './Helper';\nif(!help()) { console.error('usage'); process.exit(2); }\n",
		worker: "console.error('working'); process.exit(1);\n",
		helper: "export function help() { return true; }\n",
		other:  "export const other = 1;\n",
	}
	edited := func(changes map[string]string) map[string]string {
		files := map[string]string{}
		for name, contents := range base {
			files[name] = contents
		}
		for name, contents := range changes {
			files[name] = contents
		}
		return files
	}
	original := correctnessRequireBlockingStandardStreamsFingerprintOf(t, base, entry)

	moves := map[string]map[string]string{
		"a file becoming imported": edited(map[string]string{
			other: "import './Worker';\nexport const other = 1;\n",
		}),
		"a file becoming unimported": edited(map[string]string{
			entry: "if(!true) { console.error('usage'); process.exit(2); }\n",
		}),
		"the import moved to another file": edited(map[string]string{
			entry: "import { other } from './Other';\nif(!other) { console.error('usage'); process.exit(2); }\n",
		}),
	}
	for name, files := range moves {
		if correctnessRequireBlockingStandardStreamsFingerprintOf(t, files, entry) == original {
			t.Errorf("%s left the fingerprint unchanged, so a stale finding would replay", name)
		}
	}

	holds := map[string]map[string]string{
		"a second importer of an imported file": edited(map[string]string{
			other: "import { help } from './Helper';\nexport const other = help();\n",
		}),
		"a comment in an imported file": edited(map[string]string{
			helper: "// always true\nexport function help() { return true; }\n",
		}),
		"another exit in the entry": edited(map[string]string{
			entry: "import { help } from './Helper';\nif(!help()) { console.error('usage'); process.exit(2); }\nprocess.exit(0);\n",
		}),
		"a type-only import": edited(map[string]string{
			other: "import type { help } from './Helper';\nimport type {} from './Worker';\nexport const other: typeof help | undefined = undefined;\n",
		}),
	}
	for name, files := range holds {
		if correctnessRequireBlockingStandardStreamsFingerprintOf(t, files, entry) != original {
			t.Errorf("%s moved the fingerprint, so files whose verdict cannot change would not replay", name)
		}
	}
}
