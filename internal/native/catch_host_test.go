package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The area host producers keep their code-bearing shapes while changing the
// exception carrier. Node supplies the Error identity, message type and code.
func TestCaughtHostErrorsAndBrandMutant(t *testing.T) {
	const source = `#define ADAMIC_NODE_HOST 1
#include "adamic.h"
#include <stdio.h>
double adamic_fs_file_host_chdir(const adamic_string *directory);
int main(int argc, char **argv) {
 static adamic_string path = ADAMIC_STRING("/adamic-stricter-missing-directory-82407");
 (void)adamic_fs_file_host_chdir(&path);
 if (!adamic_exception_pending || adamic_thrown == NULL) {return 71;}
 static adamic_string message = ADAMIC_STRING("message"), code = ADAMIC_STRING("code");
 adamic_heap *m = adamic_caught_property(adamic_thrown, &message);
 adamic_heap *c = adamic_caught_property(adamic_thrown, &code);
 printf("%s %s %.*s\n", adamic_is_error(adamic_thrown) ? "true" : "false", m != NULL && m->kind == adamic_kind_string ? "string" : "wrong", (int)((adamic_string *)c)->length, ((adamic_string *)c)->bytes);
 adamic_release(m); adamic_release(c); adamic_release(adamic_thrown);
 adamic_thrown = NULL; adamic_exception_pending = false;
 return 0;
}
`
	expected, err := exec.Command("node", "-e", `try {process.chdir('/adamic-stricter-missing-directory-82407')} catch(e) {console.log(e instanceof Error, typeof e.message, e.code)}`).CombinedOutput()
	if err != nil || string(expected) != "true string ENOENT\n" {
		t.Fatalf("Node: %v %s", err, expected)
	}
	directory := t.TempDir()
	baseline := filepath.Join(directory, "baseline")
	if err := Build(source, baseline, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{nil, {"filesystem-host"}} {
		output, err := exec.Command(baseline, arguments...).CombinedOutput()
		if err != nil || string(output) != string(expected) {
			t.Fatalf("host carrier: %v %s", err, output)
		}
	}
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range files {
		if files[index].name == "exceptions.c" {
			text := string(files[index].contents)
			if strings.Count(text, "((const adamic_object *)value)->shape == &host_error_shape || ") != 1 {
				t.Fatal("host brand mutation must identify exactly one site")
			}
			files[index].contents = []byte(strings.Replace(text, "((const adamic_object *)value)->shape == &host_error_shape || ", "", 1))
			changed = true
		}
	}
	if !changed {
		t.Fatal("host brand runtime missing")
	}
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	flags := append(Flags(Options{Sanitize: true}), "-DADAMIC_NODE_HOST=1")
	library, err := cachedRuntime(files, flags, compiler, "host Error brand mutant", filepath.Join(directory, "cache"))
	if err != nil {
		t.Fatal(err)
	}
	input, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "mutant")
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	arguments := append(flags, "-I", filepath.Dir(library), "-o", binary, input)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("mutant must compile: %v %s", err, output)
	}
	for _, arguments := range [][]string{nil, {"filesystem-host"}} {
		output, err := exec.Command(binary, arguments...).CombinedOutput()
		if err != nil || string(output) != "false string ENOENT\n" {
			t.Fatalf("Node must catch brand erasure, not a build/sanitizer failure: %v %s", err, output)
		}
	}
	t.Log("Node catches erased host Error brand in the admitted area filesystem host producer; both mutants compile and finish without sanitizer errors")
}
