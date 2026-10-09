package changelog

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Each kind of change is named under its own heading, with the smallest version it allows.
func TestDiffSetNamesEachKindOfChange(t *testing.T) {
	t.Parallel()
	before := []byte(`{
		"format": {"printWidth": 120},
		"rules": {
			"stays": "error",
			"turns-off": "error",
			"loosens": "error",
			"tightens": "warn",
			"reconfigured": ["error", {"max": 1}],
			"already-off": "off"
		}
	}`)
	after := []byte(`{
		"format": {"printWidth": 100},
		"rules": {
			"stays": ["error"],
			"turns-off": "off",
			"loosens": "warn",
			"tightens": 2,
			"reconfigured": ["error", {"max": 2}],
			"already-off": ["off", {"ignored": true}],
			"new-rule": "warn"
		}
	}`)
	changes, err := DiffSet(before, after)
	if err != nil {
		t.Fatal(err)
	}
	for label, pair := range map[string][2][]string{
		"on":       {changes.TurnedOn, {"new-rule (warn)"}},
		"off":      {changes.TurnedOff, {"turns-off"}},
		"severity": {changes.Severity, {"loosens: error -> warn", "tightens: warn -> error"}},
		"options":  {changes.Options, {"reconfigured"}},
		"settings": {changes.Settings, {"format"}},
	} {
		if !slices.Equal(pair[0], pair[1]) {
			t.Errorf("%s: %q, want %q", label, pair[0], pair[1])
		}
	}
	if changes.Bump != Minor {
		t.Errorf("bump %s, want minor", changes.Bump)
	}
}

// What can only relax a set is a patch; each way a set can fail a project that passed is at least a
// minor, and a removed set, which breaks every configuration extending it, is a major.
func TestEachChangeAllowsTheSmallestVersionThePolicySays(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		before []byte
		after  []byte
		want   Bump
	}{
		{"nothing", []byte(`{"rules":{"a":"error"}}`), []byte(`{"rules":{"a":"error"}}`), Patch},
		{"a rule turned off", []byte(`{"rules":{"a":"error"}}`), []byte(`{"rules":{"a":"off"}}`), Patch},
		{"a rule removed", []byte(`{"rules":{"a":"error"}}`), []byte(`{"rules":{}}`), Patch},
		{"a severity lowered", []byte(`{"rules":{"a":"error"}}`), []byte(`{"rules":{"a":"warn"}}`), Patch},
		{"a rule turned on", []byte(`{"rules":{}}`), []byte(`{"rules":{"a":"warn"}}`), Minor},
		{"a severity raised", []byte(`{"rules":{"a":"warn"}}`), []byte(`{"rules":{"a":"error"}}`), Minor},
		{"options changed", []byte(`{"rules":{"a":["error",{}]}}`), []byte(`{"rules":{"a":["error",{"x":1}]}}`), Minor},
		{"an ignore pattern dropped", []byte(`{"ignorePatterns":["dist"]}`), []byte(`{}`), Minor},
		{"a set added", nil, []byte(`{"rules":{"a":"error"}}`), Minor},
		{"a set removed", []byte(`{"rules":{"a":"error"}}`), nil, Major},
	} {
		changes, err := DiffSet(testCase.before, testCase.after)
		if err != nil {
			t.Fatalf("%s: %v", testCase.name, err)
		}
		if changes.Bump != testCase.want {
			t.Errorf("%s: bump %s, want %s", testCase.name, changes.Bump, testCase.want)
		}
	}
}

func TestASeverityThatIsNotOneIsRefused(t *testing.T) {
	t.Parallel()
	if _, err := DiffSet([]byte(`{"rules":{"a":"error"}}`), []byte(`{"rules":{"a":"fatal"}}`)); err == nil {
		t.Fatal("a severity of \"fatal\" was read as one")
	}
}

// A version moves at least as far as the change allows, and always forward.
func TestRequireBump(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		previous, next string
		bump           Bump
		allowed        bool
	}{
		{"1.0.0", "1.0.1", Patch, true},
		{"1.0.0", "1.1.0", Patch, true},
		{"1.0.0", "1.0.1", Minor, false},
		{"1.0.0", "1.1.0", Minor, true},
		{"1.4.2", "2.0.0", Minor, true},
		{"1.0.0", "1.1.0", Major, false},
		{"1.0.0", "2.0.0", Major, true},
		{"1.0.0", "1.0.0", Patch, false},
		{"1.2.0", "1.1.9", Patch, false},
		{"2.0.0", "1.9.9", Patch, false},
		{"1.0.0", "1.1", Patch, false},
		{"1.0.0", "v1.1.0", Patch, false},
	} {
		err := RequireBump(testCase.previous, testCase.next, testCase.bump)
		if (err == nil) != testCase.allowed {
			t.Errorf("%s -> %s at %s: allowed %v, err %v", testCase.previous, testCase.next, testCase.bump, err == nil, err)
		}
	}
}

func TestRenderNamesEverySetThatChanged(t *testing.T) {
	t.Parallel()
	changed, bump, err := Diff(
		map[string][]byte{"typescript": []byte(`{"rules":{"a":"error"}}`), "react": []byte(`{"rules":{"b":"error"}}`)},
		map[string][]byte{"typescript": []byte(`{"rules":{"a":"error","c":"error"}}`), "react": []byte(`{"rules":{"b":"error"}}`)},
	)
	if err != nil {
		t.Fatal(err)
	}
	section := Render("1.1.0", "2026-10-03", changed, nil)
	want := "## 1.1.0 (2026-10-03)\n\n### `cohere:typescript`\n\n- On: `c (error)`\n"
	if section != want || bump != Minor {
		t.Fatalf("rendered (bump %s):\n%s\nwant:\n%s", bump, section, want)
	}
	if strings.Contains(section, "react") {
		t.Fatal("a set that did not change was listed")
	}
}

// The sets read from a commit are the sets the working tree holds when nothing is uncommitted: the same
// names and the same bytes, so a diff between HEAD and the tree is empty.
func TestSetsAtACommitMatchTheTree(t *testing.T) {
	t.Parallel()
	module, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if status, err := exec.Command("git", "-C", module, "status", "--porcelain", "--", SetDirectory).Output(); err != nil || len(status) > 0 {
		t.Skipf("NOT MEASURED: the sets have uncommitted changes, so HEAD and the tree differ by design (%s)", bytes.TrimSpace(status))
	}
	committed, err := SetsAt(module, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := SetsAt(module, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(committed) < 2 || len(committed) != len(tree) {
		t.Fatalf("read %d sets at HEAD and %d in the tree", len(committed), len(tree))
	}
	for name, contents := range committed {
		if !bytes.Equal(contents, tree[name]) {
			t.Errorf("cohere:%s differs between HEAD and the tree", name)
		}
	}
	if _, ok := committed["system-inc/structure"]; !ok {
		t.Error("a nested set was not named by its path under the set directory")
	}
}
