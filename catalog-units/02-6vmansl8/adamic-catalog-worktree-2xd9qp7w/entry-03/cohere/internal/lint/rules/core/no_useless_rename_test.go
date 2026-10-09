package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// uselessRenameFile is where the fixtures pretend to live.
const uselessRenameFile = "/repository/source/UselessRename.ts"

// uselessRenameOptions routes the options through the rule's own exported decoder.
func uselessRenameOptions(t *testing.T, ignoreDestructuring bool, ignoreImport bool,
	ignoreExport bool) any {
	t.Helper()
	raw, err := json.Marshal(map[string]bool{
		"ignoreDestructuring": ignoreDestructuring,
		"ignoreImport":        ignoreImport,
		"ignoreExport":        ignoreExport,
	})
	if err != nil {
		t.Fatalf("could not marshal the options: %v", err)
	}
	decoded, err := DecodeNoUselessRenameOptions(raw)
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// The corpus is ESLint's own, imported from eslint/tests/lib/rules/no-useless-rename.js:
// 55 pass and 107 fail, with 86 fix vectors and 21 deliberate declines.
//
// Extracted by loading upstream's tester with a stubbed RuleTester and rendering these literals
// from that JSON, so nothing was retyped and no escape sequence was hand-written. That matters
// more here than usual: five cases turn on a unicode ESCAPE being preserved as an escape, and a
// tool that cooked \\u0061 into a would leave them asserting the opposite of upstream while still
// passing. The generator refuses any byte outside printable ASCII, and TestNoUselessRenameCorpus
// ByteCheck compares the escape cases against upstream's own bytes.
func TestNoUselessRenameFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText          string
		ignoreDestructuring bool
		ignoreImport        bool
		ignoreExport        bool
		findings            int
	}{
		{"let {foo: foo} = obj;", false, false, false, 1},
		{"({foo: (foo)} = obj);", false, false, false, 1},
		{"let {\\u0061: a} = obj;", false, false, false, 1},
		{"let {a: \\u0061} = obj;", false, false, false, 1},
		{"let {\\u0061: \\u0061} = obj;", false, false, false, 1},
		{"let {a, foo: foo} = obj;", false, false, false, 1},
		{"let {foo: foo, bar: baz} = obj;", false, false, false, 1},
		{"let {foo: bar, baz: baz} = obj;", false, false, false, 1},
		{"let {foo: foo, bar: bar} = obj;", false, false, false, 2},
		{"let {foo: {bar: bar}} = obj;", false, false, false, 1},
		{"let {foo: {bar: bar}, baz: baz} = obj;", false, false, false, 2},
		{"let {'foo': foo} = obj;", false, false, false, 1},
		{"let {'foo': foo, 'bar': baz} = obj;", false, false, false, 1},
		{"let {'foo': bar, 'baz': baz} = obj;", false, false, false, 1},
		{"let {'foo': foo, 'bar': bar} = obj;", false, false, false, 2},
		{"let {'foo': {'bar': bar}} = obj;", false, false, false, 1},
		{"let {'foo': {'bar': bar}, 'baz': baz} = obj;", false, false, false, 2},
		{"let {foo: foo = 1, 'bar': bar = 1, baz: baz} = obj;", false, false, false, 3},
		{"let {foo: {bar: bar = 1, 'baz': baz = 1}} = obj;", false, false, false, 2},
		{"let {foo: {bar: bar = {}} = {}} = obj;", false, false, false, 1},
		{"({foo: (foo) = a} = obj);", false, false, false, 1},
		{"let {foo: foo = (a)} = obj;", false, false, false, 1},
		{"let {foo: foo = (a, b)} = obj;", false, false, false, 1},
		{"function func({foo: foo}) {}", false, false, false, 1},
		{"function func({foo: foo, bar: baz}) {}", false, false, false, 1},
		{"function func({foo: bar, baz: baz}) {}", false, false, false, 1},
		{"function func({foo: foo, bar: bar}) {}", false, false, false, 2},
		{"function func({foo: foo = 1, 'bar': bar = 1, baz: baz}) {}", false, false, false, 3},
		{"function func({foo: {bar: bar = 1, 'baz': baz = 1}}) {}", false, false, false, 2},
		{"function func({foo: {bar: bar = {}} = {}}) {}", false, false, false, 1},
		{"({foo: foo}) => {}", false, false, false, 1},
		{"({foo: foo, bar: baz}) => {}", false, false, false, 1},
		{"({foo: bar, baz: baz}) => {}", false, false, false, 1},
		{"({foo: foo, bar: bar}) => {}", false, false, false, 2},
		{"({foo: foo = 1, 'bar': bar = 1, baz: baz}) => {}", false, false, false, 3},
		{"({foo: {bar: bar = 1, 'baz': baz = 1}}) => {}", false, false, false, 2},
		{"({foo: {bar: bar = {}} = {}}) => {}", false, false, false, 1},
		{"const {foo: foo, ...stuff} = myObject;", false, false, false, 1},
		{"const {foo: foo, bar: baz, ...stuff} = myObject;", false, false, false, 1},
		{"const {foo: foo, bar: bar, ...stuff} = myObject;", false, false, false, 2},
		{"import {foo as foo} from 'foo';", false, false, false, 1},
		{"import {'foo' as foo} from 'foo';", false, false, false, 1},
		{"import {\\u0061 as a} from 'foo';", false, false, false, 1},
		{"import {a as \\u0061} from 'foo';", false, false, false, 1},
		{"import {\\u0061 as \\u0061} from 'foo';", false, false, false, 1},
		{"import {foo as foo, bar as baz} from 'foo';", false, false, false, 1},
		{"import {foo as bar, baz as baz} from 'foo';", false, false, false, 1},
		{"import {foo as foo, bar as bar} from 'foo';", false, false, false, 2},
		{"var foo = 0; export {foo as foo};", false, false, false, 1},
		{"var foo = 0; export {foo as 'foo'};", false, false, false, 1},
		{"export {foo as 'foo'} from 'bar';", false, false, false, 1},
		{"export {'foo' as foo} from 'bar';", false, false, false, 1},
		{"export {'foo' as 'foo'} from 'bar';", false, false, false, 1},
		{"export {' \U0001f44d ' as ' \U0001f44d '} from 'bar';", false, false, false, 1},
		{"export {'' as ''} from 'bar';", false, false, false, 1},
		{"var a = 0; export {a as \\u0061};", false, false, false, 1},
		{"var \\u0061 = 0; export {\\u0061 as a};", false, false, false, 1},
		{"var \\u0061 = 0; export {\\u0061 as \\u0061};", false, false, false, 1},
		{"var foo = 0; var bar = 0; export {foo as foo, bar as baz};", false, false, false, 1},
		{"var foo = 0; var baz = 0; export {foo as bar, baz as baz};", false, false, false, 1},
		{"var foo = 0; var bar = 0;export {foo as foo, bar as bar};", false, false, false, 2},
		{"export {foo as foo} from 'foo';", false, false, false, 1},
		{"export {a as \\u0061} from 'foo';", false, false, false, 1},
		{"export {\\u0061 as a} from 'foo';", false, false, false, 1},
		{"export {\\u0061 as \\u0061} from 'foo';", false, false, false, 1},
		{"export {foo as foo, bar as baz} from 'foo';", false, false, false, 1},
		{"var foo = 0; var bar = 0; export {foo as bar, baz as baz} from 'foo';", false, false, false, 1},
		{"export {foo as foo, bar as bar} from 'foo';", false, false, false, 2},
		{"({/* comment */foo: foo} = {});", false, false, false, 1},
		{"({/* comment */foo: foo = 1} = {});", false, false, false, 1},
		{"({foo, /* comment */bar: bar} = {});", false, false, false, 1},
		{"({foo/**/ : foo} = {});", false, false, false, 1},
		{"({foo/**/ : foo = 1} = {});", false, false, false, 1},
		{"({foo /**/: foo} = {});", false, false, false, 1},
		{"({foo /**/: foo = 1} = {});", false, false, false, 1},
		{"({foo://\nfoo} = {});", false, false, false, 1},
		{"({foo: /**/foo} = {});", false, false, false, 1},
		{"({foo: (/**/foo)} = {});", false, false, false, 1},
		{"({foo: (foo/**/)} = {});", false, false, false, 1},
		{"({foo: (foo //\n)} = {});", false, false, false, 1},
		{"({foo: /**/foo = 1} = {});", false, false, false, 1},
		{"({foo: (/**/foo) = 1} = {});", false, false, false, 1},
		{"({foo: (foo/**/) = 1} = {});", false, false, false, 1},
		{"({foo: foo/* comment */} = {});", false, false, false, 1},
		{"({foo: foo//comment\n,bar} = {});", false, false, false, 1},
		{"({foo: foo/* comment */ = 1} = {});", false, false, false, 1},
		{"({foo: foo // comment\n = 1} = {});", false, false, false, 1},
		{"({foo: foo = /* comment */ 1} = {});", false, false, false, 1},
		{"({foo: foo = // comment\n 1} = {});", false, false, false, 1},
		{"({foo: foo = (1/* comment */)} = {});", false, false, false, 1},
		{"import {/* comment */foo as foo} from 'foo';", false, false, false, 1},
		{"import {foo,/* comment */bar as bar} from 'foo';", false, false, false, 1},
		{"import {foo/**/ as foo} from 'foo';", false, false, false, 1},
		{"import {foo /**/as foo} from 'foo';", false, false, false, 1},
		{"import {foo //\nas foo} from 'foo';", false, false, false, 1},
		{"import {foo as/**/foo} from 'foo';", false, false, false, 1},
		{"import {foo as foo/* comment */} from 'foo';", false, false, false, 1},
		{"import {foo as foo/* comment */,bar} from 'foo';", false, false, false, 1},
		{"let foo; export {/* comment */foo as foo};", false, false, false, 1},
		{"let foo, bar; export {foo,/* comment */bar as bar};", false, false, false, 1},
		{"let foo; export {foo/**/as foo};", false, false, false, 1},
		{"let foo; export {foo as/**/ foo};", false, false, false, 1},
		{"let foo; export {foo as /**/foo};", false, false, false, 1},
		{"let foo; export {foo as//comment\n foo};", false, false, false, 1},
		{"let foo; export {foo as foo/* comment*/};", false, false, false, 1},
		{"let foo, bar; export {foo as foo/* comment*/,bar};", false, false, false, 1},
		{"let foo, bar; export {foo as foo//comment\n,bar};", false, false, false, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "unnecessarilyRenamed"
			}
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoUselessRename,
				uselessRenameFile, testCase.sourceText, uselessRenameOptions(t,
					testCase.ignoreDestructuring, testCase.ignoreImport, testCase.ignoreExport)),
				expected...)
		})
	}
}

