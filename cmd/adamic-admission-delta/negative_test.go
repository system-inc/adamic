package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
func negativeCommand(t *testing.T, corrected, alreadyAccepted bool) (observation, report) {
	t.Helper()
	dir, _, base, head := commandFixture(t)
	source, witness, list := negativeFixture(t)
	if corrected {
		source = []byte(strings.ReplaceAll(string(source), witness.RepairFrom, witness.RepairTo))
	}
	writeExecutable(t, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) echo '// STOP';;\nbuild) printf '#!/bin/sh\\necho 42\\necho \"adamic: panic: unset\" >&2\\nexit 70\\n' > \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
	for path, data := range map[string]string{
		"corpus/input.a":  string(source),
		"oracle/node.mjs": "import {readFileSync} from 'node:fs';const s=readFileSync(process.argv[2],'utf8');console.log('42');if(s.includes('STOP')){console.error('adamic: panic: unset');process.exit(70);}\n",
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
	if alreadyAccepted {
		base = head
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
	o, r := negativeCommand(t, false, false)
	if o.Exit != 0 || r.Verdict != "pass" || r.AcceptedWitnesses != 1 || r.NodeAgreements != 0 || !r.Programs[0].NegativeWitness || *r.Programs[0].Agree {
		t.Fatal("listed witness not distinguished from agreement", o, r)
	}
	if r.NegativeWitnessCount != 1 || !strings.Contains(o.Stderr, "negative witnesses: 1\n") {
		t.Fatal("list size not printed", o)
	}
}
func TestNegativeTypeCorrectMutant(t *testing.T) {
	t.Parallel()
	o, r := negativeCommand(t, true, false)
	if o.Exit == 0 || r.Verdict != "fail" || r.AcceptedWitnesses != 0 || r.Programs[0].NegativeWitness || *r.Programs[0].Agree {
		t.Fatal("type-correct mutant hidden by list", o, r)
	}
	t.Log("type-correct copy demands Node agreement; backend exit 70 is a divergence")
}
func TestNegativeValidProgramCannotBeListed(t *testing.T) {
	t.Parallel()
	_, witness, list := negativeFixture(t)
	source := []byte("let value: string = \"ready\";\nconsole.log('42');\n")
	for _, valid := range [][]byte{source, []byte("// let value: string = undefined!;\nconsole.log('42');\n"), []byte("let value: string = undefined!;\nimport './dependency.a';\n")} {
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

func TestNegativeAlreadyAcceptedRepairMutant(t *testing.T) {
	t.Parallel()
	o, r := negativeCommand(t, true, true)
	if o.Exit == 0 || r.Verdict != "fail" || r.Programs[0].Class != "accepted-by-both" || !r.Programs[0].Sampled || r.Programs[0].NegativeWitness || *r.Programs[0].Agree {
		t.Fatal("already-admitted type-correct mutant evaded Node agreement", o, r)
	}
	t.Log("type-correct input admitted by both compilers still demands Node agreement and rejects exit 70")
}
func TestNegativeEditedWitnessMandatory(t *testing.T) {
	t.Parallel()
	source, _, list := negativeFixture(t)
	changed := append(append([]byte{}, source...), []byte("\n")...)
	if !negativeRelated("corpus/input.a", changed, list) {
		t.Fatal("edit lost mandatory comparison")
	}
	if _, listed := list.Witnesses[contentHash(changed)]; listed {
		t.Fatal("edit retained exception")
	}
	p := entry{program: program{Path: "corpus/input.a"}, Corpus: "fixtures", Class: "accepted-by-both", Head: observation{Exit: 0}, NegativeRelated: true}
	selected, omitted := sample([]entry{p}, headSeed("sha"), 0.001, time.Second)
	if len(selected) != 1 || omitted != 0 {
		t.Fatal("edited witness sampled away", selected, omitted)
	}
}

// checkedBoundaryFixture is a checked view's stop on a value its cast let in, ruled on #995z2fa.
func checkedBoundaryFixture(t *testing.T) ([]byte, negativeWitness, negativeWitnessList) {
	t.Helper()
	source := []byte("interface View { count: number; }\nconst viewed = { count: 'abc' } as unknown as View;\nconsole.log('42');\n")
	witness := negativeWitness{Form: checkedBoundaryForm, RulingTask: "#995z2fa", Ruling: "Adamic, Oct 25", Declaration: "corpus/input.a:1", DeclaredType: "number", RepairFrom: "{ count: 'abc' }", RepairTo: "{ count: 5 }", TypeCorrectSHA256: contentHash([]byte(strings.ReplaceAll(string(source), "{ count: 'abc' }", "{ count: 5 }"))), Exit: 70, Stderr: "adamic: panic: field read failed: viewed.count is not a number; expected number, found string\n"}
	return source, witness, negativeWitnessList{Comment: "Adding an entry is a ruling routed to @system_adamic", Witnesses: map[string]negativeWitness{contentHash(source): witness}}
}

// Only the view check's own located message qualifies, naming the declared type, under a named ruling
// with a repair: a compiler-bug panic, another guard, a wrong exit or a message about another type is
// refused when the list is read. Mutant: dropping supportedCheckedBoundary's compiler bug test accepts
// the disguised one.
func TestNegativeCheckedBoundaryMetadata(t *testing.T) {
	t.Parallel()
	source, witness, list := checkedBoundaryFixture(t)
	path := filepath.Join(t.TempDir(), "negatives.json")
	writeNegativeList(t, path, list)
	if _, err := readNegativeWitnesses(path); err != nil {
		t.Fatal("ruled checked boundary rejected:", err)
	}
	if err := validateWitnessSource(source, witness); err != nil {
		t.Fatal("ruled checked boundary source rejected:", err)
	}
	for name, change := range map[string]func(*negativeWitness){
		"compiler bug": func(w *negativeWitness) {
			w.Stderr = "adamic: panic: compiler bug: incomplete checked field contract\n"
		},
		"disguised compiler bug": func(w *negativeWitness) {
			w.Stderr = "adamic: panic: field read failed: compiler bug: viewed.count expected number, found string\n"
		},
		"another guard": func(w *negativeWitness) { w.Stderr = "adamic: panic: index 3 is outside an array of length 1\n" },
		"another type":  func(w *negativeWitness) { w.DeclaredType = "boolean" },
		"two lines":     func(w *negativeWitness) { w.Stderr += w.Stderr },
		"wrong exit":    func(w *negativeWitness) { w.Exit = 1 },
		"no ruling":     func(w *negativeWitness) { w.Ruling = "" },
		"no repair":     func(w *negativeWitness) { w.RepairTo = w.RepairFrom },
		"unknown form":  func(w *negativeWitness) { w.Form = "other" },
	} {
		bad := witness
		change(&bad)
		writeNegativeList(t, path, negativeWitnessList{Comment: list.Comment, Witnesses: map[string]negativeWitness{contentHash(source): bad}})
		if _, err := readNegativeWitnesses(path); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

// The declaration must declare the ruled type, the repair must replace one place, and its hash must be
// the copy's.
func TestNegativeCheckedBoundarySource(t *testing.T) {
	t.Parallel()
	source, witness, _ := checkedBoundaryFixture(t)
	for name, change := range map[string]func(*negativeWitness){
		"undeclared":   func(w *negativeWitness) { w.Declaration = "corpus/input.a:3" },
		"two places":   func(w *negativeWitness) { w.RepairFrom, w.RepairTo = "count", "total" },
		"another hash": func(w *negativeWitness) { w.TypeCorrectSHA256 = contentHash(source) },
	} {
		bad := witness
		change(&bad)
		if err := validateWitnessSource(source, bad); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

// A run is the ruled stop only with its exit and its whole message on both backends.
func TestNegativeCheckedBoundaryExactStop(t *testing.T) {
	t.Parallel()
	_, witness, _ := checkedBoundaryFixture(t)
	node := observation{Exit: 0, Stdout: "42\n"}
	stop := observation{Exit: 70, Stderr: witness.Stderr}
	if !acceptsNegativeWitness(witness, "root", node, stop, stop) {
		t.Fatal("ruled stop rejected")
	}
	for _, bad := range []observation{{Exit: 1, Stderr: witness.Stderr}, {Exit: 70, Stderr: "adamic: panic: field read failed: viewed.count is not a number; expected number, found boolean\n"}, {Exit: 70, Stderr: "adamic: panic: compiler bug: unavailable checked union read\n"}} {
		if acceptsNegativeWitness(witness, "root", node, bad, stop) || acceptsNegativeWitness(witness, "root", node, stop, bad) {
			t.Errorf("wrong stop accepted: %+v", bad)
		}
	}
}

// checkedBoundaryCommand runs the command on the fixture with a compiler that stops whatever holds 'abc',
// or, when the repair misfits too, everything.
func checkedBoundaryCommand(t *testing.T, repairAgrees bool) (observation, report) {
	t.Helper()
	dir, _, base, head := commandFixture(t)
	source, _, list := checkedBoundaryFixture(t)
	stops := `grep -q "'abc'" "$2"`
	if !repairAgrees {
		stops = "true"
	}
	writeExecutable(t, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) if "+stops+"; then echo '// STOP'; else echo '// GO'; fi;;\nbuild) if "+stops+"; then printf '#!/bin/sh\\necho 42\\necho \"adamic: panic: field read failed: viewed.count is not a number; expected number, found string\" >&2\\nexit 70\\n' > \"$4.tmp\"; else printf '#!/bin/sh\\necho 42\\n' > \"$4.tmp\"; fi && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
	for path, data := range map[string]string{
		"corpus/input.a":  string(source),
		"oracle/node.mjs": "import {readFileSync} from 'node:fs';const s=readFileSync(process.argv[2],'utf8');console.log('42');if(s.includes('STOP')){console.error('adamic: panic: field read failed: viewed.count is not a number; expected number, found string');process.exit(70);}\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(data), 0755); err != nil {
			t.Fatal(err)
		}
	}
	writeNegativeList(t, filepath.Join(dir, "cloud/admission-corpus/negative-witnesses.json"), list)
	if _, err := git(dir, "add", "corpus/input.a", "oracle/node.mjs", "cloud/admission-corpus/negative-witnesses.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := git(dir, "commit", "-qm", "checked boundary witness"); err != nil {
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

// A ruled checked-boundary stop whose repair agrees with Node is a witness, not an agreement.
func TestNegativeCheckedBoundaryCommandAccepted(t *testing.T) {
	t.Parallel()
	o, r := checkedBoundaryCommand(t, true)
	if o.Exit != 0 || r.Verdict != "pass" || r.AcceptedWitnesses != 1 || !r.Programs[0].NegativeWitness || *r.Programs[0].Agree || r.Programs[0].RepairAgree == nil || !*r.Programs[0].RepairAgree {
		t.Fatal("ruled checked boundary not accepted", o, r)
	}
}

// The repair is the stop's mutant: one that stops too, where Node prints, fails the run. Mutant: not
// running the repair, or not requiring its agreement, accepts it.
func TestNegativeCheckedBoundaryRepairMustAgree(t *testing.T) {
	t.Parallel()
	o, r := checkedBoundaryCommand(t, false)
	if o.Exit == 0 || r.Verdict != "fail" || r.AcceptedWitnesses != 0 || r.Programs[0].NegativeWitness || r.Programs[0].RepairAgree == nil || *r.Programs[0].RepairAgree {
		t.Fatal("a repair that disagrees with Node was accepted", o, r)
	}
}
