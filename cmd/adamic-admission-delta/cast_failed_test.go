package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// castFailedSource lies at line 3, holding a Square where line 5's cast claims a Circle. Its repair builds a
// Circle, so the same cast is true.
const castFailedSource = "type Circle = { readonly kind: 'Circle'; readonly radius: number };\n" +
	"type Square = { readonly kind: 'Square'; readonly side: number };\n" +
	"const shape: Circle | Square = { kind: 'Square', side: 1 };\n" +
	"console.log('42');\n" +
	"const circle = shape as Circle;\n" +
	"console.log(`${circle.radius}`);\n"

// castFailedStopMessage is the tag test's stop for line 5's cast.
const castFailedStopMessage = "adamic: panic: cast failed: this Circle | Square is not a Circle\n"

// castFailedFixture is the cast_failed form's entry for castFailedSource, at name, ruled on #b5w3ycg: the as
// keyword of line 5 starts at column 22.
func castFailedFixture(name string) ([]byte, negativeWitness, negativeWitnessList) {
	source := []byte(castFailedSource)
	from, to := "kind: 'Square', side: 1", "kind: 'Circle', radius: 1"
	witness := negativeWitness{Form: castFailedForm, RulingTask: "#b5w3ycg", Ruling: "Adamic, Nov 16, applying rulings.md:9 and :385", Declaration: "corpus/" + name + ":5", Column: 22, DeclaredType: "Circle", RepairFrom: from, RepairTo: to, TypeCorrectSHA256: contentHash([]byte(strings.ReplaceAll(castFailedSource, from, to))), Exit: 70, Stderr: castFailedStopMessage}
	return source, witness, negativeWitnessList{Comment: "Adding an entry is a ruling routed to @system_adamic", Witnesses: map[string]negativeWitness{contentHash(source): witness}}
}

