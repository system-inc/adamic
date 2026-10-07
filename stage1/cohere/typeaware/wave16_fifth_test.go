package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var wave16FifthRuleNames = []string{"no-throw-literal", "no-useless-backreference", "prefer-arrow-callback"}

func wave16FifthControls() []string {
	return []string{
		"throw 'x';throw 1;throw undefined;throw new Error();throw (new Error());throw (undefined);throw true;throw null;throw {};throw [];throw ()=>{};throw class{};",
		"declare let x:any;throw x&&'bad';throw 'bad'&&new Error();throw x||'bad';throw x?'bad':'bad';throw x?'bad':new Error();throw x=1;throw x=new Error();throw x+=new Error();throw x&&=new Error();throw x??='bad';throw new Error(),1;",
		"function f(undefined:any){throw undefined;}async function g(){throw await 1;}function* gen(){throw yield 1;}",
		"declare const foo:any;foo(function bar(){bar();});foo(function bar(){const bar=1;bar;});foo(function(){function inner(){arguments;}});foo(function(){arguments;});foo(function(arguments:any){arguments;});",
		"declare const foo:any;foo(function /*keep*/ named(a:number):number {return a;});foo(function(){return this;}.bind(this));foo((function(){return this;}).bind(this));foo((function(){return this;}?.bind)(this));foo(function(){return this;}.bind(/*keep*/this));",
		"declare const foo:any;foo(function(a,a){});foo(function(this:any){});foo(async function\n(){});new foo(function(){});foo(function*(){});foo(function(){new.target;});",
		"const R=RegExp;R('\\\\1(a)');let alias;alias=RegExp;new alias('\\\\1(a)');globalThis.RegExp('\\\\1(a)');const {RegExp:r}=globalThis;r('\\\\1(a)');",
		"const pattern='\\\\1(a)';let flags='u';new RegExp(pattern,flags);let changed='\\\\1(a)';changed='x';RegExp(changed);function f(RegExp:any){RegExp('\\\\1(a)');}",
		"RegExp('\\\\1(a){','u');declare const flags:string;RegExp('\\\\1(a){',flags);RegExp('\\\\k<a>(?<a>x)\\\\k<b>');RegExp('[[]\\\\1](a)','v');",
		"/* 世界 🌍 */declare const foo:any;foo(function(){return 1;});throw 'x';/\\1(a)/;",
		"throw new Error();",
		"throw new Error('error');",
		"throw Error('error');",
		"var e = new Error(); throw e;",
		"function foo(undefined) { throw undefined; }",
		"try {throw new Error();} catch (e) {throw e;};",
		"throw a;",
		"throw foo();",
		"throw new foo();",
		"throw foo.bar;",
		"throw foo[bar];",
		"class C { #field; foo() { throw foo.#field; } }",
		"throw foo = new Error();",
		"throw foo.bar ||= 'literal'",
		"throw foo[bar] ??= 'literal'",
		"throw 1, 2, new Error();",
		"throw 'literal' && new Error();",
		"throw new Error() || 'literal';",
		"throw foo ? new Error() : 'literal';",
		"throw foo ? 'literal' : new Error();",
		"throw tag `${foo}`;",
		"function* foo() { var index = 0; throw yield index++; }",
		"async function foo() { throw await bar; }",
		"throw obj?.foo",
		"throw obj?.foo()",
		"throw 'error';",
		"throw 0;",
		"throw false;",
		"throw null;",
		"throw {};",
		"throw undefined;",
		"throw 'a' + 'b';",
		"var b = new Error(); throw 'a' + b;",
		"throw foo = 'error';",
		"throw foo += new Error();",
		"throw foo &= new Error();",
		"throw foo &&= 'literal'",
		"throw new Error(), 1, 2, 3;",
		"throw 'literal' && 'not an Error';",
		"throw foo && 'literal'",
		"throw foo ? 'not an Error' : 'literal';",
		"throw `${err}`;",
		"regExp('\\\\1(a)')",
		"new Regexp('\\\\1(a)', 'u')",
		"RegExp.foo('\\\\1(a)', 'u')",
		"new foo.RegExp('\\\\1(a)')",
		"RegExp(p)",
		"new RegExp(p, 'u')",
		"RegExp('\\\\1(a)' + suffix)",
		"new RegExp(`${prefix}\\\\1(a)`)",
		"let RegExp; new RegExp('\\\\1(a)');",
		"function foo() { var RegExp; RegExp('\\\\1(a)', 'u'); }",
		"function foo(RegExp) { new RegExp('\\\\1(a)'); }",
		"if (foo) { const RegExp = bar; RegExp('\\\\1(a)'); }",
		"/* globals RegExp:off */ new RegExp('\\\\1(a)');",
		"RegExp('\\\\1(a)');",
		"/(?:)/",
		"/(?:a)/",
		"new RegExp('')",
		"RegExp('(?:a)|(?:b)*')",
		"/^ab|[cd].\\n$/",
		"/(a)/",
		"RegExp('(a)|(b)')",
		"new RegExp('\\\\n\\\\d(a)')",
		"/\\0(a)/",
		"/\\0(a)/u",
		"/(?<=(a))(b)(?=(c))/",
		"/(?<!(a))(b)(?!(c))/",
		"/(?<foo>a)/",
		"RegExp('\\\\\\\\1(a)')",
		"/\\\\1(a)/",
		"/\\1/",
		"/^\\1$/",
		"/\\2(a)/",
		"/\\1(?:a)/",
		"/\\1(?=a)/",
		"/\\1(?!a)/",
		"/^[\\1](a)$/",
		"new RegExp('[\\\\1](a)')",
		"/\\11(a)/",
		"/\\k<foo>(a)/",
		"/^(a)\\1\\2$/",
		"/(a)\\1/",
		"/(a).\\1/",
		"RegExp('(a)\\\\1(b)')",
		"/(a)(b)\\2(c)/",
		"/(?<foo>a)\\k<foo>/",
		"new RegExp('(.)\\\\1')",
		"RegExp('(a)\\\\1(?:b)')",
		"/(a)b\\1/",
		"/((a)\\2)/",
		"/((a)b\\2c)/",
		"/^(?:(a)\\1)$/",
		"/^((a)\\2)$/",
		"/^(((a)\\3))|b$/",
		"/a(?<foo>(.)b\\2)/",
		"/(a)?(b)*(\\1)(c)/",
		"/(a)?(b)*(\\2)(c)/",
		"/(?<=(a))b\\1/",
		"/(?<=(?=(a)\\1))b/",
		"/(?<!\\1(a))b/",
		"/(?<=\\1(a))b/",
		"/(?<!\\1.(a))b/",
		"/(?<=\\1.(a))b/",
		"/(?<=(?:\\1.(a)))b/",
		"/(?<!(?:\\1)((a)))b/",
		"/(?<!(?:\\2)((a)))b/",
		"/(?=(?<=\\1(a)))b/",
		"/(?=(?<!\\1(a)))b/",
		"/(.)(?<=\\2(a))b/",
		"/^(a)\\1|b/",
		"/^a|(b)\\1/",
		"/^a|(b|c)\\1/",
		"/^(a)|(b)\\2/",
		"/^(?:(a)|(b)\\2)$/",
		"/^a|(?:.|(b)\\1)/",
		"/^a|(?:.|(b).(\\1))/",
		"/^a|(?:.|(?:(b)).(\\1))/",
		"/^a|(?:.|(?:(b)|c).(\\1))/",
		"/^a|(?:.|(?:(b)).(\\1|c))/",
		"/^a|(?:.|(?:(b)|c).(\\1|d))/",
		"/.(?=(b))\\1/",
		"/.(?<=(b))\\1/",
		"/a(?!(b)\\1)./",
		"/a(?<!\\1(b))./",
		"/a(?!(b)(\\1))./",
		"/a(?!(?:(b)\\1))./",
		"/a(?!(?:(b))\\1)./",
		"/a(?<!(?:\\1)(b))./",
		"/a(?<!(?:(?:\\1)(b)))./",
		"/(?<!(a))(b)(?!(c))\\2/",
		"/a(?!(b|c)\\1)./",
		"RegExp('\\\\1(a)[')",
		"new RegExp('\\\\1(a){', 'u')",
		"new RegExp('\\\\1(a)\\\\2', 'ug')",
		"const flags = 'gus'; RegExp('\\\\1(a){', flags);",
		"RegExp('\\\\1(a)\\\\k<foo>', 'u')",
		"new RegExp('\\\\k<foo>(?<foo>a)\\\\k<bar>')",
		"new RegExp('([[A--B]])\\\\1', 'v')",
		"new RegExp('[[]\\\\1](a)', 'v')",
		"/((?<foo>bar)\\k<foo>|(?<foo>baz))/",
		"/(b)(\\2a)/",
		"/\\k<foo>(?<foo>bar)/",
		"RegExp('(a|bc)|\\\\1')",
		"new RegExp('(?!(?<foo>\\\\n))\\\\1')",
		"/(?<!(a)\\1)b/",
		"new RegExp('(\\\\1)')",
		"/^(a\\1)$/",
		"/^((a)\\1)$/",
		"new RegExp('^(a\\\\1b)$')",
		"RegExp('^((\\\\1))$')",
		"/((\\2))/",
		"/a(?<foo>(.)b\\1)/",
		"/a(?<foo>\\k<foo>)b/",
		"/^(\\1)*$/",
		"/^(?:a)(?:((?:\\1)))*$/",
		"/(?!(\\1))/",
		"/a|(b\\1c)/",
		"/(a|(\\1))/",
		"/(a|(\\2))/",
		"/(?:a|(\\1))/",
		"/(a)?(b)*(\\3)/",
		"/(?<=(a\\1))b/",
		"/\\1(a)/",
		"/\\1.(a)/",
		"/(?:\\1)(?:(a))/",
		"/(?:\\1)(?:((a)))/",
		"/(?:\\2)(?:((a)))/",
		"/(?:\\1)(?:((?:a)))/",
		"/(\\2)(a)/",
		"RegExp('(a)\\\\2(b)')",
		"/(?:a)(b)\\2(c)/",
		"/\\k<foo>(?<foo>a)/",
		"/(?:a(b)\\2)(c)/",
		"new RegExp('(a)(b)\\\\3(c)')",
		"/\\1(?<=(a))./",
		"/\\1(?<!(a))./",
		"/(?<=\\1)(?<=(a))/",
		"/(?<!\\1)(?<!(a))/",
		"/(?=\\1(a))./",
		"/(?!\\1(a))./",
		"/(?<=(a)\\1)b/",
		"/(?<!.(a).\\1.)b/",
		"/(.)(?<!(b|c)\\2)d/",
		"/(?<=(?:(a)\\1))b/",
		"/(?<=(?:(a))\\1)b/",
		"/(?<=(a)(?:\\1))b/",
		"/(?<!(?:(a))(?:\\1))b/",
		"/(?<!(?:(a))(?:\\1)|.)b/",
		"/.(?!(?<!(a)\\1))./",
		"/.(?=(?<!(a)\\1))./",
		"/.(?!(?<=(a)\\1))./",
		"/.(?=(?<=(a)\\1))./",
		"/(a)|\\1b/",
		"/^(?:(a)|\\1b)$/",
		"/^(?:(a)|b(?:c|\\1))$/",
		"/^(?:a|b(?:(c)|\\1))$/",
		"/^(?:(a(?!b))|\\1b)+$/",
		"/^(?:(?:(a)(?!b))|\\1b)+$/",
		"/^(?:(a(?=a))|\\1b)+$/",
		"/^(?:(a)(?=a)|\\1b)+$/",
		"/.(?:a|(b)).|(?:(\\1)|c)./",
		"/.(?!(a)|\\1)./",
		"/.(?<=\\1|(a))./",
		"/a(?!(b)).\\1/",
		"/(?<!(a))b\\1/",
		"/(?<!(a))(?:\\1)/",
		"/.(?<!a|(b)).\\1/",
		"/.(?!(a)).(?!\\1)./",
		"/.(?<!(a)).(?<!\\1)./",
		"/.(?=(?!(a))\\1)./",
		"/.(?<!\\1(?!(a)))/",
		"/\\1(a)(b)\\2/",
		"/\\1(a)\\1/",
		"/\\1(a)\\2(b)/",
		"/\\1.(?<=(a)\\1)/",
		"/(?!\\1(a)).\\1/",
		"/(a)\\2(b)/; RegExp('(\\\\1)');",
		"RegExp('\\\\1(a){', flags);",
		"const r = RegExp, p = '\\\\1', s = '(a)'; new r(p + s);",
		"new RegExp('\\\\1([[A--B]])', 'v')",
		"/\\k<foo>((?<foo>bar)|(?<foo>baz))/",
		"/((?<foo>bar)|\\k<foo>(?<foo>baz))/",
		"/\\k<foo>((?<foo>bar)|(?<foo>baz)|(?<foo>qux))/",
		"/((?<foo>bar)|\\k<foo>(?<foo>baz)|(?<foo>qux))/",
		"/((?<foo>bar)|\\k<foo>|(?<foo>baz))/",
		"/((?<foo>bar)|\\k<foo>|(?<foo>baz)|(?<foo>qux))/",
		"/((?<foo>bar)|(?<foo>baz\\k<foo>)|(?<foo>qux\\k<foo>))/",
		"/(?<=((?<foo>bar)|(?<foo>baz))\\k<foo>)/",
		"/((?!(?<foo>bar))|(?!(?<foo>baz)))\\k<foo>/",
		"foo(a => a);",
		"foo(function*() {});",
		"foo(function() { this; });",
		"foo(function() { (() => this); });",
		"foo(function() { this; }.bind(obj));",
		"foo(function() { this; }.call(this));",
		"foo(a => { (function() {}); });",
		"var foo = function foo() {};",
		"(function foo() {})();",
		"foo(function bar() { bar; });",
		"foo(function bar() { arguments; });",
		"foo(function bar() { arguments; }.bind(this));",
		"foo(function bar() { new.target; });",
		"foo(function bar() { new.target; }.bind(this));",
		"foo(function bar() { this; }.bind(this, somethingElse));",
		"foo((function() {}).bind.bar)",
		"foo((function() { this.bar(); }).bind(obj).bind(this))",
		"foo(function bar() {});",
		"foo(function() {});",
		"foo(nativeCb || function() {});",
		"foo(bar ? function() {} : function() {});",
		"foo(function() { (function() { this; }); });",
		"foo(function() { this; }.bind(this));",
		"foo(bar || function() { this; }.bind(this));",
		"foo(function() { (() => this); }.bind(this));",
		"foo(function bar(a) { a; });",
		"foo(function(a) { a; });",
		"foo(function(arguments) { arguments; });",
		"qux(function(foo, bar, baz) { return foo * 2; })",
		"qux(function(foo, bar, baz) { return foo * bar; }.bind(this))",
		"qux(function(foo, bar, baz) { return foo * this.qux; }.bind(this))",
		"foo(function() {}.bind(this, somethingElse))",
		"qux(function(foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) { return foo + bar; });",
		"qux(function(baz, baz) { })",
		"qux(function( /* no params */ ) { })",
		"qux(function( /* a */ foo /* b */ , /* c */ bar /* d */ , /* e */ baz /* f */ ) { return foo; })",
		"qux(async function (foo = 1, bar = 2, baz = 3) { return baz; })",
		"qux(async function (foo = 1, bar = 2, baz = 3) { return this; }.bind(this))",
		"foo(async function /*\n*/ () { return 1; });",
		"foo(async function // c\n () { return 1; });",
		"foo(async function\n () { return 1; });",
		"foo(async function /* c */ () { return 1; });",
		"foo((bar || function() {}).bind(this))",
		"foo(function() {}.bind(this).bind(obj))",
		"foo?.(function() {});",
		"foo?.(function() { return this; }.bind(this));",
		"foo(function() { return this; }?.bind(this));",
		"foo((function() { return this; }?.bind)(this));",
		"\n            test(\n                function ()\n                { }\n            );\n            ",
		"\n            test(\n                function (\n                    ...args\n                ) /* Lorem ipsum\n                dolor sit amet. */ {\n                    return args;\n                }\n            );\n            ",
		"foo(a => a);",
		"foo((a:string) => a);",
		"foo(function*() {});",
		"foo(function() { this; });",
		"foo(function() { (() => this); });",
		"foo(function() { this; }.bind(obj));",
		"foo(function() { this; }.call(this));",
		"foo(a => { (function() {}); });",
		"var foo = function foo() {};",
		"(function foo() {})();",
		"foo(function bar() { bar; });",
		"foo(function bar() { arguments; });",
		"foo(function bar() { arguments; }.bind(this));",
		"foo(function bar() { new.target; });",
		"foo(function bar() { new.target; }.bind(this));",
		"foo(function bar() { this; }.bind(this, somethingElse));",
		"foo((function() {}).bind.bar)",
		"foo((function() { this.bar(); }).bind(obj).bind(this))",
		"test('clean', function (this: any) { this.foo = 'Cleaned!';});",
		"obj.test('clean', function (foo) { this.foo = 'Cleaned!'; });",
		"foo(function bar() {});",
		"foo(function() {});",
		"foo(nativeCb || function() {});",
		"foo(bar ? function() {} : function() {});",
		"foo(function() { (function() { this; }); });",
		"foo(function() { this; }.bind(this));",
		"foo(bar || function() { this; }.bind(this));",
		"foo(function() { (() => this); }.bind(this));",
		"foo(function bar(a:string) { a; });",
		"foo(function(a:any) { a; });",
		"foo(function(arguments:any) { arguments; });",
		"qux(function(foo:string, bar:number, baz:string) { return foo * 2; })",
		"qux(function(foo:number, bar:number, baz:number) { return foo * bar; }.bind(this))",
		"qux(function(foo:any, bar:any, baz:any) { return foo * this.qux; }.bind(this))",
		"foo(function() {}.bind(this, somethingElse))",
		"qux(function(foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) { return foo + bar; });",
		"qux(function(baz:string, baz:string) { })",
		"qux(function( /* no params */ ) { })",
		"qux(function( /* a */ foo:string /* b */ , /* c */ bar:string /* d */ , /* e */ baz:string /* f */ ) { return foo; })",
		"qux(async function (foo:number = 1, bar:number = 2, baz:number = 3) { return baz; })",
		"qux(async function (foo:number = 1, bar:number = 2, baz:number = 3) { return this; }.bind(this))",
		"foo((bar || function() {}).bind(this))",
		"foo(function() {}.bind(this).bind(obj))",
		"foo?.(function() {});",
		"foo?.(function() { return this; }.bind(this));",
		"foo(function() { return this; }?.bind(this));",
		"foo((function() { return this; }?.bind)(this));",
		"\n            test(\n                function ()\n                { }\n            );\n            ",
		"\n            test(\n                function (\n                    ...args\n                ) /* Lorem ipsum\n                dolor sit amet. */ {\n                    return args;\n                }\n            );\n            ",
		"foo(function():string { return 'foo' });",
		"test('foo', function (this: any) {});",
		"RegExp(/\\1(a)/,'');RegExp('\\\\1(a){','u'+missing);RegExp('\\\\1(a){',`${missing}`);RegExp('\\\\1(a');RegExp('\\\\1(a))');RegExp('(a)\\\\1\\\\');RegExp('(a)\\\\1[');RegExp('\\\\1(a){,}');RegExp('\\\\1(a){,}','u');",
		"let alias;alias=RegExp;alias=1;alias('\\\\1(a)');const wrap=(RegExp as any);wrap('\\\\1(a)');(false?RegExp:RegExp)('\\\\1(a)');",
		"RegExp=other;RegExp('\\\\1(a)');globalThis.RegExp('\\\\1(a)');",
		"window=other;window.RegExp('\\\\1(a)');globalThis['Reg'+'Exp']('\\\\1(a)');const root=globalThis;const {RegExp:r=other}=root;new r('\\\\1(a)');",
		"const p=p+'x';RegExp(p);let p2='\\\\1(a)';p2++;RegExp(p2);let p3='\\\\1(a)';function f(p3:any){p3='x';}RegExp(p3);",
		"new RegExp('\\u{1f30d}\\\\1(🌍)','u');new RegExp('[\\q{]x}]\\\\1(a)','v');new RegExp('\\x28\\\\1(a)');",
	}
}

