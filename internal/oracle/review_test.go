package oracle

import (
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

var reviewPending = regexp.MustCompile(`^awaits [A-Za-z0-9_./-]+: [^\r\n]+$`)

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
		t.Fatalf("invalid pending message %q; want awaits <branch>: <reason>", message)
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
	reviewPrograms(t, "agree", func(t *testing.T, path string) {
		if err := reviewAgreement(t, path, nil); err != nil {
			t.Fatal(err)
		}
	})
}

func TestReviewProgramsRefuse(t *testing.T) {
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
	if err := reviewLayout(reviewRoot); err != nil {
		t.Fatal(err)
	}
}

// Exercise the same checks used by the lane with deliberately wrong inputs.
func TestReviewProgramsSelfTest(t *testing.T) {
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
