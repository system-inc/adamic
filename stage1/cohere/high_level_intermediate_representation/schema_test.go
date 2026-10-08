package high_level_intermediate_representation

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSharedTerminalSchemas(t *testing.T) {
	t.Parallel()
	checkSharedTerminalSchemas(t, true)
}
func TestNodeSharedTerminalSchemas(t *testing.T) {
	t.Parallel()
	checkSharedTerminalSchemas(t, false)
}
func checkSharedTerminalSchemas(t *testing.T, nativeRequired bool) {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	out := t.TempDir()
	upstream := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	overlay := map[string]string{}
	for name, source := range map[string]string{"printer": "testdata/oracle_test.go", "frame": "replay/oracle_test.go", "inputs": "replay/inputs_test.go", "schemas": "testdata/schema_oracle_test.go"} {
		overlay[filepath.Join(upstream, "stage1_shared_"+name+"_test.go")] = filepath.Join(lane, source)
	}
	data, _ := json.Marshal(map[string]any{"Replace": overlay})
	path := filepath.Join(out, "overlay.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	command(t, filepath.Join(root, "cohere"), []string{"GOWORK=" + filepath.Join(root, "cohere/go.work"), "HIR_SHARED_SCHEMA_OUT=" + out}, "go", "test", "-count=1", "-timeout=3h", "-tags=lintoracle", "-overlay", path, "-run=^TestStage1SharedTerminalSchemas$", "./internal/lint/ecmascript/high_level_intermediate_representation")
	entry := filepath.Join(lane, "replay/schema_main.ts")
	binary := filepath.Join(out, "schema")
	if nativeRequired {
		command(t, root, nil, "go", "run", "./cmd/adamic", "build", entry, "-o", binary, "--sanitize")
	}
	graphText := func(data []byte) []byte {
		lines := strings.Split(string(data), "\n")
		n, err := strconv.Atoi(strings.TrimPrefix(lines[2], "graph-lines "))
		if err != nil {
			t.Fatal(err)
		}
		return []byte(strings.Join(lines[3:3+n], "\n") + "\n")
	}
	for _, kind := range []string{"scope", "sequence", "catch", "escape", "phi"} {
		before := filepath.Join(out, kind+"-before.checkpoint")
		original, err := os.ReadFile(before)
		if err != nil {
			t.Fatal(err)
		}
		after, err := os.ReadFile(filepath.Join(out, kind+"-after.checkpoint"))
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"--roundtrip", "--clone", "--finalize"} {
			want := original
			if mode == "--clone" {
				want = graphText(original)
			} else if mode == "--finalize" {
				want = graphText(after)
			}
			node := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", entry, mode, before)
			native := node
			if nativeRequired {
				native = command(t, root, nil, binary, mode, before)
			}
			if !bytes.Equal(node, want) || !bytes.Equal(native, want) {
				t.Fatalf("%s %s differs from Go: Node %s; native %s", kind, mode, firstDifference(node, want), firstDifference(native, want))
			}
		}
	}
	if !nativeRequired {
		for _, change := range []struct{ name, from, to, kind, mode string }{
			{"scope-index", "scope: fn.scopeAt(json.numberField(root,'Scope'))", "scope: fn.scopeAt(json.numberField(root,'Scope') + 1)", "scope", "--roundtrip"},
			{"sequence-body", "kind: 'Sequence',block: block('Block')", "kind: 'Sequence',block: block('Fallthrough')", "sequence", "--roundtrip"},
			{"exception-continuation", "continuation: block('Continuation')", "continuation: block('Handler')", "catch", "--roundtrip"},
		} {
			dir, err := os.MkdirTemp(filepath.Dir(lane), "hir-schema-mutant-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			for _, folder := range []string{"", "replay"} {
				files, err := filepath.Glob(filepath.Join(lane, folder, "*.ts"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(dir, folder), 0755); err != nil {
					t.Fatal(err)
				}
				for _, file := range files {
					data, err := os.ReadFile(file)
					if err != nil {
						t.Fatal(err)
					}
					if filepath.Base(file) == "decode.ts" {
						if strings.Count(string(data), change.from) != 1 {
							t.Fatal("mutant anchor moved: " + change.name)
						}
						data = []byte(strings.Replace(string(data), change.from, change.to, 1))
					}
					if err := os.WriteFile(filepath.Join(dir, folder, filepath.Base(file)), data, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			input := filepath.Join(out, change.kind+"-before.checkpoint")
			want, err := os.ReadFile(input)
			if err != nil {
				t.Fatal(err)
			}
			command := exec.Command("node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "replay/schema_main.ts"), change.mode, input)
			command.Dir = root
			got, err := command.CombinedOutput()
			if err == nil && bytes.Equal(got, want) {
				t.Fatal("mutant survived: " + change.name)
			}
			if err != nil && !strings.Contains(string(got), "unknown Go scope id") {
				t.Fatalf("mutant failed for an unrelated reason: %s: %v %s", change.name, err, got)
			}
			t.Log("Node Go-byte comparison catches " + change.name)
		}
	}
	t.Logf("15/15 Go terminal/phi roundtrip, clone and finalization comparisons pass; native executed=%t", nativeRequired)
}
