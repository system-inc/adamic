package oracle

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

type wasiFSRefusal struct{ member, reason, marker string }

var wasiFSRefusedFixtures = map[string]wasiFSRefusal{
	"open":                {"fs.openSync", "path_open has no mode parameter", "ADAMIC_WASI_FS_OPEN_MODE"},
	"write_file":          {"fs.writeFileSync", "path_open has no mode parameter", "ADAMIC_WASI_FS_WRITE_FILE_MODE"},
	"write_buffer":        {"fs.writeFileSync", "path_open has no mode parameter", "ADAMIC_WASI_FS_WRITE_FILE_MODE"},
	"mkdir":               {"fs.mkdirSync", "path_create_directory has no mode parameter", "ADAMIC_WASI_FS_MKDIR_MODE"},
	"stat":                {"fs.utimesSync", "timestamps are unsigned and Node WASI drops subsecond precision", "ADAMIC_WASI_FS_UTIMES"},
	"utimes":              {"fs.utimesSync", "timestamps are unsigned and Node WASI drops subsecond precision", "ADAMIC_WASI_FS_UTIMES"},
	"buffer":              {"fs.utimesSync", "timestamps are unsigned and Node WASI drops subsecond precision", "ADAMIC_WASI_FS_UTIMES"},
	"system":              {"fs.utimesSync", "timestamps are unsigned and Node WASI drops subsecond precision", "ADAMIC_WASI_FS_UTIMES"},
	"wasi_dynamic_open":   {"fs.openSync", "path_open has no mode parameter", "ADAMIC_WASI_FS_OPEN_MODE"},
	"wasi_dynamic_mkdir":  {"fs.mkdirSync", "path_create_directory has no mode parameter", "ADAMIC_WASI_FS_MKDIR_MODE"},
	"wasi_dynamic_utimes": {"fs.utimesSync", "timestamps are unsigned and Node WASI drops subsecond precision", "ADAMIC_WASI_FS_UTIMES"},
}

func assertWASIFSRefusal(t *testing.T, source, binary string, want wasiFSRefusal) {
	t.Helper()
	err := native.Build(source, binary, native.Options{Target: "wasm32-wasi"})
	var refused *native.TargetRefused
	if !errors.As(err, &refused) || refused.Member != want.member || !strings.Contains(refused.Reason, want.reason) || !strings.Contains(err.Error(), "wasm32-wasi refuses "+want.member) {
		t.Fatalf("expected compile-time WASI refusal for %s (%s), got %v", want.member, want.reason, err)
	}
	if _, err := os.Stat(binary); !os.IsNotExist(err) {
		t.Fatalf("refused source published a binary: %v", err)
	}
	t.Log(err)
}

// Each original runtime-skipping fixture still runs against Node natively and
// in JavaScript via TestNodeFSFileAgreesWithNode. Its WASI leg must stop here,
// without compiling or running a partial executable and without a test skip.
func TestWASIFSCompileTimeRefusals(t *testing.T) {
	t.Parallel()
	for fixture, want := range wasiFSRefusedFixtures {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_file_"+fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			assertWASIFSRefusal(t, native.C(program), filepath.Join(t.TempDir(), "refused.wasm"), want)
		})
	}
}

