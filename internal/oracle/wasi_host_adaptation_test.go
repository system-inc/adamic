package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

var wasiFileAdapters = []struct{ node, path, operation, before, after string }{
	{"fs.openSync('', 'r')", "", "(void)adamic_fs_file_open(&path,&flag,0666);", "if (path->length == 0) { system_error(ENOENT, \"open\", name);", "if (path->length == 0) { system_error(EEXIST, \"open\", name);"},
	{"fs.readFileSync('.')", ".", "adamic_release(adamic_fs_file_read_file(&path,&flag));", "S_ISDIR(information.st_mode)) { return EISDIR; }", "S_ISDIR(information.st_mode)) { return ENOENT; }"},
	{"console.log(fs.existsSync(''))", "", "puts(adamic_fs_file_exists(&path) ? \"true\" : \"false\");", "if (path->length == 0) { free(name); return false; }", "if (path->length == 0) { free(name); return true; }"},
	{"fs.statSync('')", "", "adamic_release(adamic_fs_file_stat(&path,true));", "if (throw_if_missing) { system_error(ENOENT, \"stat\", name); }", "if (throw_if_missing) { system_error(EEXIST, \"stat\", name); }"},
	{"fs.mkdirSync('')", "", "adamic_release(adamic_fs_file_mkdir(&path,false,0777));", "if (path->length == 0) { system_error(ENOENT, \"mkdir\", name);", "if (path->length == 0) { system_error(EEXIST, \"mkdir\", name);"},
}

// Same API and Node witness on both ABIs. The wasm unit also compiles the real
// runtime implementation under renamed symbols; its mutant must run cleanly.
func wasiFileAdapter(t *testing.T, index int, wasm bool) {
	t.Helper()
	if wasm && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	probe := wasiFileAdapters[index]
	path := filepath.Join(t.TempDir(), "main.a")
	source := `import * as fs from 'node:fs'; try {` + probe.node + `} catch(error) {console.log(error.code);}`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNodeWith(t, inputRun{}, path)
	code := fmt.Sprintf(`#include "adamic.h"
#include <stdio.h>
int main(void) {
 adamic_start(0,NULL);
 adamic_string path=ADAMIC_STRING(%q),flag=ADAMIC_STRING("r");
 %s
 if(adamic_thrown!=NULL) {
  adamic_string *code=adamic_thrown->slots[2].reference;
  printf("%%.*s\n",(int)code->length,code->bytes);
  adamic_release(adamic_thrown); adamic_thrown=NULL;
 }
 return 0;
}
`, probe.path, probe.operation)
	options := native.Options{Sanitize: true}
	if wasm {
		options = native.Options{Target: "wasm32-wasi"}
	}
	observe := func(code string) run {
		binary := filepath.Join(t.TempDir(), "program")
		if err := native.Build(code, binary, options); err != nil {
			t.Fatal(err)
		}
		if wasm {
			return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
		}
		return executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	}
	if difference := disagreement(truth, observe(code)); difference != "" {
		t.Fatal(difference)
	}
	if !wasm {
		return
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_fs_file.c"))
	if err != nil {
		t.Fatal(err)
	}
	implementation := string(data)
	if strings.Count(implementation, probe.before) != 1 {
		t.Fatal("adapter mutant anchor changed")
	}
	implementation = strings.Replace(implementation, probe.before, probe.after, 1)
	symbols := regexp.MustCompile(`\badamic_fs_file_[A-Za-z_]+`)
	implementation = symbols.ReplaceAllString(implementation, "mutant_$0")
	mutant := implementation + "\n" + symbols.ReplaceAllString(code, "mutant_$0")
	bad := observe(mutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatalf("mutant must fail only Node comparison: %+v", bad)
	}
	t.Logf("adapter %d mutant compiled and ran cleanly; only Node stdout caught it", index)
}

func TestWASIHostBuildBoundaryNodeMutant(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(`import process from 'node:process'; console.log(process.pid > 0 ? "true" : "false");`), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNodeWith(t, inputRun{}, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	var refused *native.TargetRefused
	err = native.Build(code, filepath.Join(t.TempDir(), "refused.wasm"), native.Options{Target: "wasm32-wasi"})
	if !errors.As(err, &refused) || refused.Member != "process.pid" {
		t.Fatalf("want pid target refusal: %v", err)
	}
	if strings.Count(code, "adamic_node_pid()") != 1 {
		t.Fatal("pid mutant anchor changed")
	}
	code = "#include \"adamic.h\"\nstatic double fake_pid(void) {return 0;}\n" + strings.Replace(code, "adamic_node_pid()", "fake_pid()", 1)
	bad := onWASI(t, code)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatalf("fake capability must fail only Node comparison: %+v", bad)
	}
	t.Log("fake WASI process id compiled and ran cleanly; only Node stdout caught it")
}

func TestWASIHostFileAdapter_000(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 0, true) }

