package oracle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const rmOptionsFixture = "internal/oracle/testdata/node_fs_rm_options.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{rmOptionsFixture, true, false})
}

// Probe Node's actual validator before checking that unsupported options cannot
// reach either backend. Every row starts with a real directory and retains it.
func TestNodeFileSystemRmOptionsValidation(test *testing.T) {
	test.Parallel()
	for _, key := range []string{"recursive", "force", "maxRetries", "retryDelay"} {
		values := []string{"undefined", "null", "'yes'"}
		if key == "recursive" || key == "force" {
			values = append(values, "0")
		} else {
			values = append(values, "true", "-1", "0.5")
		}
		for _, value := range values {
			test.Run(key+"/"+value, func(test *testing.T) {
				test.Parallel()
				directory := filepath.Join(test.TempDir(), "directory")
				options := "{" +
					strings.ReplaceAll("recursive:true,force:true,", key+":true,", "") +
					key +
					":" +
					value +
					"}"
				code := `const fs=require('node:fs'); const dir=` +
					strconv.Quote(directory) +
					`; fs.mkdirSync(dir); try { fs.rmSync(dir,` +
					options +
					`); console.log('removed'); } catch(e) { console.log(e.name+' '+e.code+' '+fs.existsSync(dir)); }`
				result := execute(test, "node", "-e", code)
				class, codeName := "TypeError", "ERR_INVALID_ARG_TYPE"
				if value == "-1" || value == "0.5" {
					class, codeName = "RangeError", "ERR_OUT_OF_RANGE"
				}
				want := class + " " + codeName + " true\n"
				if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != want {
					test.Fatalf(
						"Node validator: got exit %d stdout %q stderr %q, want %q",
						result.exitCode,
						result.stdout,
						result.stderr,
						want,
					)
				}
				// Node's declarations reject wrong types before lowering. Undefined is
				// declared legal, so it must reach a named NotYet instead of a default.
				path := filepath.Join(test.TempDir(), "main.a")
				source := "import {rmSync} from 'node:fs'; rmSync(" + strconv.Quote(directory) + "," + options + ");"
				if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
					test.Fatal(writeError)
				}
				checked, loadError := load.Load([]string{path})
				if loadError != nil {
					if value == "undefined" || value == "-1" || value == "0.5" {
						test.Fatalf("expected declared options to reach lowering: %v", loadError)
					}
					if !strings.Contains(loadError.Error(), "not assignable") {
						test.Fatalf("want a type mismatch, not an unrelated checker failure: %v", loadError)
					}
					test.Logf("Node %s %s; checker rejects wrong type", class, codeName)
					return
				}
				_, lowerError := lower.Lower(context.Background(), checked)
				var missing *lower.NotYet
				if !errors.As(lowerError, &missing) || !strings.Contains(missing.What, "rmSync") {
					test.Fatalf("Node %s %s: want named rmSync NotYet, got %v", class, codeName, lowerError)
				}
				if value == "undefined" && (!strings.Contains(missing.What, "options."+key) ||
					!strings.Contains(missing.What, "ERR_INVALID_ARG_TYPE")) {
					test.Fatalf("want present-undefined refusal naming %s and Node's code, got %v", key, lowerError)
				}
				test.Logf("Node %s %s; %s", class, codeName, missing.What)
			})
		}
	}
}

func TestNodeFileSystemRmOptionsUndefinedBinding(test *testing.T) {
	test.Parallel()
	for _, key := range []string{"recursive", "force", "maxRetries", "retryDelay"} {
		test.Run(key, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), "main.a")
			options := "{" + strings.ReplaceAll("recursive:true,force:true,", key+":true,", "") + key + ":undefined}"
			source := fmt.Sprintf(
				"import {rmSync} from 'node:fs'; const options=%s; rmSync('missing',options);",
				options,
			)
			if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			_, loweredError := lowered(test, path)
			var missing *lower.NotYet
			if !errors.As(loweredError, &missing) || !strings.Contains(missing.What, "rmSync") {
				test.Fatalf("want named rmSync refusal, got %v", loweredError)
			}
		})
	}
}

func TestNodeFileSystemRmForceUndefinedReproducer(test *testing.T) {
	test.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/fs_rm_refused/oct9_native_p32_fs_rm_force_undefined.a")
	_, loweredError := lowered(test, path)
	var missing *lower.NotYet
	if !errors.As(loweredError, &missing) || !strings.Contains(missing.What, "options.force present as undefined") {
		test.Fatalf("want force undefined refusal, got %v", loweredError)
	}
	root := test.TempDir()
	result := onNodeWith(test, inputRun{arguments: []string{root}}, path)
	if result.exitCode != 70 ||
		string(result.stdout) != "" ||
		!strings.Contains(string(result.stderr), `"options.force"`) ||
		!strings.Contains(string(result.stderr), "Received undefined") {
		test.Fatalf("Node reproducer: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	entries, readDirectoryError := os.ReadDir(root)
	if readDirectoryError != nil || len(entries) != 1 || !entries[0].IsDir() {
		test.Fatalf("Node must retain directory: entries=%v error=%v", entries, readDirectoryError)
	}
	test.Log("recreated oct9 reproducer: Node rejects before removal; lowering refuses before either backend")
}

func TestNodeFileSystemRmOptionsCounts(test *testing.T) {
	test.Parallel()
	_, measure := counted(test, rmOptionsFixture, false, nil, false, false)
	row := measure()
	test.Log(row)
	recorded, recordedCountsError := recordedCounts()
	if recordedCountsError != nil {
		test.Fatal(recordedCountsError)
	}
	if recorded[rmOptionsFixture] != row {
		test.Fatalf("rm options counts not recorded: %s", row)
	}
}

func TestNodeFileSystemRmOptionsAgreesWithNode(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, rmOptionsFixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	truth := onNode(test, path)
	want := "no-recursive=false\nno-force=false\nno-retries=false\nno-delay=false\nempty-options=false\nno-options=false\nError\nSystemError\nremoved\n"
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != want {
		test.Fatalf("Node: %+v", truth)
	}
	sanitized, binary := natively(test, program)
	results := map[string]run{
		"native release":   released(test, program),
		"native sanitized": sanitized,
		"JavaScript":       onJavaScriptBackend(test, program),
	}
	for backend, got := range results {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(truth, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		test.Run("wasm32-wasi", func(test *testing.T) {
			test.Parallel()
			actual, reason := observeWasi(test, native.C(program))
			if assertWasiRefusal(test, rmOptionsFixture, reason, wasiRefusedFixtures) {
				return
			}
			if difference := disagreement(truth, actual); difference != "" {
				test.Fatal(difference)
			}
		})
	}
}
