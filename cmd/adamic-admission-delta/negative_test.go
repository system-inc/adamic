package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func negativeFixture(test *testing.T) ([]byte, negativeWitness, negativeWitnessList) {
	test.Helper()
	source := []byte("let value: string = undefined!;\nconsole.log('42');\n")
	witness := negativeWitness{
		RulingTask:        "#9wc5q5j",
		Declaration:       "corpus/input.a:1",
		DeclaredType:      "string",
		RepairFrom:        "undefined!",
		RepairTo:          `"ready"`,
		TypeCorrectSha256: contentHash([]byte(strings.ReplaceAll(string(source), "undefined!", `"ready"`))),
		Exit:              70,
		Stderr:            "adamic: panic: unset\n",
	}
	return source, witness, negativeWitnessList{
		Comment:   "Adding an entry is a ruling routed to @system_adamic",
		Witnesses: map[string]negativeWitness{contentHash(source): witness},
	}
}
func writeNegativeList(test *testing.T, path string, list negativeWitnessList) {
	test.Helper()
	data, encodeError := json.Marshal(list)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(path, data, 0600); writeError != nil {
		test.Fatal(writeError)
	}
}
func TestNegativeContentHash(test *testing.T) {
	test.Parallel()
	source, witness, list := negativeFixture(test)
	if validateWitnessSourceError := validateWitnessSource(source, witness); validateWitnessSourceError != nil {
		test.Fatal(validateWitnessSourceError)
	}
	// Names are documentary; a renamed byte-identical program retains the ruling.
	if _, listed := list.Witnesses[contentHash(source)]; !listed {
		test.Fatal("renamed witness lost")
	}
	for _, edit := range []string{string(source) + "\n", strings.ReplaceAll(string(source), "undefined!", `"ready"`)} {
		if _, listed := list.Witnesses[contentHash([]byte(edit))]; listed {
			test.Fatal("edit retained ruling")
		}
	}
}
func TestNegativeExactBackendStop(test *testing.T) {
	test.Parallel()
	_, witness, _ := negativeFixture(test)
	node := observation{Exit: 0, Stdout: "42\n"}
	stop := observation{Exit: 70, Stdout: "42\n", Stderr: witness.Stderr}
	if !acceptsNegativeWitness(witness, "root", node, stop, stop) {
		test.Fatal("ruled stop rejected")
	}
	for _, bad := range []observation{
		{Exit: 1, Stderr: witness.Stderr},
		{Exit: 70, Stderr: "wrong\n"},
		{Exit: 70, Stderr: witness.Stderr + "extra\n"},
		{Exit: 70, Stderr: witness.Stderr, Error: "timeout"},
		{Exit: 70, Stderr: witness.Stderr, Stdout: "other\n"},
	} {
		if acceptsNegativeWitness(witness, "root", node, bad, stop) ||
			acceptsNegativeWitness(witness, "root", node, stop, bad) {
			test.Fatalf("wrong stop accepted: %+v", bad)
		}
	}
	if acceptsNegativeWitness(witness, "root", observation{Exit: 70}, stop, stop) {
		test.Fatal("Node failure listed")
	}
	witness.Stderr = "adamic: panic: {root}/input.a:1\n"
	stop.Stderr = "adamic: panic: /checkout/input.a:1\n"
	if !acceptsNegativeWitness(witness, "/checkout", node, stop, stop) {
		test.Fatal("checkout prefix expansion failed")
	}
}
func TestNegativeRulingMetadata(test *testing.T) {
	test.Parallel()
	source, witness, list := negativeFixture(test)
	path := filepath.Join(test.TempDir(), "negatives.json")
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
			bad.TypeCorrectSha256 = "bad"
		case "header":
			changed.Comment = ""
		}
		changed.Witnesses[contentHash(source)] = bad
		writeNegativeList(test, path, changed)
		if _, readNegativeWitnessesError := readNegativeWitnesses(path); readNegativeWitnessesError == nil {
			test.Fatal("bad metadata accepted", field)
		}
	}
}
func negativeCommand(test *testing.T, corrected, alreadyAccepted bool) (observation, report) {
	test.Helper()
	dir, _, base, head := commandFixture(test)
	source, witness, list := negativeFixture(test)
	if corrected {
		source = []byte(strings.ReplaceAll(string(source), witness.RepairFrom, witness.RepairTo))
	}
	writeExecutable(
		test,
		head,
		"#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) echo '// STOP';;\nbuild) printf '#!/bin/sh\\necho 42\\necho \"adamic: panic: unset\" >&2\\nexit 70\\n' > \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n",
	)
	for path, data := range map[string]string{
		"corpus/input.a":  string(source),
		"oracle/node.mjs": "import {readFileSync} from 'node:fs';const s=readFileSync(process.argv[2],'utf8');console.log('42');if(s.includes('STOP')){console.error('adamic: panic: unset');process.exit(70);}\n",
	} {
		if writeError := os.WriteFile(filepath.Join(dir, path), []byte(data), 0755); writeError != nil {
			test.Fatal(writeError)
		}
	}
	writeNegativeList(test, filepath.Join(dir, "cloud/admission-corpus/negative-witnesses.json"), list)
	if _, gitError := git(
		dir,
		"add",
		"corpus/input.a",
		"oracle/node.mjs",
		"cloud/admission-corpus/negative-witnesses.json",
	); gitError != nil {
		test.Fatal(gitError)
	}
	if _, gitError := git(dir, "commit", "-qm", "negative witness input"); gitError != nil {
		test.Fatal(gitError)
	}
	sha, gitError := git(dir, "rev-parse", "HEAD")
	if gitError != nil {
		test.Fatal(gitError)
	}
	if alreadyAccepted {
		base = head
	}
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--corpus",
		"corpus",
		"--json",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr, observation.Stdout)
	}
	return observation, report
}
func TestNegativeCommandAccepted(test *testing.T) {
	test.Parallel()
	observation, report := negativeCommand(test, false, false)
	if observation.Exit != 0 ||
		report.Verdict != "pass" ||
		report.AcceptedWitnesses != 1 ||
		report.NodeAgreements != 0 ||
		!report.Programs[0].NegativeWitness ||
		*report.Programs[0].Agree {
		test.Fatal("listed witness not distinguished from agreement", observation, report)
	}
	if report.NegativeWitnessCount != 1 || !strings.Contains(observation.Stderr, "negative witnesses: 1\n") {
		test.Fatal("list size not printed", observation)
	}
}
func TestNegativeTypeCorrectMutant(test *testing.T) {
	test.Parallel()
	observation, report := negativeCommand(test, true, false)
	if observation.Exit == 0 ||
		report.Verdict != "fail" ||
		report.AcceptedWitnesses != 0 ||
		report.Programs[0].NegativeWitness ||
		*report.Programs[0].Agree {
		test.Fatal("type-correct mutant hidden by list", observation, report)
	}
	test.Log("type-correct copy demands Node agreement; backend exit 70 is a divergence")
}
func TestNegativeValidProgramCannotBeListed(test *testing.T) {
	test.Parallel()
	_, witness, list := negativeFixture(test)
	source := []byte("let value: string = \"ready\";\nconsole.log('42');\n")
	for _, valid := range [][]byte{
		source,
		[]byte("// let value: string = undefined!;\nconsole.log('42');\n"),
		[]byte("let value: string = undefined!;\nimport './dependency.a';\n"),
	} {
		if validateWitnessSourceError := validateWitnessSource(valid, witness); validateWitnessSourceError == nil {
			test.Fatal("valid source accepted as literal type lie")
		}
	}
	list.Witnesses[witness.TypeCorrectSha256] = witness
	path := filepath.Join(test.TempDir(), "negatives.json")
	writeNegativeList(test, path, list)
	if _, readNegativeWitnessesError := readNegativeWitnesses(path); readNegativeWitnessesError == nil {
		test.Fatal("type-correct hash listed")
	}
}
func TestNegativeEmptyListPrinted(test *testing.T) {
	test.Parallel()
	dir, sha, base, head := commandFixture(test)
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--corpus",
		"corpus",
		"--json",
	)
	if !strings.Contains(observation.Stderr, "negative witnesses: 0\n") {
		test.Fatal("empty list size not printed", observation)
	}
}

