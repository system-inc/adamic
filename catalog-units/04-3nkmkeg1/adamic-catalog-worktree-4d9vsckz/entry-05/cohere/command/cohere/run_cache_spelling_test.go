package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// runFrom runs the fixture's binary in dir as a shell there would, with PWD naming dir, so the run spells its
// root the way dir does. Arguments are fixture.run's: verbose, cached, no formatting.
func (fixture *runCacheFixture) runFrom(dir string) string {
	fixture.t.Helper()
	command := exec.Command(fixture.binary, verboseArguments([]string{"--no-format"})...)
	command.Dir = dir
	command.Env = append([]string{"HOME=" + fixture.home, "PWD=" + dir}, withoutVariables(os.Environ(), "HOME", "PWD")...)
	output, err := command.CombinedOutput()
	if _, isExit := err.(*exec.ExitError); err != nil && !isExit {
		fixture.t.Fatalf("running cohere in %s: %v\n%s", dir, err, output)
	}
	return string(output)
}

// withoutVariables is the environment without the named variables, which a run sets for itself.
func withoutVariables(environment []string, names ...string) []string {
	kept := make([]string, 0, len(environment))
	for _, variable := range environment {
		name, _, _ := strings.Cut(variable, "=")
		if !contains(names, name) {
			kept = append(kept, variable)
		}
	}
	return kept
}

func contains(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}

// Every spelling of one project root is one cache (#547dhjz). /tmp and /private/tmp are one directory on macOS,
// as is a project reached through a linked directory, and the cache directory inside the project is the same
// under each. Keyed by the spelled path, a run from the other spelling missed everything and then wrote over
// what the first spelling's next run would read: alternating, every run checked every file.
//
// The spellings here are the fixture's root as made, a symbolic link to it in another temporary directory, and
// the root's resolved target, which on macOS differs from the first, since /var is a link to /private/var. After
// the first run records, each spelling in turn replays. An edit made through the link replays every other file's
// findings, and a replay prints its paths as the replaying run spells the root.
func TestEverySpellingOfTheRootIsOneCache(t *testing.T) {
	t.Parallel()
	fixture := newRunCacheFixture(t, buildCohere(t))
	fixture.write("source/debug.ts", "export function stop(): void {\n    debugger;\n}\n")
	fixture.commit("a finding, so a replay prints a path")

	link := filepath.Join(t.TempDir(), "linked-project")
	if err := os.Symlink(fixture.root, link); err != nil {
		t.Skipf("this filesystem cannot make symbolic links: %v", err)
	}
	spellings := []string{fixture.root, link}
	if resolved, err := filepath.EvalSymlinks(fixture.root); err == nil && resolved != fixture.root {
		spellings = append(spellings, resolved)
	}

	// Primed from the root as made, as establishHit does: a run over files written this second is not
	// recorded, so the second run is the one that records, and the third replays.
	fixture.runFrom(fixture.root)
	fixture.runFrom(fixture.root)
	if primed := fixture.runFrom(fixture.root); !isRunCacheReplay(primed) {
		t.Fatalf("an unchanged tree did not replay from its own spelling, so nothing below can show another was noticed:\n%s", primed)
	}
	for round := range 2 {
		for _, spelling := range append(spellings[1:], spellings[0]) {
			output := fixture.runFrom(spelling)
			if !isRunCacheReplay(output) {
				t.Fatalf("round %d: a run from %s did not replay the run recorded from another spelling of the same root:\n%s",
					round+1, spelling, output)
			}
			if want := filepath.Join(spelling, "source", "debug.ts"); !strings.Contains(output, want) {
				t.Errorf("round %d: the replay from %s does not print %s, the finding's path as this run spells the root:\n%s",
					round+1, spelling, want, output)
			}
			// At a line's start, where a finding's path is printed: /tmp/x is inside /private/tmp/x.
			for _, other := range spellings {
				printed := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(filepath.Join(other, "source", "debug.ts")+":"))
				if other != spelling && printed.MatchString(output) {
					t.Errorf("round %d: the replay from %s prints the path as %s spells it:\n%s", round+1, spelling, other, output)
				}
			}
		}
	}

	// An edit made through the link: the edited file is checked, and the files it leaves alone replay their
	// findings from the entries the other spellings recorded.
	fixture.write("source/a.ts", "// edited through the link\nexport const a: number = 1;\n")
	edited := fixture.runFrom(link)
	if isRunCacheReplay(edited) {
		t.Fatalf("a run after an edit replayed:\n%s", edited)
	}
	replayed := regexp.MustCompile(`(\d+) of (\d+) files replayed from cache`).FindStringSubmatch(edited)
	if replayed == nil || replayed[1] == "0" {
		t.Errorf("after an edit through the link, no file's findings replayed from the entries another spelling recorded:\n%s", edited)
	}
	if again := fixture.runFrom(spellings[len(spellings)-1]); !isRunCacheReplay(again) {
		t.Errorf("the run after the link's recorded nothing another spelling could replay:\n%s", again)
	}
}

// A run that could have replayed and did not says why, in --verbose and --json (#547dhjz): here, a run after an
// edit names the input that changed, and a run under another binary names the binary.
func TestARunThatMissesTheCacheSaysWhy(t *testing.T) {
	t.Parallel()
	fixture := newRunCacheFixture(t, buildCohere(t))
	fixture.establishHit()

	fixture.write("source/a.ts", "// edited\nexport const a: number = 1;\n")
	edited, _ := fixture.run(true)
	if !strings.Contains(edited, "cache: not replayed whole: ") || !strings.Contains(edited, "a.ts") {
		t.Errorf("a run after an edit does not say which input kept it from replaying:\n%s", edited)
	}
	// And --json counts the files whose findings did replay, which it used to report as none.
	fixture.write("source/a.ts", "// edited again\nexport const a: number = 1;\n")
	counted, _ := fixture.run(true, "--json")
	if !regexp.MustCompile(`"filesReplayed":[1-9]`).MatchString(counted) {
		t.Errorf("a run that replayed files' findings reports none in --json:\n%s", counted)
	}
	// And a whole replay counts every file in scope, never the 0 a full check reads.
	if replayed, _ := fixture.run(true, "--json"); !isRunCacheReplay(replayed) && !strings.Contains(replayed, `"replayed":true`) {
		t.Errorf("an unchanged tree's second --json run did not replay:\n%s", replayed)
	} else if !regexp.MustCompile(`"filesReplayed":[1-9]`).MatchString(replayed) {
		t.Errorf("a whole replay reports no file replayed in --json:\n%s", replayed)
	}

	copied := filepath.Join(t.TempDir(), "cohere")
	contents, err := os.ReadFile(fixture.binary)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(copied, contents, 0o755); err != nil {
		t.Fatal(err)
	}
	other := *fixture
	other.binary = copied
	output, _ := other.run(true, "--json")
	for _, want := range []string{`"missReason":"the run on record was taken under another key: the cohere binary changed"`,
		`"findingsMissReason":"the findings on record were taken under another key: the cohere binary changed"`} {
		if !strings.Contains(output, want) {
			t.Errorf("a run under another binary's --json does not say %s:\n%s", want, output)
		}
	}
}
