package oracle

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Ruling 11 (Adamic on #bbssdcy, Nov 24; #4cvdkk3): an untagged checked predicate's target is judged relative
// to its source. A field every object member of the source declares, as required when the target requires
// it, with a type assignable to the target's, is read nowhere at the call; every other field is refined and
// read at the call through the view's checked read, and the target's view checks every read narrowed to it.
// A refined field no membership can express (a Map, ReadonlyMap, Set or ReadonlySet, which share one
// representation) keeps main's stop and words, by Adamic's road (a). The false side is checked by the
// remainder's view at its first read, never at the branch's entry (ruling 12 A).
//
// Each row is held to Node, then lowered: agree rows run on sanitized and release native and the JavaScript
// backend and must print what Node prints; stop rows must stop at run time with exit 70, Node's stdout up to
// the stop, and stderr starting with stop; pending and refused rows must stop lowering with words containing
// stop. Language's p01, p02 and p03 are byte for byte from #hatdcsp, and p02u and p03u are their untagged
// twins (conditions (a) and (b)). The n rows read nothing after the call, so only the call's own check stops
// their lie, which is what the wrong versions in predicates_relative.mutants.md turn red.
func TestPredicateRelativeTarget(t *testing.T) {
	t.Parallel()
	const (
		agree   = "agree"
		stops   = "stops"
		pending = "pending"
		refused = "refused"
	)
	for _, probe := range []struct{ name, node, outcome, stdout, stop string }{
		{"p01.ts", "target undefined\n", refused, "", "p01.ts:7:24: Adamic 0.1 refuses a cast the runtime can't check: an `as unknown as` pair from Record<string, string>"},
		{"p02.ts", "named undefined\n", stops, "", "adamic: panic: field read failed: loose.name is not initialized; expected string, found missing\n"},
		{"p02u.ts", "named undefined\n", stops, "", "adamic: panic: field read failed: loose.name is not initialized; expected string, found missing\n"},
		{"p03.ts", "named undefined\n", pending, "", "p03.ts:3:1: stage 0 can't lower checked predicate overload target Dict & Named yet"},
		{"p03u.ts", "named undefined\n", pending, "", "p03u.ts:3:1: stage 0 can't lower checked predicate overload target Named yet"},
		{"s1_package_info.ts", "info /info/a a@1\nmissing /b false\nmissing /c true\nnot info\n", agree, "", ""},
		{"s2_missing_info.ts", "imports of a\nnone\nnone\nexists false\n", agree, "", ""},
		{"s3_source_mapping.ts", "1,2 none undefined undefined\n3,4 0:5:6 0 7\n1\n", agree, "", ""},
		// A field held as undefined alone has no checked read, so its target keeps main's road (Lower's must on
		// #4cvdkk3): a raw read would trust a value a checked view admitted (v_cast_view, v_param_view).
		{"s4b_detached_required.ts", "detached f.ts 7\nrelated 0\n", pending, "", "s4b_detached_required.ts:4:1: stage 0 can't lower checked predicate overload target Detached yet"},
		{"v_cast_view.ts", "tail\n", pending, "", "v_cast_view.ts:4:1: stage 0 can't lower checked predicate overload target Tail yet"},
		{"v_param_view.ts", "tail\nmore\n", pending, "", "v_param_view.ts:4:1: stage 0 can't lower checked predicate overload target Tail yet"},
		{"s5_custom_option.ts", "1\n", pending, "", "s5_custom_option.ts:3:1: stage 0 can't lower checked predicate overload target Custom yet"},
		{"s5b_set_field.ts", "set 1\n", pending, "", "s5b_set_field.ts:3:1: stage 0 can't lower checked predicate overload target WithSet yet"},
		{"r1_refined_absent.ts", "dst number\n", stops, "", "adamic: panic: field read failed: v.name is not a string; expected string, found number\n"},
		{"r2_refined_record_read.ts", "a undefined\n", stops, "", "adamic: panic: field read failed: v.inner.a is not initialized; expected number, found missing\n"},
		{"r3_always_true_missing.ts", "missing undefined\n", stops, "", "adamic: panic: field read failed: info.directoryExists is not initialized; expected boolean, found missing\n"},
		{"r4_always_true_info.ts", "info undefined\n", stops, "", "adamic: panic: field read failed: missing.contents is not initialized; expected Contents, found missing\n"},
		{"r5_lying_false.ts", "false branch entered for an entry\nmissing undefined\n", stops, "false branch entered for an entry\n", "adamic: panic: field read failed: entry.directoryExists is not initialized; expected boolean, found missing\n"},
		{"r6_family9a.ts", "p\n", pending, "", "r6_family9a.ts:4:1: stage 0 can't lower a checked predicate target whose check would read past the depth bound (WithProgram) yet"},
		{"r7_alias_write.ts", "dst number\n", stops, "", "adamic: panic: field read failed: v.name is not a string; expected string, found number\n"},
		{"n1_optional_unread.ts", "named\n", stops, "", "adamic: panic: field read failed: loose.name is not initialized; expected string, found missing\n"},
		{"n3_refined_unread.ts", "dst\n", stops, "", "adamic: panic: field read failed: v.name is not a string; expected string, found number\n"},
		{"n4_absent_unread.ts", "missing\n", stops, "", "adamic: panic: field read failed: info.directoryExists is not initialized; expected boolean, found missing\n"},
		{"x1_three_members.ts", "two three \n", agree, "", ""},
		{"x2_two_refined.ts", "p1 6\nloose q\n", agree, "", ""},
		{"x3_filter.ts", "2 3\n", agree, "", ""},
		// A lying false whose branch reads a field every source member holds reads it true, as Node does
		// (Language's should on #4cvdkk3): beside r5, which stops at a refined field, it pins both halves.
		{"x4_lying_false_held.ts", "missing at /a\n", agree, "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_relative", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			if source, lie := predicateRelativeLies[probe.name]; lie {
				path = filepath.Join(t.TempDir(), probe.name)
				if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.node)}, node); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			switch probe.outcome {
			case pending:
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.stop) {
					t.Fatalf("want a not yet with %q, got %v", probe.stop, err)
				}
				return
			case refused:
				var refusal *lower.Refused
				if !errors.As(err, &refusal) || !strings.Contains(err.Error(), probe.stop) {
					t.Fatalf("want a refusal with %q, got %v", probe.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				backend := []string{"sanitized native", "release native", "the JavaScript backend"}[index]
				if probe.outcome == agree {
					if difference := disagreement(node, got); difference != "" {
						t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, backend, got.exitCode, got.stdout, got.stderr)
					}
					continue
				}
				if got.exitCode != 70 || string(got.stdout) != probe.stdout || !bytes.HasPrefix(got.stderr, []byte(probe.stop)) {
					t.Fatalf("want a stop on %s: exit 70, stdout %q, stderr starting %q; got exit %d, stdout %q, stderr %q", backend, probe.stdout, probe.stop, got.exitCode, got.stdout, got.stderr)
				}
			}
			if probe.outcome == agree {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
