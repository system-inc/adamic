package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestCorrectnessNoLoadTimeImportMetaPathFires pins each shape that reads a path while the module
// loads, with the exact text reported, so a finding on the wrong node fails rather than passing on its
// id alone.
func TestCorrectnessNoLoadTimeImportMetaPathFires(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		source  string
		reports []string
	}{
		// The shape that took the dev server down (ahra 36b7fe07).
		"a module-scope join": {
			"import * as NodePath from 'node:path';\nconst directory = NodePath.join(import.meta.dirname, 'manifests');\n",
			[]string{"import.meta.dirname"},
		},
		"an exported constant": {
			"import * as NodePath from 'node:path';\nexport const Generated = NodePath.join(import.meta.dirname, 'a.ts');\n",
			[]string{"import.meta.dirname"},
		},
		// Undefined at load and thrown later, wherever it is joined.
		"a bare read into a constant": {"const directory = import.meta.dirname;\n", []string{"import.meta.dirname"}},
		"filename into a constant":    {"export const file = import.meta.filename;\n", []string{"import.meta.filename"}},
		"a top-level call's argument": {"console.log(import.meta.dirname);\n", []string{"import.meta.dirname"}},
		"a top-level statement":       {"if (import.meta.dirname) { run(); }\n", []string{"import.meta.dirname"}},
		"an element access":           {"const directory = import.meta['dirname'];\n", []string{"import.meta['dirname']"}},
		"parenthesized meta":          {"const directory = (import.meta).dirname;\n", []string{"(import.meta).dirname"}},
		"a function invoked on the spot": {
			"const directory = (() => import.meta.dirname)();\n",
			[]string{"import.meta.dirname"},
		},
		"a function expression invoked on the spot": {
			"const directory = (function() { return import.meta.dirname; })();\n",
			[]string{"import.meta.dirname"},
		},
		"a static field": {"class Paths { static root = import.meta.dirname; }\n", []string{"import.meta.dirname"}},
		"a static block": {"class Paths { static { register(import.meta.dirname); } }\n", []string{"import.meta.dirname"}},
		"a computed method name, evaluated when the class is defined": {
			"class Paths { [import.meta.filename]() {} }\n",
			[]string{"import.meta.filename"},
		},
		"a decorator on a method's parameter, run when the class is defined": {
			"class Paths { method(@inject(import.meta.dirname) value: string) {} }\n",
			[]string{"import.meta.dirname"},
		},
		"a namespace body": {"namespace Paths { export const root = import.meta.dirname; }\n", []string{"import.meta.dirname"}},
		// The guard's exemption is filename against process.argv; dirname there is an ordinary read.
		"dirname compared with process.argv": {
			"if (process.argv[1] === import.meta.dirname) { main(); }\n",
			[]string{"import.meta.dirname"},
		},
		"filename compared with something else": {
			"if (import.meta.filename === entry) { main(); }\n",
			[]string{"import.meta.filename"},
		},
		"two reads in one statement": {
			"const both = [import.meta.dirname, import.meta.filename];\n",
			[]string{"import.meta.dirname", "import.meta.filename"},
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, CorrectnessNoLoadTimeImportMetaPath, "probe.ts", testCase.source)
			wantIds := make([]string, len(testCase.reports))
			for index := range wantIds {
				wantIds[index] = "loadTimeImportMetaPath"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
			for index, diagnostic := range result.Diagnostics {
				reported := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.reports[index] {
					t.Errorf("finding %d reported %q, want %q", index, reported, testCase.reports[index])
				}
			}
		})
	}
}

func TestCorrectnessNoLoadTimeImportMetaPathStaysSilent(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		// The repair: resolved when called (ahra 36b7fe07, f43214af).
		"a function body": "import * as NodePath from 'node:path';\n" +
			"export function suiteManifestDirectory(): string {\n    return NodePath.join(import.meta.dirname, 'manifests');\n}\n",
		"an arrow body":           "const directory = () => import.meta.dirname;\n",
		"a function expression":   "const directory = function() { return import.meta.dirname; };\n",
		"a method":                "class Paths { root() { return import.meta.dirname; } }\n",
		"an object method":        "const paths = { root() { return import.meta.dirname; } };\n",
		"a getter":                "class Paths { get root() { return import.meta.dirname; } }\n",
		"a constructor":           "class Paths { root: string; constructor() { this.root = import.meta.dirname; } }\n",
		"an instance field":       "class Paths { root = import.meta.dirname; }\n",
		"a parameter default":     "function root(directory = import.meta.dirname) { return directory; }\n",
		"a callback passed along": "process.on('exit', () => console.log(import.meta.dirname));\n",
		// A function that is called later is deferred even when its call is at module scope: the body
		// runs at load, but that is the call's business, which the rule cannot follow.
		"a declared function called at load": "function root() { return import.meta.dirname; }\nroot();\n",
		// The main-module guard: false in a bundle, which is what it should be there.
		"the main-module guard": "import * as NodePath from 'node:path';\n" +
			"if (process.argv[1] && NodePath.resolve(process.argv[1]) === import.meta.filename) {\n    main();\n}\n",
		"the guard, plain":            "if (process.argv[1] === import.meta.filename) { main(); }\n",
		"the guard, reversed":         "if (import.meta.filename !== process.argv[1]) { help(); }\n",
		"the guard, loose equality":   "if (process.argv[1] == import.meta.filename) { main(); }\n",
		"the guard, through realpath": "if (realpathSync(process.argv[1]) === import.meta.filename) { main(); }\n",
		"url, which a bundle keeps":   "const here = new URL('.', import.meta.url);\n",
		"another meta property":       "const environment = import.meta.env;\n",
		"new.target is not import":    "function Paths() { return new.target; }\n",
		"the name on another object":  "const directory = paths.dirname;\n",
		"the words in a string":       "const message = 'import.meta.dirname at module scope';\n",
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, CorrectnessNoLoadTimeImportMetaPath, "probe.ts", source)
			rule_testing.ExpectClean(t, result)
		})
	}
}
