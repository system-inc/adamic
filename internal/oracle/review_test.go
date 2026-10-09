package oracle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/lower"
)

const reviewRoot = repository + "/internal/oracle/testdata/review"

// A pending sidecar names what the program awaits: a branch (the gate's census fails it once the branch is on main
// and the program still skips) or, while no branch exists yet, a task (#<id>; the Mac-side review census pages
// integration once that task is Done and the sidecar remains).
var reviewPending = regexp.MustCompile(`^awaits (#[0-9a-z]{6,9}|[A-Za-z0-9_./-]+): [^\r\n]+$`)

// A ruled difference (an intended departure from Node, such as step 24's loud stack guard where Node throws a
// catchable RangeError) is a <name>.intended sidecar, 'intended: <ruling>: <what differs>', with the pinned output in
// <name>.expected.json ({"stdout", "stderr", "exit"}). The backends are held to the pin instead of Node, and the pin
// must still differ from Node: one that matches Node is stale.
var reviewIntended = regexp.MustCompile(`^intended: [^:\r\n]+: [^\r\n]+$`)

func reviewPinned(t *testing.T, path string) *run {
	t.Helper()
	base := strings.TrimSuffix(path, ".a")
	data, err := os.ReadFile(base + ".intended")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	if message := strings.TrimSuffix(string(data), "\n"); !reviewIntended.MatchString(message) {
		t.Fatalf("invalid intended sidecar %q; want intended: <ruling>: <what differs>", message)
	}
	pin, err := os.ReadFile(base + ".expected.json")
	if err != nil {
		t.Fatalf("an intended difference needs its pinned output in %s.expected.json: %v", filepath.Base(base), err)
	}
	var expected struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Exit   int    `json:"exit"`
	}
	if err := json.Unmarshal(pin, &expected); err != nil {
		t.Fatal(err)
	}
	return &run{stdout: []byte(expected.Stdout), stderr: []byte(expected.Stderr), exitCode: expected.Exit}
}

func reviewSkip(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".pending")
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	message := strings.TrimSuffix(string(data), "\n")
	if !reviewPending.MatchString(message) {
		t.Fatalf("invalid pending message %q; want awaits <branch>: <reason> or awaits #<task>: <reason>", message)
	}
	t.Skip(message)
}

func reviewPrograms(t *testing.T, category string, check func(*testing.T, string)) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(reviewRoot, category, "*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no review programs in %s", category)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			reviewSkip(t, path)
			started := time.Now()
			t.Cleanup(func() {
				if elapsed := time.Since(started); elapsed >= 30*time.Second {
					t.Errorf("review unit took %s; must be under 30s", elapsed)
				}
			})
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			check(t, absolute)
		})
	}
}

// All execution paths use stage 0's existing lowerer and oracle build helpers.
func reviewAgreement(t *testing.T, path string, expected *run) error {
	t.Helper()
	program, err := lowered(t, path)
	if err != nil {
		return fmt.Errorf("stage 0: %w", err)
	}
	truth := onNode(t, path)
	if expected != nil {
		if expected.exitCode == truth.exitCode && string(expected.stdout) == string(truth.stdout) && string(expected.stderr) == string(truth.stderr) {
			return fmt.Errorf("the intended difference is stale: Node now gives the pinned output; drop the .intended and .expected.json")
		}
		truth = *expected
	}
	backend := onJavaScriptBackend(t, program)
	native, sanitized := natively(t, program)
	release := released(t, program)
	var failures []error
	for _, result := range []struct {
		name string
		got  run
	}{{"native", native}, {"JavaScript", backend}, {"release", release}} {
		if difference := disagreement(truth, result.got); difference != "" {
			failures = append(failures, fmt.Errorf("%s: %s; Node exit %d stdout %q stderr %q; got exit %d stdout %q stderr %q", result.name, difference, truth.exitCode, truth.stdout, truth.stderr, result.got.exitCode, result.got.stdout, result.got.stderr))
		}
	}
	if truth.exitCode == 0 {
		if report := leaks(t, program, sanitized); report != "" {
			failures = append(failures, fmt.Errorf("leaks: %s", report))
		}
	}
	return errors.Join(failures...)
}