// The clean cases, which carry most of the discrimination.
//
// Shorthand renames nothing and a computed key is not known until run time, so a port comparing
// two names without those tests reports every shorthand destructuring in the tree. The three
// option rows are the other half: the same source is a finding without the option and clean with
// it.
func TestNoUselessRenameStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText          string
		ignoreDestructuring bool
		ignoreImport        bool
		ignoreExport        bool
	}{
		{"let {foo} = obj;", false, false, false},
		{"let {foo: bar} = obj;", false, false, false},
		{"let {foo: bar, baz: qux} = obj;", false, false, false},
		{"let {foo: {bar: baz}} = obj;", false, false, false},
		{"let {foo, bar: {baz: qux}} = obj;", false, false, false},
		{"let {'foo': bar} = obj;", false, false, false},
		{"let {'foo': bar, 'baz': qux} = obj;", false, false, false},
		{"let {'foo': {'bar': baz}} = obj;", false, false, false},
		{"let {foo, 'bar': {'baz': qux}} = obj;", false, false, false},
		{"let {['foo']: bar} = obj;", false, false, false},
		{"let {['foo']: bar, ['baz']: qux} = obj;", false, false, false},
		{"let {['foo']: {['bar']: baz}} = obj;", false, false, false},
		{"let {foo, ['bar']: {['baz']: qux}} = obj;", false, false, false},
		{"let {[foo]: foo} = obj;", false, false, false},
		{"let {['foo']: foo} = obj;", false, false, false},
		{"let {[foo]: bar} = obj;", false, false, false},
		{"function func({foo}) {}", false, false, false},
		{"function func({foo: bar}) {}", false, false, false},
		{"function func({foo: bar, baz: qux}) {}", false, false, false},
		{"({foo}) => {}", false, false, false},
		{"({foo: bar}) => {}", false, false, false},
		{"({foo: bar, baz: qui}) => {}", false, false, false},
		{"import * as foo from 'foo';", false, false, false},
		{"import foo from 'foo';", false, false, false},
		{"import {foo} from 'foo';", false, false, false},
		{"import {foo as bar} from 'foo';", false, false, false},
		{"import {foo as bar, baz as qux} from 'foo';", false, false, false},
		{"import {'foo' as bar} from 'baz';", false, false, false},
		{"export {foo} from 'foo';", false, false, false},
		{"var foo = 0;export {foo as bar};", false, false, false},
		{"var foo = 0; var baz = 0; export {foo as bar, baz as qux};", false, false, false},
		{"export {foo as bar} from 'foo';", false, false, false},
		{"export {foo as bar, baz as qux} from 'foo';", false, false, false},
		{"var foo = 0; export {foo as 'bar'};", false, false, false},
		{"export {foo as 'bar'} from 'baz';", false, false, false},
		{"export {'foo' as bar} from 'baz';", false, false, false},
		{"export {'foo' as 'bar'} from 'baz';", false, false, false},
		{"export {'' as ' '} from 'baz';", false, false, false},
		{"export {' ' as ''} from 'baz';", false, false, false},
		{"export {'foo'} from 'bar';", false, false, false},
		{"const {...stuff} = myObject;", false, false, false},
		{"const {foo, ...stuff} = myObject;", false, false, false},
		{"const {foo: bar, ...stuff} = myObject;", false, false, false},
		{"let {foo: foo} = obj;", true, false, false},
		{"let {foo: foo, bar: baz} = obj;", true, false, false},
		{"let {foo: foo, bar: bar} = obj;", true, false, false},
		{"import {foo as foo} from 'foo';", false, true, false},
		{"import {foo as foo, bar as baz} from 'foo';", false, true, false},
		{"import {foo as foo, bar as bar} from 'foo';", false, true, false},
		{"var foo = 0;export {foo as foo};", false, false, true},
		{"var foo = 0;var bar = 0;export {foo as foo, bar as baz};", false, false, true},
		{"var foo = 0;var bar = 0;export {foo as foo, bar as bar};", false, false, true},
		{"export {foo as foo} from 'foo';", false, false, true},
		{"export {foo as foo, bar as baz} from 'foo';", false, false, true},
		{"export {foo as foo, bar as bar} from 'foo';", false, false, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUselessRename,
				uselessRenameFile, testCase.sourceText, uselessRenameOptions(t,
					testCase.ignoreDestructuring, testCase.ignoreImport, testCase.ignoreExport)))
		})
	}
}

