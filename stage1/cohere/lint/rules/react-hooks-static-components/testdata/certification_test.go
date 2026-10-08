package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Overlay this owned test beside the shared harness; no shared source is edited.
func TestHIRStaticComponentsCertificate(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	rows := []string{}
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "react-hooks/static-components" {
			rows = append(rows, row)
		}
	}
	upstreamCount := len(rows)
	if upstreamCount == 0 {
		t.Fatal("no upstream static-components cases")
	}
	for _, d := range prepareRegistry(t, directory) {
		if d.Name == "react-hooks/static-components" {
			rows = append(rows, ownedWitnessRows(t, directory, d)...)
		}
	}
	for _, row := range rows {
		source := strings.Split(row, "\t")[0]
		config := filepath.Join(t.TempDir(), "tsconfig.json")
		if err := os.WriteFile(config, []byte(fmt.Sprintf(`{"compilerOptions":{"strict":true,"jsx":"preserve"},"files":[%q]}`, source)), 0600); err != nil {
			t.Fatal(err)
		}
		compare(t, oracle, binary, directory, manifest(t, []string{"program " + config, row}))
	}
	t.Logf("static-components: %d upstream cases and %d owned witnesses match Go on Node, emitted JS and native", upstreamCount, len(rows)-upstreamCount)
	bad := mutant(t, "dynamic.set(target,target);", "dynamic.delete(target);", "rules/react-hooks-static-components/rule.a")
	badBinary := buildMutantPort(t, bad)
	caught := false
	for _, row := range rows {
		source := strings.Split(row, "\t")[0]
		config := filepath.Join(t.TempDir(), "tsconfig.json")
		if err := os.WriteFile(config, []byte(fmt.Sprintf(`{"compilerOptions":{"strict":true,"jsx":"preserve"},"files":[%q]}`, source)), 0600); err != nil {
			t.Fatal(err)
		}
		path := manifest(t, []string{"program " + config, row})
		want := execute(t, "", oracle, "--manifest", path).output
		prefix := filepath.Join(t.TempDir(), "mutant-transcript")
		gotNative := execute(t, "", badBinary, "--manifest", path, "--record", prefix).output
		gotNode := execute(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), filepath.Join(bad, "main.ts"), "--manifest", path, "--replay", prefix).output
		if !bytes.Equal(gotNode, want) && !bytes.Equal(gotNative, want) {
			caught = true
			break
		}
	}
	if !caught {
		t.Fatal("static-components creation mutant survived")
	}
	t.Log("static-components creation mutant caught on native and Node")
}
