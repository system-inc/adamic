package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestGraphRegionsCompiledMillion(t *testing.T) {
	// Not parallel: measure the two large resident sets sequentially.
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/graph_regions/million.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	at := strings.LastIndex(code, "\treturn 0;")
	if at < 0 {
		t.Fatal("missing main return")
	}
	code = "#include <sys/resource.h>\n#include <stdio.h>\n" + code[:at] + "\tstruct rusage usage; getrusage(RUSAGE_SELF, &usage); printf(\"rss_kib %ld\\n\", usage.ru_maxrss);\n" + code[at:]
	for _, options := range []native.Options{{}, {Count: true}} {
		binary := filepath.Join(t.TempDir(), "million")
		if err := native.Build(code, binary, options); err != nil {
			t.Fatal(err)
		}
		result := execute(t, "/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", binary)
		if result.exitCode != 0 {
			t.Fatalf("native %d: %s", result.exitCode, result.stderr)
		}
		if !strings.Contains(string(result.stdout), "members 1000000 reachable 950001") {
			t.Fatal(string(result.stdout))
		}
		if options.Count && !strings.Contains(string(result.stderr), "live 1000000 bytes 70000000 reachable 950001 bytes 66500070 unreachable 49999 bytes 3499930 metadata 16000064") {
			t.Fatal(string(result.stderr))
		}
		t.Logf("compiled Adamic count=%t flags %s:\n%s%s", options.Count, strings.Join(native.Flags(options), " "), result.stdout, result.stderr)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	measured := filepath.Join(t.TempDir(), "million.a")
	source = append(source, []byte("\nconsole.log('rss_kib ' + process.resourceUsage().maxRSS);\nconsole.log('heap_used_bytes ' + process.memoryUsage().heapUsed);\n")...)
	if err := os.WriteFile(measured, source, 0644); err != nil {
		t.Fatal(err)
	}
	result := execute(t, "/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), measured)
	if result.exitCode != 0 {
		t.Fatalf("Node: %d %s", result.exitCode, result.stderr)
	}
	t.Logf("source on Node (oracle type-stripping loader):\n%s", result.stdout)
}

func TestGraphRegionsCountsAndFree(t *testing.T) {
	t.Parallel()
	table, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		if !strings.Contains(fixture.path, "graph_regions_") || strings.Contains(fixture.path, "million") {
			continue
		}
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			row := counted(t, fixture.path, false, nil, false, false)
			if !strings.Contains(string(table), row+"\n") {
				t.Errorf("graph ownership counts changed: %s", row)
			}
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.GraphTypes) == 0 {
				t.Fatalf("%s has no graph types", fixture.path)
			}
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			name, args := pinnedStack(binary)
			result := execute(t, name, args...)
			if result.exitCode != 0 || !graphRegionLine.Match(result.stderr) {
				t.Fatalf("region free missing: %d %s", result.exitCode, result.stderr)
			}
			counts := countsLine.FindSubmatch(result.stderr)
			if counts == nil {
				t.Fatal("missing heap counts")
			}
			allocations, _ := strconv.Atoi(string(counts[1]))
			frees, _ := strconv.Atoi(string(counts[2]))
			arenas, _ := strconv.Atoi(string(counts[6]))
			if allocations != frees+arenas {
				t.Fatalf("accepted graph leaked: %s", result.stderr)
			}
			if strings.HasSuffix(fixture.path, "graph_regions_throw.a") && len(graphRegionLine.FindAll(result.stderr, -1)) != 1 {
				t.Fatalf("throw must free its region exactly once: %s", result.stderr)
			}
			t.Logf("%s", fmt.Sprintf("%s\n%s", row, result.stderr))
		})
	}
}
