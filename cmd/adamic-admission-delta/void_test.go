package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// A given seed replaces head's SHA-256: two heads sample the same programs under one seed, two seeds sample different
// ones, and a seed that isn't hex of at least 8 bytes is refused. Mutant: samplingSeed ignoring the given seed (the two
// heads sample differently).
func TestGivenSeedReplacesHead(test *testing.T) {
	test.Parallel()
	var programs []entry
	for index := 0; index < 40; index++ {
		programs = append(
			programs,
			entry{
				program: program{Path: fmt.Sprintf("input-%02d.a", index)},
				Corpus:  "fixtures",
				Class:   "newly-accepted",
			},
		)
	}
	selection := func(head, given string) []int {
		test.Helper()
		seed, samplingSeedError := samplingSeed(head, given)
		if samplingSeedError != nil {
			test.Fatal(samplingSeedError)
		}
		selected, _ := sample(programs, seed, 50, time.Second)
		return selected
	}
	if reflect.DeepEqual(selection("head-one", ""), selection("head-two", "")) {
		test.Fatal(
			"two heads sampled the same programs without a seed, so this test can't tell a given seed from head's",
		)
	}
	given := "0123456789abcdef0123456789abcdef"
	if !reflect.DeepEqual(selection("head-one", given), selection("head-two", given)) {
		test.Fatal("one given seed sampled different programs for two heads")
	}
	if reflect.DeepEqual(selection("head-one", given), selection("head-one", "fedcba9876543210")) {
		test.Fatal("two given seeds sampled the same programs")
	}
	for _, bad := range []string{"zz", "0123456789abcd", "0123456789abcdeg"} {
		if _, samplingSeedError := samplingSeed("head", bad); samplingSeedError == nil {
			test.Errorf("seed %q accepted", bad)
		}
	}
}

// The JSON records the seed used: the given one, or head's SHA-256 when none is given. Mutant: sampling_seed
// recording head's SHA-256 even when a seed is given.
func TestSeedIsRecorded(test *testing.T) {
	test.Parallel()
	dir, sha, _, head := commandFixture(test)
	args := []string{
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		head,
		"--head-binary",
		head,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
	}
	digest := sha256.Sum256([]byte(sha))
	for _, testCase := range []struct {
		given, want string
	}{{"", hex.EncodeToString(digest[:])}, {"00112233445566778899", "00112233445566778899"}} {
		run := args
		if testCase.given != "" {
			run = append(append([]string(nil), args...), "--seed", testCase.given)
		}
		observation := invokeCommand(test, dir, run...)
		var report report
		if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
			test.Fatal(decodeError, observation.Stderr)
		}
		if observation.Exit != 0 || report.SamplingSeed != testCase.want || report.SeedGiven != (testCase.given != "") {
			test.Fatalf(
				"seed %q: exit %d, sampling_seed %q seed_given %v, want %q",
				testCase.given,
				observation.Exit,
				report.SamplingSeed,
				report.SeedGiven,
				testCase.want,
			)
		}
	}
	if observation := invokeCommand(
		test,
		dir,
		append(args, "--seed", "not-hex")...,
	); observation.Exit != 1 || !strings.Contains(observation.Stderr, "must be hex") {
		test.Fatalf("a bad seed ran: exit %d %s", observation.Exit, observation.Stderr)
	}
}

