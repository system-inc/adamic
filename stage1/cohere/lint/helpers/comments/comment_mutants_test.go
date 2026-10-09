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
)

const testCommentMutantsShards = 5

// Each fixed mutant is a top-level unit; growing the witness corpus never changes
// its owner. Every unit checks all witnesses with the original sanitizer/leak checks.
// ADAMIC_TEST_SHARD=i/n selects units locally; unset runs every unit. The generated
// witnesses come from pinned cohere 7945d102a6c18dd36adf9114a758ce646e8b2359.
func TestCommentMutants_000(t *testing.T) { runCommentMutant(t, 0) }
func TestCommentMutants_001(t *testing.T) { runCommentMutant(t, 1) }
func TestCommentMutants_002(t *testing.T) { runCommentMutant(t, 2) }
func TestCommentMutants_003(t *testing.T) { runCommentMutant(t, 3) }
func TestCommentMutants_004(t *testing.T) { runCommentMutant(t, 4) }

func commentMutants() []struct{ file, old, new string } {
	return []struct{ file, old, new string }{
		{"can_begin_at.ts", "if(position === 0 && text.startsWith('#!'))", "if(false)"},
		{"collect_list_interiors.ts", "if(depth === 1)", "if(false)"},
		{"sort_by_position.ts", "> current.start", "< current.start"},
		{"all.ts", "anchors[node.end] = true;", "anchors[node.end] = false;"},
		{"for_file.ts", "if(!this.ready)", "if(true)"},
	}
}

func runCommentMutant(t *testing.T, index int) {
	t.Helper()
	t.Parallel()
	mutants := commentMutants()
	if len(mutants) != testCommentMutantsShards {
		t.Fatalf("enumerated %d mutants, declared %d shards", len(mutants), testCommentMutantsShards)
	}
	if selection := os.Getenv("ADAMIC_TEST_SHARD"); selection != "" {
		parts := strings.Split(selection, "/")
		if len(parts) != 2 {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
		}
		selected, err := strconv.Atoi(parts[0])
		if err != nil {
			t.Fatal(err)
		}
		count, err := strconv.Atoi(parts[1])
		if err != nil || count < 1 || selected < 0 || selected >= count {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", selection)
		}
		if index%count != selected {
			t.Skip("excluded by ADAMIC_TEST_SHARD")
		}
	}
	// The planted-failure subprocess supplies prepared outputs to this exact unit
	// and its production survivor assertion, without rebuilding five binaries.
	if os.Getenv("ADAMIC_COMMENT_MUTANT_SURVIVOR_PROBE") == "1" {
		got, want := []byte("mutant\n"), []byte("oracle\n")
		if index == 2 {
			got = want
		}
		requireCommentMutantKilled(t, got, want)
		return
	}
	m := mutants[index]
	t.Logf("shard-%03d: mutant %s, full witness corpus", index, m.file)
	path, err := filepath.Abs("testdata/witnesses.json")
	if err != nil {
		t.Fatal(err)
	}
	want := run(t, "", oracle(t), path)
	directory := t.TempDir()
	for _, file := range []string{"main.ts", "comment.ts", "can_begin_at.ts", "collect_list_interiors.ts", "sort_by_position.ts", "all.ts", "for_file.ts"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if file == m.file {
			if strings.Count(text, m.old) != 1 {
				t.Fatalf("mutant anchor count %d", strings.Count(text, m.old))
			}
			text = strings.Replace(text, m.old, m.new, 1)
		}
		parserRoot, err := filepath.Abs("../../../../typescript")
		if err != nil {
			t.Fatal(err)
		}
		options, err := filepath.Abs("../options_json.ts")
		if err != nil {
			t.Fatal(err)
		}
		text = strings.ReplaceAll(text, "../../../../typescript", filepath.ToSlash(parserRoot))
		text = strings.ReplaceAll(text, "../options_json.ts", filepath.ToSlash(options))
		if err := os.WriteFile(filepath.Join(directory, file), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	got := run(t, "", build(t, directory), path)
	requireCommentMutantKilled(t, got, want)
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", i+1, a[i], b[i])
			break
		}
	}
}

func requireCommentMutantKilled(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		t.Fatal("compiled semantic mutant survived")
	}
}

func TestCommentMutantsUnion(t *testing.T) {
	mutants := commentMutants()
	if len(mutants) != testCommentMutantsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(mutants), testCommentMutantsShards)
	}
	data, err := os.ReadFile("testdata/witnesses.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Name string }
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("comment mutant corpus must be non-empty")
	}
	expected := map[string]bool{}
	for _, m := range mutants {
		for row, witness := range rows {
			key := fmt.Sprintf("%s/testdata/witnesses.json/%d/%s", m.file, row, witness.Name)
			if witness.Name == "" || expected[key] {
				t.Fatalf("empty or repeated unsplit case %q", key)
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
				t.Fatalf("repeated or unexpected shard case %q", key)
			}
			actual[key] = true
			count++
		}
	}
	if count != len(expected) {
		t.Fatalf("union count %d, unsplit %d", count, len(expected))
	}
	for key := range expected {
		if !actual[key] {
			t.Fatalf("missing shard case %q", key)
		}
	}
	t.Logf("union: %d mutants x %d witnesses = %d unique cases", len(mutants), len(rows), count)
}

func TestCommentMutantsPlantedFailure(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for shard := 0; shard < testCommentMutantsShards; shard++ {
		name := fmt.Sprintf("TestCommentMutants_%03d", shard)
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		command := exec.CommandContext(ctx, binary, "-test.run=^"+name+"$", "-test.timeout=75s", "-test.v")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if err == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return err
		}
		command.WaitDelay = time.Second
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(value, "ADAMIC_COMMENT_MUTANT_SURVIVOR_PROBE=") {
				command.Env = append(command.Env, value)
			}
		}
		command.Env = append(command.Env, "ADAMIC_COMMENT_MUTANT_SURVIVOR_PROBE=1")
		output, err := command.CombinedOutput()
		cancel()
		if shard == 2 {
			if err == nil || !bytes.Contains(output, []byte("compiled semantic mutant survived")) || !bytes.Contains(output, []byte("--- FAIL: "+name)) {
				t.Fatalf("%s did not catch planted survivor: %v\n%s", name, err, output)
			}
			t.Logf("planted survivor caught only by %s (shard-002)", name)
		} else if err != nil {
			t.Fatalf("survivor leaked into %s: %v\n%s", name, err, output)
		}
	}
}
