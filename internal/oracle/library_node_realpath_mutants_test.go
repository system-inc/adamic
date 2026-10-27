package oracle

import (
	"bytes"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Mutants remain valid C and must finish cleanly. Only their output differs from
// the same fixture's Node truth, not a compiler diagnostic or sanitizer failure.
func realpathMutant(t *testing.T, fixture string, change func(*ir.Program, string) string) {
	t.Helper()
	path, _ := filepath.Abs(filepath.Join("testdata/host_realpath", fixture))
	cwd := realpathDirectory(t)
	loader, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	truth := realpathRun(t, cwd, "node", "--disable-warning=ExperimentalWarning", loader, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	original := native.C(program)
	source := change(program, original)
	if source == original {
		t.Fatal("mutant changed nothing")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := realpathRun(t, cwd, binary)
	if truth.exitCode != 0 || got.exitCode != 0 || len(truth.stderr) != 0 || len(got.stderr) != 0 {
		t.Fatalf("must finish sanitizer-clean: Node %+v mutant %+v", truth, got)
	}
	if bytes.Equal(truth.stdout, got.stdout) {
		t.Fatal("mutant survived Node comparison")
	}
	t.Logf("caught by stdout: Node %q; mutant %q", truth.stdout, got.stdout)
}

func TestHostRealpathMutantMissingThrow(t *testing.T) {
	t.Parallel()
	realpathMutant(t, "closure.a", func(_ *ir.Program, source string) string {
		return regexp.MustCompile(`if \((\w+) == NULL\) \{ adamic_fs_file_realpath_error\([^;]+; \}`).ReplaceAllString(source, `if ($1 == NULL) { $1 = adamic_decode_utf8((const unsigned char *)"not thrown", 10); }`)
	})
}
func TestHostRealpathMutantClosureIdentity(t *testing.T) {
	t.Parallel()
	realpathMutant(t, "closure.a", func(program *ir.Program, _ string) string {
		target, held := -1, -1
		for i, f := range program.Functions {
			if f.Source.Role == "host_realpath" {
				target = i
			}
		}
		for i, local := range program.Locals {
			if local.Global && local.Name == "host_realpath_value" {
				held = i
			}
		}
		if target < 0 || held < 0 {
			t.Fatal("missing canonical host closure")
		}
		mutateStringExpressions(reflect.ValueOf(program).Elem(), func(value ir.Expression) ir.Expression {
			if read, ok := value.(ir.Read); ok && read.Local == held {
				return ir.MakeClosure{Function: target}
			}
			return value
		})
		return native.C(program)
	})
}
func TestHostRealpathMutantUnresolvedInput(t *testing.T) {
	t.Parallel()
	realpathMutant(t, "paths.a", func(program *ir.Program, _ string) string {
		mutateStringExpressions(reflect.ValueOf(program).Elem(), func(value ir.Expression) ir.Expression {
			if call, ok := value.(ir.NodeFSFile); ok && call.Operation == "host_realpath" {
				return call.Arguments[0]
			}
			return value
		})
		return native.C(program)
	})
}
func TestHostRealpathMutantLexicalSymlink(t *testing.T) {
	t.Parallel()
	realpathMutant(t, "paths.a", func(_ *ir.Program, source string) string {
		source = strings.ReplaceAll(source, "adamic_host_realpath(", "realpath_lexical_mutant(")
		const helper = `
static adamic_string *realpath_lexical_mutant(const adamic_string *path) {
 if (path->length > 0 && path->bytes[0] == '/') { return adamic_fs_file_host_join(1, (adamic_string *const[]){(adamic_string *)path}); }
 adamic_string *cwd = adamic_host_cwd();
 adamic_string *result = adamic_fs_file_host_join(2, (adamic_string *const[]){cwd, (adamic_string *)path});
 adamic_release(cwd);
 return result;
}
`
		return strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+helper, 1)
	})
}
