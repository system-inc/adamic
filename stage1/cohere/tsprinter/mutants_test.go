package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mutation struct{ name, file, from, to, entry string }

var mutations = []mutation{
	{"protocol width confuses Number with parseFloat", "docMain.ts", "printWidth: Number(fields[1] ?? '')", "printWidth: Number.parseFloat(fields[1] ?? '')", "docMain.ts"},
	{"program loses its statement separator", "expressions.ts", "if(parts.length > 0) parts.push(this.docs.hardline());\n                    const previous =", "if(false) parts.push(this.docs.hardline());\n                    const previous =", "statementsMain.ts"},
	{"declaration loses its function keyword", "expressions.ts", "`${async}function${generator}${name}`", "`${async}${node.kind === 'FunctionDeclaration' ? '' : 'function'}${generator}${name}`", "statementsMain.ts"},
	{"assignment chain ignores statement boundaries", "expressions.ts", "!['ExpressionStatement', 'VariableDeclarationList', 'VariableStatement']", "!['UnrelatedStatement', 'VariableDeclarationList', 'VariableStatement']", "main.ts"},
	{"hashbang loses its refusal", "expressions.ts", "source.startsWith('#!') || source.startsWith('\\ufeff#!') || ", "", "main.ts"},
	{"variable statement loses its keyword", "expressions.ts", "this.docs.text(kind),", "this.docs.text(''),", "main.ts"},
	{"await loses its keyword", "expressions.ts", "this.docs.text('await ')", "this.docs.text('void ')", "main.ts"},
	{"yield loses delegation", "expressions.ts", "delegated ? 'yield*' : 'yield'", "delegated ? 'yield' : 'yield'", "main.ts"},
	{"tagged template loses its tag", "expressions.ts", "this.print(this.child(id, 0), id, 'tag')", "this.docs.text('')", "main.ts"},
	{"template preview loses optional stopping boundaries", "expressions.ts", "for(const boundary of this.optionalBoundaries) flatPrinter.optionalBoundaries.add(boundary);", "for(const boundary of this.optionalBoundaries) flatPrinter.sequenceBoundaries.add(boundary);", "main.ts"},
	{"computed key forgets its clean text", "expressions.ts", "const keyText = this.docs.textContent(left);", "const keyText = this.docs.get(left).kind === 'text' ? this.docs.get(left).text : undefined;", "main.ts"},
	{"method loses its key", "expressions.ts", "this.propertyKey(index, offset)", "this.docs.text(\"\")", "main.ts"},
	{"function loses its keyword", "expressions.ts", "`${async}function${generator}${name}`", "`${async}${generator}${name}`", "main.ts"},
	{"optional chain loses its stopping parentheses", "parentheses.ts", "if(optionalBoundaries.has(index) && parent >= 0)", "if(false && parent >= 0)", "main.ts"},
	{"return and throw lose their keyword", "expressions.ts", "this.docs.text(node.kind === 'ReturnStatement' ? 'return' : 'throw')", "this.docs.text(node.kind === 'ReturnStatement' ? 'throw' : 'throw')", "main.ts"},
	{"statement expression loses its semicolon", "expressions.ts", "return this.docs.concat([printed, this.docs.text(';')]);", "return this.docs.concat([printed, this.docs.text('')]);", "main.ts"},
	{"body loses its opening brace", "expressions.ts", "this.docs.text('{'),\n                    this.docs.indent(this.docs.concat([this.docs.hardline(), this.docs.concat(parts)]))", "this.docs.text('['),\n                    this.docs.indent(this.docs.concat([this.docs.hardline(), this.docs.concat(parts)]))", "main.ts"},
	{"simple statement loses debugger", "expressions.ts", "return this.docs.text('debugger;');", "return this.docs.text(';');", "main.ts"},
	{"arrow loses its arrow token", "expressions.ts", "this.docs.text(' =>'),\n                chain ?", "this.docs.text(' '),\n                chain ?", "main.ts"},
	{"template loses its interpolation dollar", "expressions.ts", "this.docs.text('${')", "this.docs.text('{')", "main.ts"},
	{"member chain loses its method dot", "expressions.ts", "this.docs.text(this.ownOptional(index) ? '?.' : '.')", "this.docs.text(this.ownOptional(index) ? '?.' : '')", "main.ts"},
	{"arguments lose their trailing comma", "expressions.ts", "const trailing = this.docs.ifBreak(this.docs.text(','), this.docs.text(''));", "const trailing = this.docs.text('');", "main.ts"},
	{"object loses its property colon", "expressions.ts", "this.docs.text(':')", "this.docs.text('')", "main.ts"},
	{"conditional swaps its separator", "expressions.ts", "this.docs.text('? ')", "this.docs.text(': ')", "main.ts"},
	{"assignment loses its operator", "expressions.ts", "this.docs.text(` ${this.operator(index)}`)", "this.docs.text(' ')", "main.ts"},
	{"sequence loses a comma", "expressions.ts", "parts.push(this.docs.text(','));", "parts.push(this.docs.text(''));", "main.ts"},
	{"group ignores width", "doc.ts", "!document.broken &&\n                        this.fits(flat", "!document.broken ||\n                        this.fits(flat", "docMain.ts"},
	{"fill never packs a pair", "doc.ts", "separatorMode = fitsPair ? 2 : 1;", "separatorMode = 1;", "docMain.ts"},
	{"expression loses required parentheses", "expressions.ts", "return this.parenthesize(id, parent, role)", "return parent < -1 && this.parenthesize(id, parent, role)", "main.ts"},
}

