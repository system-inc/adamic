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

func TestRepositoryAgreement(t *testing.T) {
	directory := os.Getenv("ADAMIC_ESTREE_CORPUS")
	if directory == "" {
		// census: required-input ADAMIC_ESTREE_CORPUS: stage1/cohere/estree/testdata/corpus.mjs supplies a frozen Go/Node audit directory with port-records.jsonl and referenced source and answer files; setup does not currently export this corpus.
		t.Skip("set ADAMIC_ESTREE_CORPUS to a completed Go/Node corpus audit directory")
	}
	data, err := os.ReadFile(filepath.Join(directory, "port-records.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Path, Answer string
		PortStatus   string `json:"portStatus"`
		Bytes        int
	}
	var records []record
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var item record
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			t.Fatal(err)
		}
		if item.PortStatus == "identical" {
			records = append(records, item)
		}
	}
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := build(t, main, true)
	total := 0
	// Bound memory by batching whole files, with all accepted records accounted for.
	for start := 0; start < len(records); start += 100 {
		end := start + 100
		if end > len(records) {
			end = len(records)
		}
		var list strings.Builder
		var want bytes.Buffer
		for _, item := range records[start:end] {
			answer, err := os.ReadFile(item.Answer)
			if err != nil {
				t.Fatal(err)
			}
			if len(answer) != item.Bytes {
				t.Fatal("Go answer byte count changed")
			}
			want.Write(answer)
			list.WriteString(item.Path + "\n")
			total += len(answer)
		}
		manifest := filepath.Join(t.TempDir(), "manifest")
		if err := os.WriteFile(manifest, []byte(list.String()), 0644); err != nil {
			t.Fatal(err)
		}
		for name, got := range map[string][]byte{"source Node": onNode(t, main, "--manifest", manifest), "sanitized native": execute(t, "", binary, "--manifest", manifest), "emitted JS": onNode(t, script, "--manifest", manifest)} {
			if diff := firstDifference(want.Bytes(), got); diff != "" {
				t.Fatalf("records %d..%d %s: %s", start, end, name, diff)
			}
		}
	}
	t.Logf("%d Go-accepted/source-Node-identical repository snapshot files, %d bytes checked on uninstrumented source Node, sanitized native and emitted JS; refusals and mismatches remain excluded and separately reported", len(records), total)
}

func TestCorpusNativeRefusals(t *testing.T) {
	directory := os.Getenv("ADAMIC_ESTREE_CORPUS")
	if directory == "" {
		// census: required-input ADAMIC_ESTREE_CORPUS: stage1/cohere/estree/testdata/corpus.mjs supplies a frozen Go/Node audit directory with port-records.jsonl and referenced source and answer files; setup does not currently export this corpus.
		t.Skip("completed frozen corpus required")
	}
	data, err := os.ReadFile(filepath.Join(directory, "port-records.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	main, _ := filepath.Abs("main.ts")
	binary, _ := build(t, main, true)
	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var item struct{ Path, Status string }
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			t.Fatal(err)
		}
		if item.Status == "ok" {
			continue
		}
		stdout.Truncate(0)
		stdout.Seek(0, 0)
		stderr.Truncate(0)
		stderr.Seek(0, 0)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		command := exec.CommandContext(ctx, binary, item.Path)
		command.Stdout = stdout
		command.Stderr = stderr
		runErr := command.Run()
		timedOut := ctx.Err() != nil
		cancel()
		outInfo, _ := stdout.Stat()
		message, readErr := os.ReadFile(stderr.Name())
		if readErr != nil {
			t.Fatal(readErr)
		}
		if timedOut || runErr == nil || outInfo.Size() != 0 || !strings.Contains(string(message), "adamic: panic:") || strings.Contains(string(message), "compiler bug:") || strings.Contains(string(message), "AddressSanitizer") {
			t.Fatalf("%s: timeout=%v exit=%v stdout=%d stderr=%s", item.Path, timedOut, runErr, outInfo.Size(), message)
		}
		count++
	}
	t.Logf("all %d frozen Go refusals explicitly refused with empty stdout on sanitized native before the per-file 2s deadline", count)
}
