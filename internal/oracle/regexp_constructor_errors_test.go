package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const regexpConstructorErrorsFixture = "internal/oracle/testdata/regexp_constructor_errors.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{regexpConstructorErrorsFixture, true, false})
}

func TestRegExpConstructorErrorsWASI(t *testing.T) {
	t.Parallel()
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("WASI_SYSROOT is required")
	}
	path, _ := filepath.Abs(filepath.Join(repository, regexpConstructorErrorsFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if diff := disagreement(expected, onWASI(t, native.C(program))); diff != "" {
		t.Fatal(diff)
	}
}

func TestRegExpConstructorErrorsNodeMutants(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, regexpConstructorErrorsFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	for _, mutant := range []struct{ name, file, old, new string }{
		{"wrong-message", "regexp_compile_runtime.c", "memcpy((char *)text->bytes,message,length);", "memcpy((char *)text->bytes,message,length); if(length>0)((char *)text->bytes)[0]='!';"},
		{"accept-duplicate-flags", "regexp_compile_parser.c", "byte >= 128 || seen[byte]", "byte >= 128"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", mutant.file))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(data), mutant.old) != 1 {
				t.Fatal("mutation site moved")
			}
			changed := strings.Replace(string(data), mutant.old, mutant.new, 1)
			source := strings.Replace(native.C(program), `#include "`+mutant.file+`"`, changed, 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant failed outside Node comparison: %+v", actual)
			}
			if bytes.Equal(expected.stdout, actual.stdout) {
				t.Fatal("Node did not catch mutant")
			}
			t.Logf("Node caught %s: stdout differs", mutant.name)
		})
	}
}

// Not parallel: updates the shared internal/oracle/counts.md file with -update-counts.
func TestRegExpConstructorErrorsCounts(t *testing.T) {
	row := counted(t, regexpConstructorErrorsFixture, false, nil, false, false)
	t.Log(row)
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, row+"\n") {
		return
	}
	if !*updateCounts {
		t.Fatalf("constructor errors counts missing or changed: %s", row)
	}
	if strings.Contains(text, "| "+regexpConstructorErrorsFixture+" |") {
		t.Fatal("existing row changed; review it explicitly")
	}
	at := strings.Index(text, "| internal/oracle/testdata/regexp_constructor_syntax.a |")
	if at < 0 {
		t.Fatal("regexp fixture boundary missing")
	}
	if err := os.WriteFile(countsPath, []byte(text[:at]+row+"\n"+text[at:]), 0644); err != nil {
		t.Fatal(err)
	}
}
