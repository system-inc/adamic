package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func negativeFixture(t *testing.T) ([]byte, negativeWitness, negativeWitnessList) {
	t.Helper()
	source := []byte("let value: string = undefined!;\nconsole.log('42');\n")
	witness := negativeWitness{RulingTask: "#9wc5q5j", Declaration: "corpus/input.a:1", DeclaredType: "string", RepairFrom: "undefined!", RepairTo: `"ready"`, TypeCorrectSHA256: contentHash([]byte(strings.ReplaceAll(string(source), "undefined!", `"ready"`))), Exit: 70, Stderr: "adamic: panic: unset\n"}
	return source, witness, negativeWitnessList{Comment: "Adding an entry is a ruling routed to @system_adamic", Witnesses: map[string]negativeWitness{contentHash(source): witness}}
}
func writeNegativeList(t *testing.T, path string, list negativeWitnessList) {
	t.Helper()
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func TestNegativeContentHash(t *testing.T) {
	t.Parallel()
	source, witness, list := negativeFixture(t)
	if err := validateWitnessSource(source, witness); err != nil {
		t.Fatal(err)
	}
	// Names are documentary; a renamed byte-identical program retains the ruling.
	if _, listed := list.Witnesses[contentHash(source)]; !listed {
		t.Fatal("renamed witness lost")
	}
	for _, edit := range []string{string(source) + "\n", strings.ReplaceAll(string(source), "undefined!", `"ready"`)} {
		if _, listed := list.Witnesses[contentHash([]byte(edit))]; listed {
			t.Fatal("edit retained ruling")
		}
	}
}
func TestNegativeExactBackendStop(t *testing.T) {
	t.Parallel()
	_, witness, _ := negativeFixture(t)
	node := observation{Exit: 0, Stdout: "42\n"}
	stop := observation{Exit: 70, Stdout: "42\n", Stderr: witness.Stderr}
	if !acceptsNegativeWitness(witness, "root", node, stop, stop) {
		t.Fatal("ruled stop rejected")
	}
	for _, bad := range []observation{{Exit: 1, Stderr: witness.Stderr}, {Exit: 70, Stderr: "wrong\n"}, {Exit: 70, Stderr: witness.Stderr + "extra\n"}, {Exit: 70, Stderr: witness.Stderr, Error: "timeout"}, {Exit: 70, Stderr: witness.Stderr, Stdout: "other\n"}} {
		if acceptsNegativeWitness(witness, "root", node, bad, stop) || acceptsNegativeWitness(witness, "root", node, stop, bad) {
			t.Fatalf("wrong stop accepted: %+v", bad)
		}
	}
	if acceptsNegativeWitness(witness, "root", observation{Exit: 70}, stop, stop) {
		t.Fatal("Node failure listed")
	}
	witness.Stderr = "adamic: panic: {root}/input.a:1\n"
	stop.Stderr = "adamic: panic: /checkout/input.a:1\n"
	if !acceptsNegativeWitness(witness, "/checkout", node, stop, stop) {
		t.Fatal("checkout prefix expansion failed")
	}
}
func TestNegativeRulingMetadata(t *testing.T) {
	t.Parallel()
	source, witness, list := negativeFixture(t)
	path := filepath.Join(t.TempDir(), "negatives.json")
	for _, field := range []string{"task", "declaration", "type", "exit", "message", "repair", "hash", "header"} {
		bad := witness
		changed := list
		changed.Witnesses = map[string]negativeWitness{}
		switch field {
		case "task":
			bad.RulingTask = ""
		case "declaration":
			bad.Declaration = "input.a"
		case "type":
			bad.DeclaredType = "boolean"
		case "exit":
			bad.Exit = 1
		case "message":
			bad.Stderr = ""
		case "repair":
			bad.RepairTo = "null!"
		case "hash":
			bad.TypeCorrectSHA256 = "bad"
		case "header":
			changed.Comment = ""
		}
		changed.Witnesses[contentHash(source)] = bad
		writeNegativeList(t, path, changed)
		if _, err := readNegativeWitnesses(path); err == nil {
			t.Fatal("bad metadata accepted", field)
		}
	}
}
func negativeCommand(t *testing.T, corrected bool) (observation, report) {
	t.Helper()
	dir, _, base, head := commandFixture(t)
	source, witness, list := negativeFixture(t)
	if corrected {
		source = []byte(strings.ReplaceAll(string(source), witness.RepairFrom, witness.RepairTo))
	}
	for path, data := range map[string]string{
		"corpus/input.a":  string(source),
		"oracle/node.mjs": "import {readFileSync} from 'node:fs';const s=readFileSync(process.argv[2],'utf8');console.log('42');if(s.includes('STOP')){console.error('adamic: panic: unset');process.exit(70);}\n",
		"head":            "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) echo '// STOP';;\nbuild) printf '#!/bin/sh\\necho 42\\necho \"adamic: panic: unset\" >&2\\nexit 70\\n' > \"$4\"; chmod +x \"$4\";;\nesac\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(data), 0755); err != nil {
			t.Fatal(err)
		}
	}
	writeNegativeList(t, filepath.Join(dir, "cloud/admission-corpus/negative-witnesses.json"), list)
	if _, err := git(dir, "add", "corpus/input.a", "oracle/node.mjs", "cloud/admission-corpus/negative-witnesses.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := git(dir, "commit", "-qm", "negative witness input"); err != nil {
		t.Fatal(err)
	}
	sha, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--corpus", "corpus", "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr, o.Stdout)
	}
	return o, r
}
func TestNegativeCommandAccepted(t *testing.T) {
	t.Parallel()
	o, r := negativeCommand(t, false)
	if o.Exit != 0 || r.Verdict != "pass" || r.AcceptedWitnesses != 1 || r.NodeAgreements != 0 || !r.Programs[0].NegativeWitness || *r.Programs[0].Agree {
		t.Fatal("listed witness not distinguished from agreement", o, r)
	}
	if r.NegativeWitnessCount != 1 || !strings.Contains(o.Stderr, "negative witnesses: 1\n") {
		t.Fatal("list size not printed", o)
	}
}
func TestNegativeTypeCorrectMutant(t *testing.T) {
	t.Parallel()
	o, r := negativeCommand(t, true)
	if o.Exit == 0 || r.Verdict != "fail" || r.AcceptedWitnesses != 0 || r.Programs[0].NegativeWitness || *r.Programs[0].Agree {
		t.Fatal("type-correct mutant hidden by list", o, r)
	}
	t.Log("type-correct copy demands Node agreement; backend exit 70 is a divergence")
}
func TestNegativeValidProgramCannotBeListed(t *testing.T) {
	t.Parallel()
	_, witness, list := negativeFixture(t)
	source := []byte("let value: string = \"ready\";\nconsole.log('42');\n")
	for _, valid := range [][]byte{source, []byte("// let value: string = undefined!;\nconsole.log('42');\n")} {
		if err := validateWitnessSource(valid, witness); err == nil {
			t.Fatal("valid source accepted as literal type lie")
		}
	}
	list.Witnesses[witness.TypeCorrectSHA256] = witness
	path := filepath.Join(t.TempDir(), "negatives.json")
	writeNegativeList(t, path, list)
	if _, err := readNegativeWitnesses(path); err == nil {
		t.Fatal("type-correct hash listed")
	}
}
func TestNegativeEmptyListPrinted(t *testing.T) {
	t.Parallel()
	dir, sha, base, head := commandFixture(t)
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--corpus", "corpus", "--json")
	if !strings.Contains(o.Stderr, "negative witnesses: 0\n") {
		t.Fatal("empty list size not printed", o)
	}
}