func TestNegativeAlreadyAcceptedRepairMutant(test *testing.T) {
	test.Parallel()
	observation, report := negativeCommand(test, true, true)
	if observation.Exit == 0 ||
		report.Verdict != "fail" ||
		report.Programs[0].Class != "accepted-by-both" ||
		!report.Programs[0].Sampled ||
		report.Programs[0].NegativeWitness ||
		*report.Programs[0].Agree {
		test.Fatal("already-admitted type-correct mutant evaded Node agreement", observation, report)
	}
	test.Log("type-correct input admitted by both compilers still demands Node agreement and rejects exit 70")
}
func TestNegativeEditedWitnessMandatory(test *testing.T) {
	test.Parallel()
	source, _, list := negativeFixture(test)
	changed := append(append([]byte{}, source...), []byte("\n")...)
	if !negativeRelated("corpus/input.a", changed, list) {
		test.Fatal("edit lost mandatory comparison")
	}
	if _, listed := list.Witnesses[contentHash(changed)]; listed {
		test.Fatal("edit retained exception")
	}
	p := entry{
		program:         program{Path: "corpus/input.a"},
		Corpus:          "fixtures",
		Class:           "accepted-by-both",
		Head:            observation{Exit: 0},
		NegativeRelated: true,
	}
	selected, omitted := sample([]entry{p}, headSeed("sha"), 0.001, time.Second)
	if len(selected) != 1 || omitted != 0 {
		test.Fatal("edited witness sampled away", selected, omitted)
	}
}

