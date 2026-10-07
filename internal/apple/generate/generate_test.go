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
	"slices"
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
	files := append(append([]File{}, output.Files...), File{Path: "bindings-check.m", Content: output.Check}, File{Path: LeavesFile, Content: output.Leaves}, File{Path: HoldsFile, Content: output.Holds})
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
	for _, want := range []string{"takeText(text: string | undefined)", "0:string?", "readonly optionalText: string | undefined", "@objc get displayText -> string", "@objc set replaceText: string", "@objc get label -> string", "@objc set setLabel: string", "@objc method copyRecord -> new object", "performAction(): void", "rawAction(): void", "reader: FixtureReadable", "@objc method withCompletion: 0:block(string) -> void", "withCompletion(completion: (argument1: string) => void): void;"} {
		if !strings.Contains(record, want) {
			t.Errorf("record missing %q", want)
		}
	}
	if err := readAST(strings.NewReader(`{"kind":"TranslationUnitDecl","inner":[{"kind":"VarDecl","inner":[{"kind":"ObjCBoolLiteralExpr","value":true}]}]}`), func(*node) error { return nil }); err != nil {
		t.Fatal("boolean AST literal rejected", err)
	}
	for _, bad := range []string{"platformOnly():", "takeArray(", "takeCallback(", "variadic(", "narrowInteger():", "narrowFloat():", "release():", "discardSelf():"} {
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
	// A protocol's optional method is an optional member a program's class may implement, never a
	// message promised present; one whose types can't cross is left out with its reason.
	readable := files["foundation/fixture-readable.d.ts"]
	if !strings.Contains(readable, "@objc implement optionalPing -> void\n\t\t */\n\t\toptionalPing?(): void;") || strings.Contains(readable, "@objc method optionalPing") {
		t.Errorf("optional protocol method was promised present, or can't be implemented:\n%s", readable)
	}
	source := files["foundation/fixture-table-source.d.ts"]
	for _, want := range []string{
		"@objc protocol NSFixtureTableSource",
		// Required: a message to an object of Apple's, and a method a class implements.
		"@objc method numberOfRowsInFixtureTable: 0:object -> integer\n\t\t * @objc implement numberOfRowsInFixtureTable: 0:object -> integer\n\t\t */\n\t\tnumberOfRows(table: FixtureRecord): number;",
		// Two methods Swift names fixtureTable(_:...) fold their labels in, positional.
		"fixtureTableTextForRow?(table: FixtureRecord, row: number): string | undefined;",
		"fixtureTableShouldSelectRow?(table: FixtureRecord, row: number): boolean;",
		"Skipped -[NSFixtureTableSource fixtureTable:finish:]: a program's class can't be handed a block() yet.",
		"Skipped -[NSFixtureTableSource copyFixtureTable:]: a result its caller owns",
	} {
		if !strings.Contains(source, want) {
			t.Errorf("table source missing %q:\n%s", want, source)
		}
	}
	// Leaves: a class whose instances, and every subclass's, keep only leaves (NSFixtureText and its
	// mutable subclass keep strings); not one whose subclass keeps a record (NSFixtureNote), nor a
	// timer, which keeps what its initializer is handed, nor the record, whose mutators keep records,
	// nor the root, which may be anything.
	for _, want := range []string{"\nleaf NSFixtureText\n", "\nleaf NSFixtureMutableText\n", "\nleaf NSString\n", "\t-[NSFixtureMutableText appendText:] NSString *: a leaf\n"} {
		if !strings.Contains(string(output.Leaves), want) {
			t.Errorf("leaf table missing %q:\n%s", want, output.Leaves)
		}
	}
	for _, refused := range []string{"leaf NSFixtureNote\n", "leaf NSFixtureStickyNote\n", "leaf NSFixtureRecord\n", "leaf NSFixtureRoot\n", "leaf NSFixtureTimer\n"} {
		if strings.Contains(string(output.Leaves), refused) {
			t.Errorf("leaf table names %q", refused)
		}
	}
	// Deprecated on the platform is left out, with its reason; deprecated only later, or only on
	// another platform, isn't.
	text := files["foundation/fixture-text.d.ts"]
	if !strings.Contains(text, "// Skipped -[NSFixtureText oldWay]: deprecated on macos.") || strings.Contains(text, "oldWay(") || !strings.Contains(text, "futureWay(): void;") || !strings.Contains(text, "phoneWay(): void;") {
		t.Errorf("deprecation misread:\n%s", text)
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
	// The loader reads the fixture's generated modules from the directory the variable names, as it
	// reads a Mac's cache.
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	bindings := filepath.Join(directory, "bindings")
	if err := fixtureOutput(t).Write(bindings); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "run", "./cmd/adamic", "c", "internal/apple/generate/testdata/bindings.a")
	command.Env = append(os.Environ(), "ADAMIC_APPLE_BINDINGS="+bindings)
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
	for _, want := range []string{"objc_msgSend", "NSFixturePanel", "setFrame:display:animate:", "configureMode:style:", "NSFixtureAdd", "adamic_apple_rectangle", "adamic_apple_options", "adamic_apple_string", "NSFixtureCreateText", "objc_release(result);", "showText:", "showCount:", "initWithTitle:", "readText", "adamic_apple_delegate(", "AdamicDelegate_Source", "\"numberOfRowsInFixtureTable:\"", "\"fixtureTable:textForRow:\"", "\"q@:@\"", "\"@@:@q\"", "adamic_apple_give_back"} {
		if !bytes.Contains(content, []byte(want)) {
			t.Errorf("lowered C missing %q", want)
		}
	}
}