func mutatedPort(t *testing.T, change mutation) string {
	t.Helper()
	directory := t.TempDir()
	parser, err := filepath.Abs("../../typescript/parser")
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	applied := false
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		if file == change.file {
			if strings.Count(source, change.from) != 1 {
				t.Fatalf("mutant %q must change one place", change.name)
			}
			source = strings.Replace(source, change.from, change.to, 1)
			applied = true
		}
		source = strings.ReplaceAll(source, "../../typescript/parser/", parser+"/")
		if err = os.WriteFile(filepath.Join(directory, file), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !applied {
		t.Fatal("mutant not applied")
	}
	return filepath.Join(directory, change.entry)
}

func TestMutants(t *testing.T) {
	t.Parallel()
	documents, documentWant, _ := documentCorpus(t)
	expressions, expressionWant, _ := expressionCorpus(t)
	statementCases, statementWant := "", ""
	for _, change := range mutations {
		if change.entry == "statementsMain.ts" {
			statementCases, statementWant, _ = statementCorpus(t)
			break
		}
	}
	for _, change := range mutations {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			cases, want := documents, documentWant
			arguments := []string{cases}
			if change.entry == "main.ts" {
				cases, want = expressions, expressionWant
				arguments = []string{"--cases", cases, "80"}
			}
			if change.entry == "statementsMain.ts" {
				cases, want = statementCases, statementWant
				arguments = []string{"--cases", cases, "80"}
			}
			if change.name == "hashbang loses its refusal" {
				cases = filepath.Join(filepath.Dir(expressions), "gaps.txt")
				gapWant, err := os.ReadFile(filepath.Join(filepath.Dir(expressions), "gap-answers.txt"))
				if err != nil {
					t.Fatal(err)
				}
				want = string(gapWant)
				arguments = []string{"--cases", cases, "80"}
			}
			path := mutatedPort(t, change)
			node := onNode(t, path, arguments...)
			program := lowered(t, path)
			native, _ := natively(t, program, arguments...)
			for _, side := range []struct {
				name   string
				result run
			}{{"Node", node}, {"native", native}} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
					t.Fatalf("%s mutant must finish normally: exit %d stderr %s", side.name, side.result.exitCode, side.result.stderr)
				}
				if string(side.result.stdout) == want {
					t.Fatalf("%s mutant escaped the byte comparison", side.name)
				}
				difference := firstDifference(string(side.result.stdout), want)
				if filepath.Base(cases) != "gaps.txt" {
					difference = corpusDifference(t, cases, string(side.result.stdout), want)
				}
				t.Logf("%s caught by successful-run output mismatch: %s", side.name, difference)
			}
		})
	}
}

type importFixture struct{ Label, Source, Path, Origin, Go, Prettier, Reason string }

func TestSideEffectImports(t *testing.T) {
	t.Parallel()
	runImportFixtures(t, "testdata/side-effect-imports.json", 45, 30, []mutation{
		{"import keyword lost", "expressions.ts", "const parts: number[] = [this.docs.text('import')];", "const parts: number[] = [this.docs.text('export')];", "statementsMain.ts"},
		{"import separator lost", "expressions.ts", "parts.push(this.docs.text(' '));\n        parts.push(this.print(this.child(index, offset), index));", "parts.push(this.docs.text(''));\n        parts.push(this.print(this.child(index, offset), index));", "statementsMain.ts"},
		{"import semicolon lost", "expressions.ts", "parts.push(this.docs.text(';'));\n        return this.docs.concat(parts);", "parts.push(this.docs.text(''));\n        return this.docs.concat(parts);", "statementsMain.ts"},
		{"import attributes silently dropped", "expressions.ts", "if(this.node(index).children.length > offset + 1) {", "if(this.node(index).children.length > offset + 2) {", "statementsMain.ts"},
	})
}

