package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var batch4Names = []string{"@typescript-eslint/adjacent-overload-signatures", "@typescript-eslint/ban-tslint-comment", "@typescript-eslint/consistent-type-definitions", "@typescript-eslint/default-param-last", "@typescript-eslint/init-declarations", "@typescript-eslint/no-dupe-class-members", "@typescript-eslint/no-duplicate-enum-values", "@typescript-eslint/no-dynamic-delete", "@typescript-eslint/no-extra-non-null-assertion", "@typescript-eslint/no-import-type-side-effects", "@typescript-eslint/no-misused-new", "@typescript-eslint/no-non-null-asserted-optional-chain", "@typescript-eslint/no-this-alias", "@typescript-eslint/no-unnecessary-parameter-property-assignment", "@typescript-eslint/no-unsafe-function-type", "@typescript-eslint/no-useless-empty-export", "@typescript-eslint/prefer-as-const", "@typescript-eslint/triple-slash-reference", "@typescript-eslint/no-explicit-any", "@typescript-eslint/no-inferrable-types"}

func batch4Selected(name string) bool {
	for _, rule := range batch4Names {
		if name == rule {
			return true
		}
	}
	return false
}
func batch4Generated(t *testing.T) []string {
	t.Helper()
	sources := []string{"function f(a: string): void; const x=1; function f(a: number): void;", "// tslint:disable-next-line\nconst x=1;", "type Shape = { value: number };", "function f(a=1,b:number){}", "let missing: number;", "class C { value=1; value=2; }", "enum E { A=1, B=1, C='a', D='a' }", "delete object[key];", "value!!; value!?.field;", "import { type A, type B as C } from 'module';", "interface I { new(): I; constructor(): void; } declare class C { new(): C; }", "value?.field!; (value?.field)!;", "const receiver=this;", "class C { constructor(public value:number) { this.value=value; } }", "let value: Function;", "import 'module'; export {};", "let value: 'a' = 'a'; const literal = 'b' as 'b';", "/// <reference path=\"module\" />\nconst value=1;", "let value: any;", "let value: number=1; function f(value: boolean=true){}"}
	if len(sources) != len(batch4Names) || len(sources) != 20 {
		t.Fatal("controls must cover twenty rules")
	}
	var rows []string
	for i, source := range sources {
		path := filepath.Join(t.TempDir(), "control.ts")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		options := ""
		if batch4Names[i] == "@typescript-eslint/init-declarations" {
			options = `{"Mode":"always"}`
		}
		rows = append(rows, path+"\t"+batch4Names[i]+"\t\t\t\t"+options)
	}
	return rows
}

// Not parallel: sanitized rebuilds precede timing and bound memory.
func TestBatch4Controls(t *testing.T) {
	directory, _ := filepath.Abs(".")
	oracle := goOracle(t)
	rows := batch4Generated(t)
	for _, row := range rows {
		answer := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
		if string(answer.output) == "0\n" {
			t.Fatalf("inert control: %s", row)
		}
	}
	path := manifest(t, rows)
	want := execute(t, "", oracle, "--manifest", path).output
	if diff := difference(node(t, directory, path, false).output, want); diff != "" {
		t.Fatal(diff)
	}
	compare(t, oracle, buildPort(t, directory, true), directory, path)
}

// Not parallel: upstream capture sets process-wide environment state.
func TestBatch4UpstreamAgree(t *testing.T) {
	directory, _ := filepath.Abs(".")
	oracle := goOracle(t)
	rows := upstream(t)
	selected := []string{}
	counts := map[string]int{}
	var refused []string
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if batch4Selected(fields[1]) {
			if strings.HasSuffix(row, "\tunsupported-recovery") {
				refused = append(refused, row)
			} else {
				selected = append(selected, row)
			}
			counts[fields[1]]++
		}
	}
	for _, name := range batch4Names {
		if counts[name] == 0 {
			t.Fatalf("no upstream cases for %s", name)
		}
		t.Logf("%s: %d captured cases", name, counts[name])
	}
	for _, name := range batch4Names {
		t.Run(name, func(t *testing.T) {
			var ruleRows []string
			for _, row := range selected {
				if strings.Split(row, "\t")[1] == name {
					ruleRows = append(ruleRows, row)
				}
			}
			path := manifest(t, ruleRows)
			want := execute(t, "", oracle, "--manifest", path).output
			if diff := difference(node(t, directory, path, false).output, want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
	if !t.Failed() {
		binary := buildPort(t, directory, true)
		for _, row := range refused {
			checkRecoveryRefusal(t, oracle, binary, directory, row)
		}
		compare(t, oracle, binary, directory, manifest(t, selected))
	}
}

// Not parallel: twenty independent sanitized rebuilds bound peak memory.
func TestBatch4Mutants(t *testing.T) {
	rows := batch4Generated(t)
	path := manifest(t, rows)
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for _, rule := range batch4Names {
		t.Run(rule, func(t *testing.T) {
			stem := strings.ReplaceAll(strings.TrimPrefix(rule, "@typescript-eslint/"), "-", "_")
			directory := mutant(t, "ctx.enabled('"+rule+"')", "ctx.enabled('omitted-batch4-rule')", stem+".ts")
			binary := buildPort(t, directory, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("mutant survived on %s", side.name)
				}
				t.Logf("selected-rule listener omission caught on %s: %s", side.name, difference(side.run.output, want))
			}
		})
	}
}

func batch4Extension(rule, path string) string {
	if !batch4Selected(rule) {
		return ".ts"
	}
	if strings.HasSuffix(path, ".d.ts") {
		return ".d.ts"
	}
	extension := filepath.Ext(path)
	if extension == "" {
		return ".ts"
	}
	return extension
}

// Not parallel: sanitized rebuild precedes timing samples.
func TestBatch4ExtraEditMutant(t *testing.T) {
	path := manifest(t, []string{batch4Generated(t)[9]})
	want := execute(t, "", goOracle(t), "--manifest", path).output
	directory := mutant(t, "for(const edit of finding.extraEdits)", "for(const edit of finding.extraEdits.slice(0, 0))")
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name string
		run  execution
	}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("extra-edit application mutant survived on %s", side.name)
		}
		if !bytes.Equal(bytes.Split(side.run.output, []byte("fixed\t"))[0], bytes.Split(want, []byte("fixed\t"))[0]) {
			t.Fatalf("%s mutant changed findings rather than only fixed source", side.name)
		}
		t.Logf("fixed-source check alone caught omitted extra edits on %s: %s", side.name, difference(side.run.output, want))
	}
}