func wave16FifthSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = regexp.MustCompile(`'\./([^']+\.ts)'`).ReplaceAllString(source, "'"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/$1'")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave16_fifth_suite.a"), archive, false)
}

// All output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestWave16FifthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE16_FIFTH_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave16_fifth_suite.a")
	binary := h.build(stage0, "wave16", entry, archive, false)
	oracle := volumeOracle(h, "wave16-oracle", "oracle_wave16_fifth.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler","jsx":"preserve","lib":["ES2022","DOM"]},"files":["node.d.a"]}`)
	h.write("node.d.a", "declare module 'node:child_process' {\n export function exec(command:string,...args:unknown[]):unknown;\n export namespace exec {export function __promisify__(command:string):unknown;}\n export function execSync(command:string,...args:unknown[]):unknown;\n export function spawn(command:string,...args:unknown[]):unknown;\n export function spawnSync(command:string,...args:unknown[]):unknown;\n export function execFile(command:string,...args:unknown[]):unknown;\n export function execFileSync(command:string,...args:unknown[]):unknown;\n}\ndeclare module 'node:util' {export function promisify(f:typeof import('node:child_process').exec):typeof import('node:child_process').exec.__promisify__;}\n")
	paths := []string{filepath.Join(directory, "node.d.a")}
	for i, source := range wave16FifthControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.a", "export let value=1; export const object={x:1}; export default object;\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range wave16FifthRuleNames {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave16-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, profile := range []struct {
		name  string
		flags []string
	}{
		{"allow-named", []string{"--allow-named"}},
		{"bound-this-only", []string{"--bound-this-only"}},
		{"named-bound", []string{"--allow-named", "--bound-this-only"}},
	} {
		arguments := append([]string{config, manifest}, profile.flags...)
		truth := h.must(profile.name+"-go", exec.Command(oracle, arguments...))
		for _, runner := range []struct{ name, binary string }{{"native", binary}, {"asan", asan}} {
			got := h.must(profile.name+"-"+runner.name, exec.Command(runner.binary, arguments...))
			if len(got.stderr) != 0 || !bytes.Equal(got.stdout, truth.stdout) {
				t.Fatalf("profile %s/%s differs at byte %d", profile.name, runner.name, firstDifference(got.stdout, truth.stdout))
			}
		}
		t.Logf("profile %s normal and sanitized: %s", profile.name, summary(truth.stdout))
	}

	// Each rule mutant changes a real judgment or report range, finishes normally,
	// and is killed exclusively by the independent cohere diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"throw-range", "no_throw_literal.a", "this.rules.byte(node.end)", "this.rules.byte(node.end) + 1"},
		{"backreference-range", "no_useless_backreference.a", "this.rules.byte(node.end)", "this.rules.byte(node.end) + 1"},
		{"arrow-fix", "prefer_arrow_callback.a", "' =>'", "' => '"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := wave16FifthSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatalf("%s mutant survived", change.name)
			}
			t.Logf("%s: exit 0, Go byte oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
			if err := os.Remove(mutant); err != nil {
				t.Fatal(err)
			}
		})
	}
	// The existing fact API is independently held by its direct checker tests.
	if manifest := os.Getenv("ADAMIC_WAVE16_REPOSITORY_MANIFEST"); manifest != "" {
		h.compare("repository", oracle, binary, filepath.Join(repository, "tsconfig.json"), manifest)
		h.compare("repository-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if manifest := os.Getenv("ADAMIC_WAVE16_COMPILER_MANIFEST"); manifest != "" {
		corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if corpus == "" {
			t.Fatal("compiler source is required")
		}
		h.compare("compiler", oracle, binary, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
		h.compare("compiler-asan", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','node-symbol-details'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released program queried using the existing fact API: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by the required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
