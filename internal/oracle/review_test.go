package oracle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
	base := strings.TrimSuffix(path, filepath.Ext(path))
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
	data, err := os.ReadFile(strings.TrimSuffix(path, filepath.Ext(path)) + ".pending")
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

// reviewProgramPaths is every review program in directory, .a and .ts alike, in name order. Each finds its sidecars
// by its own name with its extension taken off, so a .ts program's pending sidecar is found as an .a program's is.
func reviewProgramPaths(directory string) ([]string, error) {
	var paths []string
	for _, pattern := range []string{"*.a", "*.ts"} {
		matches, err := filepath.Glob(filepath.Join(directory, pattern))
		if err != nil {
			return nil, err
		}
		paths = append(paths, matches...)
	}
	slices.Sort(paths)
	return paths, nil
}

func reviewPrograms(t *testing.T, category string, check func(*testing.T, string)) {
	t.Helper()
	paths, err := reviewProgramPaths(filepath.Join(reviewRoot, category))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no review programs in %s", category)
	}
	reviewEach(t, paths, check)
}

func reviewEach(t *testing.T, paths []string, check func(*testing.T, string)) {
	t.Helper()
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
	// reviewEach passes an absolute path to load.Load without a display mapper.
	// Program.Where restores an .a program's extension and prints that full path unchanged, as it prints a .ts program's.
	return reviewRefusalDiagnostic(path, err)
}

func reviewRefusalDiagnostic(path string, err error) error {
	var refused *lower.Refused
	var notYet *lower.NotYet
	if err == nil {
		return fmt.Errorf("accepted refused review program %s", path)
	}
	if (!errors.As(err, &refused) && !errors.As(err, &notYet)) || strings.TrimSpace(err.Error()) == "" {
		return fmt.Errorf("want compiler refusal with diagnostic, got %v", err)
	}
	where := ""
	if refused != nil {
		where = refused.Where
	} else {
		where = notYet.Where
	}
	location := regexp.MustCompile(`^` + regexp.QuoteMeta(filepath.ToSlash(path)) + `:[1-9][0-9]*:[1-9][0-9]*$`)
	if !location.MatchString(where) {
		return fmt.Errorf("want refusal location %s:<positive line>:<positive column>, got %q: %w", path, where, err)
	}
	return nil
}

// reviewAgreePrograms is every agree program's file name, the key TestReviewProgramsAgreeWithNode's family
// (families_test.go) owns it by.
func reviewAgreePrograms() []string {
	paths, _ := reviewProgramPaths(filepath.Join(reviewRoot, "agree"))
	names := make([]string, len(paths))
	for index, path := range paths {
		names[index] = filepath.Base(path)
	}
	return names
}

// TestReviewProgramsAgreeWithNode is a family (families_test.go): TestReviewProgramsAgreeWithNode_000 and on each run
// the agree programs their shard owns, keeping the TestReviewPrograms prefix Judge's review lane reads.
func runReviewAgreeShard(t *testing.T, shard int) {
	t.Helper()
	planted, probing := familyProbe("TestReviewProgramsAgreeWithNode")
	names := reviewAgreePrograms()
	if len(names) == 0 {
		t.Fatal("no review programs in agree")
	}
	var paths []string
	for _, index := range familyCases(t, "TestReviewProgramsAgreeWithNode", shard, names) {
		paths = append(paths, filepath.Join(reviewRoot, "agree", names[index]))
	}
	reviewEach(t, paths, func(t *testing.T, path string) {
		expected := reviewPinned(t, path)
		if probing && filepath.Base(path) == planted {
			// TestFamiliesCatchAPlantedCase's mutant: TestReviewProgramsSelfTest's wrong output, one byte more than Node's.
			wrong := onNode(t, path)
			wrong.stdout = append(append([]byte(nil), wrong.stdout...), '!')
			expected = &wrong
		}
		if err := reviewAgreement(t, path, expected); err != nil {
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

// reviewLayout refuses a program outside agree/ and refused/, and an .a and a .ts of one name, which would share
// their sidecars.
func reviewLayout(root string) error {
	stems := map[string]bool{}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !slices.Contains([]string{".a", ".ts"}, filepath.Ext(path)) {
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
		stem := strings.TrimSuffix(relative, filepath.Ext(relative))
		if stems[stem] {
			return fmt.Errorf("review program %s is both .a and .ts, which would share its sidecars", stem)
		}
		stems[stem] = true
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
	t.Run("refusal_location", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "program[1].a")
		for _, where := range []string{path + ":1:2", path + ":123:456"} {
			for _, err := range []error{&lower.Refused{Where: where, What: "test", Fix: "test"}, &lower.NotYet{Where: where, What: "test"}} {
				if got := reviewRefusalDiagnostic(path, err); got != nil {
					t.Errorf("valid location rejected: %v", got)
				}
			}
		}
		for _, where := range []string{"", filepath.Base(path) + ":1:2", "/other" + path + ":1:2", path + ".ts:1:2", path + ":0:1", path + ":1:0", path + ":-1:2", path + ":1", path + ":1:2:3", path + ":line:2", path + ":1:2\n"} {
			for _, err := range []error{&lower.Refused{Where: where, What: "test", Fix: "test"}, &lower.NotYet{Where: where, What: "test"}} {
				if reviewRefusalDiagnostic(path, err) == nil {
					t.Errorf("invalid location passed: %q", where)
				}
			}
		}
	})
	t.Run("loose_file", func(t *testing.T) {
		for _, name := range []string{"loose.a", "loose.ts"} {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, name), []byte("console.log(\"accepted\");\n"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := reviewLayout(root); err == nil {
				t.Errorf("loose program %s passed", name)
			}
		}
	})
	t.Run("one_name_twice", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "agree"), 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"twin.a", "twin.ts"} {
			if err := os.WriteFile(filepath.Join(root, "agree", name), []byte("console.log(\"accepted\");\n"), 0644); err != nil {
				t.Fatal(err)
			}
		}
		if err := reviewLayout(root); err == nil || !strings.Contains(err.Error(), "both .a and .ts") {
			t.Fatalf("an .a and a .ts of one name passed: %v", err)
		}
	})
}
