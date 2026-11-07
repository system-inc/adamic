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
	{
		"fs.openSync('', 'r')",
		"",
		"(void)adamic_fs_file_open(&path,&flag,0666);",
		"if (path->length == 0) { system_error(ENOENT, \"open\", name);",
		"if (path->length == 0) { system_error(EEXIST, \"open\", name);",
	},
	{
		"fs.readFileSync('.')",
		".",
		"adamic_release(adamic_fs_file_read_file(&path,&flag));",
		"S_ISDIR(information.st_mode)) { return EISDIR; }",
		"S_ISDIR(information.st_mode)) { return ENOENT; }",
	},
	{
		"console.log(fs.existsSync(''))",
		"",
		"puts(adamic_fs_file_exists(&path) ? \"true\" : \"false\");",
		"if (path->length == 0) { free(name); return false; }",
		"if (path->length == 0) { free(name); return true; }",
	},
	{
		"fs.statSync('')",
		"",
		"adamic_release(adamic_fs_file_stat(&path,true));",
		"if (throw_if_missing) { system_error(ENOENT, \"stat\", name); }",
		"if (throw_if_missing) { system_error(EEXIST, \"stat\", name); }",
	},
	{
		"fs.mkdirSync('')",
		"",
		"adamic_release(adamic_fs_file_mkdir(&path,false,0777));",
		"if (path->length == 0) { system_error(ENOENT, \"mkdir\", name);",
		"if (path->length == 0) { system_error(EEXIST, \"mkdir\", name);",
	},
}

// Same API and Node witness on both ABIs. The wasm unit also compiles the real
// runtime implementation under renamed symbols; its mutant must run cleanly.
func wasiFileAdapter(test *testing.T, index int, wasm bool) {
	test.Helper()
	if wasm && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	probe := wasiFileAdapters[index]
	path := filepath.Join(test.TempDir(), "main.a")
	source := `import * as fs from 'node:fs'; try {` + probe.node + `} catch(error) {console.log(error.code);}`
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	truth := onNodeWith(test, inputRun{}, path)
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
		binary := filepath.Join(test.TempDir(), "program")
		if buildError := native.Build(code, binary, options); buildError != nil {
			test.Fatal(buildError)
		}
		if wasm {
			return execute(
				test,
				"node",
				"--disable-warning=ExperimentalWarning",
				filepath.Join(repository, "oracle/wasi.mjs"),
				binary,
			)
		}
		return executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	}
	if difference := disagreement(truth, observe(code)); difference != "" {
		test.Fatal(difference)
	}
	if !wasm {
		return
	}
	data, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_fs_file.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	implementation := wasiFileMutantImplementation(test, data)
	if strings.Count(implementation, probe.before) != 1 {
		test.Fatal("adapter mutant anchor changed")
	}
	implementation = strings.Replace(implementation, probe.before, probe.after, 1)
	symbols := regexp.MustCompile(`\badamic_fs_file_[A-Za-z_]+`)
	implementation = symbols.ReplaceAllString(implementation, "mutant_$0")
	mutant := implementation + "\n" + symbols.ReplaceAllString(code, "mutant_$0")
	bad := observe(mutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		test.Fatalf("mutant must fail only Node comparison: %+v", bad)
	}
	test.Logf("adapter %d mutant compiled and ran cleanly; only Node stdout caught it", index)
}

