package javascript

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/prettier"
)

// The corpus differential (internal/format/differential, env-gated) is the acceptance test. These run in
// every `go test` instead: each snippet goes through this printer and through the embedded Prettier fork,
// and the two outputs must be equal. Snippets are written unformatted, so the oracle rewrites each one
// and an identity printer cannot pass, and the fork's own customizations are asserted on the output as
// well, so a printer and an oracle that both lost one cannot agree their way past it.

type formatCase struct {
	name     string
	fileName string
	source   string
	// contains are substrings the fork's output must hold: its customizations, spelled out.
	contains []string
}

var formatCases = []formatCase{
	{
		name:     "anonymous function has no space before its parameters (fork 5f95bdb0e)",
		fileName: "Probe.ts",
		source:   "const a = function (x) { return x }\nfunction named (y) { return y }\nconst g = async function* () {}\n",
		contains: []string{"function(x) {", "function named(y) {", "async function*() {}"},
	},
	{
		name:     "catch and finally start their own lines (fork d99223a24)",
		fileName: "Probe.ts",
		source:   "try { a() } catch (error) { b() } finally { c() }\ntry { a() } catch { b() }\ntry { a() } finally { c() }\n",
		contains: []string{"}\ncatch(error) {", "}\nfinally {", "}\ncatch {"},
	},
	{
		name:     "control keywords have no space before their parenthesis",
		fileName: "Probe.ts",
		source: "if (a) { b() } else if (c) { d() } else { e() }\nfor (let i = 0; i < 3; i++) {}\nfor (const x of y) {}\n" +
			"while (a) {}\ndo { a() } while (b)\nswitch (a) { case 1: break; default: }\nfor (;;) {}\n",
		contains: []string{"if(a) {", "}\nelse if(c) {", "}\nelse {", "for(let i", "for(const x of y)", "while(a)", "} while(b);", "switch(a)", "for(;;)"},
	},
	{
		name:     "member chains, call arguments and arrow chains break as the fork breaks them",
		fileName: "Probe.ts",
		source: "const result = items.filter((item) => item.enabled && item.visible).map((item) => item.identifier).reduce((total, identifier) => total + identifier.length, 0);\n" +
			"const curried = (a: number) => (b: number) => (c: number) => (d: number) => a + b + c + d + someVeryLongIdentifierName + anotherVeryLongIdentifierName;\n" +
			"useEffect(() => { subscribe(); return () => unsubscribe(); }, [subscribe, unsubscribe]);\n",
	},
	{
		name:     "types: unions, intersections, mapped, conditional, generics and overloads",
		fileName: "Probe.ts",
		source: "type Union = 'Moonlight' | 'Starlight' | 'Sunlight' | 'Twilight' | 'Daylight' | 'Midnight' | 'Highnoon' | 'Dawn';\n" +
			"type Mapped<T> = { readonly [Key in keyof T]?: T[Key] extends Function ? never : T[Key] };\n" +
			"type Conditional<T> = T extends string ? 'String' : T extends number ? 'Number' : T extends boolean ? 'Boolean' : 'Other';\n" +
			"interface Shape extends Base, Other { area(): number; readonly name?: string; [key: string]: unknown }\n" +
			"export abstract class Square<T extends object = {}> extends Shape implements Area { private readonly side: number = 1; constructor(public value: T) { super() } abstract area(): number; static create<U>(this: void): Square<U> { return null! } }\n" +
			"enum Direction { Up = 1, Down, Left, Right }\ndeclare module 'probe' { export function f(): void }\n",
	},
	{
		name:     "comments keep their places",
		fileName: "Probe.ts",
		source: "// leading\nconst a = 1; // trailing\n/**\n * Documented.\n */\nfunction f(/* inline */ x: number /* after */) {\n  // dangling\n}\n" +
			"const o = {\n  // own line\n  a: 1, // end of line\n  b: 2,\n};\nswitch (a) {\n  // before case\n  case 1:\n}\n",
	},
	{
		name:     "imports, exports and template literals",
		fileName: "Probe.ts",
		source: "import Default, { a, type B, c as d } from './module';\nimport type { E } from \"./types\";\nexport * as namespace from './namespace';\n" +
			"export { a, d as renamed };\nexport default class {}\nconst text = `value ${a + b} and ${ call(c) }`;\nconst tagged = tag`raw ${value}`;\n",
	},
	{
		name:     "JSX: elements, attributes, expressions, conditional rendering and text",
		fileName: "Probe.tsx",
		source: "export function View(properties: { title: string; items: string[] }) { return <div className=\"view\" onClick={() => properties.onClick()} data-long-attribute-name=\"a long attribute value that forces breaking\">\n" +
			"{properties.title ? <h1>{properties.title}</h1> : null}{\" \"}some text that is long enough to make the children fill across lines and wrap at the width\n" +
			"{properties.items.map((item) => <span key={item}>{item}</span>)}<></></div> }\nconst generic = <T,>(value: T) => value;\n",
	},
	{
		name:     "javascript: JSDoc types stay comments, numeric keys unquote under babel, JSX parses",
		fileName: "Probe.js",
		source: "/**\n * @typedef {{ a: number }} Shape\n * @param {string} name\n * @returns {Shape}\n */\nexport function make (name) { return /** @type {Shape} */ ({ 'a': 1, '2': name, 'b-c': 3 }) }\n" +
			"const view = <div className='x'>{make('y')}</div>\n",
		contains: []string{"export function make(name) {", "/** @type {Shape} */ ({ a: 1, 2: name, 'b-c': 3 })", "<div className=\"x\">"},
	},
	{
		name:     "json: comments, trailing commas, quoting and numbers through the JavaScript printer",
		fileName: "settings.json",
		source: "// leading\n{ unquoted: 1, 'single': 'it\\'s', \"nested\": { \"array\": [1.50, -2, +3, 0x1F, 1e3,], /* inside */ \"empty\": {} },\n" +
			"  \"long\": [\"a value long enough\", \"to make this array break across lines\", \"at the print width of one hundred twenty\"], 10: true, n: null, }\n",
		contains: []string{`"unquoted": 1`, `"single": "it's"`, `"10": true`, "/* inside */"},
	},
	{
		name:     "json-stringify: package.json breaks every object and array",
		fileName: "package.json",
		source:   "{\"name\":\"probe\",\"private\":true,\"files\":[],\"scripts\":{\"build\":\"tsc\"},\"keywords\":[\"a\",\"b\"],\"version\":1.50}\n",
		contains: []string{"{\n    \"name\": \"probe\",", "\"files\": [],", "\"keywords\": [\n        \"a\",", "\"version\": 1.50"},
	},
}

