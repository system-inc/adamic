package lower

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestNodeRealpathNativeForms(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`import fs from 'node:fs'; const f = fs.realpathSync.native; if (!!f) console.log(f('.')); console.log(String(f !== fs.realpathSync.native));`,
		`import * as _fs from 'node:fs'; console.log(_fs.realpathSync.native('.'));`,
		`import {realpathSync as resolve} from 'node:fs'; console.log(resolve.native('.'));`,
		`import fs from 'node:fs'; const fsRealpath: (path: string) => string = fs.realpathSync.native; console.log(fsRealpath('.'));`,
		`import fs from 'node:fs'; console.log(String(!!fs.realpathSync.native)); if (!!fs.realpathSync.native) console.log('present');`,
		`import fs from 'node:fs'; const fsRealpath: (path: string) => string = !!fs.realpathSync.native ? fs.realpathSync.native : (p: string): string => p; console.log(fsRealpath('.'));`,
		`import fs from 'node:fs'; const a = fs.realpathSync.native; console.log(String(a === fs.realpathSync.native));`,
	} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			program, err := lowerSource(test, source)
			if err != nil {
				test.Fatal(err)
			}
			count := 0
			for _, f := range program.Functions {
				if f.Source.Role == "host_realpath" {
					count++
				}
			}
			if count > 1 {
				test.Fatalf("got %d canonical host closures", count)
			}
		})
	}
	if !(ir.NodeFSFile{Operation: "host_realpath"}).MayThrow() {
		test.Fatal("realpath must throw")
	}
}

func TestNodeRealpathNativeRefusals(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ source, reason string }{
		{`fs.realpathSync.native('.', 'utf8');`, "encoding, options, or callback"},
		{`fs.realpathSync.native('.', {encoding:'utf8'});`, "encoding, options, or callback"},
		{`fs.realpathSync.native(...(['.'] as [string]));`, "spread arguments"},
		{`fs.realpathSync.native(Buffer.from('.'));`, "non-string path"},
		{
			`const f: (this: {tag: number}, path: string) => string = fs.realpathSync.native;`,
			"needs exactly (path: string) => string",
		},
		{
			`const f: (path: string | Buffer) => string = fs.realpathSync.native;`,
			"needs exactly (path: string) => string",
		},
		{
			`const f: (path: string) => string | undefined = fs.realpathSync.native;`,
			"needs exactly (path: string) => string",
		},
		{
			`const f: (path: string, options?: "utf8") => string = fs.realpathSync.native;`,
			"needs exactly (path: string) => string",
		},
	} {
		test.Run(one.source, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, `import fs from 'node:fs'; import {Buffer} from 'node:buffer'; `+one.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, one.reason) {
				test.Fatalf("want %s, got %v", one.reason, err)
			}
		})
	}
}

// These boundaries are deliberate: neither host binding weakens the checker or
// admits generic Error descriptor reads before the Library follow-up.
func TestNodeRealpathNativeKeepsCheckerBoundaries(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ name, reason string }{
		{"raw_if.a", "error TS2774:"},
		{"error_fields.a", "in on an Error internal or optional property (host descriptor metadata)"},
	} {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata/realpath/refused", one.name))
			if err != nil {
				test.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err == nil {
				_, err = Lower(context.Background(), program)
			}
			if err == nil || !strings.Contains(err.Error(), one.reason) {
				test.Fatalf("want %s, got %v", one.reason, err)
			}
		})
	}
}