// The fix vectors, which are upstream's own output fields.
//
// These assert what the repair WRITES rather than that a finding appeared, and they are the only
// thing that can see a fixer repairing the right span with the wrong text. The escape rows are
// the sharpest: let {a: \\u0061} = obj fixes to let {\\u0061} = obj, keeping the escape, while
// let {\\u0061: a} = obj fixes to let {a} = obj. Which side survives is decided by the repair
// keeping the LOCAL text, and nothing but an output comparison could tell the two apart.
func TestNoUselessRenameFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText          string
		wantSource          string
		ignoreDestructuring bool
		ignoreImport        bool
		ignoreExport        bool
	}{
		{"let {foo: foo} = obj;", "let {foo} = obj;", false, false, false},
		{"({foo: (foo)} = obj);", "({foo} = obj);", false, false, false},
		{"let {\\u0061: a} = obj;", "let {a} = obj;", false, false, false},
		{"let {a: \\u0061} = obj;", "let {\\u0061} = obj;", false, false, false},
		{"let {\\u0061: \\u0061} = obj;", "let {\\u0061} = obj;", false, false, false},
		{"let {a, foo: foo} = obj;", "let {a, foo} = obj;", false, false, false},
		{"let {foo: foo, bar: baz} = obj;", "let {foo, bar: baz} = obj;", false, false, false},
		{"let {foo: bar, baz: baz} = obj;", "let {foo: bar, baz} = obj;", false, false, false},
		{"let {foo: foo, bar: bar} = obj;", "let {foo, bar} = obj;", false, false, false},
		{"let {foo: {bar: bar}} = obj;", "let {foo: {bar}} = obj;", false, false, false},
		{"let {foo: {bar: bar}, baz: baz} = obj;", "let {foo: {bar}, baz} = obj;", false, false, false},
		{"let {'foo': foo} = obj;", "let {foo} = obj;", false, false, false},
		{"let {'foo': foo, 'bar': baz} = obj;", "let {foo, 'bar': baz} = obj;", false, false, false},
		{"let {'foo': bar, 'baz': baz} = obj;", "let {'foo': bar, baz} = obj;", false, false, false},
		{"let {'foo': foo, 'bar': bar} = obj;", "let {foo, bar} = obj;", false, false, false},
		{"let {'foo': {'bar': bar}} = obj;", "let {'foo': {bar}} = obj;", false, false, false},
		{"let {'foo': {'bar': bar}, 'baz': baz} = obj;", "let {'foo': {bar}, baz} = obj;", false, false, false},
		{"let {foo: foo = 1, 'bar': bar = 1, baz: baz} = obj;", "let {foo = 1, bar = 1, baz} = obj;", false, false, false},
		{"let {foo: {bar: bar = 1, 'baz': baz = 1}} = obj;", "let {foo: {bar = 1, baz = 1}} = obj;", false, false, false},
		{"let {foo: {bar: bar = {}} = {}} = obj;", "let {foo: {bar = {}} = {}} = obj;", false, false, false},
		{"let {foo: foo = (a)} = obj;", "let {foo = (a)} = obj;", false, false, false},
		{"let {foo: foo = (a, b)} = obj;", "let {foo = (a, b)} = obj;", false, false, false},
		{"function func({foo: foo}) {}", "function func({foo}) {}", false, false, false},
		{"function func({foo: foo, bar: baz}) {}", "function func({foo, bar: baz}) {}", false, false, false},
		{"function func({foo: bar, baz: baz}) {}", "function func({foo: bar, baz}) {}", false, false, false},
		{"function func({foo: foo, bar: bar}) {}", "function func({foo, bar}) {}", false, false, false},
		{"function func({foo: foo = 1, 'bar': bar = 1, baz: baz}) {}", "function func({foo = 1, bar = 1, baz}) {}", false, false, false},
		{"function func({foo: {bar: bar = 1, 'baz': baz = 1}}) {}", "function func({foo: {bar = 1, baz = 1}}) {}", false, false, false},
		{"function func({foo: {bar: bar = {}} = {}}) {}", "function func({foo: {bar = {}} = {}}) {}", false, false, false},
		{"({foo: foo}) => {}", "({foo}) => {}", false, false, false},
		{"({foo: foo, bar: baz}) => {}", "({foo, bar: baz}) => {}", false, false, false},
		{"({foo: bar, baz: baz}) => {}", "({foo: bar, baz}) => {}", false, false, false},
		{"({foo: foo, bar: bar}) => {}", "({foo, bar}) => {}", false, false, false},
		{"({foo: foo = 1, 'bar': bar = 1, baz: baz}) => {}", "({foo = 1, bar = 1, baz}) => {}", false, false, false},
		{"({foo: {bar: bar = 1, 'baz': baz = 1}}) => {}", "({foo: {bar = 1, baz = 1}}) => {}", false, false, false},
		{"({foo: {bar: bar = {}} = {}}) => {}", "({foo: {bar = {}} = {}}) => {}", false, false, false},
		{"const {foo: foo, ...stuff} = myObject;", "const {foo, ...stuff} = myObject;", false, false, false},
		{"const {foo: foo, bar: baz, ...stuff} = myObject;", "const {foo, bar: baz, ...stuff} = myObject;", false, false, false},
		{"const {foo: foo, bar: bar, ...stuff} = myObject;", "const {foo, bar, ...stuff} = myObject;", false, false, false},
		{"import {foo as foo} from 'foo';", "import {foo} from 'foo';", false, false, false},
		{"import {'foo' as foo} from 'foo';", "import {foo} from 'foo';", false, false, false},
		{"import {\\u0061 as a} from 'foo';", "import {a} from 'foo';", false, false, false},
		{"import {a as \\u0061} from 'foo';", "import {\\u0061} from 'foo';", false, false, false},
		{"import {\\u0061 as \\u0061} from 'foo';", "import {\\u0061} from 'foo';", false, false, false},
		{"import {foo as foo, bar as baz} from 'foo';", "import {foo, bar as baz} from 'foo';", false, false, false},
		{"import {foo as bar, baz as baz} from 'foo';", "import {foo as bar, baz} from 'foo';", false, false, false},
		{"import {foo as foo, bar as bar} from 'foo';", "import {foo, bar} from 'foo';", false, false, false},
		{"var foo = 0; export {foo as foo};", "var foo = 0; export {foo};", false, false, false},
		{"var foo = 0; export {foo as 'foo'};", "var foo = 0; export {foo};", false, false, false},
		{"export {foo as 'foo'} from 'bar';", "export {foo} from 'bar';", false, false, false},
		{"export {'foo' as foo} from 'bar';", "export {'foo'} from 'bar';", false, false, false},
		{"export {'foo' as 'foo'} from 'bar';", "export {'foo'} from 'bar';", false, false, false},
		{"export {' \U0001f44d ' as ' \U0001f44d '} from 'bar';", "export {' \U0001f44d '} from 'bar';", false, false, false},
		{"export {'' as ''} from 'bar';", "export {''} from 'bar';", false, false, false},
		{"var a = 0; export {a as \\u0061};", "var a = 0; export {a};", false, false, false},
		{"var \\u0061 = 0; export {\\u0061 as a};", "var \\u0061 = 0; export {\\u0061};", false, false, false},
		{"var \\u0061 = 0; export {\\u0061 as \\u0061};", "var \\u0061 = 0; export {\\u0061};", false, false, false},
		{"var foo = 0; var bar = 0; export {foo as foo, bar as baz};", "var foo = 0; var bar = 0; export {foo, bar as baz};", false, false, false},
		{"var foo = 0; var baz = 0; export {foo as bar, baz as baz};", "var foo = 0; var baz = 0; export {foo as bar, baz};", false, false, false},
		{"var foo = 0; var bar = 0;export {foo as foo, bar as bar};", "var foo = 0; var bar = 0;export {foo, bar};", false, false, false},
		{"export {foo as foo} from 'foo';", "export {foo} from 'foo';", false, false, false},
		{"export {a as \\u0061} from 'foo';", "export {a} from 'foo';", false, false, false},
		{"export {\\u0061 as a} from 'foo';", "export {\\u0061} from 'foo';", false, false, false},
		{"export {\\u0061 as \\u0061} from 'foo';", "export {\\u0061} from 'foo';", false, false, false},
		{"export {foo as foo, bar as baz} from 'foo';", "export {foo, bar as baz} from 'foo';", false, false, false},
		{"var foo = 0; var bar = 0; export {foo as bar, baz as baz} from 'foo';", "var foo = 0; var bar = 0; export {foo as bar, baz} from 'foo';", false, false, false},
		{"export {foo as foo, bar as bar} from 'foo';", "export {foo, bar} from 'foo';", false, false, false},
		{"({/* comment */foo: foo} = {});", "({/* comment */foo} = {});", false, false, false},
		{"({/* comment */foo: foo = 1} = {});", "({/* comment */foo = 1} = {});", false, false, false},
		{"({foo, /* comment */bar: bar} = {});", "({foo, /* comment */bar} = {});", false, false, false},
		{"({foo: foo/* comment */} = {});", "({foo/* comment */} = {});", false, false, false},
		{"({foo: foo//comment\n,bar} = {});", "({foo//comment\n,bar} = {});", false, false, false},
		{"({foo: foo/* comment */ = 1} = {});", "({foo/* comment */ = 1} = {});", false, false, false},
		{"({foo: foo // comment\n = 1} = {});", "({foo // comment\n = 1} = {});", false, false, false},
		{"({foo: foo = /* comment */ 1} = {});", "({foo = /* comment */ 1} = {});", false, false, false},
		{"({foo: foo = // comment\n 1} = {});", "({foo = // comment\n 1} = {});", false, false, false},
		{"({foo: foo = (1/* comment */)} = {});", "({foo = (1/* comment */)} = {});", false, false, false},
		{"import {/* comment */foo as foo} from 'foo';", "import {/* comment */foo} from 'foo';", false, false, false},
		{"import {foo,/* comment */bar as bar} from 'foo';", "import {foo,/* comment */bar} from 'foo';", false, false, false},
		{"import {foo as foo/* comment */} from 'foo';", "import {foo/* comment */} from 'foo';", false, false, false},
		{"import {foo as foo/* comment */,bar} from 'foo';", "import {foo/* comment */,bar} from 'foo';", false, false, false},
		{"let foo; export {/* comment */foo as foo};", "let foo; export {/* comment */foo};", false, false, false},
		{"let foo, bar; export {foo,/* comment */bar as bar};", "let foo, bar; export {foo,/* comment */bar};", false, false, false},
		{"let foo; export {foo as foo/* comment*/};", "let foo; export {foo/* comment*/};", false, false, false},
		{"let foo, bar; export {foo as foo/* comment*/,bar};", "let foo, bar; export {foo/* comment*/,bar};", false, false, false},
		{"let foo, bar; export {foo as foo//comment\n,bar};", "let foo, bar; export {foo//comment\n,bar};", false, false, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t, rule_testing.RunWithOptions(t, NoUselessRename,
				uselessRenameFile, testCase.sourceText, uselessRenameOptions(t,
					testCase.ignoreDestructuring, testCase.ignoreImport, testCase.ignoreExport)),
				testCase.wantSource)
		})
	}
}