// checkedBoundaryFixture is a checked view's stop on a value its cast let in, ruled on #995z2fa.
func checkedBoundaryFixture(test *testing.T) ([]byte, negativeWitness, negativeWitnessList) {
	test.Helper()
	source := []byte(
		"interface View { count: number; }\nconst viewed = { count: 'abc' } as unknown as View;\nconsole.log('42');\n",
	)
	witness := negativeWitness{
		Form:              checkedBoundaryForm,
		RulingTask:        "#995z2fa",
		Ruling:            "Adamic, Oct 25",
		Declaration:       "corpus/input.a:1",
		DeclaredType:      "number",
		RepairFrom:        "{ count: 'abc' }",
		RepairTo:          "{ count: 5 }",
		TypeCorrectSha256: contentHash([]byte(strings.ReplaceAll(string(source), "{ count: 'abc' }", "{ count: 5 }"))),
		Exit:              70,
		Stderr:            "adamic: panic: field read failed: viewed.count is not a number; expected number, found string\n",
	}
	return source, witness, negativeWitnessList{
		Comment:   "Adding an entry is a ruling routed to @system_adamic",
		Witnesses: map[string]negativeWitness{contentHash(source): witness},
	}
}

// Only the view check's own located message qualifies, naming the declared type, under a named ruling
// with a repair: a compiler-bug panic, another guard, a wrong exit or a message about another type is
// refused when the list is read. Mutant: dropping supportedCheckedBoundary's compiler bug test accepts
// the disguised one.
func TestNegativeCheckedBoundaryMetadata(test *testing.T) {
	test.Parallel()
	source, witness, list := checkedBoundaryFixture(test)
	path := filepath.Join(test.TempDir(), "negatives.json")
	writeNegativeList(test, path, list)
	if _, readNegativeWitnessesError := readNegativeWitnesses(path); readNegativeWitnessesError != nil {
		test.Fatal("ruled checked boundary rejected:", readNegativeWitnessesError)
	}
	if validateWitnessSourceError := validateWitnessSource(source, witness); validateWitnessSourceError != nil {
		test.Fatal("ruled checked boundary source rejected:", validateWitnessSourceError)
	}
	for name, change := range map[string]func(*negativeWitness){
		"compiler bug": func(witness *negativeWitness) {
			witness.Stderr = "adamic: panic: compiler bug: incomplete checked field contract\n"
		},
		"disguised compiler bug": func(witness *negativeWitness) {
			witness.Stderr = "adamic: panic: field read failed: compiler bug: viewed.count expected number, found string\n"
		},
		"another guard": func(
			witness *negativeWitness,
		) {
			witness.Stderr = "adamic: panic: index 3 is outside an array of length 1\n"
		},
		"another type": func(witness *negativeWitness) { witness.DeclaredType = "boolean" },
		"two lines":    func(witness *negativeWitness) { witness.Stderr += witness.Stderr },
		"wrong exit":   func(witness *negativeWitness) { witness.Exit = 1 },
		"no ruling":    func(witness *negativeWitness) { witness.Ruling = "" },
		"no repair":    func(witness *negativeWitness) { witness.RepairTo = witness.RepairFrom },
		"unknown form": func(witness *negativeWitness) { witness.Form = "other" },
	} {
		bad := witness
		change(&bad)
		writeNegativeList(
			test,
			path,
			negativeWitnessList{Comment: list.Comment, Witnesses: map[string]negativeWitness{contentHash(source): bad}},
		)
		if _, readNegativeWitnessesError := readNegativeWitnesses(path); readNegativeWitnessesError == nil {
			test.Errorf("%s accepted", name)
		}
	}
}

