package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: checker archives and sanitizer/corpus runs share scratch.
func TestWave14ThirdAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE14_THIRD_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_14_third.a")
	binary := h.build(stage0, "third", entry, archive, false)
	oracle := volumeOracle(h, "third-oracle", "oracle_wave_14_third.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	controls := []string{
		"var x = foo; function bar() { x: for(;;) { break x; } }",
		"function bar() { var x = foo; x: for(;;) { break x; } }",
		"function bar(x:number) { x: for(;;) { break x; } }",
		"function bar() { q: for(;;) { break q; } } function foo () { var q = t; }",
		"function bar() { var x = foo; q: for(;;) { break q; } }",
		"function x() {} x: for(;;) { break x; }",
		"class x {} x: for(;;) { break x; }",
		"let x = 1; x: for(;;) { break x; }",
		"const x = 1; x: for(;;) { break x; }",
		"Object: for(;;) { break Object; }",
		"function bar() { x: for(;;) { break x; } var x; }",
		"x: for(;;) { break x; }",
		"{ let x = 1; } x: for(;;) { break x; }",
		"x: for(;;) { let x = 1; break x; }",
		"x: y: for(;;) { break x; }",
		"x: { let x = 1; }",
		"x: { function x() {} }",
		"x: { class x {} }",
		"x: for (let x = 0;;) { break x; }",
		"x: { var x = 1; }",
		"/* 世界 🌍 */\r\nconst café=1;café:{break café;}\r\n",
		"declare const p:string;RegExp(p,'z');new RegExp(p,'ii');RegExp(p,'uv');RegExp(p,'zug');RegExp(p,'zii');",
		"declare const p:string;RegExp(p,'gimy');new RegExp(p);RegExp();new RegExp();RegExp(p,'v');",
		"function local(RegExp:(pattern:string,flags:string)=>unknown){RegExp('[','z');}const p:any={RegExp(pattern:string,flags:string){}};p.RegExp('[','z');",
		"declare const p:string;RegExp(p,'🌍');RegExp(p,'\\n');RegExp(p,\"'\");RegExp(p,'\\x07');",
		`declare const p:string;RegExp(p,'\u0080');RegExp(p,'\u00a0');RegExp(p,'\ue000');RegExp(p,'\u{10ffff}');RegExp(p,'é');`,
		"declare const p:string;declare const flags:string;RegExp(p,flags);RegExp(p,('ii'));RegExp((p),'ii');",
		"const value=/[Á]/u;",
		"const value=/[👍]/;",
		"const value=/[👶🏻]/u;",
		"const value=/[🇯🇵]/u;",
		"const value=/[👨‍👩‍👦]/u;",
		"const value=/[👨‍👩‍👦]/;",
		`const value=/[\u{d83d}\u{dc4d}]/u;`,
		`const value=/[\ud83d\udc4d]/;`,
		`const value=/[\ud83d\udc4d]/u;`,
		"const value=/[👍-￿]/;",
		"const value=/[a-z👍]/;",
		`const value=/[🇯\d🇵]/u;`,
		`const value=/[🇯\q{abc}🇵]/v;`,
		"const value=/[🇯[A]🇵]/v;",
		"const value=/[[🇯🇵]&&[abc]]/v;",
		"const value=/[Á][🇯🇵]/u;",
		`const value=/[👍]\a/;`,
		"const value=/[‌][‍][a]/;",
		"const value=/[a‍b‍c]/u;",
		"const value=/[a‍‍b]/u;",
		"const value=/[́]/u;",
		"const value=/[à́]/u;",
		`const value=/[\n̅]/u;`,
		"/* 世界 🌍 */\r\nconst value=/[Á👍]/;\r\n",
		`declare const p:string;RegExp(p,'\ud800');RegExp(p,'\udc00');RegExp(p,'\ud800\udc00');RegExp(p,'\u{10000}');`,
		`declare const p:string;RegExp(p,'\ud800gii');RegExp(p,'\udc00uv');RegExp(p,'\ud800\ud800');RegExp(p,'\udc00\ud800');`,
		`declare const p:string;RegExp(p,'\ud800🌍\udc00');RegExp(p,'�');RegExp(p,'\ufffd');`,
	}
	var paths []string
	for i, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, id := range []string{"invalidRegexp", "identifierClashWithLabel", "combiningClass", "surrogatePair", "surrogatePairWithoutUFlag", "regionalIndicatorSymbol", "emojiModifier", "zwj", "suggestUnicodeFlag"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+id+"\t")) {
			t.Fatal("missing positive", id)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "third-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	constructorSources := []string{
		"new RegExp('[Á]');",
		"new RegExp('[👍]');new RegExp('[👶🏻]','u');new RegExp('[🇯🇵]','u');new RegExp('[👨‍👩‍👦]','u');",
		`new RegExp('[\\u{d83d}\\u{dc4d}]','u');`,
		`new RegExp('[\u0041\u0301]','u');new RegExp('[\x41\u0301]');`,
		`new RegExp('[\uD83D\uDC4D]');new RegExp('[\\uD83D\\uDC4D]');`,
		`new RegExp('[\u{1F44D}]');new RegExp('[\\u{41}\\u{301}]','u');`,
		"/* 世界 🌍 */\r\nnew RegExp('é👍[Á]');\r\n",
		"new RegExp(`[👍]`);new RegExp(`\n\t[❇️]`);",
		`const flags='u';new RegExp('[🇯🇵]',flags);`,
		`const pattern='[Á][Á]';new RegExp(pattern,'u');`,
		`const a='[A';const b='́]';const p=a+b;new RegExp((p));`,
		"new RegExp('[A'+'́]'+'[B'+'̀]');",
		"new RegExp(`[${'A'}́]`);new RegExp(`[${1}́]`);",
		`let p='[Á]';new RegExp(p);`,
		`let p='[Á]';p='abc';new RegExp(p);`,
		`let p='[Á]';p+='abc';new RegExp(p);`,
		`let p='[Á]';({p}={p:'abc'});new RegExp(p);`,
		`let p='[Á]';[...p]=[];new RegExp(p);`,
		`let p='[Á]';for(p of []){}new RegExp(p);`,
		`let p='[Á]';function f(){let p='abc';p='x';}new RegExp(p);`,
		`const p='[Á]';const value={p: 'abc'};new RegExp(p);`,
		`const r=/[👍]/u;new RegExp(r);`,
		`new RegExp(/[👍]/u,'');new RegExp(/[👍]/,'u');`,
		`declare const flags:string;new RegExp(/[👍]/,flags);new RegExp('[Á]',flags);`,
		`new RegExp(/[👍]/);new RegExp(/[Á]/u);`,
		`new RegExp('[Á] [');new RegExp('{ [Á]','u');new RegExp('[abc]');`,
		`new RegExp('[Á]','u'+'');new RegExp('[👍]',('u'));`,
		`const p=p+'[Á]';new RegExp(p);`,
		`function f(RegExp:any){new RegExp('[Á]');}new RegExp('[B̀]');`,
		`new RegExp('[🇯[A]🇵]','v');new RegExp('[[🇯🇵]&&[abc]]','v');`,
		`new RegExp('[\ud800\u0301]','u');new RegExp('[\udc00\u0301]','u');`,
		"new RegExp('[A\\\ń]');new RegExp('[Á]\\\n');",
		`const R=RegExp;new R('[Á]');`,
		`let R=RegExp;R=()=>0;new R('[Á]');`,
		`let R;R=RegExp;new R('[Á]');`,
		`let R;R ||= RegExp;new R('[Á]');`,
		`const R=RegExp;const S=R;new S('[Á]');`,
		`const R=RegExp;R('[👍]','u');`,
		`globalThis.RegExp('[Á]');new globalThis.RegExp('[👍]');`,
		`globalThis['Reg'+'Exp']('[Á]');`,
		`const g=globalThis;new g.RegExp('[Á]');`,
		`const {RegExp:R}=globalThis;new R('[Á]');`,
		`const {RegExp}=globalThis;new RegExp('[Á]');`,
		`let R;({RegExp:R}=globalThis);new R('[Á]');`,
		`function f(R=RegExp){new R('[Á]');}`,
		`let R;({R=RegExp}={});new R('[Á]');`,
		`let R;({RegExp:R=()=>0}=globalThis);new R('[Á]');`,
		`const {RegExp:R=RegExp}=globalThis;new R('[Á]');`,
		`const R=(true?RegExp:RegExp);new R('[Á]');`,
		`new (true&&RegExp)('[Á]');new (false||RegExp)('[Á]');`,
		`new (RegExp as any)('[Á]');new (RegExp!)('[Á]');`,
		`new (0,RegExp)('[Á]');`,
		`RegExp=(()=>0) as any;new RegExp('[Á]');`,
		`globalThis=({}) as any;new globalThis.RegExp('[Á]');`,
		`globalThis.RegExp=(()=>0) as any;new globalThis.RegExp('[Á]');`,
		`const object={RegExp};object.RegExp('[Á]');`,
		`function f(RegExp:any){new RegExp('[Á]');}const R=RegExp;new R('[B̀]');`,
		`const key='RegExp';globalThis[key]('[Á]');`,
		`const R=RegExp;let S=R;S=S;new S('[Á]');`,
		`const R=RegExp;function f(R:any){new R('[Á]');}new R('[B̀]');`,
		`const {RegExp:R,...rest}=globalThis;new R('[Á]');rest.RegExp('[Á]');`,
		`const [R]=[RegExp];new R('[Á]');`,
	}
	var constructorPaths []string
	for i, source := range constructorSources {
		constructorPaths = append(constructorPaths, h.write(fmt.Sprintf("constructor-%03d.a", i), source+"\nexport {};\n"))
	}
	constructorManifest := h.write("constructors.manifest", strings.Join(constructorPaths, "\n")+"\n")
	constructorTruth := h.must("constructors-go", exec.Command(oracle, config, constructorManifest, "--class-only"))
	compareConstructors := func(name string, program string) {
		got := h.must(name, exec.Command(program, config, constructorManifest, "--class-only"))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, constructorTruth.stdout) {
			t.Fatalf("constructor bytes differ at %d: %s", firstDifference(got.stdout, constructorTruth.stdout), got.stderr)
		}
		t.Logf("%s: %d identical finding bytes; %s", name, len(got.stdout), strings.TrimSpace(string(got.stdout[bytes.LastIndex(bytes.TrimSpace(got.stdout), []byte("\n"))+1:])))
	}
	compareConstructors("constructors-native", binary)
	compareConstructors("constructors-asan", asan)
	for _, m := range []struct{ name, file, from, to string }{
		{"cooked-mapping", "regexp_cooked_offsets.a", "byteOffsets.push(source);\n            }", "byteOffsets.push(source + 1);\n            }"},
		{"constant-write", "regexp_constant.a", "this.writes(occurrence)", "false"},
		{"call-flags", "no_misleading_character_class.a", "this.unicode = hasFlags && (flags.text.includes('u') || flags.text.includes('v'));", "this.unicode = false;"},
		{"call-literal", "no_misleading_character_class.a", "checked.push(patternIndex);", "checked.push(-1);"},
		{"constant-dedup", "no_misleading_character_class.a", "reported.includes(item.id)", "false"},
		{"tracker-alias", "regexp_reference_tracker.a", "this.followTarget(new Shadow(this.rules).name(parentIndex), root);", "this.followTarget(new Shadow(this.rules).name(parentIndex), true);"},
		{"tracker-write", "regexp_reference_tracker.a", "global && this.constants.writes(index)", "false"},
	} {
		unused := wave14NextMutant(h, stage0, archive, m.name, m.file, m.from, m.to)
		os.Remove(unused)
		mutant := h.build(stage0, m.name+"-third", filepath.Join(directory, m.name+"-source/wave_14_third.a"), archive, false)
		got := h.must(m.name+"-run", exec.Command(mutant, config, constructorManifest, "--class-only"))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, constructorTruth.stdout) {
			t.Fatal("constructor mutant survived", m.name)
		}
		t.Logf("%s mutant exits 0 with empty stderr; byte oracle catches byte %d", m.name, firstDifference(got.stdout, constructorTruth.stdout))
		os.Remove(mutant)
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"label", "no_label_var.a", "!new ScopeSymbols(this.rules, at).names.includes(label)", "new ScopeSymbols(this.rules, at).names.includes(label)"},
		{"flags", "no_invalid_regexp.a", "if('dgimsuvy'.includes(flag)) {", "if('never'.includes(flag)) {"},
		{"unicode-quote", "no_invalid_regexp.a", "code >= 128 && !this.printable(code)", "code >= 128 && this.printable(code)"},
		{"surrogate-decoding", "no_invalid_regexp.a", `decoded += '\ufffd\ufffd\ufffd';`, `decoded += '\ufffd';`},
		{"class", "no_misleading_character_class.a", "this.combining(current.value) && !this.combining(previous.value)", "this.combining(previous.value) && !this.combining(current.value)"},
	} {
		unused := wave14NextMutant(h, stage0, archive, m.name, m.file, m.from, m.to)
		os.Remove(unused)
		mutant := h.build(stage0, m.name+"-third", filepath.Join(directory, m.name+"-source/wave_14_third.a"), archive, false)
		got := h.must(m.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", m.name)
		}
		t.Logf("%s mutant exits 0 with empty stderr; full byte oracle catches byte %d", m.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	scopeOverlay := h.overlay("scope-meaning", "bridge/tsgo/checker/scope_symbols.go", "c.GetSymbolsInScope(node, ast.SymbolFlagsValue)", "c.GetSymbolsInScope(node, ast.SymbolFlagsVariable)")
	scopeArchive := h.archive("scope-meaning", scopeOverlay, false)
	scopeMutant := h.build(stage0, "scope-meaning", entry, scopeArchive, false)
	scopeResult := h.must("scope-meaning-run", exec.Command(scopeMutant, config, manifest))
	if len(scopeResult.stderr) != 0 || bytes.Equal(scopeResult.stdout, truth.stdout) {
		t.Fatal("scope meaning mutant survived")
	}
	t.Logf("scope meaning mutant exits 0 with empty stderr; full byte oracle catches byte %d", firstDifference(scopeResult.stdout, truth.stdout))
	os.Remove(scopeArchive)
	os.Remove(scopeMutant)
	for _, corpus := range []struct{ name, config, manifest string }{
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST")},
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE14_NEXT_REPOSITORY_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
	for _, gap := range []struct{ name, source, id, reason string }{
		{"label-parser", "undefined: for(;;) {break undefined;}", "identifierClashWithLabel", "parser slice expected semicolon"},
		{"pattern", "new RegExp('[');", "invalidRegexp", "native ECMAScript pattern validation"},
	} {
		witness := h.write(gap.name+"-witness.a", gap.source+"\nexport {};\n")
		roots := h.write(gap.name+"-witness.manifest", witness+"\n")
		want := h.must(gap.name+"-go-positive", exec.Command(oracle, config, roots))
		if !bytes.Contains(want.stdout, []byte("\t"+gap.id+"\t")) {
			t.Fatal("gap lacks positive Go control", gap.name)
		}
		args := []string{config, roots}
		if gap.name == "constructor" {
			args = append(args, "--class-only")
		}
		got := h.run(gap.name+"-native-refusal", exec.Command(binary, args...))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte(gap.reason)) {
			t.Fatalf("gap changed: %v %s", got.err, got.stderr)
		}
		t.Logf("%s: Go positive %s; native explicit NotYet panic 70", gap.name, gap.id)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,3,'LabeledStatement','scope-symbols'));
`)
	probe := h.write("probe.a", "x:;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released handle panics 70; retaining it exits 0 and is caught")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