// Only the tag test's own message qualifies, its claimed type as the declared type, under a named ruling with
// a column and a .ts declaration; and only this form names a column. Mutant: dropping supportedCastFailed's
// .ts test accepts the .a declaration.
func TestNegativeCastFailedMetadata(t *testing.T) {
	t.Parallel()
	source, witness, list := castFailedFixture("input.ts")
	path := filepath.Join(t.TempDir(), "negatives.json")
	writeNegativeList(t, path, list)
	if _, err := readNegativeWitnesses(path); err != nil {
		t.Fatal("ruled cast_failed witness rejected:", err)
	}
	if err := validateWitnessSource(source, witness); err != nil {
		t.Fatal("ruled cast_failed source rejected:", err)
	}
	for name, change := range map[string]func(*negativeWitness){
		"compiler bug": func(w *negativeWitness) {
			w.Stderr = "adamic: panic: compiler bug: cast failed: this Circle | Square is not a Circle\n"
		},
		"another guard": func(w *negativeWitness) {
			w.Stderr = "adamic: panic: field read failed: shape.radius is not a number; expected number, found undefined\n"
		},
		"another claim":  func(w *negativeWitness) { w.DeclaredType = "Square" },
		"two lines":      func(w *negativeWitness) { w.Stderr += w.Stderr },
		"wrong exit":     func(w *negativeWitness) { w.Exit = 1 },
		"no ruling":      func(w *negativeWitness) { w.Ruling = "" },
		"no column":      func(w *negativeWitness) { w.Column = 0 },
		"no repair":      func(w *negativeWitness) { w.RepairTo = w.RepairFrom },
		".a declaration": func(w *negativeWitness) { w.Declaration = "corpus/input.a:5" },
		"checked with column": func(w *negativeWitness) {
			_, checked, _ := checkedBoundaryFixture(t)
			*w = checked
			w.Column = 3
		},
	} {
		bad := witness
		change(&bad)
		writeNegativeList(t, path, negativeWitnessList{Comment: list.Comment, Witnesses: map[string]negativeWitness{contentHash(source): bad}})
		if _, err := readNegativeWitnesses(path); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

// The stop names no position, so the entry's column must be where an as cast's keyword starts, and the repair
// must replace one place and keep that cast. Mutant: castAt accepting any column whose line holds " as "
// accepts the shifted columns.
func TestNegativeCastFailedPosition(t *testing.T) {
	t.Parallel()
	source, witness, _ := castFailedFixture("input.ts")
	for name, change := range map[string]func(*negativeWitness){
		"column before": func(w *negativeWitness) { w.Column = 21 },
		"column after":  func(w *negativeWitness) { w.Column = 23 },
		"another line":  func(w *negativeWitness) { w.Declaration = "corpus/input.ts:4" },
		"cast removed":  func(w *negativeWitness) { w.RepairFrom, w.RepairTo = "shape as Circle", "shape" },
		"two places": func(w *negativeWitness) {
			w.RepairFrom, w.RepairTo = "readonly", "readonly "
		},
	} {
		bad := witness
		change(&bad)
		bad.TypeCorrectSHA256 = contentHash(repairedSource(source, bad))
		if err := validateWitnessSource(source, bad); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	for name, line := range map[string]string{
		"as in a string": "const note = 'shape as Circle';",
		"a name":         "const alias = shape;",
	} {
		if column := strings.Index(line, "as") + 1; castAt(line, column) {
			t.Errorf("%s: castAt accepted %q at %d", name, line, column)
		}
	}
	for _, line := range []string{"const circle = shape as Circle;", "return (value) as Circle;", "const first = list[0]\tas Circle;"} {
		if column := strings.LastIndex(line, "as ") + 1; !castAt(line, column) {
			t.Errorf("castAt refused the cast in %q at %d", line, column)
		}
	}
}

// The stop names no position, and it names the concrete target rather than the text written, so a witness
// holds one cast: any other word as refuses it, wherever it sits, strings and comments included, and only
// as const and names holding the letters are kept (Compiler on #s1gqkzm). Mutant: otherAs finding nothing
// accepts every "refused" row.
func TestNegativeCastFailedOneCast(t *testing.T) {
	t.Parallel()
	_, witness, _ := castFailedFixture("input.ts")
	for name, test := range map[string]struct {
		added   string
		refused bool
	}{
		"same cast after":       {"const again = shape as Circle;\n", true},
		"same cast in a call":   {"console.log((shape as Circle).kind);\n", true},
		"opening target":        {"const either = shape as Circle | Square;\n", true},
		"another target":        {"const other = shape as Square;\n", true},
		"an alias":              {"type C = Circle;\nconst aliased = shape as C;\n", true},
		"a generic":             {"function claim<T>(value: unknown): T { return value as T; }\n", true},
		"a parenthesized type":  {"const wrapped = shape as (Circle);\n", true},
		"a line-final as":       {"const broken = shape as\n  Circle;\n", true},
		"a template before":     {"const text = `s` as string;\n", true},
		"the word in a comment": {"// read as the circle it claims to be\n", true},
		"as const":              {"const sizes = [1, 2] as const;\nconst tight = [3]  as\tconst;\n", false},
		"names holding as":      {"const alias = 1, cast = 2, has = 3, asCircle = 4, $as = 5, as_ = 6;\n", false},
	} {
		source := []byte(castFailedSource + test.added)
		bad := witness
		bad.TypeCorrectSHA256 = contentHash(repairedSource(source, bad))
		err := validateWitnessSource(source, bad)
		if test.refused && (err == nil || !strings.Contains(err.Error(), "holds another as")) {
			t.Errorf("%s: a second as accepted: %v", name, err)
		}
		if !test.refused && err != nil {
			t.Errorf("%s: refused: %v", name, err)
		}
	}
}

// castFailedCommand runs the command on castFailedSource held at corpus/<name> under an entry declaring
// corpus/<declared>, with a compiler whose backends print printed and then stop at the cast whenever the lie
// is in the program.
func castFailedCommand(t *testing.T, name, declared, printed string) (observation, report) {
	t.Helper()
	dir, _, base, head := commandFixture(t)
	source, witness, list := castFailedFixture(declared)
	list.Witnesses[contentHash(source)] = witness
	lie := `grep -q "kind: 'Square', side: 1" "$2"`
	writeExecutable(t, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) if "+lie+"; then echo '// STOP'; else echo '// GO'; fi;;\nbuild) if "+lie+"; then printf '#!/bin/sh\\nprintf \""+strings.ReplaceAll(printed, "\n", "\\n")+"\"\\necho \""+strings.TrimSuffix(castFailedStopMessage, "\n")+"\" >&2\\nexit 70\\n' > \"$4.tmp\"; else printf '#!/bin/sh\\necho 42\\necho 1\\n' > \"$4.tmp\"; fi && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
	node := "import {readFileSync} from 'node:fs';const s=readFileSync(process.argv[2],'utf8');" +
		"if(s.includes('// STOP')){process.stdout.write(" + jsString(printed) + ");console.error(" + jsString(strings.TrimSuffix(castFailedStopMessage, "\n")) + ");process.exit(70);}" +
		"console.log('42');console.log(s.includes(\"kind: 'Square', side: 1\")?'undefined':'1');\n"
	for path, data := range map[string]string{"corpus/" + name: string(source), "oracle/node.mjs": node} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(data), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if name != "input.a" {
		if _, err := git(dir, "rm", "-q", "corpus/input.a"); err != nil {
			t.Fatal(err)
		}
	}
	writeNegativeList(t, filepath.Join(dir, "cloud/admission-corpus/negative-witnesses.json"), list)
	if _, err := git(dir, "add", "corpus/"+name, "oracle/node.mjs", "cloud/admission-corpus/negative-witnesses.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := git(dir, "commit", "-qm", "cast_failed witness"); err != nil {
		t.Fatal(err)
	}
	sha, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--corpus", "corpus", "--json")
	var r report
	if name != declared {
		return o, r
	}
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr, o.Stdout)
	}
	return o, r
}

// jsString is text as a JavaScript string literal.
func jsString(text string) string {
	data, _ := json.Marshal(text)
	return string(data)
}

// A ruled cast stop whose backends printed what Node printed up to the cast, and whose repair keeps the cast
// and agrees with Node, is a witness. Backends that printed what Node never did fail the run, and so does a
// .a program holding the same bytes. Mutants: dropping acceptsNegativeWitness's prefix test accepts the
// "41" row; dropping run's .ts test accepts the .a row.
func TestNegativeCastFailedCommand(t *testing.T) {
	t.Parallel()
	o, r := castFailedCommand(t, "input.ts", "input.ts", "42\n")
	if o.Exit != 0 || r.Verdict != "pass" || r.AcceptedWitnesses != 1 || len(r.Programs) != 1 || !r.Programs[0].NegativeWitness || r.Programs[0].RepairAgree == nil || !*r.Programs[0].RepairAgree {
		t.Fatal("ruled cast_failed witness not accepted", o, r)
	}
	o, r = castFailedCommand(t, "input.ts", "input.ts", "41\n")
	if o.Exit == 0 || r.Verdict != "fail" || r.AcceptedWitnesses != 0 || r.Programs[0].NegativeWitness {
		t.Fatal("backends that printed what Node never did were accepted", o, r)
	}
	o, _ = castFailedCommand(t, "input.a", "input.ts", "42\n")
	if o.Exit != 1 || !strings.Contains(o.Stderr, "a cast_failed witness is .ts alone") {
		t.Fatal("a .a program holding a cast_failed witness's bytes was accepted", o)
	}
}
