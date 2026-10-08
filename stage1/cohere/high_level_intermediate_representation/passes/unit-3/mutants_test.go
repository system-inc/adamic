package unit3

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type semanticMutant struct{ pass, file, before, after string }

func TestNodeSemanticMutants(t *testing.T) {
	t.Parallel()
	root := unit3Root(t)
	records := loadFixtures(t)
	own := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-3")
	mutants := []semanticMutant{
		{"outline_functions", "outline_functions.ts", "? '_temp' :", "? '_wrong' :"},
		{"drop_manual_memoization", "drop_manual_memoization.ts", "kind === 'useMemo' ? instruction.lvalue : first.place", "first.place"},
		{"inline_iife_including_memo_callbacks", "inline_iife.ts", "functions.set(instruction.lvalue.identifier,iid);", "functions.delete(instruction.lvalue.identifier);"},
		{"inline_remap", "inline_remap.ts", "identifiers.set(context.identifier,capture.identifier);", "identifiers.delete(context.identifier);"},
		{"invoked_functions", "invoked_functions.ts", "if(!grew) { return invoked; }", "if(!grew) { return new Map<number,FunctionIndex>(); }"},
		{"dead_code_elimination", "dead_code_elimination.ts", "const yes = liveIdentifier(fn,used,names,fn.instruction(id).lvalue.identifier);", "const yes = true;"},
		{"merge_consecutive_blocks", "merge_consecutive_blocks.ts", "block.predecessors.length !== 1", "block.predecessors.length !== 0"},
	}
	native := os.Getenv("HIR_UNIT3_NATIVE")
	buildSlots := make(chan struct{}, 1)
	for _, mutant := range mutants {
		if os.Getenv("HIR_UNIT3_PASS_GROUP") == "available" && mutant.pass == "drop_manual_memoization" {
			continue
		}
		if selected := os.Getenv("HIR_UNIT3_MUTANT_PASS"); selected != "" && mutant.pass != selected {
			continue
		}
		if native != "" && os.Getenv("HIR_UNIT3_PASS_GROUP") == "independent" {
			switch mutant.pass {
			case "outline_functions", "merge_consecutive_blocks", "inline_remap", "invoked_functions", "dead_code_elimination":
			default:
				continue
			}
		}
		t.Run(mutant.pass, func(t *testing.T) {
			t.Parallel()
			temporary := t.TempDir()
			cohere := filepath.Join(temporary, "stage1/cohere")
			hir := filepath.Join(cohere, "high_level_intermediate_representation")
			lane := filepath.Join(hir, "passes/unit-3")
			if err := os.MkdirAll(lane, 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"arena", "static_single_assignment", "typeaware"} {
				if err := os.Symlink(filepath.Join(root, "stage1/cohere", name), filepath.Join(cohere, name)); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Symlink(filepath.Join(root, "stage1/typescript"), filepath.Join(temporary, "stage1/typescript")); err != nil {
				t.Fatal(err)
			}
			shared, err := os.ReadDir(filepath.Join(root, "stage1/cohere/high_level_intermediate_representation"))
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range shared {
				if entry.Name() == "passes" {
					continue
				}
				if err := os.Symlink(filepath.Join(root, "stage1/cohere/high_level_intermediate_representation", entry.Name()), filepath.Join(hir, entry.Name())); err != nil {
					t.Fatal(err)
				}
			}
			files, err := filepath.Glob(filepath.Join(own, "*.ts"))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if filepath.Base(file) == mutant.file {
					if strings.Count(string(data), mutant.before) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), mutant.before, mutant.after, 1))
				}
				if err := os.WriteFile(filepath.Join(lane, filepath.Base(file)), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			var mutantBinary string
			if native != "" {
				mutantBinary = filepath.Join(temporary, "mutant")
				entry := "main.ts"
				if os.Getenv("HIR_UNIT3_PASS_GROUP") != "" {
					entry = "independent_main.ts"
				}
				buildSlots <- struct{}{}
				func() {
					defer func() { <-buildSlots }()
					unit3Command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(lane, entry), "-o", mutantBinary, "--sanitize")
				}()
			}
			attempts := 0
			for _, record := range records {
				if strings.Split(record.Pass, ":")[0] != mutant.pass {
					continue
				}
				if record.Probe {
					continue
				}
				// Select a real effect witness; all inputs still run in the census test.
				if !strings.Contains(record.After, "sidecar\tunit3.result\t$\tfunction\t-\tresult\t") {
					t.Fatal("missing result")
				}
				marker := "sidecar\tunit3.result\t$\tfunction\t-\tresult\t"
				result := strings.Split(strings.Split(record.After, marker)[1], "\n")[0]
				if result == "" || result == "false" || (mutant.pass != "invoked_functions" && (result == "0" || result == "0,0")) {
					continue
				}
				if mutant.pass == "drop_manual_memoization" && (strings.Split(result, ",")[1] == "0" || !strings.Contains(record.After, "FinishMemoize")) {
					continue
				}
				if mutant.pass == "dead_code_elimination" && strings.Split(result, ",")[0] == "0" {
					continue
				}
				if mutant.pass == "drop_manual_memoization" && !strings.Contains(record.Before, "string:useMemo") && !strings.Contains(record.Before, "\"Name\":\"useMemo\"") {
					continue
				}
				if mutant.pass == "inline_remap" && strings.Contains(record.Before, "\tkey\tcaptures\t") {
					continue
				}
				if mutant.pass == "inline_remap" && strings.Contains(record.Before, "function\t-\tcaptures\t\n") {
					continue
				}
				attempts++
				input := filepath.Join(temporary, "before.checkpoint")
				if err := os.WriteFile(input, []byte(record.Before), 0600); err != nil {
					t.Fatal(err)
				}
				var baseline []byte
				if native != "" {
					baseline = unit3Command(t, root, nil, native, "--checkpoint", input)
				} else {
					baseline = unit3Command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(own, "main.ts"), "--checkpoint", input)
				}
				if !bytes.Equal(baseline, []byte(record.After)) {
					if attempts >= 32 {
						t.Fatal("no exact-byte baseline witness")
					}
					continue
				}
				c := exec.Command("node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), filepath.Join(lane, "main.ts"), "--checkpoint", input)
				if native != "" {
					c = exec.Command(mutantBinary, "--checkpoint", input)
				}
				c.Dir = root
				output, err := c.CombinedOutput()
				if err != nil {
					t.Fatalf("mutant must change an answer, not crash: %v\n%s", err, output)
				}
				if bytes.Equal(output, baseline) {
					if attempts >= 32 {
						t.Fatal("semantic mutant survived")
					}
					continue
				}
				backend := "Node"
				if native != "" {
					backend = "sanitized native"
				}
				t.Logf("semantic mutant caught on %s: %s/%s", backend, record.Key, record.Pass)
				return
			}
			t.Fatal("no semantic witness caught the mutant")
		})
	}
}