func TestWASIHostBuildBoundaryNodeMutant(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(
		path,
		[]byte(`import process from 'node:process'; console.log(process.pid > 0 ? "true" : "false");`),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	truth := onNodeWith(test, inputRun{}, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	code := native.C(program)
	var refused *native.TargetRefused
	buildError := native.Build(
		code,
		filepath.Join(test.TempDir(), "refused.wasm"),
		native.Options{Target: "wasm32-wasi"},
	)
	if !errors.As(buildError, &refused) || refused.Member != "process.pid" {
		test.Fatalf("want pid target refusal: %v", buildError)
	}
	if strings.Count(code, "adamic_node_pid()") != 1 {
		test.Fatal("pid mutant anchor changed")
	}
	code = "#include \"adamic.h\"\nstatic double fake_pid(void) {return 0;}\n" +
		strings.Replace(code, "adamic_node_pid()", "fake_pid()", 1)
	bad := onWASI(test, code)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		test.Fatalf("fake capability must fail only Node comparison: %+v", bad)
	}
	test.Log("fake WASI process id compiled and ran cleanly; only Node stdout caught it")
}

func TestWASIHostFileAdapter_000(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 0, true) }

func TestNativeHostFileAdapter_000(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 0, false) }

func TestWASIHostFileAdapter_001(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 1, true) }

func TestNativeHostFileAdapter_001(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 1, false) }

func TestWASIHostFileAdapter_002(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 2, true) }

func TestNativeHostFileAdapter_002(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 2, false) }

func TestWASIHostFileAdapter_003(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 3, true) }

func TestNativeHostFileAdapter_003(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 3, false) }

func TestWASIHostFileAdapter_004(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 4, true) }

func TestNativeHostFileAdapter_004(test *testing.T) { test.Parallel(); wasiFileAdapter(test, 4, false) }

func TestWASIHostTimestampPrecisionNodeMutant(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	directory := test.TempDir()
	file := filepath.Join(directory, "timestamp")
	path := filepath.Join(directory, "main.a")
	if writeError := os.WriteFile(path, []byte(
		`import * as fs from 'node:fs'; const path=process.argv[2]; fs.writeFileSync(path,''); fs.utimesSync(path,1000.5,1000.5); console.log(fs.statSync(path).mtimeMs === 1000500 ? 'true' : 'false');`,
	), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	truth := onNodeWith(test, inputRun{arguments: []string{file}}, path)
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
	if buildError := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); buildError != nil {
		test.Fatal(buildError)
	}
	actual := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/wasi.mjs"),
		binary,
	)
	if actual.exitCode != 70 ||
		wasiRefusalReason(actual) != "fs.utimesSync timestamp precision or range is unavailable" {
		test.Fatalf("want precision refusal: %+v", actual)
	}
	data, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_fs_file.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	implementation := wasiFileMutantImplementation(test, data)
	base := strings.Index(implementation, "static double fs_utimes")
	if base < 0 {
		test.Fatal("timestamp function missing")
	}
	start := base + strings.Index(implementation[base:], "#ifdef ADAMIC_TARGET_WASI")
	end := strings.Index(implementation[start:], "#endif\n") + start + len("#endif\n")
	if start < 0 || end <= start {
		test.Fatal("timestamp mutant anchor changed")
	}
	implementation = implementation[:start] + implementation[end:]
	symbols := regexp.MustCompile(`\badamic_fs_file_[A-Za-z_]+`)
	implementation = symbols.ReplaceAllString(implementation, "mutant_$0")
	bad := onWASI(test, implementation+"\n"+symbols.ReplaceAllString(source, "mutant_$0"))
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		test.Fatalf("silent precision loss must fail only Node comparison: %+v", bad)
	}
	test.Log("removed precision guard compiled and ran cleanly; only Node stdout caught silent rounding")
}

// The private include defines public fs entry points too. Expand it before
// renaming so each probe owns one complete, renamed implementation. Kept at
// the end so existing optional-WASI skip sites retain their source positions.
func wasiFileMutantImplementation(test *testing.T, data []byte) string {
	test.Helper()
	const include = `#include "node_watch_file_impl.h"`
	if strings.Count(string(data), include) != 1 {
		test.Fatal("watch implementation include changed")
	}
	header, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_watch_file_impl.h"))
	if readError != nil {
		test.Fatal(readError)
	}
	return strings.Replace(string(data), include, string(header), 1)
}