// A compile timeout, a runtime timeout and a missing native toolchain are void: exit 3, "void": true and a reason,
// never a red. Mutants: a compile timeout charged as a red again (strings.HasPrefix catching compiler-timeout, exit
// 1); timedOut always empty (the native timeout is a mismatch, exit 1); the clang check removed (the program runs and
// mismatches, exit 1); exitCode returning 1 for a void.
func TestTimeoutsAndMissingToolchainAreVoid(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		name        string
		head        string
		environment func(test *testing.T) []string
		extra       []string
		reason      string
	}{
		{
			name:   "compile timeout",
			head:   "#!/bin/sh\nsleep 5\n",
			extra:  []string{"--compile-timeout", "300ms"},
			reason: "corpus/input.a: compile timeout",
		},
		{
			name:   "runtime timeout",
			head:   "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) echo \"console.log('42');\";;\nbuild) printf '#!/bin/sh\\nsleep 5\\n' > \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n",
			extra:  []string{"--timeout", "1s"},
			reason: "corpus/input.a: native timeout",
		},
		{
			// The fixture's head builds a program printing 41 against Node's 42, so a run that went ahead would be red.
			name: "missing toolchain",
			environment: func(test *testing.T) []string {
				return []string{"ADAMIC_SANDBOX=1", "ADAMIC_TOOLS=" + test.TempDir()}
			},
			reason: "native toolchain missing: clang",
		},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			dir, sha, base, head := commandFixture(test)
			if testCase.head != "" {
				writeExecutable(test, head, testCase.head)
			}
			var environment []string
			if testCase.environment != nil {
				environment = testCase.environment(test)
			}
			args := append([]string{
				"--base",
				sha,
				"--head",
				sha,
				"--base-binary",
				base,
				"--head-binary",
				head,
				"--manifest",
				filepath.Join(dir, "manifest.json"),
				"--json",
			}, testCase.extra...)
			observation := invokeCommandWith(test, dir, environment, args...)
			var report report
			if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
				test.Fatalf("%v: %s %s", decodeError, observation.Stdout, observation.Stderr)
			}
			if observation.Exit != 3 ||
				!report.Void ||
				report.Verdict != "void" ||
				!strings.HasPrefix(report.VoidReason, testCase.reason) {
				test.Fatalf(
					"exit %d, void %v, verdict %s, reason %q, want exit 3 and a void for %q: %s",
					observation.Exit,
					report.Void,
					report.Verdict,
					report.VoidReason,
					testCase.reason,
					observation.Stderr,
				)
			}
		})
	}
}

// A compiler error the tool doesn't recognise stays a red, alone or beside another program's compile timeout: exit 1,
// "void": false. Mutant: the void winning over a red (verdict void whenever a reason exists, exit 3).
func TestUnrecognisedCompilerErrorStaysRed(test *testing.T) {
	test.Parallel()
	dir, sha, base, head := commandFixture(test)
	writeExecutable(test, base, "#!/bin/sh\necho 'missing input' >&2\nexit 1\n")
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
	)
	var r report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &r); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	if observation.Exit != 1 || r.Void || r.Verdict != "fail" || r.Programs[0].Class != "compiler-error" {
		test.Fatalf(
			"an unrecognised compiler error: exit %d, void %v, verdict %s, want a red: %s",
			observation.Exit,
			r.Void,
			r.Verdict,
			observation.Stderr,
		)
	}
	// A second program whose compile times out, beside the first's error.
	if writeError := os.WriteFile(
		filepath.Join(dir, "corpus/slow.a"),
		[]byte("console.log('slow');\n"),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	for _, args := range [][]string{{"add", "corpus/slow.a"}, {"commit", "-qm", "slow"}} {
		if _, gitError := git(dir, args...); gitError != nil {
			test.Fatal(gitError)
		}
	}
	writeExecutable(
		test,
		base,
		"#!/bin/sh\ncase \"$2\" in\n*slow.a) sleep 5;;\nesac\necho 'missing input' >&2\nexit 1\n",
	)
	observation = invokeCommand(
		test,
		dir,
		"--base",
		"HEAD",
		"--base-binary",
		base,
		"--head-binary",
		base,
		"--corpus",
		"corpus",
		"--json",
		"--compile-timeout",
		"300ms",
	)
	r = report{}
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &r); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	if observation.Exit != 1 ||
		r.Void ||
		r.Verdict != "fail" ||
		!strings.Contains(r.VoidReason, "corpus/slow.a: compile timeout") {
		test.Fatalf(
			"an error beside a compile timeout: exit %d, void %v, verdict %s, reason %q, want a red: %s",
			observation.Exit,
			r.Void,
			r.Verdict,
			r.VoidReason,
			observation.Stderr,
		)
	}
}

// headSeed is the seed a run with no --seed samples by: head's SHA-256.
func headSeed(head string) []byte {
	digest := sha256.Sum256([]byte(head))
	return digest[:]
}