// The cases upstream reports and deliberately declines to repair.
//
// Each carries output: null upstream for one of two reasons, and both are ported. A comment in
// the discarded part would be deleted by a rewrite the engine applies unattended, and a
// parenthesized left side of a default cannot become a shorthand property, because parentheses
// are not legal there.
//
// Asserted as the ABSENCE of a proposal rather than through ExpectFixedSource, which refuses a
// result carrying no fixes rather than treating it as an unchanged rewrite. The count is what
// separates a withheld repair from one that lands and happens to write the same bytes.
func TestNoUselessRenameDeclinesToFix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText          string
		findings            int
		ignoreDestructuring bool
		ignoreImport        bool
		ignoreExport        bool
	}{
		{"({foo: (foo) = a} = obj);", 1, false, false, false},
		{"({foo/**/ : foo} = {});", 1, false, false, false},
		{"({foo/**/ : foo = 1} = {});", 1, false, false, false},
		{"({foo /**/: foo} = {});", 1, false, false, false},
		{"({foo /**/: foo = 1} = {});", 1, false, false, false},
		{"({foo://\nfoo} = {});", 1, false, false, false},
		{"({foo: /**/foo} = {});", 1, false, false, false},
		{"({foo: (/**/foo)} = {});", 1, false, false, false},
		{"({foo: (foo/**/)} = {});", 1, false, false, false},
		{"({foo: (foo //\n)} = {});", 1, false, false, false},
		{"({foo: /**/foo = 1} = {});", 1, false, false, false},
		{"({foo: (/**/foo) = 1} = {});", 1, false, false, false},
		{"({foo: (foo/**/) = 1} = {});", 1, false, false, false},
		{"import {foo/**/ as foo} from 'foo';", 1, false, false, false},
		{"import {foo /**/as foo} from 'foo';", 1, false, false, false},
		{"import {foo //\nas foo} from 'foo';", 1, false, false, false},
		{"import {foo as/**/foo} from 'foo';", 1, false, false, false},
		{"let foo; export {foo/**/as foo};", 1, false, false, false},
		{"let foo; export {foo as/**/ foo};", 1, false, false, false},
		{"let foo; export {foo as /**/foo};", 1, false, false, false},
		{"let foo; export {foo as//comment\n foo};", 1, false, false, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUselessRename, uselessRenameFile,
				testCase.sourceText, uselessRenameOptions(t, testCase.ignoreDestructuring,
					testCase.ignoreImport, testCase.ignoreExport))
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "unnecessarilyRenamed"
			}
			rule_testing.ExpectFindings(t, result, expected...)
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("expected no repair to be offered, got %d", len(diagnostic.Fixes))
				}
			}
		})
	}
}

