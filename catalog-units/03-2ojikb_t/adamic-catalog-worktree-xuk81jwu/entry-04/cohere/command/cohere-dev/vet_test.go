//go:build unix

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// unkeyedLiteral is a module whose package user builds another package's struct with unkeyed fields, which
// go vet's composites check reports and go test's curated vet does not (#kr2yp54). keyed is the same
// literal with its fields named, which passes both.
var (
	unkeyedLiteral = map[string]string{
		"go.mod":            "module example.com/root\n\ngo 1.21\n",
		"other/other.go":    "package other\n\ntype Range struct{ Start, End int }\n",
		"user/user.go":      "package user\n\nimport \"example.com/root/other\"\n\nvar Span = other.Range{1, 2}\n",
		"user/user_test.go": "package user\n\nimport \"testing\"\n\nfunc TestUser(t *testing.T) {}\n",
	}
	keyed = "package user\n\nimport \"example.com/root/other\"\n\nvar Span = other.Range{Start: 1, End: 2}\n"
)

func writeFiles(t *testing.T, directory string, files map[string]string) {
	t.Helper()
	for name, contents := range files {
		path := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// The landing gate runs every check plain go vet runs: an unkeyed literal of another package's struct
// refuses the landing, and the same landing lands once its fields are keyed. This runs the real go through
// the land fixture.
func TestTheGateRefusesWhatPlainGoVetReports(t *testing.T) {
	t.Parallel()
	fixture := newLandFixture(t)
	// The real go: the stand-in's directory comes off the front of PATH.
	fixture.environment = append(fixture.environment, "PATH="+os.Getenv("PATH"))
	a := fixture.worktree("a")
	writeFiles(t, a, unkeyedLiteral)
	fixture.git(a, "add", "go.mod", "other", "user")
	fixture.git(a, "commit", "-q", "-m", "an unkeyed literal of another package's struct")
	land := func() (string, error) {
		landing := exec.Command(fixture.wrapper, "land")
		landing.Dir, landing.Env = a, fixture.environment
		output, err := landing.CombinedOutput()
		return string(output), err
	}
	if output, err := land(); err == nil || !strings.Contains(output, "struct literal uses unkeyed fields") {
		t.Fatalf("an unkeyed literal of another package's struct did not refuse the landing (%v):\n%s", err, output)
	}
	writeFiles(t, a, map[string]string{"user/user.go": keyed})
	fixture.git(a, "add", "user/user.go")
	fixture.git(a, "commit", "-q", "-m", "its fields keyed")
	if output, err := land(); err != nil || !strings.Contains(output, "landed") {
		t.Errorf("the landing did not land once the fields were keyed (%v):\n%s", err, output)
	}
}

// cohere-dev test vets with every check the gate's go vet runs, not go test's curated subset, so the edit
// loop fails where the gate would: 130 unkeyed literals passed cohere-dev test and failed plain go vet
// (#kr2yp54). A caller's own -vet is kept.
func TestCohereDevTestVetsWithEveryCheck(t *testing.T) {
	t.Parallel()
	wrapper := filepath.Join(t.TempDir(), "cohere-dev")
	if output, err := exec.Command("go", "build", "-o", wrapper, ".").CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, output)
	}
	module := t.TempDir()
	writeFiles(t, module, unkeyedLiteral)
	home := privatePool(t)
	test := func(arguments ...string) (string, error) {
		command := exec.Command(wrapper, append([]string{"test"}, arguments...)...)
		command.Dir = module
		command.Env = append(outsideThePool(os.Environ()), homeVariable+"="+home, "GOWORK=off")
		output, err := command.CombinedOutput()
		return string(output), err
	}
	if output, err := test("./user"); err == nil || !strings.Contains(output, "struct literal uses unkeyed fields") {
		t.Errorf("cohere-dev test passed an unkeyed literal that go vet reports (%v):\n%s", err, output)
	}
	if output, err := test("-vet=off", "./user"); err != nil {
		t.Errorf("cohere-dev test did not keep the caller's -vet=off (%v):\n%s", err, output)
	}
}