func TestNativeHostFileAdapter_000(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 0, false) }

func TestWASIHostFileAdapter_001(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 1, true) }

func TestNativeHostFileAdapter_001(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 1, false) }

func TestWASIHostFileAdapter_002(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 2, true) }

func TestNativeHostFileAdapter_002(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 2, false) }

func TestWASIHostFileAdapter_003(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 3, true) }

func TestNativeHostFileAdapter_003(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 3, false) }

func TestWASIHostFileAdapter_004(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 4, true) }

func TestNativeHostFileAdapter_004(t *testing.T) { t.Parallel(); wasiFileAdapter(t, 4, false) }

func TestWASIHostTimestampPrecisionNodeMutant(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	directory := t.TempDir()
	file := filepath.Join(directory, "timestamp")
	path := filepath.Join(directory, "main.a")
	if err := os.WriteFile(path, []byte(`import * as fs from 'node:fs'; const path=process.argv[2]; fs.writeFileSync(path,''); fs.utimesSync(path,1000.5,1000.5); console.log(fs.statSync(path).mtimeMs === 1000500 ? 'true' : 'false');`), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNodeWith(t, inputRun{arguments: []string{file}}, path)
	source := fmt.Sprintf(`#include "adamic.h"
#include <stdio.h>
int main(void) {
 adamic_start(0,NULL); adamic_string path=ADAMIC_STRING(%q);
 (void)adamic_fs_file_utimes(&path,1000.5,1000.5);
 adamic_object *stat=adamic_fs_file_stat(&path,true);
 puts(stat->slots[1].number==1000500 ? "true" : "false");
 adamic_release(stat); return 0;
}
`, file)
	// The control must refuse the unsupported precision, never round it silently.
	binary := filepath.Join(directory, "control.wasm")
	if err := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
	if actual.exitCode != 70 || wasiRefusalReason(actual) != "fs.utimesSync timestamp precision or range is unavailable" {
		t.Fatalf("want precision refusal: %+v", actual)
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_fs_file.c"))
	if err != nil {
		t.Fatal(err)
	}
	implementation := string(data)
	base := strings.Index(implementation, "static double fs_utimes")
	if base < 0 {
		t.Fatal("timestamp function missing")
	}
	start := base + strings.Index(implementation[base:], "#ifdef ADAMIC_TARGET_WASI")
	end := strings.Index(implementation[start:], "#endif\n") + start + len("#endif\n")
	if start < 0 || end <= start {
		t.Fatal("timestamp mutant anchor changed")
	}
	implementation = implementation[:start] + implementation[end:]
	symbols := regexp.MustCompile(`\badamic_fs_file_[A-Za-z_]+`)
	implementation = symbols.ReplaceAllString(implementation, "mutant_$0")
	bad := onWASI(t, implementation+"\n"+symbols.ReplaceAllString(source, "mutant_$0"))
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatalf("silent precision loss must fail only Node comparison: %+v", bad)
	}
	t.Log("removed precision guard compiled and ran cleanly; only Node stdout caught silent rounding")
}