// The escape cases still hold real escape sequences on disk.
//
// Five corpus cases turn on a unicode escape being preserved rather than cooked, and a tool that
// cooked one would leave the fixture asserting the opposite of upstream WHILE STILL PASSING,
// because a cooked and an uncooked name compare equal through this rule. So the bytes are checked
// rather than the behaviour: the literal must still contain a backslash-u sequence.
func TestNoUselessRenameCorpusByteCheck(t *testing.T) {
	t.Parallel()

	escapeCases := []string{
		"let {\\u0061: a} = obj;",
		"let {a: \\u0061} = obj;",
		"let {\\u0061: \\u0061} = obj;",
		"import {\\u0061 as a} from 'foo';",
		"import {a as \\u0061} from 'foo';",
		"import {\\u0061 as \\u0061} from 'foo';",
		"var a = 0; export {a as \\u0061};",
		"var \\u0061 = 0; export {\\u0061 as a};",
		"var \\u0061 = 0; export {\\u0061 as \\u0061};",
		"export {a as \\u0061} from 'foo';",
		"export {\\u0061 as a} from 'foo';",
		"export {\\u0061 as \\u0061} from 'foo';",
	}
	if len(escapeCases) != 12 {
		t.Fatalf("expected 12 escape cases, got %d", len(escapeCases))
	}
	for _, sourceText := range escapeCases {
		if !strings.Contains(sourceText, "\\u") {
			t.Errorf("the escape was cooked away in %q, so this case no longer tests what "+
				"upstream asserts", sourceText)
		}
	}
}

