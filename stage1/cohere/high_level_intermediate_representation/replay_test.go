package high_level_intermediate_representation

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCheckpointReplayOracle(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	destination := exportConstructionCensus(t, os.Getenv("HIR_REPLAY_CENSUS_EXPORT"))
	manifest := filepath.Join(destination, "checkpoint-manifest.tsv")
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(lane, "replay/main.ts")
	node := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", entry, "--census", manifest)
	binary := filepath.Join(t.TempDir(), "replay")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", entry, "-o", binary)
	native := command(t, root, nil, binary, "--census", manifest)
	if !bytes.Equal(native, node) {
		t.Fatal("native/Node checkpoint output differs")
	}
	cases := map[string]string{}
	for _, part := range strings.Split(string(native), "checkpoint\t")[1:] {
		key, body, ok := strings.Cut(part, "\n")
		if !ok {
			t.Fatal("missing case body")
		}
		if _, exists := cases[key]; exists {
			t.Fatal("duplicate case")
		}
		cases[key] = body
	}
	originals, probes := 0, 0
	var mutant string
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		f := strings.Split(line, "\t")
		want, err := os.ReadFile(f[1])
		if err != nil {
			t.Fatal(err)
		}
		got, exists := cases[f[0]]
		if !exists {
			t.Fatal("missing case " + f[0])
		}
		delete(cases, f[0])
		if got != string(want) {
			t.Fatalf("checkpoint %s differs: %s", f[0], firstDifference([]byte(got), want))
		}
		count, err := strconv.Atoi(f[2])
		if err != nil {
			t.Fatal(err)
		}
		if f[3] == "true" {
			probes += count
		} else {
			originals += count
		}
		if mutant == "" {
			rows := strings.Split(string(want), "\n")
			max := -1
			at := -1
			for i, row := range rows {
				parts := strings.Split(row, "\t")
				if len(parts) == 7 && parts[0] == "sidecar" && parts[1] == "replay-fixture" && parts[2] == "$" && parts[3] == "instruction" {
					id, _ := strconv.Atoi(parts[4])
					if id > max {
						max = id
						at = i
					}
				}
			}
			if at >= 0 {
				parts := strings.Split(rows[at], "\t")
				parts[4] = strconv.Itoa(max + 1)
				rows[at] = strings.Join(parts, "\t")
				mutant = strings.Join(rows, "\n")
			}
		}
	}
	if len(cases) != 0 || originals != 1465 || probes != 72 {
		t.Fatalf("census changed: originals=%d probes=%d extras=%d", originals, probes, len(cases))
	}
	t.Logf("Go checkpoint framing and fresh HIR replay: native = Node %d/%d originals including all 23 Flow graphs; %d probes", originals, originals, probes)
	if mutant == "" {
		t.Fatal("no instruction anchor for index mutant")
	}
	bad := filepath.Join(t.TempDir(), "off-by-one.checkpoint")
	if err := os.WriteFile(bad, []byte(mutant), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), entry, "--checkpoint", bad}, {binary, "--checkpoint", bad}} {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = root
		output, err := c.CombinedOutput()
		if err == nil || !strings.Contains(string(output), "unknown instruction $:") {
			t.Fatalf("corrupt-index mutant survived: %v %s", err, output)
		}
	}
	t.Log("off-by-one instruction anchor mutant caught by identity reader on native and Node")
	variants := filepath.Join(lane, "replay/variants.ts")
	want, err := os.ReadFile(filepath.Join(destination, "central-instructions.dump"))
	if err != nil {
		t.Fatal(err)
	}
	vNode := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", variants)
	vBinary := filepath.Join(t.TempDir(), "variants")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", variants, "-o", vBinary)
	vNative := command(t, root, nil, vBinary)
	if !bytes.Equal(vNode, want) || !bytes.Equal(vNative, want) {
		t.Fatalf("central variants differ: Go %s Node %s native %s", want, vNode, vNative)
	}
	t.Log("DeclareContext, absent/empty/local/global StartMemoize and FinishMemoize variants match Go; clone owns memo arrays")
	contracts := filepath.Join(lane, "replay/contracts.ts")
	cNode := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", contracts)
	cBinary := filepath.Join(t.TempDir(), "contracts")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", contracts, "-o", cBinary)
	cNative := command(t, root, nil, cBinary)
	if string(cNode) != "declaration/copy/rewrite/cache contracts pass\n" || !bytes.Equal(cNode, cNative) {
		t.Fatalf("shared API contracts differ: Node %s native %s", cNode, cNative)
	}

}