func TestNodeRealpathNativeClosedUses(test *testing.T) {
	test.Parallel()
	cases := map[string]string{
		"map.a":              "passed as an argument",
		"for_each.a":         "passed as an argument",
		"user_argument.a":    "passed as an argument",
		"return.a":           "returned",
		"object.a":           "stored in an object",
		"array.a":            "stored in an array",
		"another_binding.a":  "stored in another binding",
		"call.a":             "reads .call",
		"apply.a":            "reads .apply",
		"bind.a":             "reads .bind",
		"two_arguments.a":    "zero or multiple arguments",
		"zero_arguments.a":   "zero or multiple arguments",
		"non_string.a":       "non-string path",
		"capture_escape.a":   "passed as an argument",
		"export.a":           "exported",
		"export_specifier.a": "exported",
		"let.a":              "let or var",
		"var.a":              "let or var",
		"spread.a":           "spread arguments",
		"unknown_kind.a":     "KindTypeOfExpression",
	}
	for name, reason := range cases {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata/realpath/refused", name))
			if err != nil {
				test.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err == nil {
				_, err = Lower(context.Background(), program)
			}
			var gap *NotYet
			if !errors.As(err, &gap) ||
				!strings.Contains(gap.What, "as a value is admitted only as a const called with one path; this use") ||
				!strings.Contains(gap.What, reason) ||
				!strings.Contains(gap.What, "extra arguments Node reads as options") {
				test.Fatalf("want native use refusal %s, got %v", reason, err)
			}
		})
	}
}

func TestNodeRealpathNativePlainArmIsAdmitted(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("testdata/realpath/refused/plain_arm.a")
	if err != nil {
		test.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	if _, err = Lower(context.Background(), program); err != nil {
		test.Fatal(err)
	}
}

func TestNodeRealpathNativeExportAndOptionsRefusals(test *testing.T) {
	test.Parallel()
	const exported = "node:fs.realpathSync.native as a value is admitted only as a const called with one path; this use is exported could pass it extra arguments Node reads as options"
	const options = "node:fs.realpathSync.native with encoding, options, or callback arguments; need exactly one string path"
	for name, reason := range map[string]string{
		"export_renamed.a": exported,
		"export_default.a": exported,
		"options_call.a":   options,
		"options_string.a": options,
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata/realpath/refused", name))
			if err != nil {
				test.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err == nil {
				_, err = Lower(context.Background(), program)
			}
			var gap *NotYet
			if !errors.As(err, &gap) || gap.What != reason {
				test.Fatalf("want native refusal %q, got %v", reason, err)
			}
		})
	}
}

func TestNodeRealpathPlainUsesAndOptions(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`import fs from 'node:fs'; const f = fs.realpathSync; console.log(f('.')); console.log(String(f === fs.realpathSync)); console.log(String(!!f));`,
		`import fs from 'node:fs'; const f = !!fs.realpathSync.native ? fs.realpathSync.native : fs.realpathSync; console.log(f('.'));`,
	} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			if _, err := lowerSource(test, source); err != nil {
				test.Fatal(err)
			}
		})
	}
	for _, source := range []string{
		`import fs from 'node:fs'; fs.realpathSync('.', 'utf8');`,
		`import fs from 'node:fs'; const f = fs.realpathSync; f('.', 'utf8');`,
	} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, "options") {
				test.Fatalf("want named options NotYet, got %v", err)
			}
		})
	}
}

func TestNodeRealpathPlainCallRefusals(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ name, source, reason string }{
		{"options", `fs.realpathSync('.', {encoding: 'utf8'});`, "options"},
		{"two arguments", `fs.realpathSync('.', undefined);`, "options"},
		{"spread", `fs.realpathSync(...(['.'] as [string]));`, "spread"},
		{"Buffer path", `fs.realpathSync(Buffer.from('.'));`, "non-string path"},
		{"URL path", `fs.realpathSync(new URL('file:///tmp/'));`, "non-string path"},
	} {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(
				test,
				`import fs from 'node:fs'; import {Buffer} from 'node:buffer'; import {URL} from 'node:url'; `+
					one.source,
			)
			var gap *NotYet
			if !errors.As(err, &gap) ||
				!strings.Contains(gap.What, "node:fs.realpathSync") ||
				!strings.Contains(gap.What, one.reason) {
				test.Fatalf("want named plain realpath refusal %q, got %v", one.reason, err)
			}
		})
	}
}