func reviewRefusal(t *testing.T, path string) error {
	t.Helper()
	_, err := lowered(t, path)
	var refused *lower.Refused
	var notYet *lower.NotYet
	if err == nil {
		return fmt.Errorf("accepted refused review program %s", path)
	}
	if (!errors.As(err, &refused) && !errors.As(err, &notYet)) || strings.TrimSpace(err.Error()) == "" {
		return fmt.Errorf("want compiler refusal with diagnostic, got %v", err)
	}
	return nil
}

func TestReviewProgramsAgreeWithNode(t *testing.T) {
	t.Parallel()
	reviewPrograms(t, "agree", func(t *testing.T, path string) {
		if err := reviewAgreement(t, path, reviewPinned(t, path)); err != nil {
			t.Fatal(err)
		}
	})
}

func TestReviewProgramsRefuse(t *testing.T) {
	t.Parallel()
	reviewPrograms(t, "refused", func(t *testing.T, path string) {
		if err := reviewRefusal(t, path); err != nil {
			t.Fatal(err)
		}
	})
}

func reviewLayout(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".a" {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parts := strings.Split(relative, string(filepath.Separator))
		if len(parts) != 2 || (parts[0] != "agree" && parts[0] != "refused") {
			return fmt.Errorf("loose review program: %s", relative)
		}
		return nil
	})
}

func TestReviewProgramsNoLooseFiles(t *testing.T) {
	t.Parallel()
	if err := reviewLayout(reviewRoot); err != nil {
		t.Fatal(err)
	}
}

// Exercise the same checks used by the lane with deliberately wrong inputs.
func TestReviewProgramsSelfTest(t *testing.T) {
	t.Parallel()
	t.Run("wrong_output", func(t *testing.T) {
		path, err := filepath.Abs(filepath.Join(reviewRoot, "agree", "smoke.a"))
		if err != nil {
			t.Fatal(err)
		}
		wrong := onNode(t, path)
		wrong.stdout = append(append([]byte(nil), wrong.stdout...), '!')
		err = reviewAgreement(t, path, &wrong)
		if err == nil {
			t.Fatal("wrong expected output passed")
		}
		t.Logf("caught planted expected output: %v", err)
		for _, name := range []string{"native", "JavaScript", "release"} {
			if !strings.Contains(err.Error(), name+": stdout differs") {
				t.Errorf("%s comparison did not catch planted output: %v", name, err)
			}
		}
	})
	t.Run("stale_intended", func(t *testing.T) {
		path, err := filepath.Abs(filepath.Join(reviewRoot, "agree", "smoke.a"))
		if err != nil {
			t.Fatal(err)
		}
		same := onNode(t, path)
		if err := reviewAgreement(t, path, &same); err == nil || !strings.Contains(err.Error(), "stale") {
			t.Fatalf("a pin equal to Node's output passed: %v", err)
		}
	})
	t.Run("sidecar_forms", func(t *testing.T) {
		for message, valid := range map[string]bool{
			"awaits compiler/step09-predicates: refuse the helper's second parameter": true,
			"awaits #cxr5x2v: catchable RangeError, step 21":                          true,
			"awaits: no target":          false,
			"awaits #CXR: not a task id": false,
		} {
			if reviewPending.MatchString(message) != valid {
				t.Errorf("pending %q: valid = %v", message, !valid)
			}
		}
		if !reviewIntended.MatchString("intended: step 24's ruling #2yr8d0q: deep recursion is a loud stack guard, exit 70") ||
			reviewIntended.MatchString("intended: no ruling named") {
			t.Error("intended sidecar form misread")
		}
	})
	t.Run("accepted_refused", func(t *testing.T) {
		root := t.TempDir()
		root = filepath.Join(root, "refused")
		if err := os.Mkdir(root, 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, "accepted.a")
		if err := os.WriteFile(path, []byte("console.log(\"accepted\");\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := reviewRefusal(t, path); err == nil || !strings.Contains(err.Error(), "accepted refused review program") {
			t.Fatalf("accepted program was not caught: %v", err)
		}
	})
	t.Run("loose_file", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "loose.a"), []byte("console.log(\"accepted\");\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := reviewLayout(root); err == nil {
			t.Fatal("loose program passed")
		}
	})
}