// The finding points at the whole property or specifier.
//
// Upstream reports the Property, the ImportSpecifier and the ExportSpecifier, which is the span the
// repair replaces. Measured against the installed rule: `let {foo: foo} = obj;` reports columns 6 to
// 14, which is `foo: foo` and not `foo`. Every message-id assertion above is satisfied by a rule
// reporting only one of the two names, and on a fixable rule the span is what says the repair lands
// where the reader was shown.
func TestNoUselessRenamePointsAtTheWholeRename(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a destructured property", "let {foo: foo} = obj;", "foo: foo"},
		{"an import specifier", "import {foo as foo} from 'foo';", "foo as foo"},
		{"an export specifier", "let foo; export {foo as foo};", "foo as foo"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUselessRename, uselessRenameFile,
				testCase.sourceText, uselessRenameOptions(t, false, false, false))
			rule_testing.ExpectFindings(t, result, "unnecessarilyRenamed")
			source := result.SourceFile.Text()
			reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("expected the finding on %q, pointed at %q", testCase.want, reported)
			}
			if result.Diagnostics[0].Message.Id != "unnecessarilyRenamed" {
				t.Errorf("expected the message id \"unnecessarilyRenamed\", got %q",
					result.Diagnostics[0].Message.Id)
			}
		})
	}
}