// The declaration must declare the ruled type, the repair must replace one place, and its hash must be
// the copy's.
func TestNegativeCheckedBoundarySource(test *testing.T) {
	test.Parallel()
	source, witness, _ := checkedBoundaryFixture(test)
	for name, change := range map[string]func(*negativeWitness){
		"undeclared":   func(witness *negativeWitness) { witness.Declaration = "corpus/input.a:3" },
		"two places":   func(witness *negativeWitness) { witness.RepairFrom, witness.RepairTo = "count", "total" },
		"another hash": func(witness *negativeWitness) { witness.TypeCorrectSha256 = contentHash(source) },
	} {
		bad := witness
		change(&bad)
		if validateWitnessSourceError := validateWitnessSource(source, bad); validateWitnessSourceError == nil {
			test.Errorf("%s accepted", name)
		}
	}
}

// An optional field's read expects its declared type plus undefined, so a stop that expected
// Identifier | undefined is declared by name?: Identifier, written as tsc writes it. A required field
// declares exactly its type, and an optional field of another type declares nothing here. Mutant:
// declaresRuledType demanding the bare declared type, which refuses the optional field.
func TestNegativeCheckedBoundaryOptionalField(test *testing.T) {
	test.Parallel()
	witness := negativeWitness{
		Form:         checkedBoundaryForm,
		RulingTask:   "#1yr7rdh",
		Ruling:       "Adamic, Oct 28",
		Declaration:  "corpus/input.ts:1",
		DeclaredType: "Identifier | undefined",
		RepairFrom:   "name: 5",
		RepairTo:     "name: { escapedText: 'five' }",
		Exit:         70,
		Stderr:       "adamic: panic: field read failed: declaration.name matches no member of Identifier | undefined; expected Identifier | undefined, found number\n",
	}
	for _, testCase := range []struct {
		name, declaration string
		declares          bool
	}{
		{"optional", "interface Declaration { readonly name?: Identifier; }", true},
		{"spelled", "interface Declaration { readonly name?: Identifier | undefined; }", true},
		{"required", "interface Declaration { readonly name: Identifier; }", false},
		{"optional of another type", "interface Declaration { readonly name?: StringLiteral; }", false},
		{"wider", "interface Declaration { readonly name?: Identifier | StringLiteral; }", false},
	} {
		source := []byte(testCase.declaration + "\nconst raw = { name: 5 };\n")
		witness := witness
		witness.TypeCorrectSha256 = contentHash(
			[]byte(strings.ReplaceAll(string(source), witness.RepairFrom, witness.RepairTo)),
		)
		validateWitnessSourceError := validateWitnessSource(source, witness)
		if testCase.declares && validateWitnessSourceError != nil {
			test.Errorf("%s: refused: %v", testCase.name, validateWitnessSourceError)
		}
		if !testCase.declares && (validateWitnessSourceError == nil ||
			validateWitnessSourceError.Error() != "violated declaration does not declare its ruled type") {
			test.Errorf("%s: want the declaration refused, got %v", testCase.name, validateWitnessSourceError)
		}
	}
}

// A run is the ruled stop only with its exit and its whole message on both backends.
func TestNegativeCheckedBoundaryExactStop(test *testing.T) {
	test.Parallel()
	_, witness, _ := checkedBoundaryFixture(test)
	node := observation{Exit: 0, Stdout: "42\n"}
	stop := observation{Exit: 70, Stderr: witness.Stderr}
	if !acceptsNegativeWitness(witness, "root", node, stop, stop) {
		test.Fatal("ruled stop rejected")
	}
	for _, bad := range []observation{{Exit: 1, Stderr: witness.Stderr}, {
		Exit:   70,
		Stderr: "adamic: panic: field read failed: viewed.count is not a number; expected number, found boolean\n",
	}, {Exit: 70, Stderr: "adamic: panic: compiler bug: unavailable checked union read\n"}} {
		if acceptsNegativeWitness(witness, "root", node, bad, stop) ||
			acceptsNegativeWitness(witness, "root", node, stop, bad) {
			test.Errorf("wrong stop accepted: %+v", bad)
		}
	}
}

