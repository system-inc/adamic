package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func programParseBuild(t *testing.T, enabled bool) string {
	return programParseBuildKind(t, enabled, false)
}
func programParseBuildKind(t *testing.T, enabled, sanitized bool) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprint("program-parse-", enabled, "-", sanitized), Files: []string{"internal/oracle/program_region_parse_test.go", "internal/oracle/testdata/program_region", "internal/lower", "internal/native", "stage1/typescript/parser", "stage1/typescript/scanner"}}, func(directory string) error {
		data, err := os.ReadFile(filepath.Join(programParseLower(t, enabled), "code.c"))
		if err != nil {
			return err
		}
		code := string(data)
		at := strings.LastIndex(code, "\treturn adamic_host_exit_status();")
		if at < 0 {
			return fmt.Errorf("main return missing")
		}
		code = "#include <sys/resource.h>\n#include <stdio.h>\n" + code[:at] + "\tstruct rusage graph_usage; getrusage(RUSAGE_SELF, &graph_usage); printf(\"graph_peak_rss_kib %ld\\n\", graph_usage.ru_maxrss);\n" + code[at:]
		options := native.Options{ProgramRegion: enabled, Sanitize: sanitized, Count: !sanitized}
		if enabled {
			return programRegionBuild(t, code, filepath.Join(directory, "parse"), options)
		}
		return native.Build(code, filepath.Join(directory, "parse"), options)
	})
}
func TestProduct_ProgramParseOn(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programParseBuild(t, true)
}
func TestProduct_ProgramParseOff(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programParseBuild(t, false)
}

func programParseArguments(t *testing.T) []string {
	t.Helper()
	source := programRegionSource(t)
	var manifest struct{ Files []struct{ Path string } }
	data, err := os.ReadFile("testdata/program_region/parse_manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 77 {
		t.Fatal("parse input inventory changed")
	}
	lines := []string{}
	for _, file := range manifest.Files {
		lines = append(lines, filepath.Join(source, strings.TrimPrefix(file.Path, "typescript/")))
	}
	path := filepath.Join(t.TempDir(), "manifest")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return []string{"--manifest", path, "--count"}
}
func programParseMeasurement(t *testing.T, sample int) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux getrusage measurement")
	}
	args := programParseArguments(t)
	absolute, err := filepath.Abs("testdata/program_region/parse.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", append([]string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", absolute}, args...)...)
	if truth.exitCode != 0 {
		t.Fatalf("source Node parse: %d %s", truth.exitCode, truth.stderr)
	}
	for _, enabled := range []bool{false, true} {
		binary := filepath.Join(programParseBuild(t, enabled), "parse")
		name, argv := pinnedStack("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", binary)
		started := time.Now()
		result := execute(t, name, append(argv, args...)...)
		seconds := time.Since(started).Seconds()
		match := countsLine.FindSubmatch(result.stderr)
		rss := graphRSS.FindSubmatch(result.stdout)
		if result.exitCode != 0 || match == nil || rss == nil {
			t.Fatalf("parse measurement: %d %s %s", result.exitCode, result.stdout, result.stderr)
		}
		if string(graphRSS.ReplaceAll(result.stdout, nil)) != string(truth.stdout) {
			t.Fatalf("parse disagrees: %s", result.stdout)
		}
		if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
			t.Fatal(report)
		}
		t.Logf("sample=%d Program=%t retains=%s releases=%s peak-live=%s peak-RSS-KiB=%s seconds=%.6f", sample, enabled, match[3], match[4], match[5], rss[1], seconds)
	}
}
func TestProgramRegionParseMeasurement0(t *testing.T) { t.Parallel(); programParseMeasurement(t, 0) }
func TestProgramRegionParseMeasurement1(t *testing.T) { t.Parallel(); programParseMeasurement(t, 1) }
func TestProgramRegionParseMeasurement2(t *testing.T) { t.Parallel(); programParseMeasurement(t, 2) }
func TestProgramRegionParseMeasurement3(t *testing.T) { t.Parallel(); programParseMeasurement(t, 3) }
func TestProgramRegionParseMeasurement4(t *testing.T) { t.Parallel(); programParseMeasurement(t, 4) }

func TestProduct_ProgramParseSanitizedOn(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programParseBuildKind(t, true, true)
}
func TestProduct_ProgramParseSanitizedOff(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programParseBuildKind(t, false, true)
}

func programParseLower(t *testing.T, enabled bool) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprint("program-parse-source-", enabled), Files: []string{"internal/oracle/program_region_parse_test.go", "internal/oracle/testdata/program_region", "internal/lower", "internal/native", "internal/javascript", "stage1/typescript/parser", "stage1/typescript/scanner"}}, func(directory string) error {
		source, err := load.Load([]string{"testdata/program_region/parse.a"})
		if err != nil {
			return err
		}
		program, err := lower.LowerWithOptions(context.Background(), source, lower.Options{ProgramRegion: enabled})
		if err != nil {
			return err
		}
		if len(program.ProgramTypes) != 0 {
			return fmt.Errorf("acyclic parser selected members")
		}
		if err := os.WriteFile(filepath.Join(directory, "code.c"), []byte(native.C(program)), 0600); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0600)
	})
}
func TestProduct_ProgramParseLowerOn(t *testing.T)  { t.Parallel(); programParseLower(t, true) }
func TestProduct_ProgramParseLowerOff(t *testing.T) { t.Parallel(); programParseLower(t, false) }
