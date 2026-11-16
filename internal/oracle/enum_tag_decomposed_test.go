package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A numeric enum holds any number, so a whole open Kind, or two or more of its members (what excluding
// one by !== leaves), reaches a discriminated union's tag slots only by tsc decomposing the union: each
// member it's built as promises one literal the value may not hold, and a switch's never read past them
// traps on valid input (#egmxf20). Every road is refused as an enum member slot, in .a and .ts alike, at
// the value written into the slot: a literal's property, a shorthand, a literal nested in another, an
// optional field's member, an array's element, two of three members (memberLiteralSlotWidened), and a
// variable bound to a literal, which no one member takes (widened's union arm). Mutants:
// openEnumTagDecomposed's check dropped from memberLiteralSlotWidened lowers the literal roads; dropped
// from widened's union arm, the variable lowers.
func TestEnumTagDecomposedIsRefused(t *testing.T) {
	t.Parallel()
	for name, position := range map[string]string{
		"literal":   "13:26",
		"shorthand": "12:20",
		"nested":    "13:35",
		"variable":  "14:18",
		"optional":  "13:20",
		"excluded":  "12:22",
		"array":     "12:37",
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(name+extension, func(t *testing.T) {
				t.Parallel()
				path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enum_tag_decomposed", name+extension))
				if err != nil {
					t.Fatal(err)
				}
				_, err = lowered(t, path)
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), "enum member slot") || !strings.Contains(err.Error(), name+extension+":"+position+":") {
					t.Fatalf("want an enum member slot refused at %s%s:%s, got %v", name, extension, position, err)
				}
			})
		}
	}
}

// What stays open: one member literal written into its own slot, a slot typed the whole Kind taking the
// open value, and a literal another member takes whole beside a member with a literal slot (control_mixed:
// the literal isn't built as that member, so its slot promises nothing; memberLiteralSlotWidened's
// takenWhole). Each lowers and agrees with Node on the JavaScript backend and natively under the leak
// check.
func TestEnumTagDecomposedControlsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for name, stdout := range map[string]string{
		"control_literal": "circle 1 square 20 7\n",
		"control_whole":   "tagged 7 2 tagged 7 3 named x\n",
		"control_mixed":   "7 7 0\n",
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(name+extension, func(t *testing.T) {
				t.Parallel()
				path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enum_tag_decomposed", name+extension))
				if err != nil {
					t.Fatal(err)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if truth.exitCode != 0 || string(truth.stdout) != stdout {
					t.Fatalf("Node must print %q: exit %d, stdout %q, stderr %q", stdout, truth.exitCode, truth.stdout, truth.stderr)
				}
				backend := onJavaScriptBackend(t, program)
				if diff := disagreement(truth, backend); diff != "" {
					t.Fatalf("javascript: %s; backend stdout %q stderr %q exit %d", diff, backend.stdout, backend.stderr, backend.exitCode)
				}
				actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
				if leaked != "" {
					t.Fatalf("leaks: %s", leaked)
				}
				if comparable {
					if diff := disagreement(truth, actual); diff != "" {
						t.Fatalf("native: %s; stdout %q stderr %q exit %d", diff, actual.stdout, actual.stderr, actual.exitCode)
					}
				}
			})
		}
	}
}

// The open value narrowed by === to one member before it goes into the slot (control_narrowed) is refused
// on main as well, at the literal, in the same words and at the same place: measured with both of
// #egmxf20's arms dropped (J-ty3bkbb). It's main's enum-literal rule, not this one, so it's pinned as a
// stop here, for whoever lifts it to see.
func TestEnumTagDecomposedNarrowedStaysMainsStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run("control_narrowed"+extension, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enum_tag_decomposed", "control_narrowed"+extension))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "enum member slot") || !strings.Contains(err.Error(), "control_narrowed"+extension+":13:17:") {
				t.Fatalf("want main's enum member slot refusal at control_narrowed%s:13:17, got %v", extension, err)
			}
		})
	}
}
