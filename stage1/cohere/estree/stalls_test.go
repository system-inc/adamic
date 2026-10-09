package estree

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func refusedBeforeDeadline(t *testing.T, argv []string, diagnostic string) {
	t.Helper()
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	err, timedOut := runWithCPUBudget(t, argv, output, &stderr, 2*time.Second)
	output.Close()
	info, _ := os.Stat(output.Name())
	if timedOut || err == nil || info.Size() != 0 || !strings.Contains(stderr.String(), diagnostic) {
		t.Fatalf("%v: timeout=%v exit=%v stdout=%d stderr=%s", argv, timedOut, err, info.Size(), &stderr)
	}
}

const testBoundedPortParserShards = 16

type boundedPortCase struct {
	id, text, filename string
	audit              bool
}

func boundedPortCases(t *testing.T) []boundedPortCase {
	t.Helper()
	cases := []boundedPortCase{
		{"eof-type", "type X = {", "input.ts", false},
		{"eof-interface", "interface I {", "input.ts", false},
		{"eof-method", "type X = { m(a: string): void;", "input.ts", false},
	}
	fixtures, err := filepath.Glob("validation/followup/stalls/*.input")
	if err != nil || len(fixtures) != 13 {
		t.Fatalf("stall fixtures: %d %v", len(fixtures), err)
	}
	for _, fixture := range fixtures {
		body, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, boundedPortCase{fixture, string(body), strings.TrimSuffix(filepath.Base(fixture), ".input"), true})
	}
	if len(cases) != testBoundedPortParserShards {
		t.Fatal("bounded parser enumeration changed")
	}
	return cases
}

func boundedPortIDs(cases []boundedPortCase) []string {
	ids := make([]string, len(cases))
	for i := range cases {
		ids[i] = cases[i].id
	}
	return ids
}

// ADAMIC_TEST_SHARD=i/n selects shards; unset runs all 13 fixtures and three EOF cases.
func TestBoundedPortParser(t *testing.T) {
	finishSetup := miscStart(t)
	cases := boundedPortCases(t)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := miscBuild(t, main)
	oracle := miscOracle(t)
	finishSetup()
	miscRunShards(t, testBoundedPortParserShards, boundedPortIDs(cases), func(t *testing.T, i int) {
		item := cases[i]
		path := filepath.Join(t.TempDir(), item.filename)
		if err := os.WriteFile(path, []byte(item.text), 0644); err != nil {
			t.Fatal(err)
		}
		accepted := false
		if item.audit {
			list := filepath.Join(t.TempDir(), "manifest")
			if err := os.WriteFile(list, []byte(path+"\n"), 0644); err != nil {
				t.Fatal(err)
			}
			var record struct{ Status string }
			if err := json.Unmarshal(execute(t, "", oracle, "--audit", list, t.TempDir()), &record); err != nil {
				t.Fatal(err)
			}
			accepted = record.Status == "ok"
		}
		if accepted {
			want := execute(t, "", oracle, path)
			for name, got := range map[string][]byte{"Node": onNode(t, main, path), "native": execute(t, "", binary, path), "emitted": onNode(t, script, path)} {
				if err := miscCompare(want, got, false); err != nil {
					t.Fatalf("%s %s: %v", item.id, name, err)
				}
			}
		} else {
			for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
				refusedBeforeDeadline(t, argv, "ESTree parser")
			}
		}
	})
}

func TestBoundedPortParserPlantedDisagreement(t *testing.T) {
	miscPlantedProof(t, testBoundedPortParserShards, boundedPortIDs(boundedPortCases(t)), func(planted bool) error {
		got := []byte("agree")
		if planted {
			got = []byte("disagree")
		}
		return miscCompare([]byte("agree"), got, false)
	})
}

func TestPortStallControl(t *testing.T) {
	main := mutantPort(t, "sourceStatements.ts", "if(this.parser.scanner.fullStart === start)", "if(false)")
	binary, _ := build(t, main, true)
	path := filepath.Join(t.TempDir(), "input.ts")
	os.WriteFile(path, []byte("class C { ) }"), 0644)
	for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}} {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
		output, _ := os.CreateTemp(t.TempDir(), "control-log")
		cmd.Stdout = output
		cmd.Stderr = output
		err := cmd.Run()
		output.Close()
		timedOut := ctx.Err() == context.DeadlineExceeded
		cancel()
		if !timedOut {
			t.Fatalf("stall control must hit deadline, got %v", err)
		}
		t.Logf("%s guard-disabled control caught by 500ms deadline", argv[0])
	}
}
