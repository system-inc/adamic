package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

type step09Fixture struct {
	File          string `json:"file"`
	Failing       bool   `json:"failing"`
	NodeOutput    string `json:"node_stdout"`
	RuntimeOutput string `json:"runtime_stdout"`
	RuntimeExit   int    `json:"runtime_exit"`
}

func step09Fixtures(t *testing.T) []step09Fixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/fixtures/checked-casts/fixtures.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []step09Fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func TestStep09AcceptanceTagged(t *testing.T) {
	for _, fixture := range step09Fixtures(t) {
		if !strings.HasPrefix(fixture.File, "01_") && !strings.HasPrefix(fixture.File, "02_") {
			continue
		}
		t.Run(fixture.File, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/checked-casts", fixture.File))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != fixture.NodeOutput || len(truth.stderr) != 0 {
				t.Fatalf("Node golden: %#v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := truth
			if fixture.Failing {
				message := "cast failed: this Node is not a EnumDeclaration"
				if strings.HasPrefix(fixture.File, "02_") {
					message = "cast failed: this DeclarationName is not a Identifier | StringLiteral"
				}
				expected = run{exitCode: 70, stdout: []byte(fixture.RuntimeOutput), stderr: []byte("adamic: panic: " + message + "\n")}
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(expected, got); diff != "" {
					t.Fatalf("contract: %s: %#v", diff, got)
				}
			}
			if !fixture.Failing {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			// Omit only the actual cast-point tag test. Later field checks remain, so
			// the exact cast diagnostic pins the obligation at the cast itself.
			changed := 0
			omit := func(value ir.Expression) ir.Expression {
				if cast, ok := value.(ir.CheckedCast); ok {
					changed++
					return cast.Value
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), omit)
			if changed != 1 {
				t.Fatalf("want one tag check, got %d", changed)
			}
			for backend, got := range map[string]run{"native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if disagreement(expected, got) == "" {
					t.Fatalf("%s tag omission survived", backend)
				}
				t.Logf("%s tag omission caught by exact cast-point diagnostic: exit=%d stderr=%q", backend, got.exitCode, got.stderr)
			}
		})
	}
}

