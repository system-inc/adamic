package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

func parserConstructionProgram(fixture string) *ir.Program {
	p := &ir.Program{Source: fixture}
	text := func(value string) ir.Expression {
		p.Strings = append(p.Strings, value)
		return ir.StringConstant{Index: len(p.Strings) - 1}
	}
	print := func(value ir.Expression) { p.Main = append(p.Main, ir.WriteLine{Stream: ir.Stdout, Value: value}) }
	number := func(value float64) ir.Expression { return ir.NumberConstant{Value: value} }
	boolText := func(value ir.Expression) ir.Expression { return ir.BooleanToString{Value: value} }
	keys := func(object ir.Expression) ir.Expression {
		return ir.ArrayJoin{Array: ir.ObjectCall{Method: "keys", Arguments: []ir.Expression{object}, Returns: ir.Array}, Separator: text(","), Element: ir.String}
	}
	if fixture == "own-key-order.a" {
		p.Locals = []ir.Local{{Name: "node", Type: ir.Object, Function: -1}}
		node := ir.Read{Local: 0, Of: ir.Object}
		fields := []ir.Field{{Name: "kind", Value: number(1)}}
		for _, name := range []string{"alpha", "beta", "2", "10"} {
			fields = append(fields, ir.Field{Name: name, Value: number(0), Absent: true, Uninitialized: true})
		}
		p.Main = append(p.Main, ir.Declare{Local: 0, Value: ir.ObjectLiteral{Fields: fields}})
		print(keys(node))
		print(boolText(ir.HasProperty{Object: ir.Box{Value: node}, Name: "beta"}))
		for _, field := range []struct {
			name  string
			value float64
		}{{"beta", 20}, {"10", 10}, {"alpha", 3}, {"2", 2}, {"beta", 21}} {
			p.Main = append(p.Main, ir.SetProperty{Object: node, Name: field.name, Value: number(field.value)})
		}
		print(keys(node))
		print(ir.ArrayJoin{Array: ir.ObjectKeys{Object: node}, Separator: text(","), Element: ir.String})
		print(boolText(ir.HasOwn{Object: node, Key: text("beta")}))
		schema := &ir.JSONSchema{Kind: "object"}
		for index, field := range fields {
			name := text(field.Name).(ir.StringConstant).Index
			schema.Fields = append(schema.Fields, ir.JSONField{Name: name, Slot: index, Schema: &ir.JSONSchema{Kind: "number"}})
		}
		print(ir.JSONStringify{Value: node, Schema: schema})
		return p
	}
	p.Locals = []ir.Local{{Name: "nodes", Type: ir.Array, Function: -1}}
	nodes := ir.Read{Local: 0, Of: ir.Array}
	metadata := []ir.Field{}
	for _, name := range []string{"pos", "end", "hasTrailingComma", "transformFlags"} {
		value := number(0)
		if name == "hasTrailingComma" {
			value = ir.BooleanConstant{}
		}
		metadata = append(metadata, ir.Field{Name: name, Value: value, Absent: true})
	}
	p.Main = append(p.Main, ir.Declare{Local: 0, Value: ir.ArrayLiteral{Element: ir.Number, Elements: []ir.Expression{number(1), number(2)}, Metadata: metadata}})
	if fixture == "node-array-keys.a" {
		print(keys(nodes))
		print(boolText(ir.HasProperty{Object: ir.Box{Value: nodes}, Name: "pos"}))
	}
	if fixture == "node-array-length.a" {
		print(ir.NumberToString{Value: ir.Length{Array: nodes}})
	}
	for _, field := range []struct {
		name  string
		value ir.Expression
	}{{"end", number(7)}, {"pos", number(-1)}, {"transformFlags", number(1)}, {"hasTrailingComma", ir.BooleanConstant{Value: true}}} {
		p.Main = append(p.Main, ir.SetProperty{Object: nodes, Name: field.name, Value: field.value})
	}
	switch fixture {
	case "node-array-keys.a":
		print(keys(nodes))
		print(boolText(ir.ArrayIsArray{Value: ir.Box{Value: nodes}}))
	case "node-array-json.a":
		print(ir.JSONStringify{Value: nodes, Schema: &ir.JSONSchema{Kind: "array", Element: &ir.JSONSchema{Kind: "number"}}})
	case "node-array-length.a":
		print(ir.NumberToString{Value: ir.Length{Array: nodes}})
		print(ir.MaybeToString{Value: ir.ArrayIndex{Array: nodes, Index: number(1), Element: ir.Number}})
	default:
		panic("unknown construction fixture")
	}
	return p
}

