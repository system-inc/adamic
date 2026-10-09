package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * vars, caughtErrors and destructuredArrayIgnorePattern, each replayed through upstream (#6esg2nx).
 *
 * cohere decoded all three and read none of them, so a config setting one changed nothing. Every row
 * below was run through the installed `@typescript-eslint/no-unused-vars` 8.71.0 twice, once with its
 * options and once with the three removed, and both columns are asserted, so an option has to silence
 * exactly what it silences upstream and nothing when it is off.
 *
 * The first 25 rows are upstream's own: every row in its v8.71.0 test files that sets one of the
 * three, extracted by transpiling the files and stubbing RuleTester. Those files run under
 * `sourceType: "script"`, as cohere reads a file with no import or export. Three of them also set
 * `reportUsedIgnorePattern`. Here the key is dropped from both columns, and upstream's answer without
 * it is what they assert; with it, they are replayed by TestNoUnusedVarsOptionsUpstreamCorpus. The rest are edge
 * rows written for this port, each replayed under the source type cohere gives it: a script unless
 * it imports or exports. Findings in their prelude are left out, as in
 * TestNoUnusedVarsIgnoreRestSiblings. Upstream's id `unusedVar` is this rule's `noUnusedVars`.
 *
 * The programs are built under TypeScript's default module detection rather than the typed harness's
 * `moduleDetection: "force"`, which makes every file a module and so leaves `vars: "local"` nothing
 * to skip. A real repository's tsconfig does not force it, and there a file with no import or export
 * is a script, which is the reading the rows were replayed under.
 */

// noUnusedVarsScriptTsConfig is the harness's tsconfig without `moduleDetection: "force"`.
const noUnusedVarsScriptTsConfig = `{
	"compilerOptions": {"strict": true, "target": "ES2022", "lib": ["ES2022"], "types": []},
	"include": ["**/*.ts"]
}`
const noUnusedVarsOptionPrelude = "declare const o: Record<string, number>;\ndeclare const n2: { x: number[] };\ndeclare const list: number[];\ndeclare const list2: [number, { e: number }];\ndeclare const nested: number[][];\ndeclare const entries: [string, number][];\ndeclare function use(...values: unknown[]): void;\n"

func TestNoUnusedVarsVarsCaughtErrorsAndDestructuredArrayIgnorePattern(t *testing.T) {
	t.Parallel()

	cases := []struct {
		withPrelude bool
		options     string
		without     string
		body        string
		want        []string
		wantWithout []string
	}{
		{false, "\"local\"", "", "var a = 10;", nil, []string{"unusedVar a"}},
		{false, "{\"args\":\"all\",\"vars\":\"local\"}", "{\"args\":\"all\"}", "\nfunction g(bar, baz) {\n  return bar + baz;\n}\ng();\n      ", nil, nil},
		{false, "{\"vars\":\"local\",\"varsIgnorePattern\":\"^_\"}", "{\"varsIgnorePattern\":\"^_\"}", "\nvar a;\nfunction foo() {\n  var _b;\n}\nfoo();\n      ", nil, []string{"unusedVar a"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nconst [a, _b, c] = items;\nconsole.log(a + c);\n      ", nil, []string{"unusedVar _b"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nconst [[a, _b, c]] = items;\nconsole.log(a + c);\n      ", nil, []string{"unusedVar _b"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nconst {\n  x: [_a, foo],\n} = bar;\nconsole.log(foo);\n      ", nil, []string{"unusedVar _a"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nfunction baz([_b, foo]) {\n  foo;\n}\nbaz();\n      ", nil, []string{"unusedVar _b"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nfunction baz({ x: [_b, foo] }) {\n  foo;\n}\nbaz();\n      ", nil, []string{"unusedVar _b"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nfunction baz([\n  {\n    x: [_b, foo],\n  },\n]) {\n  foo;\n}\nbaz();\n      ", nil, []string{"unusedVar _b"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nlet _a, b;\nfoo.forEach(item => {\n  [_a, b] = item;\n  doSomething(b);\n});\n      ", nil, []string{"unusedVar _a"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\n// doesn't report _x\nlet _x, y;\n_x = 1;\n[_x, y] = foo;\ny;\n\n// doesn't report _a\nlet _a, b;\n[_a, b] = foo;\n_a = 1;\nb;\n      ", nil, []string{"unusedVar _x", "unusedVar _a"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\",\"ignoreRestSiblings\":true}", "{\"ignoreRestSiblings\":true}", "\n// doesn't report _x\nlet _x, y;\n_x = 1;\n[_x, y] = foo;\ny;\n\n// doesn't report _a\nlet _a, b;\n_a = 1;\n({ _a, ...b } = foo);\nb;\n      ", nil, []string{"unusedVar _x"}},
		{false, "{\"caughtErrors\":\"none\"}", "{}", "\ntry {\n} catch (err) {}\n      ", nil, []string{"unusedVar err"}},
		{false, "{\"args\":\"all\",\"caughtErrors\":\"none\",\"vars\":\"all\"}", "{\"args\":\"all\"}", "\ntry {\n} catch (err) {}\n      ", nil, []string{"unusedVar err"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nconst [a, _b, c] = items;\nconsole.log(a + c);\n      ", nil, []string{"unusedVar _b"}},
		{false, "{\"vars\":\"local\",\"varsIgnorePattern\":\"^_\"}", "{\"varsIgnorePattern\":\"^_\"}", "\nvar a;\nfunction foo() {\n  var _b;\n  var c_;\n}\nfoo();\n      ", []string{"unusedVar c_"}, []string{"unusedVar a", "unusedVar c_"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nconst array = ['a', 'b', 'c'];\nconst [a, _b, c] = array;\nconst newArray = [a, c];\n      ", []string{"unusedVar newArray"}, []string{"unusedVar _b", "unusedVar newArray"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nconst array = ['a', 'b', 'c', 'd', 'e'];\nconst [a, _b, c] = array;\n      ", []string{"unusedVar a", "unusedVar c"}, []string{"unusedVar a", "unusedVar _b", "unusedVar c"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\",\"varsIgnorePattern\":\"ignore\"}", "{\"varsIgnorePattern\":\"ignore\"}", "\nconst array = ['a', 'b', 'c'];\nconst [a, _b, c] = array;\nconst fooArray = ['foo'];\nconst barArray = ['bar'];\nconst ignoreArray = ['ignore'];\n      ", []string{"unusedVar a", "unusedVar c", "unusedVar fooArray", "unusedVar barArray"}, []string{"unusedVar a", "unusedVar _b", "unusedVar c", "unusedVar fooArray", "unusedVar barArray"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nconst array = [obj];\nconst [{ _a, foo }] = array;\nconsole.log(foo);\n      ", []string{"unusedVar _a"}, []string{"unusedVar _a"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nfunction foo([{ _a, bar }]) {\n  bar;\n}\nfoo();\n      ", []string{"unusedVar _a"}, []string{"unusedVar _a"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nlet _a, b;\n\nfoo.forEach(item => {\n  [a, b] = item;\n});\n      ", []string{"unusedVar _a", "unusedVar b"}, []string{"unusedVar _a", "unusedVar b"}},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "\nconst [a, _b] = items;\nconsole.log(a + _b);\n      ", nil, nil},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\",\"varsIgnorePattern\":\"[iI]gnored\"}", "{\"varsIgnorePattern\":\"[iI]gnored\"}", "\nlet _x;\n[_x] = arr;\nfoo(_x);\n      ", nil, nil},
		{false, "{\"destructuredArrayIgnorePattern\":\"^_\",\"varsIgnorePattern\":\"[iI]gnored\"}", "{\"varsIgnorePattern\":\"[iI]gnored\"}", "\nconst [ignored] = arr;\nfoo(ignored);\n      ", nil, nil},
		{true, "{\"caughtErrors\":\"none\"}", "{}", "try {} catch (err) {}", nil, []string{"unusedVar err"}},
		{true, "{\"caughtErrors\":\"none\"}", "{}", "try {} catch ({ message }) {}", nil, []string{"unusedVar message"}},
		{true, "{\"caughtErrors\":\"none\"}", "{}", "try {} catch ({ message: m, ...rest }) { use(rest); }", nil, []string{"unusedVar m"}},
		{true, "{\"caughtErrors\":\"none\"}", "{}", "try {} catch ([first]) {}", nil, []string{"unusedVar first"}},
		{true, "{\"caughtErrors\":\"none\"}", "{}", "try {} catch (err) { use(err); }", nil, nil},
		{true, "{\"caughtErrors\":\"none\"}", "{}", "try {} catch { const inner = 1; }", []string{"unusedVar inner"}, []string{"unusedVar inner"}},
		{true, "{\"caughtErrors\":\"none\"}", "{}", "try {} catch (err) { const inner = 1; }", []string{"unusedVar inner"}, []string{"unusedVar err", "unusedVar inner"}},
		{true, "{\"caughtErrors\":\"none\"}", "{}", "try {} catch (err) {}\nexport {};", nil, []string{"unusedVar err"}},
		{true, "{\"caughtErrorsIgnorePattern\":\"^_\"}", "{\"caughtErrorsIgnorePattern\":\"^_\"}", "try {} catch ({ _message }) {}", nil, nil},
		{true, "{\"caughtErrorsIgnorePattern\":\"^_\"}", "{\"caughtErrorsIgnorePattern\":\"^_\"}", "try {} catch ({ message: _m }) {}", nil, nil},
		{true, "{\"varsIgnorePattern\":\"^_\"}", "{\"varsIgnorePattern\":\"^_\"}", "try {} catch ({ _message }) {}", []string{"unusedVar _message"}, []string{"unusedVar _message"}},
		{true, "{\"vars\":\"local\"}", "{}", "var a = 1;", nil, []string{"unusedVar a"}},
		{true, "{\"vars\":\"local\"}", "{}", "let b = 1;", nil, []string{"unusedVar b"}},
		{true, "{\"vars\":\"local\"}", "{}", "const c = 1;", nil, []string{"unusedVar c"}},
		{true, "{\"vars\":\"local\"}", "{}", "function f() {}", nil, []string{"unusedVar f"}},
		{true, "{\"vars\":\"local\"}", "{}", "class C {}", nil, []string{"unusedVar C"}},
		{true, "{\"vars\":\"local\"}", "{}", "interface I {}", nil, []string{"unusedVar I"}},
		{true, "{\"vars\":\"local\"}", "{}", "type T = 1;", nil, []string{"unusedVar T"}},
		{true, "{\"vars\":\"local\"}", "{}", "enum E { A }", nil, []string{"unusedVar E"}},
		{true, "{\"vars\":\"local\"}", "{}", "namespace N { export const x = 1; }", nil, []string{"unusedVar N"}},
		{true, "{\"vars\":\"local\"}", "{}", "const [d, { e }] = list2;", nil, []string{"unusedVar d", "unusedVar e"}},
		{true, "{\"vars\":\"local\"}", "{}", "{\n  let blockScoped = 1;\n  var hoisted = 2;\n  function inBlock() {}\n}", []string{"unusedVar blockScoped", "unusedVar inBlock"}, []string{"unusedVar blockScoped", "unusedVar hoisted", "unusedVar inBlock"}},
		{true, "{\"vars\":\"local\"}", "{}", "if (list.length) {\n  var conditional = 1;\n}", nil, []string{"unusedVar conditional"}},
		{true, "{\"vars\":\"local\"}", "{}", "for (var i of list) {}", nil, []string{"unusedVar i"}},
		{true, "{\"vars\":\"local\"}", "{}", "for (let j of list) {}", []string{"unusedVar j"}, []string{"unusedVar j"}},
		{true, "{\"vars\":\"local\"}", "{}", "function outer() {\n  var inner = 1;\n  let other = 2;\n}\nouter();", []string{"unusedVar inner", "unusedVar other"}, []string{"unusedVar inner", "unusedVar other"}},
		{true, "{\"vars\":\"local\"}", "{}", "function g(p: number) {}\ng(1);", []string{"unusedVar p"}, []string{"unusedVar p"}},
		{true, "{\"vars\":\"local\"}", "{}", "type A<T> = number;\nuse(1 as A<string>);", []string{"unusedVar T"}, []string{"unusedVar T"}},
		{true, "{\"vars\":\"local\"}", "{}", "namespace M {\n  var insideNamespace = 1;\n}", []string{"unusedVar insideNamespace"}, []string{"unusedVar M", "unusedVar insideNamespace"}},
		{true, "{\"vars\":\"local\"}", "{}", "const local = 1;\nexport {};", []string{"unusedVar local"}, []string{"unusedVar local"}},
		{true, "{\"vars\":\"local\"}", "{}", "function h() {}\nexport {};", []string{"unusedVar h"}, []string{"unusedVar h"}},
		{true, "{\"vars\":\"local\"}", "{}", "import { x } from './other';", []string{"unusedVar x"}, []string{"unusedVar x"}},
		{true, "{\"vars\":\"local\"}", "{}", "const typeOnly = 1;\ntype U = typeof typeOnly;\nuse(1 as U);", nil, []string{"usedOnlyAsType typeOnly"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "const [_a, b] = list;\nuse(b);", nil, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "const [_a = 1, b] = list;\nuse(b);", []string{"unusedVar _a"}, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "const [b, ..._rest] = list;\nuse(b);", []string{"unusedVar _rest"}, []string{"unusedVar _rest"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "const { _a, b } = o;\nuse(b);", []string{"unusedVar _a"}, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "const { x: [_a] } = n2;", nil, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\n[_a] = list;", nil, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\n[(_a)] = list;", nil, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\n[_a = 1] = list;", []string{"unusedVar _a"}, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\n[..._a] = list;", []string{"unusedVar _a"}, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\n({ x: _a } = o);", []string{"unusedVar _a"}, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\n[[_a]] = nested;", nil, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\nfor ([_a] of nested) {}", nil, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\n_a = 1;\n[_a] = list;", nil, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "let _a;\nconst copy = [_a];\nuse(copy);", nil, nil},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "function f([_a, b]: number[]) {\n  return b;\n}\nuse(f);", nil, []string{"unusedVar _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "function f(_a: number, b: number) {\n  return b;\n}\nuse(f);", nil, nil},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "for (const [_k, v] of entries) {\n  use(v);\n}", nil, []string{"unusedVar _k"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "try {} catch ([_e]) {}", nil, []string{"unusedVar _e"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "const [_a] = list;\ntype T = typeof _a;\nuse(1 as T);", nil, []string{"usedOnlyAsType _a"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\"}", "{}", "const [notMatched] = list;", []string{"unusedVar notMatched"}, []string{"unusedVar notMatched"}},
		{true, "{\"destructuredArrayIgnorePattern\":\"^_\",\"varsIgnorePattern\":\"^ignored\"}", "{\"varsIgnorePattern\":\"^ignored\"}", "const [_a, ignoredToo] = list;\nconst ignoredAlone = 1;\nconst [kept] = list;", []string{"unusedVar kept"}, []string{"unusedVar _a", "unusedVar kept"}},
	}
	if len(cases) != 79 {
		t.Fatalf("%d rows, and 79 were replayed", len(cases))
	}

	for _, testCase := range cases {
		for _, column := range []struct {
			label   string
			options string
			want    []string
		}{{"with", testCase.options, testCase.want}, {"without", testCase.without, testCase.wantWithout}} {
			t.Run(column.label+" "+column.options+": "+testCase.body, func(t *testing.T) {
				t.Parallel()

				options, err := DecodeNoUnusedVarsOptions(json.RawMessage(column.options))
				if err != nil {
					t.Fatalf("decoding %s: %v", column.options, err)
				}
				prelude := ""
				if testCase.withPrelude {
					prelude = noUnusedVarsOptionPrelude
				}
				// The harness writes the file trimmed, so positions are in the trimmed text.
				source := strings.TrimSpace(prelude+testCase.body) + "\n"
				result := rule_testing.RunTypedFilesWithSetupAndOptions(t, NoUnusedVars, map[string]string{"a.ts": source}, "a.ts", options,
					func(directory string) {
						if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(noUnusedVarsScriptTsConfig), 0o644); err != nil {
							t.Fatalf("writing the tsconfig: %v", err)
						}
					})
				var got []string
				for _, diagnostic := range result.Diagnostics {
					if diagnostic.Range.Pos() < len(prelude) {
						continue
					}
					id := diagnostic.Message.Id
					if id == "noUnusedVars" {
						id = "unusedVar"
					}
					got = append(got, id+" "+source[diagnostic.Range.Pos():diagnostic.Range.End()])
				}
				if strings.Join(got, ",") != strings.Join(column.want, ",") {
					t.Fatalf("reported %v, want %v", got, column.want)
				}
			})
		}
	}
}

// The decoder reads upstream's two forms and refuses a mode spelling upstream's schema refuses, so a
// typo fails at load rather than running the default (#6esg2nx, the class of #p9s1131).
func TestDecodeNoUnusedVarsOptionsModesAndShorthand(t *testing.T) {
	t.Parallel()

	accepted := map[string]NoUnusedVarsOptions{
		`"local"`:                  {Vars: "local", Args: "after-used", CaughtErrors: "all"},
		`"all"`:                    {Vars: "all", Args: "after-used", CaughtErrors: "all"},
		`{"vars": "local"}`:        {Vars: "local", Args: "after-used", CaughtErrors: "all"},
		`{"args": "none"}`:         {Vars: "all", Args: "none", CaughtErrors: "all"},
		`{"caughtErrors": "none"}`: {Vars: "all", Args: "after-used", CaughtErrors: "none"},
		`{"destructuredArrayIgnorePattern": "^_"}`: {
			Vars: "all", Args: "after-used", CaughtErrors: "all", DestructuredArrayIgnorePattern: "^_",
		},
		`{"ignoreClassWithStaticInitBlock": true, "ignoreUsingDeclarations": true, "reportUsedIgnorePattern": true}`: {
			Vars: "all", Args: "after-used", CaughtErrors: "all",
			IgnoreClassWithStaticInitBlock: true, IgnoreUsingDeclarations: true, ReportUsedIgnorePattern: true,
		},
	}
	for raw, want := range accepted {
		decoded, err := DecodeNoUnusedVarsOptions(json.RawMessage(raw))
		if err != nil {
			t.Errorf("%s was refused: %v", raw, err)
			continue
		}
		if decoded != want {
			t.Errorf("%s decoded to %+v, want %+v", raw, decoded, want)
		}
	}

	refused := map[string]string{
		`"locals"`:                           `"locals"`,
		`{"vars": "locals"}`:                 `vars is one of`,
		`{"vars": "Local"}`:                  `"Local"`,
		`{"args": "after_used"}`:             `args is one of`,
		`{"caughtErrors": "never"}`:          `caughtErrors is one of`,
		`{"reportUsedIgnorePatterns": true}`: `reportUsedIgnorePatterns`,
		`{"Vars": "local"}`:                  `"Vars"`,
	}
	for raw, naming := range refused {
		_, err := DecodeNoUnusedVarsOptions(json.RawMessage(raw))
		if err == nil {
			t.Errorf("%s decoded, and upstream's schema refuses it", raw)
			continue
		}
		if !strings.Contains(err.Error(), naming) {
			t.Errorf("%s was refused, but the error does not name %s: %v", raw, naming, err)
		}
	}
}

// A name destructured out of a parameter is a parameter, as upstream's definition type says (#s48y8eg).
//
// cohere filed it as a variable, so `args: "none"` and `argsIgnorePattern` did not reach it and
// `varsIgnorePattern` did. Every row was replayed through the installed
// `@typescript-eslint/no-unused-vars` 8.71.0 in a module, with the prelude's findings left out. The
// `after-used` rows pin the two halves of upstream's isAfterLastUsedArg: a destructured name is never
// skipped by it, and a later destructured name that is read, or given a default anywhere in its
// pattern, keeps the parameters before it.
func TestNoUnusedVarsJudgesADestructuredParameterAsAParameter(t *testing.T) {
	t.Parallel()

	const prelude = "declare function use(...values: unknown[]): void;\n"
	cases := []struct {
		options string
		body    string
		want    []string
	}{
		{"{\"args\":\"none\"}", "function f({ a }: { a: number }) {}\nuse(f);", nil},
		{"{\"args\":\"none\"}", "function f([a]: number[]) {}\nuse(f);", nil},
		{"{\"args\":\"none\"}", "function f({ x: { a } }: { x: { a: number } }) {}\nuse(f);", nil},
		{"{\"args\":\"none\"}", "function f({ a }: { a?: number } = {}) {}\nuse(f);", nil},
		{"{\"args\":\"none\"}", "const g = ({ a }: { a: number }) => 1;\nuse(g);", nil},
		{"{\"args\":\"none\"}", "function f(p: number) {\n  const { a } = { a: p };\n}\nuse(f);", []string{"unusedVar a"}},
		{"{\"argsIgnorePattern\":\"^_\"}", "function f({ _a }: { _a: number }) {}\nuse(f);", nil},
		{"{\"argsIgnorePattern\":\"^_\"}", "function f([_a]: number[]) {}\nuse(f);", nil},
		{"{\"argsIgnorePattern\":\"^_\"}", "function f({ x: _a }: { x: number }) {}\nuse(f);", nil},
		{"{\"argsIgnorePattern\":\"^_\"}", "function f({ a, ..._rest }: Record<string, number>) {\n  return a;\n}\nuse(f);", nil},
		{"{\"argsIgnorePattern\":\"^_\"}", "function f(p: number) {\n  const { _a } = { _a: p };\n}\nuse(f);", []string{"unusedVar _a"}},
		{"{\"varsIgnorePattern\":\"^_\"}", "function f({ _a }: { _a: number }) {}\nuse(f);", []string{"unusedVar _a"}},
		{"{\"varsIgnorePattern\":\"^_\"}", "function f([_a]: number[]) {}\nuse(f);", []string{"unusedVar _a"}},
		{"{\"varsIgnorePattern\":\"^_\"}", "function f(p: number) {\n  const { _a } = { _a: p };\n}\nuse(f);", nil},
		{"{}", "function f(a: number, { b }: { b: number }) {\n  return b;\n}\nuse(f);", nil},
		{"{}", "function f({ a }: { a: number }, b: number) {\n  return b;\n}\nuse(f);", []string{"unusedVar a"}},
		{"{}", "function f({ a, b }: { a: number; b: number }) {\n  return b;\n}\nuse(f);", []string{"unusedVar a"}},
		{"{}", "function f(a: number, { b = 1 }: { b?: number }) {}\nuse(f);", []string{"unusedVar b"}},
		{"{}", "function f(a: number, [b]: number[]) {\n  return b;\n}\nuse(f);", nil},
		{"{}", "function f(a: number, { x: { b } }: { x: { b: number } }) {\n  return b;\n}\nuse(f);", nil},
		{"{}", "function f(a: number, { b }: { b: number }, c: number) {\n  return c;\n}\nuse(f);", []string{"unusedVar b"}},
		{"{\"args\":\"all\"}", "function f(a: number, { b }: { b: number }) {\n  return b;\n}\nuse(f);", []string{"unusedVar a"}},
		{"{\"args\":\"after-used\",\"argsIgnorePattern\":\"^_\"}", "function f(a: number, { _b }: { _b: number }) {}\nuse(f);", []string{"unusedVar a"}},
		{"{\"argsIgnorePattern\":\"^_\"}", "function f([_a, b]: number[]) {\n  return b;\n}\nuse(f);", nil},
		{"{\"destructuredArrayIgnorePattern\":\"^_\"}", "function f([_a]: number[]) {}\nuse(f);", nil},
		{"{\"destructuredArrayIgnorePattern\":\"^_\",\"args\":\"all\"}", "function f({ _a }: { _a: number }) {}\nuse(f);", []string{"unusedVar _a"}},
		{"{}", "class C {\n  m({ a }: { a: number }) {}\n}\nuse(C);", []string{"unusedVar a"}},
		{"{}", "class C {\n  set v({ a }: { a: number }) {}\n}\nuse(C);", nil},
		{"{\"args\":\"none\"}", "const { a } = { a: 1 };", []string{"unusedVar a"}},
		{"{}", "function f(a: number, { b }: { b?: number } = {}) {}\nuse(f);", []string{"unusedVar b"}},
		{"{}", "function f(a: number, { x: { b } = { b: 1 } }: { x?: { b: number } }) {}\nuse(f);", []string{"unusedVar b"}},
		{"{}", "function f(a: number, [b = 1]: number[]) {}\nuse(f);", []string{"unusedVar b"}},
		{"{}", "class C {\n  set v(a: number) {}\n}\nuse(C);", nil},
		{"{\"argsIgnorePattern\":\"^_\"}", "function f(a: number, { b }: { b: number }) {}\nuse(f);", []string{"unusedVar a", "unusedVar b"}},
	}
	if len(cases) != 34 {
		t.Fatalf("%d rows, and 34 were replayed", len(cases))
	}

	for _, testCase := range cases {
		t.Run(testCase.options+" "+testCase.body, func(t *testing.T) {
			t.Parallel()

			options, err := DecodeNoUnusedVarsOptions(json.RawMessage(testCase.options))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.options, err)
			}
			source := prelude + testCase.body + "\nexport {};\n"
			result := rule_testing.RunTypedWithOptions(t, NoUnusedVars, "a.ts", source, options)
			var got []string
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Range.Pos() < len(prelude) {
					continue
				}
				id := diagnostic.Message.Id
				if id == "noUnusedVars" {
					id = "unusedVar"
				}
				got = append(got, id+" "+source[diagnostic.Range.Pos():diagnostic.Range.End()])
			}
			if strings.Join(got, ",") != strings.Join(testCase.want, ",") {
				t.Fatalf("reported %v, want %v", got, testCase.want)
			}
		})
	}
}
