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

func TestNodeRealpathNativeForms(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`import fs from 'node:fs'; const f = fs.realpathSync.native; if (!!f) console.log(f('.')); console.log(String(f !== fs.realpathSync.native));`,
		`import * as _fs from 'node:fs'; console.log(_fs.realpathSync.native('.'));`,
		`import {realpathSync as resolve} from 'node:fs'; console.log(resolve.native('.'));`,
		`import fs from 'node:fs'; const fsRealpath: (path: string) => string = fs.realpathSync.native; console.log(fsRealpath('.'));`,
		`import fs from 'node:fs'; console.log(String(!!fs.realpathSync.native)); if (!!fs.realpathSync.native) console.log('present');`,
		`import fs from 'node:fs'; const fsRealpath: (path: string) => string = !!fs.realpathSync.native ? fs.realpathSync.native : (p: string): string => p; console.log(fsRealpath('.'));`,
		`import fs from 'node:fs'; const a = fs.realpathSync.native; console.log(String(a === fs.realpathSync.native));`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, source)
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, f := range program.Functions {
				if f.Source.Role == "host_realpath" {
					count++
				}
			}
			if count > 1 {
				t.Fatalf("got %d canonical host closures", count)
			}
		})
	}
	if !(ir.NodeFSFile{Operation: "host_realpath"}).MayThrow() {
		t.Fatal("realpath must throw")
	}
}

func TestNodeRealpathNativeRefusals(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ source, reason string }{
		{`fs.realpathSync.native('.', 'utf8');`, "encoding, options, or callback"},
		{`fs.realpathSync.native('.', {encoding:'utf8'});`, "encoding, options, or callback"},
		{`fs.realpathSync.native(...(['.'] as [string]));`, "spread arguments"},
		{`fs.realpathSync.native(Buffer.from('.'));`, "non-string path"},
		{`const f: (this: {tag: number}, path: string) => string = fs.realpathSync.native;`, "needs exactly (path: string) => string"},
		{`const f: (path: string | Buffer) => string = fs.realpathSync.native;`, "needs exactly (path: string) => string"},
		{`const f: (path: string) => string | undefined = fs.realpathSync.native;`, "needs exactly (path: string) => string"},
		{`const f: (path: string, options?: "utf8") => string = fs.realpathSync.native;`, "needs exactly (path: string) => string"},
	} {
		t.Run(one.source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, `import fs from 'node:fs'; import {Buffer} from 'node:buffer'; `+one.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, one.reason) {
				t.Fatalf("want %s, got %v", one.reason, err)
			}
		})
	}
}

// These boundaries are deliberate: neither host binding weakens the checker or
// admits generic Error descriptor reads before the Library follow-up.
func TestNodeRealpathNativeKeepsCheckerBoundaries(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, reason string }{{"raw_if.a", "error TS2774:"}, {"error_fields.a", "in on an Error internal or optional property (host descriptor metadata)"}} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata/realpath/refused", one.name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err == nil {
				_, err = Lower(context.Background(), program)
			}
			if err == nil || !strings.Contains(err.Error(), one.reason) {
				t.Fatalf("want %s, got %v", one.reason, err)
			}
		})
	}
}

func TestNodeRealpathNativeClosedUses(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"map.a": "passed as an argument", "for_each.a": "passed as an argument", "user_argument.a": "passed as an argument",
		"return.a": "returned", "object.a": "stored in an object", "array.a": "stored in an array", "another_binding.a": "stored in another binding",
		"call.a": "reads .call", "apply.a": "reads .apply", "bind.a": "reads .bind", "two_arguments.a": "zero or multiple arguments", "zero_arguments.a": "zero or multiple arguments", "non_string.a": "non-string path",
		"capture_escape.a": "passed as an argument", "export.a": "exported", "export_specifier.a": "exported", "let.a": "let or var", "var.a": "let or var", "spread.a": "spread arguments", "unknown_kind.a": "KindTypeOfExpression",
	}
	for name, reason := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata/realpath/refused", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err == nil {
				_, err = Lower(context.Background(), program)
			}
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, "as a value is admitted only as a const called with one path; this use") || !strings.Contains(gap.What, reason) || !strings.Contains(gap.What, "extra arguments Node reads as options") {
				t.Fatalf("want native use refusal %s, got %v", reason, err)
			}
		})
	}
}

func TestNodeRealpathNativePlainArmIsAdmitted(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/realpath/refused/plain_arm.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Lower(context.Background(), program); err != nil {
		t.Fatal(err)
	}
}

func TestNodeRealpathNativeExportAndOptionsRefusals(t *testing.T) {
	t.Parallel()
	const exported = "node:fs.realpathSync.native as a value is admitted only as a const called with one path; this use is exported could pass it extra arguments Node reads as options"
	const options = "node:fs.realpathSync.native with encoding, options, or callback arguments; need exactly one string path"
	for name, reason := range map[string]string{
		"export_renamed.a": exported,
		"export_default.a": exported,
		"options_call.a":   options,
		"options_string.a": options,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata/realpath/refused", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err == nil {
				_, err = Lower(context.Background(), program)
			}
			var gap *NotYet
			if !errors.As(err, &gap) || gap.What != reason {
				t.Fatalf("want native refusal %q, got %v", reason, err)
			}
		})
	}
}

func TestNodeRealpathPlainUsesAndOptions(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`import fs from 'node:fs'; const f = fs.realpathSync; console.log(f('.')); console.log(String(f === fs.realpathSync)); console.log(String(!!f));`,
		`import fs from 'node:fs'; const f = !!fs.realpathSync.native ? fs.realpathSync.native : fs.realpathSync; console.log(f('.'));`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, source := range []string{
		`import fs from 'node:fs'; fs.realpathSync('.', 'utf8');`,
		`import fs from 'node:fs'; const f = fs.realpathSync; f('.', 'utf8');`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, "options") {
				t.Fatalf("want named options NotYet, got %v", err)
			}
		})
	}
}