// formatFor is the native entry point a file name reaches through internal/format/native.
func formatFor(fileName string) func(string, string, formatoptions.Options) (string, error) {
	switch {
	case strings.HasSuffix(fileName, ".json"):
		return FormatJSON
	case strings.HasSuffix(fileName, ".js"):
		return func(fileName string, text string, options formatoptions.Options) (string, error) {
			return FormatJavaScript(fileName, text, options, nil)
		}
	}
	return func(fileName string, text string, options formatoptions.Options) (string, error) {
		return Format(fileName, text, options, nil)
	}
}

// TestJavaScriptRefusesWhatTypeScriptReadsDifferently: `a < b > (c)` is two comparisons to babel and a
// generic call to TypeScript, so printing the TSX parse would rewrite the program's meaning.
func TestJavaScriptRefusesWhatTypeScriptReadsDifferently(t *testing.T) {
	t.Parallel()
	if _, err := FormatJavaScript("Probe.js", "const result = a < b > (c);\n", formatoptions.Default(), nil); err == nil {
		t.Fatal("a JavaScript file that parses as a TypeScript generic call was printed instead of refused")
	}
	if _, err := FormatJavaScript("Probe.js", "const result = a < b;\n", formatoptions.Default(), nil); err != nil {
		t.Fatalf("plain JavaScript was refused: %v", err)
	}
}

func TestFormatMatchesTheFork(t *testing.T) {
	t.Parallel()
	sameLine := formatoptions.Default()
	sameLine.BracketSameLine = true
	for _, options := range []formatoptions.Options{formatoptions.Default(), sameLine} {
		oracle, err := prettier.New(options)
		if err != nil {
			t.Fatal(err)
		}
		for _, testCase := range formatCases {
			expected, err := oracle.Format(testCase.fileName, testCase.source)
			if err != nil {
				t.Fatalf("%s: the oracle failed: %v", testCase.name, err)
			}
			if expected == testCase.source {
				t.Fatalf("%s: the oracle left the snippet unchanged, so an identity printer would pass", testCase.name)
			}
			for _, substring := range testCase.contains {
				if !strings.Contains(expected, substring) {
					t.Errorf("%s: the fork's output lacks %q:\n%s", testCase.name, substring, expected)
				}
			}
			actual, err := formatFor(testCase.fileName)(testCase.fileName, testCase.source, options)
			if err != nil {
				t.Errorf("%s (bracketSameLine %v): %v", testCase.name, options.BracketSameLine, err)
				continue
			}
			if actual != expected {
				t.Errorf("%s (bracketSameLine %v):\n--- fork\n%s--- native\n%s", testCase.name, options.BracketSameLine, expected, actual)
			}
		}
	}
}