// A delegate the compiler can't hand Apple is refused at compile time: one that can reach back to
// the object holding it (a cycle neither count sees), and an object of a class naming no Apple
// protocol, which Apple would only ever see as an object of no methods.
func TestDelegateRefusals(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	bindings := filepath.Join(directory, "bindings")
	if err := fixtureOutput(t).Write(bindings); err != nil {
		t.Fatal(err)
	}
	header := "import { FixtureRecord } from 'apple/foundation/fixture-record';\nimport type { FixtureTableSource } from 'apple/foundation/fixture-table-source';\n"
	for name, test := range map[string]struct{ source, want string }{
		"cycle": {
			header + "class Owner implements FixtureTableSource {\n\treadonly record: FixtureRecord;\n\tconstructor(record: FixtureRecord) {\n\t\tthis.record = record;\n\t}\n\tnumberOfRows(table: FixtureRecord): number {\n\t\treturn table.text.length;\n\t}\n}\nconst record = new FixtureRecord({ count: 1 });\nrecord.source = new Owner(record);\n",
			"refuses Owner, handed to Apple as a delegate the FixtureRecord holds from then on, which can reach back to that FixtureRecord",
		},
		"weak": {
			header + "import type { Weak } from 'adamic';\nclass Owner implements FixtureTableSource {\n\treadonly record: Weak<FixtureRecord>;\n\tconstructor(record: FixtureRecord) {\n\t\tthis.record = record;\n\t}\n\tnumberOfRows(table: FixtureRecord): number {\n\t\treturn table.text.length;\n\t}\n}\nconst record = new FixtureRecord({ count: 1 });\nrecord.source = new Owner(record);\n",
			"",
		},
		"no protocol": {
			header + "class Plain {\n\tnumberOfRows(table: FixtureRecord): number {\n\t\treturn table.text.length;\n\t}\n}\nconst record = new FixtureRecord({ count: 1 });\nrecord.source = new Plain();\n",
			"an object of the program's own class handed to Apple without an Apple protocol it implements",
		},
		// A leaf holds only leaves: a delegate may hold one, and no class of the program's extends one.
		"holds a leaf": {
			header + "import type { FixtureText } from 'apple/foundation/fixture-text';\nclass Owner implements FixtureTableSource {\n\treadonly note: FixtureText;\n\tconstructor(note: FixtureText) {\n\t\tthis.note = note;\n\t}\n\tnumberOfRows(table: FixtureRecord): number {\n\t\treturn this.note.text.length + table.text.length;\n\t}\n}\nexport function attach(record: FixtureRecord, note: FixtureText): void {\n\trecord.source = new Owner(note);\n}\n",
			"",
		},
		// What a class's methods keep counts like its properties: a note's subclass keeps a record
		// (stickTo:), so a note held by the record's delegate may reach back.
		"holds what a subclass keeps a record in": {
			header + "import type { FixtureNote } from 'apple/foundation/fixture-note';\nclass Owner implements FixtureTableSource {\n\treadonly note: FixtureNote;\n\tconstructor(note: FixtureNote) {\n\t\tthis.note = note;\n\t}\n\tnumberOfRows(table: FixtureRecord): number {\n\t\treturn this.note.body.length + table.text.length;\n\t}\n}\nexport function attach(record: FixtureRecord, note: FixtureNote): void {\n\trecord.source = new Owner(note);\n}\n",
			"refuses Owner, handed to Apple as a delegate the FixtureRecord holds",
		},
		"extends a leaf": {
			header + "import { FixtureText } from 'apple/foundation/fixture-text';\nclass Pinned extends FixtureText {\n\trecord: FixtureRecord | undefined = undefined;\n}\nexport function pin(pinned: Pinned): void {\n\tconsole.log(pinned.text);\n}\n",
			"a class extending Apple's FixtureText",
		},
		"a result of the wrong type": {
			header + "class Source implements FixtureTableSource {\n\tnumberOfRows(table: FixtureRecord): string {\n\t\treturn table.text;\n\t}\n}\n",
			"error TS",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			program := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(program, []byte(test.source), 0o644); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("go", "run", "./cmd/adamic", "c", program)
			command.Env = append(os.Environ(), "ADAMIC_APPLE_BINDINGS="+bindings)
			command.Dir = root
			output, runError := command.CombinedOutput()
			if test.want == "" {
				if runError != nil {
					t.Fatalf("refused a delegate it should take: %v\n%s", runError, output)
				}
				return
			}
			if runError == nil || !bytes.Contains(output, []byte(test.want)) {
				t.Fatalf("want a refusal containing %q, got %v:\n%s", test.want, runError, output)
			}
		})
	}
}

