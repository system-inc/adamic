package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/apple/naming"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func fixtureConfiguration(t *testing.T) Configuration {
	t.Helper()
	root, err := filepath.Abs("testdata/SDK")
	if err != nil {
		t.Fatal(err)
	}
	return Configuration{Frameworks: []Framework{{Name: "AppKit", Headers: filepath.Join(root, "AppKit.framework/Headers")}, {Name: "Foundation", Headers: filepath.Join(root, "Foundation.framework/Headers")}}, Platform: "macos", Umbrella: "umbrella.h"}
}
func clangPath(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal("clang is required: run cloud/setup.sh", err)
	}
	return path
}
func fixtureOutput(t *testing.T) Output {
	t.Helper()
	output, err := RunClang(context.Background(), clangPath(t), []string{"-x", "objective-c", "-fobjc-runtime=macosx-10.13", "-fsyntax-only", "-fblocks", "-Werror", "-Wno-nullability-completeness", "-F", "testdata/SDK", "-Xclang", "-ast-dump=json", "testdata/umbrella.h"}, fixtureConfiguration(t))
	if err != nil {
		t.Fatal(err)
	}
	return output
}
func TestClangGolden(t *testing.T) {
	t.Parallel()
	output := fixtureOutput(t)
	golden := map[string][]byte{}
	err := filepath.WalkDir("testdata/golden", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel("testdata/golden", path)
		if err != nil {
			return err
		}
		golden[filepath.ToSlash(relative)] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	files := append(append([]File{}, output.Files...), File{Path: "bindings-check.m", Content: output.Check})
	if len(files) != len(golden) {
		t.Fatalf("generated %d files, golden has %d", len(files), len(golden))
	}
	for _, file := range files {
		if !bytes.Equal(file.Content, golden[file.Path]) {
			t.Errorf("golden mismatch: %s\n%s", file.Path, file.Content)
		}
	}
}
func TestPatterns(t *testing.T) {
	t.Parallel()
	output := fixtureOutput(t)
	files := map[string]string{}
	for _, file := range output.Files {
		files[file.Path] = string(file.Content)
	}
	record := files["foundation/fixture-record.d.ts"]
	for _, want := range []string{"takeText(text: string | undefined)", "0:string?", "readonly optionalText: string | undefined", "@objc get displayText -> string", "@objc set replaceText: string", "@objc get label -> string", "@objc set setLabel: string", "@objc method copyRecord -> new object", "performAction(): void", "rawAction(): void", "reader: FixtureReadable"} {
		if !strings.Contains(record, want) {
			t.Errorf("record missing %q", want)
		}
	}
	if err := readAST(strings.NewReader(`{"kind":"TranslationUnitDecl","inner":[{"kind":"VarDecl","inner":[{"kind":"ObjCBoolLiteralExpr","value":true}]}]}`), func(*node) error { return nil }); err != nil {
		t.Fatal("boolean AST literal rejected", err)
	}
	for _, bad := range []string{"platformOnly():", "withCompletion(", "takeArray(", "takeCallback(", "variadic(", "narrowInteger():", "narrowFloat():", "release():", "discardSelf():"} {
		if strings.Contains(record, bad) {
			t.Errorf("unsupported method emitted: %s", bad)
		}
	}
	if !strings.Contains(record, "phoneOnly(): void") {
		t.Error("availability for another platform suppressed a method")
	}
	if !strings.Contains(record, "consumed parameters cannot use") || strings.Contains(record, "@objc method consumeRecord:") {
		t.Error("consumed parameter was passed borrowed")
	}
	if !strings.Contains(files["foundation/fixture-create-text.d.ts"], "-> new string") {
		t.Error("retained C string ownership lost")
	}
	if !strings.Contains(record, "C arrays are not carried") {
		t.Error("array omission has no specific reason")
	}
	panel := files["appkit/fixture-panel.d.ts"]
	for _, want := range []string{"0:rectangle 1.display:boolean 1.animate:boolean", "0:enum(Calm=0,Loud=7)", "1.style:options(Plain=0,Bright=2,Quiet=8,Flipped=9223372036854775808)", "readonly style: readonly FixtureStyle[]", "paint(): void", "constructor(record: FixtureRecord)", "FixtureRecord | undefined"} {
		if !strings.Contains(panel, want) {
			t.Errorf("panel missing %q", want)
		}
	}
	readable := files["foundation/fixture-readable.d.ts"]
	if strings.Contains(readable, "readonly optionalPing:") || !strings.Contains(readable, "optional protocol methods are not proven present") {
		t.Error("optional protocol method was promised present")
	}
	loose := files["foundation/fixture-loose.d.ts"]
	if !strings.Contains(loose, "string | undefined") {
		t.Error("unannotated reference narrowed to nonnull")
	}
	if files["appkit/fixture-after-include.d.ts"] == "" {
		t.Error("declaration after included excluded header was lost")
	}
	for path, content := range files {
		if strings.Contains(path, "excluded") || strings.Contains(content, "neverExport") {
			t.Error("excluded header was bound", path)
		}
	}
	if !strings.Contains(files["foundation/foundation-error.d.ts"], "class FoundationError") {
		t.Error("global naming amendment not used")
	}
}
func TestCanonicalModuleOrder(t *testing.T) {
	t.Parallel()
	first := fixtureOutput(t)
	paths := []string{}
	for _, file := range first.Files {
		paths = append(paths, file.Path)
	}
	if !sort.StringsAreSorted(paths) {
		t.Fatalf("noncanonical module order: %v", paths)
	}
	for i := 0; i < 4; i++ {
		next := fixtureOutput(t)
		if !reflect.DeepEqual(first, next) {
			t.Fatal("identical SDK produced different bytes")
		}
	}
}
func TestHeaderWitness(t *testing.T) {
	t.Parallel()
	output := fixtureOutput(t)
	path := filepath.Join(t.TempDir(), "check.m")
	if err := os.WriteFile(path, output.Check, 0644); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(clangPath(t), "-x", "objective-c", "-fobjc-runtime=macosx-10.13", "-fblocks", "-fsyntax-only", "-Werror", "-Wno-nullability-completeness", "-F", "testdata/SDK", "-I", "testdata", path)
	diagnostics, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("header witness: %v\n%s", err, diagnostics)
	}
	// A wrong enum-width tag must fail the independent header witness, rather
	// than being accepted through an unchecked objc_msgSend function cast.
	mutant := bytes.Replace(output.Check, []byte("long argument0, unsigned long argument1"), []byte("double argument0, unsigned long argument1"), 1)
	// Change the witness assertion's claimed parameter type too, so it checks
	// the mutated tag against the actual header, not just against its own text.
	mutant = bytes.Replace(mutant, []byte("__builtin_types_compatible_p(long, NSFixtureMode)"), []byte("__builtin_types_compatible_p(double, NSFixtureMode)"), 1)
	if err := os.WriteFile(path, mutant, 0644); err != nil {
		t.Fatal(err)
	}
	command = exec.Command(command.Path, command.Args[1:]...)
	diagnostics, err = command.CombinedOutput()
	if err == nil || !bytes.Contains(diagnostics, []byte("configureMode:style: parameter 0 ABI")) {
		t.Fatalf("wrong-width tag was not held by header assertion: %v\n%s", err, diagnostics)
	}
}
func TestDocumentOrderFileDelta(t *testing.T) {
	t.Parallel()
	// The first node's child ends in another file. Its parent must retain the
	// file resolved at loc, while the following omitted file belongs to B.
	input := `{"kind":"TranslationUnitDecl","inner":[
 {"kind":"ObjCInterfaceDecl","name":"First","loc":{"file":"A.h","offset":1},"range":{"begin":{"offset":1},"end":{"offset":2}},"inner":[{"kind":"ObjCMethodDecl","name":"change","loc":{"file":"B.h","offset":3}}]},
 {"kind":"ObjCInterfaceDecl","name":"Second","loc":{"offset":4,"includedFrom":{"file":"wrong.h"}}},
 {"kind":"ObjCInterfaceDecl","name":"Third","loc":{"file":"A.h","offset":5}}
 ]}`
	var files []string
	if err := readAST(strings.NewReader(input), func(n *node) error { files = append(files, n.Location.File); return nil }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, []string{"A.h", "B.h", "A.h"}) {
		t.Fatalf("delta locations: %v", files)
	}
}

