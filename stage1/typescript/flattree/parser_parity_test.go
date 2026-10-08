package flattree

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestParserParityFollowup(t *testing.T) {
	t.Parallel()
	binary := oracle(t)
	runner, _ := filepath.Abs(filepath.Join(repo, "oracle/node.mjs"))
	driver, _ := filepath.Abs("../parser/main.ts")
	for _, name := range []string{"shortest", "object", "expression", "array", "assignment"} {
		path, _ := filepath.Abs("testdata/parser-parity/" + name + ".ts.txt")
		goAnswer := run(t, "", binary, path, "--whole")
		nodeAnswer := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, path, "--whole", "--recovery")
		for _, side := range []struct {
			name   string
			answer []byte
		}{{"go", goAnswer}, {"node", nodeAnswer}} {
			expected, e := os.ReadFile("testdata/parser-parity/" + name + "." + side.name + ".txt")
			if e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(side.answer, expected) {
				t.Fatalf("%s %s parser boundary changed", name, side.name)
			}
		}
		wantParity := name != "shortest" && name != "object"
		if bytes.Equal(goAnswer, nodeAnswer) != wantParity {
			t.Fatalf("%s classification changed", name)
		}
		t.Logf("%s: Go clean; Node/Go tree parity=%t", name, wantParity)
	}
}

func TestParserSpeculationTrace(t *testing.T) {
	t.Parallel()
	runner, _ := filepath.Abs(filepath.Join(repo, "oracle/node.mjs"))
	trace, _ := filepath.Abs("testdata/parser-parity/trace.mjs")
	for _, name := range []string{"shortest", "object", "expression"} {
		path, _ := filepath.Abs("testdata/parser-parity/" + name + ".ts.txt")
		args := []string{"--disable-warning=ExperimentalWarning", runner, trace, path}
		if name == "expression" {
			args = append(args, "expression")
		}
		actual := run(t, "", "node", args...)
		expected, e := os.ReadFile("testdata/parser-parity/" + name + ".trace.json")
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(actual, expected) {
			t.Fatalf("%s trace changed", name)
		}
		var observed struct {
			Root  string `json:"root_kind"`
			Trace []struct {
				Method string            `json:"method"`
				Args   []json.RawMessage `json:"args"`
				Before struct {
					Token    string   `json:"token"`
					Contexts []string `json:"contexts"`
				} `json:"before"`
				Result json.RawMessage `json:"result"`
			} `json:"trace"`
		}
		if e := json.Unmarshal(actual, &observed); e != nil {
			t.Fatal(e)
		}
		failedClose, acceptedArrow := false, false
		for _, row := range observed.Trace {
			if row.Method == "expect" && len(row.Args) == 1 && string(row.Args[0]) == `"CloseParenToken"` && string(row.Result) == "false" {
				failedClose = true
			}
			if row.Method == "arrowCandidate" && row.Before.Token == "OpenParenToken" && string(row.Result) == "true" && len(row.Before.Contexts) == 2 && row.Before.Contexts[0] == "source" && row.Before.Contexts[1] == "variables" {
				acceptedArrow = true
			}
		}
		if name != "expression" && (!failedClose || !acceptedArrow) {
			t.Fatal("failed close-paren / accepted-arrow evidence lost")
		}
		if name == "expression" && (acceptedArrow || observed.Root != "ParenthesizedExpression") {
			t.Fatal("outside-list entry control changed")
		}
	}
}

func TestShortestParserDeletionControls(t *testing.T) {
	t.Parallel()
	data, e := os.ReadFile("testdata/parser-parity/single-deletions.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Index    int    `json:"deleted_index"`
		Source   string `json:"source"`
		GoClean  bool   `json:"go_clean"`
		NodeExit int    `json:"node_exit"`
		Parity   *bool  `json:"parity"`
	}
	if e := json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	source, e := os.ReadFile("testdata/parser-parity/shortest.ts.txt")
	if e != nil {
		t.Fatal(e)
	}
	if len(rows) != 14 || len(source) != 14 {
		t.Fatal("minimality inventory changed")
	}
	binary := oracle(t)
	runner, _ := filepath.Abs(filepath.Join(repo, "oracle/node.mjs"))
	driver, _ := filepath.Abs("../parser/main.ts")
	for _, row := range rows {
		if row.Index < 0 || row.Index >= len(source) || row.Source != string(source[:row.Index])+string(source[row.Index+1:]) {
			t.Fatal("not a single-character deletion")
		}
		path := filepath.Join(t.TempDir(), "deletion.ts")
		if e := os.WriteFile(path, []byte(row.Source), 0644); e != nil {
			t.Fatal(e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		command := exec.CommandContext(ctx, binary, path, "--whole")
		var stderr bytes.Buffer
		command.Stderr = &stderr
		goAnswer, e := command.Output()
		cancel()
		if (e == nil) != row.GoClean {
			t.Fatalf("Go deletion %d clean classification changed: %v %s", row.Index, e, &stderr)
		}
		if e != nil && !bytes.Contains(stderr.Bytes(), []byte("parser diagnostic")) {
			t.Fatalf("unexpected Go failure: %v %s", e, &stderr)
		}
		nodeAnswer := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, path, "--whole", "--recovery")
		if row.NodeExit != 0 {
			t.Fatal("nonzero recorded Node exit")
		}
		if row.GoClean && (row.Parity == nil || !*row.Parity || !bytes.Equal(goAnswer, nodeAnswer)) {
			t.Fatalf("shorter valid discrepancy found at deletion %d", row.Index)
		}
	}
}