func TestStep09AcceptanceScalars(t *testing.T) {
	for _, fixture := range step09Fixtures(t) {
		if !strings.HasPrefix(fixture.File, "05_") && !strings.HasPrefix(fixture.File, "06_") && !strings.HasPrefix(fixture.File, "09_") && !strings.HasPrefix(fixture.File, "10_") {
			continue
		}
		t.Run(fixture.File, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/checked-casts", fixture.File))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != fixture.NodeOutput || len(truth.stderr) != 0 {
				t.Fatalf("Node golden: %#v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := truth
			if fixture.Failing {
				message := "cast failed: value expected string, found number"
				if strings.HasPrefix(fixture.File, "10_") {
					message = "cast failed: args[0] expected number, found undefined"
				}
				if strings.HasPrefix(fixture.File, "05_") {
					message = "cast failed: tryExtractTSExtension(candidate) expected Extension, found string"
				}
				if strings.HasPrefix(fixture.File, "06_") {
					message = "cast failed: currentToken expected SyntaxKind.WithKeyword | SyntaxKind.AssertKeyword, found number"
				}
				expected = run{exitCode: 70, stdout: []byte(fixture.RuntimeOutput), stderr: []byte("adamic: panic: " + message + "\n")}
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(expected, got); diff != "" {
					t.Fatalf("contract: %s: %#v", diff, got)
				}
			}
			if !fixture.Failing {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			changed := 0
			for index, function := range program.Functions {
				if function.Name != "checked_scalar_cast" {
					continue
				}
				outer := function.Body[0].(ir.If)
				if strings.HasPrefix(fixture.File, "09_") || strings.HasPrefix(fixture.File, "10_") {
					program.Functions[index].Body = outer.Then
				} else {
					program.Functions[index].Body = outer.Then[0].(ir.If).Then
				}
				changed++
			}
			if changed != 1 {
				t.Fatalf("want one owned scalar check, got %d", changed)
			}
			for backend, got := range map[string]run{"native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if disagreement(expected, got) == "" {
					t.Fatalf("%s scalar omission survived", backend)
				}
				t.Logf("%s scalar omission caught by exact contract: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}

func TestStep09AcceptanceViews(t *testing.T) {
	for _, fixture := range step09Fixtures(t) {
		prefix := fixture.File[:3]
		if prefix != "03_" && prefix != "04_" && prefix != "07_" && prefix != "08_" {
			continue
		}
		t.Run(fixture.File, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/checked-casts", fixture.File))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != fixture.NodeOutput || len(truth.stderr) != 0 {
				t.Fatalf("Node golden: %#v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := truth
			if fixture.Failing {
				message := map[string]string{
					"03_": "cast failed: field read failed: result.info is not a number; expected number, found string",
					"04_": "cast failed: field read failed: result.value is not a string; expected string, found number",
					"07_": "cast failed: element read failed: result[0] expected number, found string",
					"08_": "cast failed: field read failed: result.text is not a string; expected string, found number",
				}[prefix]
				expected = run{exitCode: 70, stdout: []byte(fixture.RuntimeOutput), stderr: []byte("adamic: panic: " + message + "\n")}
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(expected, got); diff != "" {
					t.Fatalf("contract: %s: %#v", diff, got)
				}
			}
			if !fixture.Failing {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			seen, changed := 0, 0
			omit := func(value ir.Expression) ir.Expression {
				if read, ok := value.(ir.ArrayIndex); ok && prefix == "07_" && read.View == "result[0]" {
					seen++
					if seen == 2 {
						read.View = ""
						changed++
						return read
					}
					return value
				}
				if read, ok := value.(ir.Property); ok && read.View != "" && read.Name == map[string]string{"03_": "info", "04_": "value", "08_": "text"}[prefix] {
					seen++
					if prefix != "04_" || seen == 2 {
						read.View = ""
						changed++
						return read
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), omit)
			if changed != 1 {
				t.Fatalf("want one consumed view check, got %d (seen %d)", changed, seen)
			}
			mutant, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"sanitized native": mutant, "JavaScript": onJavaScriptBackend(t, program)} {
				if disagreement(expected, got) == "" {
					t.Fatalf("%s read-check omission survived", backend)
				}
				t.Logf("%s consumed-read omission caught: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}

func TestStep09PrimitiveArraySourceWriteLiar(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/checked-casts/07_generic_next.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	lie := func(value ir.Expression) ir.Expression {
		if boxed, ok := value.(ir.Box); ok {
			if number, ok := boxed.Value.(ir.NumberConstant); ok && number.Value == 13 {
				boxed.Value = ir.BooleanConstant{Value: true}
				changed++
				return boxed
			}
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), lie)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), lie)
	if changed != 1 {
		t.Fatalf("want one incoming primitive lie, got %d", changed)
	}
	expected := run{exitCode: 70, stdout: []byte("before\nview\nelement=12\n"), stderr: []byte("adamic: panic: cast failed: field read failed: <array write> matches no member of string | number; expected string | number, found boolean\n")}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(expected, got); diff != "" {
			t.Fatalf("source-write contract: %s: exit=%d stdout=%q stderr=%q", diff, got.exitCode, got.stdout, got.stderr)
		}
	}
}

func TestStep09PrimitiveObjectSourceWriteLiar(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/checked-casts/04_literal_payload_fails.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	contract := ir.ViewContractID(len(program.ViewContracts) + 1)
	program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Boolean, Name: "boolean"})
	changed := 0
	lie := func(statement ir.Statement) ir.Statement {
		if write, ok := statement.(ir.SetProperty); ok && write.Name == "value" && write.WriteContract != 0 {
			write.Value = ir.Box{Value: ir.BooleanConstant{Value: true}}
			write.WriteContract = contract
			changed++
			return write
		}
		return statement
	}
	mutateFallthroughStatements(reflect.ValueOf(&program.Main).Elem(), lie)
	mutateFallthroughStatements(reflect.ValueOf(&program.Functions).Elem(), lie)
	if changed != 1 {
		t.Fatalf("want one incoming object-slot lie, got %d", changed)
	}
	expected := run{exitCode: 70, stdout: []byte("before\nview\nvalue=Ken\n"), stderr: []byte("adamic: panic: field write failed: property 'value' on record { flags: number; value: string; } at 04_literal_payload_fails.a:9 has no compatible declared slot\n")}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(expected, got); diff != "" {
			t.Fatalf("object-slot source-write contract: %s: exit=%d stdout=%q stderr=%q", diff, got.exitCode, got.stdout, got.stderr)
		}
	}
	omit := func(statement ir.Statement) ir.Statement {
		if write, ok := statement.(ir.SetProperty); ok && write.Name == "value" && write.WriteContract == contract {
			write.WriteProven = true
			return write
		}
		return statement
	}
	mutateFallthroughStatements(reflect.ValueOf(&program.Main).Elem(), omit)
	mutateFallthroughStatements(reflect.ValueOf(&program.Functions).Elem(), omit)
	for backend, got := range map[string]run{"native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if disagreement(expected, got) == "" {
			t.Fatalf("%s object-slot source guard omission survived", backend)
		}
		t.Logf("%s object-slot source guard omission caught: exit=%d stderr=%q", backend, got.exitCode, got.stderr)
	}
}

func TestStep09NullableArrayReceiver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nullable-array.a")
	source := "interface Chain { readonly next:(number|string)[]|undefined; } function next<T>(chain:Chain):T[]{return chain.next as T[];} console.log(\"before\");const result=next<number>({next:undefined});console.log(\"view\");console.log(`element=${result[0]}`);"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 70 || string(truth.stdout) != "before\nview\n" || !strings.Contains(string(truth.stderr), "TypeError") {
		t.Fatalf("Node receiver observation: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := run{exitCode: 70, stdout: []byte("before\nview\n"), stderr: []byte("adamic: panic: cast failed: element read failed: result[0] expected number, found undefined\n")}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(expected, got); diff != "" {
			t.Fatalf("receiver check: %s: exit=%d stdout=%q stderr=%q", diff, got.exitCode, got.stdout, got.stderr)
		}
	}
}

func TestStep09ScalarIndexRootLiar(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/checked-casts/10_generator_label.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	lie := func(value ir.Expression) ir.Expression {
		if _, ok := value.(ir.ArrayLiteral); ok {
			changed++
			return ir.Undefined{Of: ir.Array}
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), lie)
	if changed != 1 {
		t.Fatalf("want one array root lie, got %d", changed)
	}
	expected := run{exitCode: 70, stdout: []byte("before\n"), stderr: []byte("adamic: panic: cast failed: field read failed: args[0] matches no member of number | Expression | undefined; expected number | Expression | undefined, found unsupported representation\n")}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(expected, got); diff != "" {
			t.Fatalf("snapshot receiver check: %s: exit=%d stdout=%q stderr=%q", diff, got.exitCode, got.stdout, got.stderr)
		}
	}
}