func TestImportClausesAndAttributes(t *testing.T) {
	t.Parallel()
	runImportFixtures(t, "testdata/import-clauses.json", 61, 0, []mutation{
		{"default binding lost", "expressions.ts", "else standalone.push(this.print(child, index));", "else standalone.push(this.docs.text('lost'));", "statementsMain.ts"},
		{"namespace alias lost", "expressions.ts", "this.docs.text('* as ')", "this.docs.text('* ')", "statementsMain.ts"},
		{"named specifier alias lost", "expressions.ts", "parts.push(this.docs.text(' as '));", "parts.push(this.docs.text(' '));", "statementsMain.ts"},
		{"clause type modifier lost", "expressions.ts", "if(phase === 'TypeKeyword') parts.push(this.docs.text(' type'));", "if(phase === 'TypeKeyword') parts.push(this.docs.text(''));", "statementsMain.ts"},
		{"specifier type modifier lost", "expressions.ts", "if(node.semantic === '1') parts.push(this.docs.text('type '));", "if(node.semantic === '1') parts.push(this.docs.text(''));", "statementsMain.ts"},
		{"attribute value lost", "expressions.ts", "this.print(this.child(child, 1), child),", `this.docs.text("'lost'"),`, "statementsMain.ts"},
		{"single type attribute flattening lost", "expressions.ts", "? this.docs.removeLines(content)", "? content", "statementsMain.ts"},
	})
}

func runImportFixtures(t *testing.T, fixtureFile string, count, realCount int, changes []mutation) {
	t.Helper()
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	var cases []importFixture
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != count {
		t.Fatalf("fixture count %d, want %d", len(cases), count)
	}
	directory := t.TempDir()
	root, _ := filepath.Abs(repository)
	fixtures, _ := filepath.Abs(fixtureFile)
	side, _ := filepath.Abs("testdata/imports_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{root + "/cohere/internal/format/javascript/import_witnesses_test.go": side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err = os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlayPath, "-run=^TestImportFormatterWitnesses$", "./internal/format/javascript")
	command.Dir = root + "/cohere"
	command.Env = append(os.Environ(), "ADAMIC_IMPORT_FIXTURES="+fixtures, "ADAMIC_IMPORT_ANSWERS="+directory+"/go.json")
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go oracle: %v %s", err, output)
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Fatal("ADAMIC_TS_PRETTIER must name prettier@3.9.6; no import oracle skips")
	}
	script, _ := filepath.Abs("testdata/imports.mjs")
	result := execute(t, nil, "node", script, library, fixtures, directory+"/npm.json")
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("Prettier: exit %d stderr %s", result.exitCode, result.stderr)
	}
	npm, err := os.ReadFile(directory + "/npm.json")
	if err != nil {
		t.Fatal(err)
	}
	var fresh []importFixture
	if err = json.Unmarshal(npm, &fresh); err != nil {
		t.Fatal(err)
	}
	if len(fresh) != len(cases) {
		t.Fatal("npm lost cases")
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var batch, want strings.Builder
	accepted, real := 0, 0
	for i, c := range cases {
		if fresh[i].Prettier != c.Prettier {
			t.Fatalf("%s upstream bytes changed", c.Label)
		}
		if strings.HasPrefix(c.Label, "real-source-") {
			real++
		}
		batch.WriteString("0" + escape.Replace(c.Source) + "\n")
		if c.Reason == "" {
			accepted++
			if c.Go != c.Prettier {
				t.Fatalf("accepted upstream mismatch %s", c.Label)
			}
			want.WriteString("ok\t" + escape.Replace(c.Go) + "\n")
		} else {
			want.WriteString("notyet\t" + c.Reason + "\n")
		}
	}
	if real != realCount || accepted < 20 {
		t.Fatalf("fixture families missing: real=%d accepted=%d", real, accepted)
	}
	path := filepath.Join(directory, "batch.txt")
	if err = os.WriteFile(path, []byte(batch.String()), 0644); err != nil {
		t.Fatal(err)
	}
	entry, _ := filepath.Abs("statementsMain.ts")
	program := lowered(t, entry)
	native, binary := natively(t, program, "--cases", path, "120")
	compare := func(name string, result run) {
		t.Helper()
		if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != want.String() {
			t.Fatalf("%s exit %d stderr %s difference %s", name, result.exitCode, result.stderr, firstDifference(string(result.stdout), want.String()))
		}
	}
	compare("Node", onNode(t, entry, "--cases", path, "120"))
	compare("native", native)
	compare("JavaScript backend", onJavaScriptBackend(t, program, "--cases", path, "120"))
	if report := leaks(t, program, binary, "--cases", path, "120"); report != "" {
		t.Fatal(report)
	}

	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			mutated := mutatedPort(t, change)
			p := lowered(t, mutated)
			for _, side := range []struct {
				name   string
				result run
			}{{"Node", onNode(t, mutated, "--cases", path, "120")}, {"native", nativelyRun(t, p, "--cases", path, "120")}} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
					t.Fatalf("%s mutant must finish normally: %d %s", side.name, side.result.exitCode, side.result.stderr)
				}
				if string(side.result.stdout) == want.String() {
					t.Fatalf("%s mutant survived", side.name)
				}
				t.Logf("%s killed by bytes: %s", side.name, firstDifference(string(side.result.stdout), want.String()))
			}
		})
	}
	t.Logf("%d fixtures, %d accepted, %d full pinned real sources; Go/npm/Node/native/backend and %d mutants", len(cases), accepted, real, len(changes))
}
