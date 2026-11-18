package oracle

import (
	"path/filepath"
	"testing"
)

// An object literal in a union slot that holds one dictionary the program declares, and no fixed object it fits, is
// built as that dictionary (lower's recordUnionDictionaryMember, #6c4swvt), so a read that narrows the slot's value
// to the dictionary reads a record, as Node reads the object. Built fixed, Language's g5 p06 on #6c4swvt stopped
// natively on an honest value. The road here needs no view: a record literal's computed entry under a declared
// dictionary, left to the named read (recordComputedEntry). A library's union (Object.entries' { [s: string]: T } | ArrayLike<T>) keeps today's
// build, which internal/lower's TestObjectRefusalsExplainSoundness pins. Mutant: build the literal fixed again
// (drop the recordUnionDictionaryMember call in recordLiteralElement), and these disagree with Node natively.
func TestRecordUnionDictionaryLiteralAgreesWithNode(t *testing.T) {
	t.Parallel()
	for name, printed := range map[string]string{
		"computed_entry.ts": "2:b\nmissing\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_union_literal", name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != printed {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			native, sanitized := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node exit %d stdout %q stderr %q; got exit %d stdout %q stderr %q", backend, difference, node.exitCode, node.stdout, node.stderr, got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, sanitized); report != "" {
				t.Fatal(report)
			}
		})
	}
}