// Which of the two names the repair keeps differs between an import and an export.
//
// Upstream writes the text of node.local in both cases, and local is the name AFTER `as` in an
// import and the name BEFORE it in an export. So the same-looking pair of inputs repairs to
// opposite sides, and a port keeping one consistent side passes every message-id assertion while
// writing the wrong text half the time. It did exactly that here until these vectors ran.
//
// Both rows are upstream's own outputs; this test states them side by side because the asymmetry
// reads as an inconsistency when the two are pages apart in a generated table.
func TestNoUselessRenameKeepsTheLocalNameAtBothSites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"an import keeps the name after as",
			"import {'foo' as foo} from 'foo';", "import {foo} from 'foo';"},
		{"an export keeps the name before as",
			"export {'foo' as foo} from 'bar';", "export {'foo'} from 'bar';"},
		{"and the other way round for an export",
			"export {foo as 'foo'} from 'bar';", "export {foo} from 'bar';"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t, rule_testing.RunWithOptions(t, NoUselessRename,
				uselessRenameFile, testCase.sourceText,
				uselessRenameOptions(t, false, false, false)), testCase.wantSource)
		})
	}
}

// A leading comment is not inside the specifier, so the repair is still offered.
//
// Upstream's getCommentsInside asks about the node's own range, where a comment written before the
// specifier is leading trivia rather than content. Our Pos() reaches back over that trivia, so
// counting from it withheld seven of upstream's fix vectors. The rows here pin both directions:
// a comment before the rename fixes, a comment inside the discarded half does not.
func TestNoUselessRenameCountsOnlyContainedComments(t *testing.T) {
	t.Parallel()

	t.Run("a comment before the specifier still fixes", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFixedSource(t, rule_testing.RunWithOptions(t, NoUselessRename,
			uselessRenameFile, "import {/* comment */foo as foo} from 'foo';",
			uselessRenameOptions(t, false, false, false)),
			"import {/* comment */foo} from 'foo';")
	})

	t.Run("a comment inside the discarded half withholds the repair", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, NoUselessRename, uselessRenameFile,
			"({foo/**/ : foo} = {});", uselessRenameOptions(t, false, false, false))
		rule_testing.ExpectFindings(t, result, "unnecessarilyRenamed")
		if len(result.Diagnostics[0].Fixes) != 0 {
			t.Errorf("expected no repair over a discarded comment, got %d",
				len(result.Diagnostics[0].Fixes))
		}
	})
}

