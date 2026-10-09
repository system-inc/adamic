package estree

import (
	"bytes"
	"context"
	"fmt"
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

const testBoundedPortParserShards = 64

type boundedPortCase struct {
	id, text, filename string
	audit              bool
}

func boundedPortCases(t *testing.T) []boundedPortCase {
	t.Helper()
	cases := []boundedPortCase{
		{"stage1/cohere/estree/stalls_test.go:0:eof-type", "type X = {", "input.ts", false},
		{"stage1/cohere/estree/stalls_test.go:0:eof-interface", "interface I {", "input.ts", false},
		{"stage1/cohere/estree/stalls_test.go:0:eof-method", "type X = { m(a: string): void;", "input.ts", false},
	}
	fixtures, err := filepath.Glob("validation/followup/stalls/*.input")
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("stall fixtures: %d %v", len(fixtures), err)
	}
	for _, fixture := range fixtures {
		body, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, boundedPortCase{"stage1/cohere/estree/" + filepath.ToSlash(fixture) + ":0:audit", string(body), strings.TrimSuffix(filepath.Base(fixture), ".input"), true})
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

// 64 fixed hash shards leave headroom for the live repository corpus. Keys are
// repository-relative fixture paths plus case index/mode; new files move no cases.
// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every live fixture and EOF case.
func TestBoundedPortParser(t *testing.T) {
	finishSetup := miscStart(t)
	cases := boundedPortCases(t)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := miscBuild(t, main)
	oracle := miscOracle(t)
	audited := make([]boundedPortCase, 0, len(cases))
	for _, item := range cases {
		if item.audit {
			audited = append(audited, item)
		}
	}
	answers := miscAnswers(t, oracle, audited)
	byID := make(map[string]miscAnswer, len(audited))
	for i, item := range audited {
		byID[item.id] = answers[i]
	}
	finishSetup()
	miscRunShards(t, testBoundedPortParserShards, boundedPortIDs(cases), func(t *testing.T, i int) {
		item := cases[i]
		path := filepath.Join(t.TempDir(), item.filename)
		if err := os.WriteFile(path, []byte(item.text), 0644); err != nil {
			t.Fatal(err)
		}
		accepted := item.audit && byID[item.id].Status == "ok"
		if accepted {
			want := byID[item.id].Data
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

const testPortStallControlShards = 1

func portStallControlResult(timedOut bool, err error) error {
	if !timedOut {
		return fmt.Errorf("stall control must hit deadline, got %v", err)
	}
	return nil
}

// ADAMIC_TEST_SHARD=i/n selects shards; unset runs the single guard-disabled
// mutant case. Both source Node and sanitized native stay together in its leaf.
func TestPortStallControl(t *testing.T) {
	finishSetup := miscStart(t)
	main := miscMutant(t, "sourceStatements.ts", "if(this.parser.scanner.fullStart === start)", "if(false)")
	binary, _ := miscBuild(t, main)
	finishSetup()
	miscRunShards(t, testPortStallControlShards, []string{"stage1/cohere/estree/stalls_test.go:0:guard-disabled"}, func(t *testing.T, _ int) {
		path := filepath.Join(t.TempDir(), "input.ts")
		if err := os.WriteFile(path, []byte("class C { ) }"), 0644); err != nil {
			t.Fatal(err)
		}
		for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}} {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
			output, err := os.CreateTemp(t.TempDir(), "control-log")
			if err != nil {
				cancel()
				t.Fatal(err)
			}
			cmd.Stdout, cmd.Stderr = output, output
			err = cmd.Run()
			output.Close()
			timedOut := ctx.Err() == context.DeadlineExceeded
			cancel()
			if err := portStallControlResult(timedOut, err); err != nil {
				t.Fatal(err)
			}
			t.Logf("%s guard-disabled control caught by 500ms deadline", argv[0])
		}
	})
}

func TestPortStallControlPlantedSurvivor(t *testing.T) {
	miscPlantedProof(t, testPortStallControlShards, []string{"stage1/cohere/estree/stalls_test.go:0:guard-disabled"}, func(planted bool) error {
		return portStallControlResult(!planted, nil)
	})
}
