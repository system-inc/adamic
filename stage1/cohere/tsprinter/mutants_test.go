package tsprinter

import (
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