// Not parallel: sanitized rebuild precedes timing samples.
func TestBatch4FilenameMutant(t *testing.T) {
	source := filepath.Join(t.TempDir(), "definition.d.ts")
	if err := os.WriteFile(source, []byte("debugger; import 'module'; export {};"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{source})
	directory, _ := filepath.Abs(".")
	oracle := goOracle(t)
	want := compare(t, oracle, buildPort(t, directory, true), directory, path)
	directory = mutant(t, "new Parser(result, this.parser.path)", "new Parser(result, 'fixed source')")
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name string
		run  execution
	}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("filename mutant survived on %s", side.name)
		}
		if !bytes.Equal(bytes.Split(side.run.output, []byte("fixed\t"))[0], bytes.Split(want, []byte("fixed\t"))[0]) {
			t.Fatalf("%s filename mutant changed initial findings", side.name)
		}
		t.Logf("fixed-source check alone caught lost declaration-file gate on %s: %s", side.name, difference(side.run.output, want))
	}
}

// Not parallel: sanitized rebuild precedes timing samples.
func TestBatch4NoopReasonMutant(t *testing.T) {
	source := filepath.Join(t.TempDir(), "noop.ts")
	if err := os.WriteFile(source, []byte("type Shape = { value: number }"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{source + "\t@typescript-eslint/consistent-type-definitions"})
	want := execute(t, "", goOracle(t), "--manifest", path).output
	if !bytes.Contains(want, []byte("the fix replaces text with itself")) {
		t.Fatal("Go control lacks no-op rejection")
	}
	directory := mutant(t, "the fix replaces text with itself", "the fix was ignored")
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name string
		run  execution
	}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("no-op reason mutant survived on %s", side.name)
		}
		if !bytes.Equal(bytes.ReplaceAll(side.run.output, []byte("the fix was ignored"), []byte("the fix replaces text with itself")), want) {
			t.Fatalf("%s mutant changed more than rejection metadata", side.name)
		}
		t.Logf("rejection-metadata check alone caught wrong no-op reason on %s: %s", side.name, difference(side.run.output, want))
	}
}

// Not parallel: sanitized rebuild precedes timing samples.
func TestBatch4UnicodeWhitespace(t *testing.T) {
	var rows []string
	for _, source := range []string{"/* tslint:disable \u0085*/", "/* tslint:disable \ufeff*/"} {
		path := filepath.Join(t.TempDir(), "unicode.ts")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\t@typescript-eslint/ban-tslint-comment")
	}
	path := filepath.Join(t.TempDir(), "reference.ts")
	if err := os.WriteFile(path, []byte("/// <reference ignored\u0085path=\"module\" />\n"), 0644); err != nil {
		t.Fatal(err)
	}
	rows = append(rows, path+"\t@typescript-eslint/triple-slash-reference\t\t\t\t\trecovery")
	manifestPath := manifest(t, rows)
	oracle := goOracle(t)
	directory, _ := filepath.Abs(".")
	want := compare(t, oracle, buildPort(t, directory, true), directory, manifestPath)
	for _, change := range []struct{ name, file, from, to string }{
		{"JavaScript trim replaces Go trim", "ban_tslint_comment.ts", "return text.slice(start, end);", "return text.trim();"},
		{"JavaScript whitespace replaces Go Fields", "triple_slash_reference.ts", "if(space(character))", "if(character.trim() === '')"},
	} {
		t.Run(change.name, func(t *testing.T) {
			directory := mutant(t, change.from, change.to, change.file)
			binary := buildPort(t, directory, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, manifestPath, false)}, {"native", execute(t, "", binary, "--manifest", manifestPath)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("Unicode whitespace mutant survived on %s", side.name)
				}
				t.Logf("Unicode whitespace mutant caught on %s: %s", side.name, difference(side.run.output, want))
			}
		})
	}
}
