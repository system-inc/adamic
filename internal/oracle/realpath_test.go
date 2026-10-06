package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A direct fs call holds the adapter as well as both compiler backends. The two mutants
// change only successful path identity or dangling-link errors, and must execute cleanly.
func TestRealPathAgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/realpath.a"))
	if err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Dir(path)}
	const reference = `const fs = require('node:fs');
for (const path of ['walking/link-to-inner', 'walking/dangling', '', 'walking/dangling/..', 'walking/link-to-inner/../link-to-file']) {
 try { console.log('Ok ' + fs.realpathSync(path)); }
 catch (error) {
  if (error.code !== 'ENOENT') throw error;
  console.log('Error cannot resolve path ' + path + ': no such file');
 }
}`
	want := executeInput(t, how, nil, "node", "-e", reference)
	if want.exitCode != 0 || len(want.stderr) != 0 || !strings.Contains(string(want.stdout), "/walking/inner\nError cannot resolve path walking/dangling: no such file\n") {
		t.Fatalf("Node fs.realpathSync did not exercise directory and dangling links: %+v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	got, binary := inputNatively(t, how, program, shared)
	for _, side := range []struct {
		name string
		got  run
	}{{"native", got}, {"Node source", onNodeWith(t, how, path)}, {"backend", inputBackend(t, how, program, shared)}} {
		if difference := disagreement(want, side.got); difference != "" {
			t.Fatalf("%s: %s; want %+v; got %+v", side.name, difference, want, side.got)
		}
	}
	if leaked := inputLeaks(t, how, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	t.Logf("fs.realpathSync reference: %s", want.stdout)
	for _, kind := range []string{"Ok", "Error"} {
		t.Run("mutant "+kind, func(t *testing.T) {
			code := native.C(program)
			if !strings.Contains(code, "adamic_real_path(") {
				t.Fatal("realpath mutant changed no call")
			}
			code = strings.ReplaceAll(code, "adamic_real_path(", "realpath_mutant(")
			code = insertCollectionMutant(code, `static adamic_object *realpath_mutant(const adamic_string *path) {
 adamic_object *result = adamic_real_path(path);
 static adamic_string selected = ADAMIC_STRING("`+kind+`");
 static adamic_string suffix = ADAMIC_STRING("/wrong");
 if (adamic_string_equal(result->slots[0].reference, &selected)) {
  adamic_string *before = result->slots[1].reference;
  result->slots[1].reference = adamic_string_concat(2, (adamic_string *const[]){before, &suffix});
  adamic_release(before);
 }
 return result;
}`)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeInput(t, how, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must run cleanly: %+v", result)
			}
			if difference := disagreement(want, result); difference != "stdout differs" {
				t.Fatalf("realpath mutant survived or failed outside output comparison: %s", difference)
			}
			t.Logf("clean exit 0, no sanitizer/leak finding; fs.realpathSync stdout comparison caught %s mutant: %s", kind, result.stdout)
		})
	}
}
