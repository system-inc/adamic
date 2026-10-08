package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

const clockStringFixture = "internal/oracle/testdata/representation_clock_string_null_undefined.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{clockStringFixture, true, false})
}

func TestRepresentationClockStringNullUndefined(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, clockStringFixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	got, sanitized := natively(t, program)
	for backend, result := range map[string]run{"native": got, "JavaScript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(node, result); diff != "" {
			t.Fatalf("%s: %s; %+v", backend, diff, result)
		}
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
}

func TestRepresentationClockStringNullUndefinedMutant(t *testing.T) {
	t.Run("clock-string-null-undefined-collapse-null", func(t *testing.T) {
		path, err := filepath.Abs(filepath.Join(repository, clockStringFixture))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		changed := 0
		for at, statement := range program.Main {
			expression, ok := statement.(ir.Evaluate)
			if !ok {
				continue
			}
			call, ok := expression.Value.(ir.Call)
			if !ok {
				continue
			}
			for index, argument := range call.Arguments {
				if null, ok := argument.(ir.Null); ok && null.Of == ir.NullishString {
					call.Arguments[index] = ir.Undefined{Of: null.Of}
					changed++
				}
			}
			expression.Value = call
			program.Main[at] = expression
		}
		if changed != 1 {
			t.Fatalf("want one null encoding, changed %d: %#v", changed, program.Main)
		}
		node := onNode(t, path)
		got, sanitized := natively(t, program)
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("unclean mutant: %+v", got)
		}
		if diff := disagreement(node, got); diff != "stdout differs" {
			t.Fatalf("mutant escaped: %q", diff)
		}
		nodeLines, mutantLines := strings.Split(string(node.stdout), "\n"), strings.Split(string(got.stdout), "\n")
		if len(nodeLines) != 5 || len(mutantLines) != 5 || nodeLines[2] != "object:true:false:fallback" || mutantLines[2] != "undefined:false:true:fallback" {
			t.Fatal("missing null/undefined witness")
		}
		t.Logf("Node null input: %s; mutant: %s; clean stdout difference kills mutant", nodeLines[2], mutantLines[2])
		if report := leaks(t, program, sanitized); report != "" {
			t.Fatal(report)
		}
	})
}

func TestRepresentationClockStringNullUndefinedNarrowing(t *testing.T) {
	for _, slot := range []string{"local", "field"} {
		for _, absent := range []string{"null", "undefined"} {
			t.Run(slot+"/"+absent, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "stale.a")
				declaration, read := "let value: string | null | undefined = 'live';", "value"
				if slot == "field" {
					declaration, read = "let box: { value: string | null | undefined } = { value: 'live' };", "box.value"
				}
				source := declaration + "\nfunction change(): void { " + read + " = " + absent + "; }\nif (typeof " + read + " === 'string') { change(); console.log(String(" + read + ".length)); }\n"
				if err := os.WriteFile(path, []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				got, sanitized := natively(t, program)
				js := onJavaScriptBackend(t, program)
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "union member where the checker narrowed it away") {
					t.Fatalf("stale narrowing escaped: %+v", got)
				}
				if diff := disagreement(js, got); diff != "" {
					t.Fatal(diff)
				}
				if report := leaks(t, program, sanitized); report != "" {
					t.Fatal(report)
				}
			})
		}
	}
}

func TestRepresentationClockStringNullUndefinedStorage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage.a")
	source := `function hold(value: string | null | undefined): string | null | undefined { return value; }
console.log(String(hold('t'.repeat(2)) === hold('tt')));
console.log(typeof hold(null));
console.log(typeof hold(undefined));
type Holder = { value: string | null | undefined };
let box: Holder = { value: 'live' };
function changeField(): void { box.value = null; }
if (typeof box.value === 'string') {
 changeField();
 console.log(typeof (box.value));
 console.log(String(box.value === null));
 console.log(box.value ?? 'fallback');
}
let value: string | null | undefined = 'live';
function changeLocal(): void { value = undefined; }
if (typeof value === 'string') {
 changeLocal();
 console.log(typeof value);
 console.log(value ?? 'fallback');
}
function truth(value: string | null | undefined): void { console.log(value ? 'true' : 'false'); }
truth('text'); truth(''); truth(null); truth(undefined);
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	got, sanitized := natively(t, program)
	for backend, result := range map[string]run{"native": got, "JavaScript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(node, result); diff != "" {
			t.Fatalf("%s: %s; %+v", backend, diff, result)
		}
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
}