func parserSourceObservation(t *testing.T, fixture string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("../../stage3/parser-ahead/rulings", fixture))
	if err != nil {
		t.Fatal(err)
	}
	code := `const fs = require('node:fs'); const {stripTypeScriptTypes}=require('node:module'); eval(stripTypeScriptTypes(fs.readFileSync(process.argv[1],'utf8'),{mode:'transform'}));`
	command := exec.Command("node", "--disable-warning=ExperimentalWarning", "-e", code, path)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("source Node: %v\n%s", err, stderr.String())
	}
	return stdout.String()
}

func parserConstructionRun(t *testing.T, fixture string, mutationFile, before, after string) string {
	t.Helper()
	source := C(parserConstructionProgram(fixture))
	binary := filepath.Join(t.TempDir(), "native")
	options := Options{Sanitize: true}
	if mutationFile == "" {
		if err := Build(source, binary, options); err != nil {
			t.Fatal(err)
		}
	} else {
		files, err := readRuntime(runtime, "runtime")
		if err != nil {
			t.Fatal(err)
		}
		changed := false
		for index := range files {
			if files[index].name != mutationFile {
				continue
			}
			content := string(files[index].contents)
			if strings.Count(content, before) != 1 {
				t.Fatalf("mutant needs one site in %s", mutationFile)
			}
			files[index].contents = []byte(strings.Replace(content, before, after, 1))
			changed = true
		}
		if !changed {
			t.Fatal("mutant changed no runtime file")
		}
		directory := t.TempDir()
		compiler, err := exec.LookPath("clang")
		if err != nil {
			t.Fatal(err)
		}
		flags := Flags(options)
		library, err := cachedRuntime(files, flags, compiler, "parser construction mutant", directory)
		if err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(directory, "main.c")
		if err := os.WriteFile(main, []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
		arguments := append(append([]string{}, flags...), "-I", filepath.Dir(library), main)
		arguments = append(arguments, RuntimeLinkFlags(library)...)
		arguments = append(arguments, "-lm", "-o", binary)
		if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
			t.Fatalf("mutant compile: %v\n%s", err, output)
		}
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("native: %v\n%s", err, stderr.String())
	}
	return stdout.String()
}

// Each test owns its IR and its native runtime copy.
func TestParserConstructionBackendOwnKeyOrder(t *testing.T) {
	t.Parallel()
	assertParserConstructionBackend(t, "own-key-order.a")
}

func TestParserConstructionBackendNodeArrayKeys(t *testing.T) {
	t.Parallel()
	assertParserConstructionBackend(t, "node-array-keys.a")
}

func TestParserConstructionBackendNodeArrayJson(t *testing.T) {
	t.Parallel()
	assertParserConstructionBackend(t, "node-array-json.a")
}

func TestParserConstructionBackendNodeArrayLength(t *testing.T) {
	t.Parallel()
	assertParserConstructionBackend(t, "node-array-length.a")
}

func assertParserConstructionBackend(t *testing.T, fixture string) {
	t.Helper()
	want := parserSourceObservation(t, fixture)
	if got := parserConstructionRun(t, fixture, "", "", ""); got != want {
		t.Fatalf("native stdout %q; Node %q", got, want)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(parserConstructionProgram(fixture))), 0o600); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", "node", runner, path); got != want {
		t.Fatalf("JavaScript stdout %q; Node %q", got, want)
	}
	t.Logf("Node, native ASan/UBSan/leaks, JavaScript agree: %q; independent IR test; source lowering covered by oracle.TestParserConstructionSource", want)
}

func TestParserConstructionRuntimeMutantSynthesizedKey(t *testing.T) {
	t.Parallel()
	assertParserConstructionRuntimeMutant(t, "own-key-order.a", "construction.c", "object->write_order[index] = SIZE_MAX;", "object->write_order[index] = index;")
}

func TestParserConstructionRuntimeMutantDeclarationOrder(t *testing.T) {
	t.Parallel()
	assertParserConstructionRuntimeMutant(t, "own-key-order.a", "library_object.c", "comparison == 0 && object->write_order != NULL", "comparison == 0 && object->write_order != NULL && false")
}

func TestParserConstructionRuntimeMutantArrayExtrasInJSON(t *testing.T) {
	t.Parallel()
	assertParserConstructionRuntimeMutant(t, "node-array-json.a", "json_stringify.c", "if (count != 0) { indent(w, depth); }\n\t\tascii(w, \"]\");", "if (!tuple && array->metadata != NULL) { ascii(w, \",\"); (void)write_value(w, array->metadata->slots[0], schema->element, depth + 1); }\n\t\tif (count != 0) { indent(w, depth); }\n\t\tascii(w, \"]\");")
}

func assertParserConstructionRuntimeMutant(t *testing.T, fixture, file, before, after string) {
	t.Helper()
	want := parserSourceObservation(t, fixture)
	got := parserConstructionRun(t, fixture, file, before, after)
	if got == want {
		t.Fatalf("runtime mutant escaped: %q", got)
	}
	t.Logf("Node disagreement caught runtime mutant: got %q, Node %q", got, want)
}
