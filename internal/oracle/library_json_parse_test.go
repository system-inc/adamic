package oracle

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"grammar", "scalars", "shapes", "inputs"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/json_parse_" + name + ".a", true, false})
	}
	for _, name := range []string{"array", "buildinfo", "missing", "nested", "literal"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/json_parse_boundary/" + name + ".a", true, true})
	}
}
func TestJSONParseAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"grammar", "scalars", "shapes", "inputs"} {
		t.Run(name, func(t *testing.T) {
			path, binary, script := sanitized(t, "internal/oracle/testdata/json_parse_"+name+".a")
			truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node failed: %+v", truth)
			}
			for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script}} {
				got := execute(t, command[0], command[1:]...)
				if diff := disagreement(truth, got); diff != "" {
					t.Fatalf("%s: Node %q %q backend %q %q", diff, truth.stdout, truth.stderr, got.stdout, got.stderr)
				}
			}
		})
	}
}
func TestJSONParseBoundaryNamedDifference(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/json_parse_boundary/*.a"))
	if err != nil || len(files) < 3 {
		t.Fatal(files, err)
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(source), "// boundary check: Adamic stops, Node continues\n// Adamic: ") {
				t.Fatal("missing known-difference header")
			}
			message := strings.Split(string(source), "\n")[1][len("// Adamic: "):]
			relative, _ := filepath.Rel(repository, path)
			_, binary, script := sanitized(t, relative)
			truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if truth.exitCode != 0 || string(truth.stdout) != "Node continues\n" || len(truth.stderr) != 0 {
				t.Fatalf("unexpected Node behavior: %+v", truth)
			}
			for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script}} {
				got := execute(t, command[0], command[1:]...)
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message+"\n" {
					t.Fatalf("want terminal named difference %q; got %+v", message, got)
				}
			}
		})
	}
}
func TestJSONParseWASI(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, name := range []string{"grammar", "scalars", "shapes"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/oracle/testdata/json_parse_"+name+".a")
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "program.wasm")
			if err = native.Build(native.C(p), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
				t.Fatal(err)
			}
			truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			got := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
			if diff := disagreement(truth, got); diff != "" {
				t.Fatalf("%s Node %q backend %q %q", diff, truth.stdout, got.stdout, got.stderr)
			}
		})
	}
}
func TestJSONParseMutants(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, helper string }{
		{"grammar", `static adamic_value parse_mutant(const adamic_string *text,const adamic_json_parse_schema *check,const adamic_json_parse_schema *layout){adamic_value r=adamic_json_parse(text,check,layout);if(adamic_thrown!=NULL){static adamic_string wrong=ADAMIC_STRING("WrongError");adamic_release(adamic_thrown->slots[0].reference);adamic_thrown->slots[0].reference=&wrong;}return r;}`},
		{"scalars", `static adamic_value parse_mutant(const adamic_string *text,const adamic_json_parse_schema *check,const adamic_json_parse_schema *layout){adamic_value r=adamic_json_parse(text,check,layout);if(check!=NULL&&check->of==1){r.number+=1;}return r;}`},
		{"shapes", `static adamic_value parse_mutant(const adamic_string *text,const adamic_json_parse_schema *check,const adamic_json_parse_schema *layout){adamic_value r=adamic_json_parse(text,check,layout);if(check!=NULL&&check->of==4&&check->kind==json_object&&check->count>0){adamic_object *o=r.reference;for(size_t i=0;i<o->shape->count;i++){if(strcmp(o->shape->names[i],"version")==0&&o->shape->references[i]){static adamic_string wrong=ADAMIC_STRING("wrong");adamic_release(o->slots[i].reference);o->slots[i].reference=&wrong;}}}return r;}`},
	} {
		t.Run(one.name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/oracle/testdata/json_parse_"+one.name+".a")
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(p)
			changed := strings.ReplaceAll(original, "adamic_json_parse(", "parse_mutant(")
			if original == changed {
				t.Fatal("mutant changed nothing")
			}
			changed = strings.Replace(changed, `#include "json_parse.h"`, `#include "json_parse.h"`+"\n#include <string.h>\n"+one.helper, 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err = native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			env := []string{"UBSAN_OPTIONS=halt_on_error=1"}
			if runtime.GOOS == "linux" {
				env = append(env, "ASAN_OPTIONS=detect_leaks=1")
			}
			truth := executeWith(t, env, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			got := executeWith(t, env, binary)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
				t.Fatalf("mutant not caught solely by Node: truth %+v mutant %+v", truth, got)
			}
			t.Log("caught only by Node stdout; ASan/UBSan/leak checks clean")
		})
	}
}

func TestJSONParseSyntaxSweep(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/json_parse_inputs.a")
	random := rand.New(rand.NewSource(22))
	alphabet := []rune("{}[]:,\"\\tfnruea01-+. \t\r\n🌍é")
	documents := []string{"", "\ufeff0", "[\u2028]", "\"\u2028\"", "{\"\\u0000\":7}", "{\"\\ud800\":null}"}
	for i := 0; i < 1800; i++ {
		length := random.Intn(45) + 1
		var b strings.Builder
		for j := 0; j < length; j++ {
			b.WriteRune(alphabet[random.Intn(len(alphabet))])
		}
		documents = append(documents, b.String())
	}
	for _, base := range []string{`{"key":[true,false,null,-0,1.25e-3,"escaped\n🌍"]}`, `[{},[],"text",{"a":1,"a":2}]`} {
		for i := 0; i < len(base); i++ {
			for _, replacement := range []string{"", "x", "\r\n", "\""} {
				documents = append(documents, base[:i]+replacement+base[i+1:])
			}
		}
	}
	for start := 0; start < len(documents); start += 100 {
		end := start + 100
		if end > len(documents) {
			end = len(documents)
		}
		args := documents[start:end]
		truth := execute(t, "node", append([]string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path}, args...)...)
		for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script}} {
			got := execute(t, command[0], append(command[1:], args...)...)
			if diff := disagreement(truth, got); diff != "" {
				for _, document := range args {
					want := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path, document)
					one := execute(t, command[0], append(command[1:], document)...)
					if d := disagreement(want, one); d != "" {
						t.Fatalf("input %q: %s, Node %q %q backend %q %q", document, d, want.stdout, want.stderr, one.stdout, one.stderr)
					}
				}
				t.Fatalf("batch %d: %s", start, diff)
			}
		}
	}
	t.Logf("%d syntax documents agree with Node", len(documents))
}

func TestJSONParseBoundaryMutant(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/json_parse_boundary/buildinfo.a")
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	original := native.C(p)
	changed := strings.ReplaceAll(original, "adamic_json_parse(", "boundary_mutant(")
	helper := `static adamic_value boundary_mutant(const adamic_string *text,const adamic_json_parse_schema *check,const adamic_json_parse_schema *layout){(void)check;(void)layout;return adamic_json_parse(text,NULL,NULL);}`
	changed = strings.Replace(changed, `#include "json_parse.h"`, `#include "json_parse.h"`+"\n"+helper, 1)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err = native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	env := []string{"UBSAN_OPTIONS=halt_on_error=1", "ASAN_OPTIONS=detect_leaks=1"}
	truth := executeWith(t, env, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	got := executeWith(t, env, binary)
	if truth.exitCode != 0 || got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "" {
		t.Fatalf("boundary mutant failed for another reason: Node %+v mutant %+v", truth, got)
	}
	// This is precisely the named difference: removing the check agrees with Node,
	// and is rejected by the fixture's required exit 70 and exact header message.
	t.Log("removing the boundary is caught only by the named-difference header contract; Node continues")
}
