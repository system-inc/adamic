package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
			t.Fatalf("%s: %s; want stdout %q, exit %d, stderr %q; got stdout %q, exit %d, stderr %q", side.name, difference, want.stdout, want.exitCode, want.stderr, side.got.stdout, side.got.exitCode, side.got.stderr)
		}
	}
	if leaked := inputLeaks(t, func() inputRun { return how }, program, binary); leaked != "" {
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
			result := executeInput(t, how, append(leakOptions(), "UBSAN_OPTIONS=halt_on_error=1"), binary)
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

// Private trees avoid changing the shared walking fixture or its recorded counts.
func TestRealPathWalk(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for _, name := range []string{"links", "target/inner"} {
		if err := os.MkdirAll(filepath.Join(directory, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "target/file"), []byte("file"), 0644); err != nil {
		t.Fatal(err)
	}
	for name, target := range map[string]string{
		"links/relative": "../target/inner", "links/chain": "relative",
		"links/cycle-a": "cycle-b", "links/cycle-b": "cycle-a",
		"links/middle": "../target",
	} {
		if err := os.Symlink(target, filepath.Join(directory, name)); err != nil {
			t.Fatal(err)
		}
	}
	realPathCompare(t, directory, []string{"links/relative", "links/chain", "links/cycle-a", "links/middle/inner/../file", "target/file/"}, false)
}

func TestRealPathPreservesCase(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "Probe"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, "PROBE")); os.IsNotExist(err) {
		t.Skip("case-sensitive file system: stat PROBE cannot find Probe; libc case mutant cannot be distinguished here")
	} else if err != nil {
		t.Fatal(err)
	}
	realPathCompare(t, directory, []string{"PROBE"}, true)
}

func realPathCompare(t *testing.T, directory string, paths []string, caseMutant bool) {
	t.Helper()
	quoted := make([]string, len(paths))
	for index, path := range paths {
		quoted[index] = fmt.Sprintf("%q", path)
	}
	list := "[" + strings.Join(quoted, ", ") + "]"
	fixture, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/realpath.a"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(fixture)
	start := strings.Index(source, "['walking/")
	end := strings.Index(source[start:], "]") + start + 1
	source = source[:start] + list + source[end:]
	path := filepath.Join(directory, "walk.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: directory}
	reference := `const fs = require('node:fs');
for (const path of ` + list + `) {
 try { console.log('Ok ' + fs.realpathSync(path)); }
 catch (error) {
  if (path === 'links/cycle-a' && error.code !== 'ELOOP') throw error;
  const reason = error.code === 'ENOENT' || error.code === 'ENOTDIR' ? 'no such file' : error.code === 'EACCES' || error.code === 'EPERM' ? 'permission denied' : 'failed';
  console.log('Error cannot resolve path ' + path + ': ' + reason);
 }
}`
	want := executeInput(t, how, nil, "node", "-e", reference)
	if want.exitCode != 0 || len(want.stderr) != 0 {
		t.Fatalf("Node reference: %+v", want)
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
			t.Fatalf("%s: %s; want stdout %q, exit %d, stderr %q; got stdout %q, exit %d, stderr %q", side.name, difference, want.stdout, want.exitCode, want.stderr, side.got.stdout, side.got.exitCode, side.got.stderr)
		}
	}
	if leaked := inputLeaks(t, func() inputRun { return how }, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	t.Logf("fs.realpathSync reference: %s", want.stdout)
	if caseMutant {
		code := strings.ReplaceAll(native.C(program), "adamic_real_path(", "realpath_case_mutant(")
		code = insertCollectionMutant(code, `#include <stdlib.h>
#include <string.h>
extern char *realpath(const char *, char *);
static adamic_object *realpath_case_mutant(const adamic_string *path) {
 adamic_object *result = adamic_real_path(path);
 if (result->slots[0].reference != NULL) {
  char *name = adamic_path_bytes(path);
  char *resolved = realpath(name, NULL);
  free(name);
  if (resolved != NULL) {
   adamic_release(result->slots[1].reference);
   result->slots[1].reference = adamic_decode_utf8((const unsigned char *)resolved, strlen(resolved));
   free(resolved);
  }
 }
 return result;
}`)
		mutant := filepath.Join(shared, "case-mutant")
		if err := native.Build(code, mutant, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := executeInput(t, how, leakOptions(), mutant)
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("mutant must run cleanly: %+v", result)
		}
		if difference := disagreement(want, result); difference != "stdout differs" {
			t.Fatalf("libc mutant survived: %s", difference)
		}
		t.Log("libc realpath mutant caught by case-preserving stdout comparison")
	}
}

// leakOptions asks for LeakSanitizer where there is one: macOS's AddressSanitizer aborts when asked.
func leakOptions() []string {
	if runtime.GOOS == "linux" {
		return []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	return nil
}