// The decoder's own shapes, which no source fixture can reach.
//
// All three options default to false, so the zero value is right here and the usual
// default-inversion hazard does not apply. They are decoded through pointers anyway, and this test
// is what would notice if a default moved and the pointers were dropped as redundant.
func TestDecodeNoUselessRenameOptions(t *testing.T) {
	t.Parallel()

	t.Run("an absent option leaves all three off", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoUselessRenameOptions(nil)
		if err != nil {
			t.Fatalf("the decoder refused an absent option: %v", err)
		}
		settings := decoded.(NoUselessRenameSettings)
		if settings.IgnoreDestructuring || settings.IgnoreImport || settings.IgnoreExport {
			t.Errorf("expected every option to default to false, got %+v", settings)
		}
	})

	t.Run("each option is read independently", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoUselessRenameOptions(
			json.RawMessage(`{"ignoreImport":true}`))
		if err != nil {
			t.Fatalf("the decoder refused a partial object: %v", err)
		}
		settings := decoded.(NoUselessRenameSettings)
		if !settings.IgnoreImport || settings.IgnoreDestructuring || settings.IgnoreExport {
			t.Errorf("expected only ignoreImport to be set, got %+v", settings)
		}
	})

	t.Run("a bare severity leaves the rule on its defaults", func(t *testing.T) {
		t.Parallel()
		// A rule configured as "error" is handed nil options, which arrives at Run as an untyped
		// nil rather than as settings. Asserting through the rule covers the fallback inside Run,
		// which no decoder test can reach.
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoUselessRename,
			uselessRenameFile, "let {foo: foo} = obj;", nil), "unnecessarilyRenamed")
	})
}

// An ordinary object literal is not a destructuring pattern, and only the context says which.
//
// Our parser produces an object LITERAL for `({foo: foo} = obj)` and only the surrounding assignment
// makes it a pattern, where upstream's parser reinterprets and hands the rule an ObjectPattern. So
// this port has to ask whether the literal is an assignment target, and a port that skipped the
// question would report every `{foo: foo}` written as a value.
//
// Measured against the installed rule: the first three rows are clean upstream and only the fourth
// reports. Written for a surviving mutant, since the imported corpus does not pair a reporting
// destructuring assignment with a non-target literal of the same shape.
func TestNoUselessRenameIgnoresAnOrdinaryObjectLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"a parenthesized literal that is not assigned", "({foo: foo});", nil},
		{"a literal used as an initializer", "let x = {foo: foo};", nil},
		{"a literal passed as an argument", "foo({bar: bar});", nil},
		{"the same shape as an assignment target", "({foo: foo} = obj);",
			[]string{"unnecessarilyRenamed"}},
		{"a nested target inside a real pattern", "({a: {foo: foo}} = obj);",
			[]string{"unnecessarilyRenamed"}},
		{"a for-of target", "for ({foo: foo} of list) {}", []string{"unnecessarilyRenamed"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoUselessRename,
				uselessRenameFile, testCase.sourceText,
				uselessRenameOptions(t, false, false, false)), testCase.messages...)
		})
	}
}