// The deprecation reader takes each way Apple's headers write it, for the platform generated.
func TestDeprecatedIn(t *testing.T) {
	t.Parallel()
	g := &generator{configuration: Configuration{Platform: "macos"}}
	for text, want := range map[string]bool{
		`API_DEPRECATED("Use NSURLSession", macos(10.0, 10.4), ios(2.0, 9.0))`:              true,
		`API_DEPRECATED("x", ios(2.0, 9.0))`:                                                false,
		`API_DEPRECATED("x", macos(10.0, API_TO_BE_DEPRECATED))`:                            false,
		`API_DEPRECATED_WITH_REPLACEMENT("newWay", macos(10.0, 11.0))`:                      true,
		`API_DEPRECATED_WITH_REPLACEMENT("newWay", macosx(10.0, 11.0))`:                     true,
		`NS_DEPRECATED_MAC(10_0, 10_4, "x")`:                                                true,
		`NS_DEPRECATED_IOS(2_0, 9_0)`:                                                       false,
		`NS_DEPRECATED(10_0, 10_4, 2_0, 9_0)`:                                               true,
		`NS_DEPRECATED(10_0, NA, 2_0, 9_0)`:                                                 false,
		`NS_CLASS_DEPRECATED_MAC(10_0, 10_4)`:                                               true,
		`AVAILABLE_MAC_OS_X_VERSION_10_0_AND_LATER_BUT_DEPRECATED_IN_MAC_OS_X_VERSION_10_4`: true,
		`DEPRECATED_ATTRIBUTE`:                                                              true,
		`availability(macos, introduced=10.0, deprecated=10.4)`:                             true,
		`availability(macos, introduced=10.0, deprecated=100000)`:                           false,
		`availability(ios, introduced=2.0, deprecated=9.0)`:                                 false,
		`API_AVAILABLE(macos(10.0))`:                                                        false,
	} {
		if got := g.deprecatedIn(text, ""); got != want {
			t.Errorf("deprecatedIn(%s) = %t, want %t", text, got, want)
		}
	}
	for macro, want := range map[string]bool{"API_DEPRECATED_BEGIN": true, "APPKIT_API_DEPRECATED_BEGIN_IOS": false, "APPKIT_API_DEPRECATED_BEGIN_MACOS": true, "API_TO_BE_DEPRECATED_BEGIN": false} {
		if got := g.deprecatedByName(macro); got != want {
			t.Errorf("deprecatedByName(%s) = %t, want %t", macro, got, want)
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
		"foundation/fixture-ambiguous-protocol.d.ts": {" * @objc protocol NSFixtureAmbiguous\n\t */\n\texport interface FixtureAmbiguousProtocol {", "doThing(): void;"},
		"foundation/fixture-element.d.ts":            {"export class FixtureElement", "activate(): void;"},
		"foundation/fixture-holder.d.ts":             {"readonly element: FixtureElementProtocol;"},
		"foundation/fixture-element-protocol.d.ts":   {"export interface FixtureElementProtocol {", "describe(): void;"},
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

// Each pattern here stopped the real Foundation, AppKit or CoreGraphics once, or failed to
// type-check across inheritance; Inheritance.h holds them.
func TestInheritedMembers(t *testing.T) {
	t.Parallel()
	output, err := RunClang(context.Background(), clangPath(t), []string{"-x", "objective-c", "-fobjc-runtime=macosx-10.13", "-fsyntax-only", "-fblocks", "-Werror", "-Wno-nullability-completeness", "-F", "testdata/SDK", "-Xclang", "-ast-dump=json", "testdata/SDK/Foundation.framework/Headers/Inheritance.h"}, fixtureConfiguration(t))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, file := range output.Files {
		files[file.Path] = string(file.Content)
	}
	for path, wants := range map[string][]string{
		"foundation/fixture-base.d.ts": {
			"add(sender: FoundationObject | undefined): void;",
			"addObject(object: FoundationObject): void;",
			"@objc method storeBacking:mask: 0:enum(Retained=0,Buffered=2) 1.mask:options(Down=2,Up=4) -> void",
			"// Skipped -[NSFixtureBase captureOld]: unavailable on macos.",
			"drawOld(): void;",
		},
		"foundation/fixture-advanced.d.ts": {
			"@objc method removeCount: 0:integer -> void\n\t\t */\n\t\tremove(count: number): void;",
			"@objc method removeItem: 0:string -> void\n\t\t */\n\t\tremove(item: string): void;",
			"// Skipped NSFixtureAdvanced.selectedCell: an ancestor declares selectedCell as a method, which stands.",
			"// Skipped NSFixtureAdvanced.title: an ancestor declares title as string, which stands.",
			"// Skipped +[NSFixtureAdvanced shared]: an ancestor declares shared as a property, which stands.",
			"cellAtRow(row: number, options: { readonly column: number }): FixtureCell | undefined;",
		},
		"foundation/foundation-object.d.ts": {
			"// Skipped category NSFixtureInformal: an informal protocol on NSObject, not its own methods.",
		},
	} {
		for _, want := range wants {
			if !strings.Contains(files[path], want) {
				t.Errorf("%s missing %q", path, want)
			}
		}
	}
	_, body, _ := strings.Cut(files["foundation/foundation-object.d.ts"], "declare module")
	if strings.Contains(body, "fixtureDidFinish") || strings.Contains(body, "initialize") {
		t.Errorf("an informal protocol or a runtime-only message was bound:\n%s", files["foundation/foundation-object.d.ts"])
	}
	if !strings.Contains(string(output.Check), "|| __builtin_types_compatible_p(unsigned long long, NSFixtureMask)") {
		t.Error("check file doesn't hold 64-bit option bits")
	}
	if t.Failed() {
		for _, path := range []string{"foundation/fixture-base.d.ts", "foundation/fixture-advanced.d.ts", "foundation/foundation-object.d.ts"} {
			t.Log(path + "\n" + files[path])
		}
	}
}

// Apple moved CGRect, CGPoint and CGSize into CoreFoundation's CFCGTypes.h; they're generated with
// CoreGraphics, where Swift and every caller find them.
func TestMovedHeaders(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	configuration := Configuration{Frameworks: []Framework{{Name: "CoreGraphics", Headers: filepath.Join(root, "CoreGraphics.framework/Headers")}}, Platform: "macos", ReadSource: func(string) ([]byte, error) { return nil, nil }}
	moved, _ := json.Marshal(filepath.Join(root, "CoreFoundation.framework/Headers/CFCGTypes.h"))
	elsewhere, _ := json.Marshal(filepath.Join(root, "CoreFoundation.framework/Headers/CFBase.h"))
	input := `{"kind":"TranslationUnitDecl","inner":[
 {"kind":"RecordDecl","name":"CGRect","tagUsed":"struct","completeDefinition":true,"loc":{"file":MOVED,"offset":1}},
 {"kind":"RecordDecl","name":"CFRange","tagUsed":"struct","completeDefinition":true,"loc":{"file":ELSEWHERE,"offset":1}}
 ]}`
	input = strings.ReplaceAll(strings.ReplaceAll(input, "MOVED", string(moved)), "ELSEWHERE", string(elsewhere))
	output, err := Generate(strings.NewReader(input), configuration)
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{}
	for _, file := range output.Files {
		paths = append(paths, file.Path)
	}
	if !slices.Contains(paths, "coregraphics/rectangle.d.ts") || len(paths) != 1 {
		t.Fatalf("moved header not generated with CoreGraphics alone: %v", paths)
	}
}

// Closures Apple calls, setters for getters that can't cross, and names a subclass keeps: each held
// the real AppKit or Foundation back once. Closures.h holds them.
func TestClosures(t *testing.T) {
	t.Parallel()
	output, err := RunClang(context.Background(), clangPath(t), []string{"-x", "objective-c", "-fobjc-runtime=macosx-10.13", "-fsyntax-only", "-fblocks", "-Werror", "-Wno-nullability-completeness", "-F", "testdata/SDK", "-Xclang", "-ast-dump=json", "testdata/SDK/AppKit.framework/Headers/Closures.h"}, fixtureConfiguration(t))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, file := range output.Files {
		files[file.Path] = string(file.Content)
	}
	session, control := files["appkit/fixture-session.d.ts"], files["appkit/fixture-control.d.ts"]
	for _, want := range []string{
		"@objc method taskWithRequest:completionHandler: 0.with:object 0.completionHandler:block(string?,object?,integer) -> object",
		"readonly completionHandler: (text: string | undefined, request: FixtureRequest | undefined, length: number) => void",
		"// Skipped -[NSFixtureSession sortWith:]: blocks returning long are not carried by the bridge.",
		"@objc get expectedLength -> integer",
		"readonly expectedLength: number;",
		"@objc get completedCount -> integer",
	} {
		if !strings.Contains(session, want) {
			t.Errorf("fixture-session missing %q", want)
		}
	}
	if closure, plain := strings.Index(session, "completionHandler:]"), strings.Index(session, "taskWithRequest:]"); closure < 0 || plain < 0 || closure > plain {
		t.Error("the overload taking the closure isn't written first")
	}
	for _, want := range []string{
		"@objc static controlWithTitle:target:action: 0.title:string 0.action:action -> object\n\t\t */\n\t\tconstructor(options: { readonly title: string; readonly action: () => void });",
		"@objc method setFrame: 0:rectangle -> void\n\t\t */\n\t\tsetFrame(frame: Rectangle): void;",
		"@objc method paintRecord: 0:object -> void\n\t\t */\n\t\tpaintRecord(record: FixtureRecord): void;",
	} {
		if !strings.Contains(control, want) {
			t.Errorf("fixture-control missing %q", want)
		}
	}
	for _, want := range []string{"void (^argument1)(id, id, long)", "id argument1, SEL argument2"} {
		if !strings.Contains(string(output.Check), want) {
			t.Errorf("check file missing %q", want)
		}
	}
	if t.Failed() {
		t.Log(session + "\n" + control)
	}
}