type oneByteReader struct {
	reader   io.Reader
	consumed int
}

func (r *oneByteReader) Read(data []byte) (int, error) {
	if len(data) > 1 {
		data = data[:1]
	}
	n, err := r.reader.Read(data)
	r.consumed += n
	return n, err
}
func TestStreamVisitsBeforeEOF(t *testing.T) {
	t.Parallel()
	input := `{"kind":"TranslationUnitDecl","inner":[{"kind":"ObjCInterfaceDecl","name":"First"},{"kind":"ObjCInterfaceDecl","name":"Second"}]}`
	reader := &oneByteReader{reader: strings.NewReader(input)}
	count := 0
	if err := readAST(reader, func(n *node) error {
		count++
		if count == 1 && reader.consumed >= len(input) {
			t.Error("whole AST buffered before first declaration")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("lost declarations")
	}
	for _, bad := range []string{`{}`, `{"kind":"NotAnAST","inner":[]}`, input + `{}`, input[:len(input)-2]} {
		if err := readAST(strings.NewReader(bad), func(*node) error { return nil }); err == nil {
			t.Errorf("invalid AST accepted: %s", bad)
		}
	}
}
func TestGenerationErrors(t *testing.T) {
	t.Parallel()
	configuration := fixtureConfiguration(t)
	ast := `{"kind":"TranslationUnitDecl","inner":[{"kind":"ObjCInterfaceDecl","name":"NSError","loc":{"file":FILE,"offset":1},"inner":[{"kind":"ObjCRootClassAttr"}]},{"kind":"ObjCInterfaceDecl","name":"FoundationError","loc":{"offset":2},"inner":[{"kind":"ObjCRootClassAttr"}]}]}`
	file, _ := json.Marshal(filepath.Join(configuration.Frameworks[1].Headers, "Fixture.h"))
	ast = strings.ReplaceAll(ast, "FILE", string(file))
	if _, err := Generate(strings.NewReader(ast), configuration); err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatal("global collision accepted", err)
	}
	if _, err := RunClang(context.Background(), "false", nil, configuration); err == nil {
		t.Fatal("failed clang produced output")
	}
	if err := (Output{Files: []File{{Path: "../escape.d.ts"}}}).Write(t.TempDir()); err == nil {
		t.Fatal("escaped output directory")
	}
}

func TestCompilerLowersGeneratedModules(t *testing.T) {
	t.Parallel()
	// Go's overlay substitutes an existing embedded seed file at build time.
	// It contains the generated bytes verbatim plus a type-only loader anchor;
	// the repository's seeds and compiler implementation are never modified.
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	var bindings strings.Builder
	output := fixtureOutput(t)
	for _, file := range output.Files {
		bindings.Write(file.Content)
		bindings.WriteByte('\n')
	}
	bindings.WriteString("declare module 'apple/appkit/window' { export interface Window { readonly anchor: boolean; } }\n")
	replacement := filepath.Join(directory, "bindings.d.ts")
	if err := os.WriteFile(replacement, []byte(bindings.String()), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "internal/load/apple/appkit/window.d.ts"): replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "run", "-overlay", overlayPath, "./cmd/adamic", "c", "internal/apple/generate/testdata/bindings.a")
	command.Dir = root
	// Keep compiler stdout and stderr in a log, as required by the repository.
	logPath := filepath.Join(directory, "lower.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	command.Stdout = log
	command.Stderr = log
	runError := command.Run()
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if runError != nil {
		t.Fatalf("go run ./cmd/adamic c: %v\n%s", runError, content)
	}
	for _, want := range []string{"objc_msgSend", "NSFixturePanel", "setFrame:display:animate:", "configureMode:style:", "NSFixtureAdd", "adamic_apple_rectangle", "adamic_apple_options", "adamic_apple_string", "NSFixtureCreateText", "objc_release(result);", "showText:", "showCount:", "initWithTitle:"} {
		if !bytes.Contains(content, []byte(want)) {
			t.Errorf("lowered C missing %q", want)
		}
	}
}

func TestImportNameCollision(t *testing.T) {
	t.Parallel()
	configuration := fixtureConfiguration(t)
	first, _ := json.Marshal(filepath.Join(configuration.Frameworks[0].Headers, "Fixture.h"))
	second, _ := json.Marshal(filepath.Join(configuration.Frameworks[1].Headers, "Fixture.h"))
	input := `{"kind":"TranslationUnitDecl","inner":[
 {"kind":"ObjCInterfaceDecl","name":"NSView","loc":{"file":FIRST,"offset":1},"inner":[{"kind":"ObjCRootClassAttr"},{"kind":"ObjCMethodDecl","name":"takeOther:","instance":true,"returnType":{"qualType":"void"},"inner":[{"kind":"ParmVarDecl","name":"other","type":{"qualType":"UIView * _Nonnull"}}]}]},
 {"kind":"ObjCInterfaceDecl","name":"UIView","loc":{"file":SECOND,"offset":2},"inner":[{"kind":"ObjCRootClassAttr"}]}
 ]}`
	input = strings.ReplaceAll(strings.ReplaceAll(input, "FIRST", string(first)), "SECOND", string(second))
	if _, err := Generate(strings.NewReader(input), configuration); err == nil || !strings.Contains(err.Error(), "already exports") {
		t.Fatal("import replaced local type silently", err)
	}
	g := &generator{types: map[string]*definition{}}
	module := &module{imports: map[string]string{"View": "apple/appkit/view"}}
	g.addImport(module, "apple/foundation/host", &definition{output: naming.Output{Name: "View", Module: "apple/uikit/view"}})
	if g.importError == nil {
		t.Fatal("one import replaced another silently")
	}
}

func TestWrittenTagTypeSpellings(t *testing.T) {
	t.Parallel()
	rectangle := &definition{node: &node{Name: "CGRect"}, declaration: naming.Declaration{Kind: naming.Struct}, output: naming.Output{Name: "Rectangle", Module: "apple/coregraphics/rectangle"}}
	mode := &definition{node: &node{Name: "NSFixtureMode", Underlying: writtenType{Qual: "long"}}, declaration: naming.Declaration{Kind: naming.Enum}, output: naming.Output{Name: "FixtureMode", Module: "apple/appkit/fixture-mode"}, members: []string{"'Calm'"}, values: []string{"0"}}
	g := &generator{types: map[string]*definition{"CGRect": rectangle, "NSFixtureMode": mode}}
	for _, written := range []string{"struct CGRect", "enum NSFixtureMode"} {
		declaration := &node{Name: "consume:", Framework: "AppKit", Result: writtenType{Qual: "void"}, Children: []*node{{Kind: "ParmVarDecl", Name: "value", Type: writtenType{Qual: written}}}}
		_, output, natives, _, reason, err := g.callable(declaration, naming.InstanceMethod, "NSFixtureOwner", "", nil, nil)
		if err != nil || reason != "" || len(natives) != 1 || output.Arguments == nil {
			t.Fatalf("written tag type %s failed: %s %v", written, reason, err)
		}
	}
}

// A protocol sharing a class's name imports as ...Protocol, as Swift does, whichever comes first,
// and id<Name> means the protocol while Name * means the class.
func TestDeclarationKindCollision(t *testing.T) {
	t.Parallel()
	output, err := RunClang(context.Background(), clangPath(t), []string{"-x", "objective-c", "-fobjc-runtime=macosx-10.13", "-fsyntax-only", "-fblocks", "-Werror", "-Wno-nullability-completeness", "-F", "testdata/SDK", "-Xclang", "-ast-dump=json", "testdata/SDK/Foundation.framework/Headers/Collision.h"}, fixtureConfiguration(t))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, file := range output.Files {
		files[file.Path] = string(file.Content)
	}
	for path, wants := range map[string][]string{
		"foundation/fixture-ambiguous.d.ts":          {"export class FixtureAmbiguous", "takeProtocol(protocol: FixtureAmbiguousProtocol)", "takeClass(value: FixtureAmbiguous)", "protocolValue(): FixtureAmbiguousProtocol;", "import type { FixtureAmbiguousProtocol } from 'apple/foundation/fixture-ambiguous-protocol';"},
		"foundation/fixture-ambiguous-protocol.d.ts": {"/** NSFixtureAmbiguous */\n\texport interface FixtureAmbiguousProtocol {", "readonly doThing: () => void;"},
		"foundation/fixture-element.d.ts":            {"export class FixtureElement", "activate(): void;"},
		"foundation/fixture-holder.d.ts":             {"readonly element: FixtureElementProtocol;"},
		"foundation/fixture-element-protocol.d.ts":   {"export interface FixtureElementProtocol {", "readonly describe: () => void;"},
	} {
		content, ok := files[path]
		if !ok {
			t.Errorf("missing %s", path)
			continue
		}
		for _, want := range wants {
			if !strings.Contains(content, want) {
				t.Errorf("%s missing %q:\n%s", path, want, content)
			}
		}
	}
	if strings.Contains(files["foundation/fixture-element-protocol.d.ts"], "activate") || strings.Contains(files["foundation/fixture-element.d.ts"], "describe") {
		t.Error("a class and its same-named protocol shared members")
	}
	for _, want := range []string{"id<NSFixtureAmbiguous> _Nonnull receiver", "NSFixtureAmbiguous * _Nonnull receiver", "id<NSFixtureElement> _Nonnull receiver"} {
		if !strings.Contains(string(output.Check), want) {
			t.Errorf("check file missing %q", want)
		}
	}
}

// Each pattern here stopped the real SDK once; Shapes.h holds them so Linux keeps them working.
func TestMemberShapes(t *testing.T) {
	t.Parallel()
	output, err := RunClang(context.Background(), clangPath(t), []string{"-x", "objective-c", "-fobjc-runtime=macosx-10.13", "-fsyntax-only", "-fblocks", "-Werror", "-Wno-nullability-completeness", "-F", "testdata/SDK", "-Xclang", "-ast-dump=json", "testdata/SDK/Foundation.framework/Headers/Shapes.h"}, fixtureConfiguration(t))
	if err != nil {
		t.Fatal(err)
	}
	var zone string
	for _, file := range output.Files {
		if file.Path == "foundation/fixture-zone.d.ts" {
			zone = string(file.Content)
		}
	}
	for _, want := range []string{
		"@objc init initWithLabel:count: 0.label:string 0.count:integer\n\t\t */\n\t\tconstructor(options: { readonly label: string; readonly count: number });",
		"@objc init initWithKind:count: 0.kind:integer 0.count:integer\n\t\t */\n\t\tconstructor(options: { readonly kind: number; readonly count: number });",
		"@objc init init\n\t\t */\n\t\tconstructor();",
		"// Skipped +[NSFixtureZone zone]: an initializer of the same shape is the constructor.",
		"static path(options: { readonly forResource: string; readonly ofType: string }): string | undefined;",
		"\t\tpath(options: { readonly forResource: string; readonly ofType: string }): string | undefined;",
		"static readonly highlight: FixtureZone;",
		"\t\thighlight(options: { readonly withLevel: number }): FixtureZone;",
		"static readonly interval: number;",
		"\t\treadonly interval: number;",
		"readonly abbreviation: string;",
		"@objc method abbreviationForDate: 0:object -> string?\n\t\t */\n\t\tabbreviationFor(date: FixtureDate): string | undefined;",
		"readonly isDaylightSavingTime: boolean;",
		"isDaylightSavingTimeFor(date: FixtureDate): boolean;",
		"// Skipped -[NSFixtureZone initToMemory]: no Adamic name yet:",
		"@objc set setVertical: boolean\n\t\t */\n\t\tisVertical: boolean;",
	} {
		if !strings.Contains(zone, want) {
			t.Errorf("fixture-zone missing %q", want)
		}
	}
	if strings.Count(zone, "isVertical: boolean;") != 1 {
		t.Error("a category's redeclared property was emitted beside the class's")
	}
	if t.Failed() {
		t.Log(zone)
	}
}