// checkedBoundaryCommand runs the command on the fixture, held at corpus/<name>, with a compiler that stops
// whatever holds 'abc', or, when the repair misfits too, everything, and refuses a program whose extension
// isn't name's, as stage 1 refuses in .a the tagless cast a .ts witness makes.
func checkedBoundaryCommand(test *testing.T, repairAgrees bool, name string) (observation, report) {
	test.Helper()
	dir, _, base, head := commandFixture(test)
	source, witness, list := checkedBoundaryFixture(test)
	witness.Declaration = "corpus/" + name + ":1"
	list.Witnesses[contentHash(source)] = witness
	stops := `grep -q "'abc'" "$2"`
	if !repairAgrees {
		stops = "true"
	}
	writeExecutable(test, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\nesac\ncase \"$2\" in\n*"+
		filepath.Ext(name)+
		") ;;\n*) echo \"adamic: $2: Adamic 0.1 refuses a downcast\" >&2; exit 1;;\nesac\ncase \"$1\" in\njs) if "+
		stops+
		"; then echo '// STOP'; else echo '// GO'; fi;;\nbuild) if "+
		stops+
		"; then printf '#!/bin/sh\\necho 42\\necho \"adamic: panic: field read failed: viewed.count is not a number; expected number, found string\" >&2\\nexit 70\\n' > \"$4.tmp\"; else printf '#!/bin/sh\\necho 42\\n' > \"$4.tmp\"; fi && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
	for path, data := range map[string]string{
		"corpus/" + name:  string(source),
		"oracle/node.mjs": "import {readFileSync} from 'node:fs';const s=readFileSync(process.argv[2],'utf8');console.log('42');if(s.includes('STOP')){console.error('adamic: panic: field read failed: viewed.count is not a number; expected number, found string');process.exit(70);}\n",
	} {
		if writeError := os.WriteFile(filepath.Join(dir, path), []byte(data), 0755); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if name != "input.a" {
		if _, gitError := git(dir, "rm", "-q", "corpus/input.a"); gitError != nil {
			test.Fatal(gitError)
		}
	}
	writeNegativeList(test, filepath.Join(dir, "cloud/admission-corpus/negative-witnesses.json"), list)
	if _, gitError := git(
		dir,
		"add",
		"corpus/"+name,
		"oracle/node.mjs",
		"cloud/admission-corpus/negative-witnesses.json",
	); gitError != nil {
		test.Fatal(gitError)
	}
	if _, gitError := git(dir, "commit", "-qm", "checked boundary witness"); gitError != nil {
		test.Fatal(gitError)
	}
	sha, gitError := git(dir, "rev-parse", "HEAD")
	if gitError != nil {
		test.Fatal(gitError)
	}
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--corpus",
		"corpus",
		"--json",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr, observation.Stdout)
	}
	return observation, report
}

// A ruled checked-boundary stop whose repair agrees with Node is a witness, not an agreement.
func TestNegativeCheckedBoundaryCommandAccepted(test *testing.T) {
	test.Parallel()
	observation, report := checkedBoundaryCommand(test, true, "input.a")
	if observation.Exit != 0 ||
		report.Verdict != "pass" ||
		report.AcceptedWitnesses != 1 ||
		!report.Programs[0].NegativeWitness ||
		*report.Programs[0].Agree ||
		report.Programs[0].RepairAgree == nil ||
		!*report.Programs[0].RepairAgree {
		test.Fatal("ruled checked boundary not accepted", observation, report)
	}
}

// A .ts witness's repair is written and compiled as .ts, as the witness is: stage 1 (#q97a1rr) refuses in
// .a the tagless cast view_logical_write_stops' four .ts witnesses make, so a repair copied as .a never
// compiled and the run failed. Mutant: writing the repair as repair-N.a again refuses it, and the run
// fails.
func TestNegativeCheckedBoundaryRepairKeepsItsExtension(test *testing.T) {
	test.Parallel()
	observation, report := checkedBoundaryCommand(test, true, "input.ts")
	if observation.Exit != 0 ||
		report.Verdict != "pass" ||
		report.AcceptedWitnesses != 1 ||
		len(report.Programs) != 1 ||
		report.Programs[0].Path != "corpus/input.ts" ||
		!report.Programs[0].NegativeWitness ||
		report.Programs[0].RepairAgree == nil ||
		!*report.Programs[0].RepairAgree {
		test.Fatal("a .ts witness's repair did not compile as .ts", observation, report)
	}
}

// The repair is the stop's mutant: one that stops too, where Node prints, fails the run. Mutant: not
// running the repair, or not requiring its agreement, accepts it.
func TestNegativeCheckedBoundaryRepairMustAgree(test *testing.T) {
	test.Parallel()
	observation, report := checkedBoundaryCommand(test, false, "input.a")
	if observation.Exit == 0 ||
		report.Verdict != "fail" ||
		report.AcceptedWitnesses != 0 ||
		report.Programs[0].NegativeWitness ||
		report.Programs[0].RepairAgree == nil ||
		*report.Programs[0].RepairAgree {
		test.Fatal("a repair that disagrees with Node was accepted", observation, report)
	}
}
