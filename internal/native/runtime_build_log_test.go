package native

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// loggedSteps sets ADAMIC_BUILD_LOG to a new file for the length of step and returns every line step wrote there, each
// matched as buildLogLine reads it. Not for a parallel test: it sets the process's environment.
func loggedSteps(test *testing.T, step func()) [][]string {
	test.Helper()
	log := filepath.Join(test.TempDir(), "build.log")
	test.Setenv("ADAMIC_BUILD_LOG", log)
	step()
	test.Setenv("ADAMIC_BUILD_LOG", "")
	written, readError := os.ReadFile(log)
	if os.IsNotExist(readError) {
		return nil
	}
	if readError != nil {
		test.Fatal(readError)
	}
	var lines [][]string
	for _, line := range strings.Split(strings.TrimSuffix(string(written), "\n"), "\n") {
		match := buildLogLine.FindStringSubmatch(line)
		if match == nil {
			test.Fatalf("%q isn't a build log line", line)
		}
		lines = append(lines, match)
	}
	return lines
}

// loggedSeconds reads a matched line's seconds.
func loggedSeconds(test *testing.T, match []string) float64 {
	test.Helper()
	seconds, parseError := strconv.ParseFloat(match[5], 64)
	if parseError != nil {
		test.Fatal(parseError)
	}
	return seconds
}

// The runtime's archive writes a cCompile line per unit it compiles and a cArchive line for ar, each a miss under the
// archive's key with the seconds it took, and a cArchive hit when the cache holds it; the archive is the same bytes
// with the log and without it. A mutant that drops the outcome, says hit where it compiled, or writes 0 seconds fails.
//
// Not parallel: it sets ADAMIC_BUILD_LOG for the process.
func TestRuntimeArchiveLogsItsCompilesAndItsHit(test *testing.T) {
	nonce := strconv.FormatInt(time.Now().UnixNano(), 36)
	var files []runtimeFile
	for index := range 4 {
		files = append(files, runtimeFile{
			fmt.Sprintf("unit%d.c", index),
			[]byte(fmt.Sprintf("int unit%d_%s(void) { return %d; }\n", index, nonce, index)),
		})
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		test.Fatal(err)
	}
	quiet := test.TempDir()
	if compileRuntimeError := compileRuntime(
		files,
		Flags(Options{}),
		compiler,
		quiet,
		2,
		nil,
		"0123456789abcdef",
	); compileRuntimeError != nil {
		test.Fatal(compileRuntimeError)
	}
	logged := test.TempDir()
	lines := loggedSteps(test, func() {
		if compileRuntimeError := compileRuntime(
			files,
			Flags(Options{}),
			compiler,
			logged,
			2,
			nil,
			"0123456789abcdef",
		); compileRuntimeError != nil {
			test.Fatal(compileRuntimeError)
		}
	})
	quietArchive, err := os.ReadFile(filepath.Join(quiet, "runtime.a"))
	if err != nil {
		test.Fatal(err)
	}
	loggedArchive, err := os.ReadFile(filepath.Join(logged, "runtime.a"))
	if err != nil {
		test.Fatal(err)
	}
	if !bytes.Equal(quietArchive, loggedArchive) {
		test.Fatal("the archive compiled with ADAMIC_BUILD_LOG set differs from the one compiled without it")
	}
	compiled, archived := map[string]bool{}, 0
	for _, match := range lines {
		if match[3] != "0123456789ab" || match[4] != "miss" || loggedSeconds(test, match) <= 0 {
			test.Fatalf("%q, want a miss under the archive's key with its seconds", match[0])
		}
		switch match[1] {
		case "cCompile":
			compiled[match[2]] = true
		case "cArchive":
			archived++
			if match[2] != "runtime.a" {
				test.Fatalf("%q, want runtime.a", match[0])
			}
		default:
			test.Fatalf("%q from compiling an archive", match[0])
		}
	}
	if len(compiled) != len(files) || !compiled["runtime/unit0.c"] || archived != 1 {
		test.Fatalf(
			"%d units compiled and %d archives logged, want %d and 1: %q",
			len(compiled),
			archived,
			len(files),
			lines,
		)
	}

	cache := test.TempDir()
	for _, want := range []string{"miss", "hit"} {
		var library string
		lines := loggedSteps(test, func() {
			if library, err = cachedRuntime(files, Flags(Options{}), compiler, "build log test", cache); err != nil {
				test.Fatal(err)
			}
		})
		key := filepath.Base(filepath.Dir(library))[:12]
		archives := 0
		for _, match := range lines {
			if match[1] == "cArchive" {
				archives++
				if match[3] != key || match[4] != want {
					test.Fatalf("%q, want the archive %s under %s", match[0], want, key)
				}
			} else if want == "hit" {
				test.Fatalf("%q from an archive the cache held", match[0])
			}
		}
		if archives != 1 {
			test.Fatalf("%d archive lines for a %s, want 1: %q", archives, want, lines)
		}
	}
	test.Logf(
		"%d units and ar logged as misses under the archive's key, the cache's archive as a hit, one archive's bytes",
		len(files),
	)
}

// A one-unit build (Build's default) logs its clang: main.c's compile and then the link, each its own step on every
// platform since the split (#2r9xhqs). The binary is the same bytes with the log and without it. A mutant that logs the
// compile as a cLink, or writes 0 seconds, fails.
//
// Not parallel: it sets ADAMIC_BUILD_LOG for the process.
func TestOneUnitBuildLogsItsClang(test *testing.T) {
	source := "#include \"adamic.h\"\n#include <stdio.h>\nint main(void) { printf(\"logged\\n\"); return 0; }\n"
	directory := test.TempDir()
	quiet := filepath.Join(directory, "quiet", "program")
	logged := filepath.Join(directory, "logged", "program")
	for _, binary := range []string{quiet, logged} {
		if createDirectoryError := os.MkdirAll(filepath.Dir(binary), 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
	}
	if buildError := Build(source, quiet, Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	lines := loggedSteps(test, func() {
		if buildError := Build(source, logged, Options{}); buildError != nil {
			test.Fatal(buildError)
		}
	})
	quietBinary, readError := os.ReadFile(quiet)
	if readError != nil {
		test.Fatal(readError)
	}
	loggedBinary, readError := os.ReadFile(logged)
	if readError != nil {
		test.Fatal(readError)
	}
	if !bytes.Equal(quietBinary, loggedBinary) {
		test.Fatal("the program built with ADAMIC_BUILD_LOG set differs from the one built without it")
	}
	var steps []string
	for _, match := range lines {
		if match[1] == "cArchive" || strings.HasPrefix(match[2], "runtime/") {
			continue
		}
		if match[3] != "-" || match[4] != "off" || loggedSeconds(test, match) <= 0 {
			test.Fatalf("%q, want an unkeyed step that read no cache, with its seconds", match[0])
		}
		steps = append(steps, match[1]+" "+match[2])
	}
	want := "cCompile main.c, cLink program"
	if strings.Join(steps, ", ") != want {
		test.Fatalf("logged %q, want %s", steps, want)
	}
	test.Logf("logged %s, and the binary is the same bytes without the log", want)
}
