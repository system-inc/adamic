package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

const testCommentMutantsShards = 5

// Each fixed mutant is a top-level unit; growing the witness corpus never changes
// its owner. Every unit checks all witnesses with the original sanitizer/leak checks.
// ADAMIC_TEST_SHARD=i/n selects units locally; unset runs every unit. The generated
// witnesses come from pinned cohere 7945d102a6c18dd36adf9114a758ce646e8b2359.
func TestCommentMutants_000(test *testing.T) {
	test.Parallel()
	runCommentMutant(test, 0)
}
func TestCommentMutants_001(test *testing.T) {
	test.Parallel()
	runCommentMutant(test, 1)
}
func TestCommentMutants_002(test *testing.T) {
	test.Parallel()
	runCommentMutant(test, 2)
}
func TestCommentMutants_003(test *testing.T) {
	test.Parallel()
	runCommentMutant(test, 3)
}
func TestCommentMutants_004(test *testing.T) {
	test.Parallel()
	runCommentMutant(test, 4)
}

// Each mutant's sanitized build is a product made ahead, so its unit only runs it.
func TestProduct_CommentMutants_000(test *testing.T) {
	test.Parallel()
	helpersNative(test, commentMutants()[0])
}
func TestProduct_CommentMutants_001(test *testing.T) {
	test.Parallel()
	helpersNative(test, commentMutants()[1])
}
func TestProduct_CommentMutants_002(test *testing.T) {
	test.Parallel()
	helpersNative(test, commentMutants()[2])
}
func TestProduct_CommentMutants_003(test *testing.T) {
	test.Parallel()
	helpersNative(test, commentMutants()[3])
}
func TestProduct_CommentMutants_004(test *testing.T) {
	test.Parallel()
	helpersNative(test, commentMutants()[4])
}

func commentMutants() []nativeMutation {
	return []nativeMutation{
		{"can_begin_at.ts", "if(position === 0 && text.startsWith('#!'))", "if(false)"},
		{"collect_list_interiors.ts", "if(depth === 1)", "if(false)"},
		{"sort_by_position.ts", "> current.start", "< current.start"},
		{"all.ts", "anchors[node.end] = true;", "anchors[node.end] = false;"},
		{"for_file.ts", "if(!this.ready)", "if(true)"},
	}
}

func runCommentMutant(test *testing.T, index int) {
	test.Helper()
	mutants := commentMutants()
	if len(mutants) != testCommentMutantsShards {
		test.Fatalf("enumerated %d mutants, declared %d shards", len(mutants), testCommentMutantsShards)
	}
	if selection := os.Getenv("ADAMIC_TEST_SHARD"); selection != "" {
		parts := strings.Split(selection, "/")
		if len(parts) != 2 {
			test.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
		}
		selected, parseError := strconv.Atoi(parts[0])
		if parseError != nil {
			test.Fatal(parseError)
		}
		count, parseError := strconv.Atoi(parts[1])
		if parseError != nil || count < 1 || selected < 0 || selected >= count {
			test.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
		}
		if index%count != selected {
			test.Skip("excluded by ADAMIC_TEST_SHARD")
		}
	}
	// The planted-failure subprocess supplies prepared outputs to this exact unit
	// and its production survivor assertion, without rebuilding five binaries.
	if os.Getenv("ADAMIC_COMMENT_MUTANT_SURVIVOR_PROBE") == "1" {
		got, want := []byte("mutant\n"), []byte("oracle\n")
		if index == 2 {
			got = want
		}
		requireCommentMutantKilled(test, got, want)
		return
	}
	mutant := mutants[index]
	test.Logf("shard-%03d: mutant %s, full witness corpus", index, mutant.file)
	path, absolutePathError := filepath.Abs("testdata/witnesses.json")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := run(test, "", oracle(test), path)
	got := run(test, "", helpersNative(test, mutant), path)
	requireCommentMutantKilled(test, got, want)
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for index := 0; index < len(a) && index < len(b); index++ {
		if a[index] != b[index] {
			test.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", index+1, a[index], b[index])
			break
		}
	}
}

func requireCommentMutantKilled(test *testing.T, got, want []byte) {
	test.Helper()
	if bytes.Equal(got, want) {
		test.Fatal("compiled semantic mutant survived")
	}
}

func TestCommentMutantsUnion(test *testing.T) {
	test.Parallel()
	mutants := commentMutants()
	if len(mutants) != testCommentMutantsShards {
		test.Fatalf("enumerated %d shards, declared %d", len(mutants), testCommentMutantsShards)
	}
	data, readError := os.ReadFile("testdata/witnesses.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var rows []struct{ Name string }
	if decodeError := json.Unmarshal(data, &rows); decodeError != nil {
		test.Fatal(decodeError)
	}
	if len(rows) == 0 {
		test.Fatal("comment mutant corpus must be non-empty")
	}
	expected := map[string]bool{}
	for _, mutant := range mutants {
		for row, witness := range rows {
			key := fmt.Sprintf("%s/testdata/witnesses.json/%d/%s", mutant.file, row, witness.Name)
			if witness.Name == "" || expected[key] {
				test.Fatalf("empty or repeated unsplit case %q", key)
			}
			expected[key] = true
		}
	}
	actual := map[string]bool{}
	count := 0
	for shard := 0; shard < testCommentMutantsShards; shard++ {
		for row, witness := range rows {
			key := fmt.Sprintf("%s/testdata/witnesses.json/%d/%s", mutants[shard].file, row, witness.Name)
			if actual[key] || !expected[key] {
				test.Fatalf("repeated or unexpected shard case %q", key)
			}
			actual[key] = true
			count++
		}
	}
	if count != len(expected) {
		test.Fatalf("union count %d, unsplit %d", count, len(expected))
	}
	for key := range expected {
		if !actual[key] {
			test.Fatalf("missing shard case %q", key)
		}
	}
	test.Logf("union: %d mutants x %d witnesses = %d unique cases", len(mutants), len(rows), count)
}

func TestCommentMutantsPlantedFailure(test *testing.T) {
	test.Parallel()
	binary, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	for shard := 0; shard < testCommentMutantsShards; shard++ {
		name := fmt.Sprintf("TestCommentMutants_%03d", shard)
		timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		command := exec.CommandContext(timeoutContext, binary, "-test.run=^"+name+"$", "-test.timeout=75s", "-test.v")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if killError == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return killError
		}
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") &&
				!strings.HasPrefix(value, "ADAMIC_COMMENT_MUTANT_SURVIVOR_PROBE=") {
				command.Env = append(command.Env, value)
			}
		}
		command.Env = append(command.Env, "ADAMIC_COMMENT_MUTANT_SURVIVOR_PROBE=1")
		// Read through childguard.Capture, to its end with no timer deciding when: exec's own copying under a
		// one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after the
		// shard exited, so a shard that had finished could fail this check under load (#n3gk22h's shape). ctx's 90 s
		// deadline ends a hung shard. Stdout comes before stderr; the checks below look for their lines anywhere in
		// either.
		stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
		output := append(stdout, stderr...)
		cancel()
		if shard == 2 {
			if captureError == nil ||
				!bytes.Contains(output, []byte("compiled semantic mutant survived")) ||
				!bytes.Contains(output, []byte("--- FAIL: "+name)) {
				test.Fatalf("%s did not catch planted survivor: %v\n%s", name, captureError, output)
			}
			test.Logf("planted survivor caught only by %s (shard-002)", name)
		} else if captureError != nil {
			test.Fatalf("survivor leaked into %s: %v\n%s", name, captureError, output)
		}
	}
}
