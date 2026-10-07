package lower

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRequireUsesTheFSImportIntrinsic(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const fs = require('fs'); console.log(` + "`${fs.existsSync('.')}`" + `);`,
		`const fs = (require('node:fs')); console.log(` + "`${fs.existsSync('.')}`" + `);`,
		`function check(): boolean { const fs: typeof import('fs') = require('fs'); return fs.existsSync('.'); } console.log(` + "`${check()}`" + `);`,
	} {
		program, err := lowerSource(t, source)
		if err != nil {
			t.Fatal(err)
		}
		if len(program.Locals) != 0 {
			t.Errorf("namespace binding emitted as a runtime local: %+v", program.Locals)
		}
	}
}

func TestRequireAndNamespaceImportHaveTheSameIR(t *testing.T) {
	t.Parallel()
	body := "console.log(fs.existsSync('.') ? 'yes' : 'no');"
	imported, err := lowerSource(t, "import * as fs from 'node:fs';"+body)
	if err != nil {
		t.Fatal(err)
	}
	for _, specifier := range []string{"fs", "node:fs"} {
		required, err := lowerSource(t, "const fs = require('"+specifier+"');"+body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(required, imported) {
			t.Fatalf("require(%q) IR differs from namespace import\nrequire: %#v\nimport: %#v", specifier, required, imported)
		}
	}
}

func TestCommonJSRefusals(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const specifier = 'fs'; require(specifier);`,
		`const specifier = 'fs'; const fs = require(specifier); console.log(fs.existsSync('.') ? 'yes' : 'no');`,
		`require('test');`, `require('sea');`, `require('sqlite');`,
		`require('some-package');`, `require('./local.a');`, `require('node:made-up');`,
		`require.resolve('fs');`, `const cache = require.cache;`,
		`module.exports = {};`, `module['exports'] = {};`, `const later = require;`,
		`require('fs', 'path');`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "use an import") {
				t.Fatalf("want CommonJS refusal and import fix, got %v", err)
			}
		})
	}
}

func TestRequireMissingHostsAreNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{`require('node:crypto');`} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "builtin host module") {
			t.Fatalf("want named host gap, got %v", err)
		}
	}
}

func TestLocalRequireAndModuleAreOrdinaryNames(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function require(name: string): string { return name; } const module = { exports: 'local' }; console.log(require(module.exports));`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequirePathAndImportHaveTheSameIR(t *testing.T) {
	t.Parallel()
	body := "console.log(path.join('a','b')); console.log(path.dirname('a/b'));"
	imported, err := lowerSource(t, "import * as path from 'node:path';"+body)
	if err != nil {
		t.Fatal(err)
	}
	for _, specifier := range []string{"path", "node:path"} {
		required, err := lowerSource(t, "const path = require('"+specifier+"');"+body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(required, imported) {
			t.Fatalf("require(%q) IR differs from import", specifier)
		}
		typed, err := lowerSource(t, "const path: typeof import('path') = require('"+specifier+"');"+body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(typed, imported) {
			t.Fatalf("typed require(%q) IR differs from import", specifier)
		}
	}
}

func TestRequirePerformanceAndImportHaveTheSameIR(t *testing.T) {
	t.Parallel()
	body := `console.log(hooks.performance.now() >= 0 ? 'yes' : 'no');`
	imported, err := lowerSource(t, "import * as hooks from 'node:perf_hooks';"+body)
	if err != nil {
		t.Fatal(err)
	}
	for _, specifier := range []string{"perf_hooks", "node:perf_hooks"} {
		required, err := lowerSource(t, "const hooks = require('"+specifier+"');"+body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(required, imported) {
			t.Fatalf("require(%q) IR differs from import", specifier)
		}
		source := "function obtain(): number { if (true) { try { const { performance } = require('" + specifier + "') as Partial<typeof import('node:perf_hooks')>; if (performance !== undefined) return performance.now(); } catch {} } return -1; } console.log(obtain() >= 0 ? 'yes' : 'no');"
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRequirePerformanceViewCannotEscape(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const hooks = require('perf_hooks') as Partial<typeof import('node:perf_hooks')>;`,
		`const { performance } = require('perf_hooks') as { performance: { now(): number } };`,
		`const { Performance } = require('perf_hooks');`,
	} {
		if _, err := lowerSource(t, source); err == nil {
			t.Fatalf("accepted unsupported module view: %s", source)
		}
	}
}