// Disable one real target-check identifier using a Go overlay. The compiler
// and C still build; only the fixture's refusal assertion may kill this mutant.
// One per original case, including callers sharing the same intrinsic.
func TestWASIFSRefusalMutants(t *testing.T) {
	t.Parallel()
	original, err := filepath.Abs(filepath.Join(repository, "internal/native/host_target.go"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []string{"open", "write_file", "write_buffer", "mkdir", "stat", "utimes", "system"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			want := wasiFSRefusedFixtures[fixture]
			before := []byte(`"` + want.marker + `"`)
			if bytes.Count(source, before) != 1 {
				t.Fatal("target mutation site moved")
			}
			directory := t.TempDir()
			replacement := filepath.Join(directory, "host_target.go")
			changed := bytes.Replace(source, before, []byte(`"`+want.marker+`_MUTANT"`), 1)
			if err := os.WriteFile(replacement, changed, 0600); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(map[string]map[string]string{"Replace": {original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(directory, "overlay.json")
			if err := os.WriteFile(overlay, data, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("go", "test", "-overlay", overlay, "./internal/oracle", "-run", "^TestWASIFSCompileTimeRefusals$/^"+fixture+"$", "-count=1", "-timeout=5m", "-v")
			command.Dir = repository
			output, err := command.CombinedOutput()
			if err == nil || !bytes.Contains(output, []byte("expected compile-time WASI refusal")) || !bytes.Contains(output, []byte("got <nil>")) || bytes.Contains(output, []byte("build failed")) || bytes.Contains(output, []byte("clang failed")) {
				t.Fatalf("target mutant failed outside the refusal assertion: %v\n%s", err, output)
			}
			t.Log("disabled target check caught by refusal assertion; Go and WASI C build, no runtime used")
		})
	}
}

var wasiFSNewFixtures = []string{"wasi_defaults", "wasi_dynamic_open", "wasi_dynamic_mkdir", "wasi_dynamic_utimes"}

// Not parallel: update-counts appends this area's new input rows to the table.
func TestWASIFSCounts(t *testing.T) {
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, name := range wasiFSNewFixtures {
		path := "internal/oracle/testdata/node_fs_file_" + name + ".a"
		row := counted(t, path, true, nil, false, false)
		t.Log(row)
		if strings.Contains(text, row+"\n") {
			continue
		}
		if !*updateCounts {
			t.Errorf("missing or changed counts: %s", row)
			continue
		}
		if strings.Contains(text, "| "+path+" |") {
			t.Fatal("existing counts changed; review explicitly")
		}
		text += row + "\n"
	}
	if *updateCounts && !t.Failed() {
		if err := os.WriteFile(countsPath, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// The supported whole-second path must still have an independent Node oracle.
// This mutant changes a real timestamp, stays in range and leaks nothing.
func TestWASIFSWholeSecondsMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_file_wasi_defaults.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	changed := strings.ReplaceAll(source, "adamic_fs_file_utimes(", "wasi_fs_wrong_time(")
	if changed == source {
		t.Fatal("timestamp mutation site absent")
	}
	helper := `#include "adamic.h"
static double wasi_fs_wrong_time(const adamic_string *path,double atime,double mtime) {
 return adamic_fs_file_utimes(path,atime,mtime == 1700000000 ? mtime+1 : mtime);
}
`
	shared := sharedDirectory(t)
	truth := onNodeWith(t, fsFilePrepare(t, shared, "node"), path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("invalid Node witness: %+v", truth)
	}
	for _, target := range []string{"", "wasm32-wasi"} {
		binary := filepath.Join(shared, "mutant-native")
		options := native.Options{Sanitize: true}
		if target != "" {
			binary = filepath.Join(shared, "mutant.wasm")
			options = native.Options{Target: target}
		}
		if err := native.Build(helper+changed, binary, options); err != nil {
			t.Fatal(err)
		}
		how := fsFilePrepare(t, shared, "input-"+filepath.Base(binary))
		var actual run
		if target == "" {
			actual = executeInput(t, how, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, how.arguments...)
		} else {
			runner := filepath.Join(repository, "oracle/wasi.mjs")
			runner, err = filepath.Abs(runner)
			if err != nil {
				t.Fatal(err)
			}
			actual = executeInput(t, how, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, binary}, how.arguments...)...)
		}
		if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(truth, actual) != "stdout differs" {
			t.Fatalf("timestamp mutant failed outside Node comparison on %q: %+v", target, actual)
		}
		t.Logf("%q whole-second mutation caught only by Node stdout; exit 0, no runtime refusal", target)
	}
}
