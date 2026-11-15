package native

import (
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// A probe that never lands (#1vb0x5s): a program whose one record became a graph record reports one
// more counted release, with allocations, frees, retains and peak the same. Each mode below is one
// program that frees everything by its end, run once in the counted build (record_graph_test.go's
// Options for it), its counts line printed. A second build of the same harness, its runtime's count.h
// rewritten by exact match so ADAMIC_COUNT_RELEASE also hands adamic_trace_release (defined in the
// harness) its caller's return address and that caller's own, says for counted-record and
// graph-record which function made each release. adamic_release stays out of line there because the
// build is address-sanitized: adamic.h's ADAMIC_OWNERSHIP is static noinline under ASan, a copy in
// each translation unit, so __builtin_return_address(0) inside it is the call site.
const recordReleaseProbeHarness = `#include "adamic.h"
#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

#ifdef ADAMIC_TRACE_RELEASE
// The trace build's count.h calls this from adamic_release: where adamic_release was called from, and
// where that caller was called from.
void adamic_trace_release(void *caller, void *outer) {
 fprintf(stderr, "release-from %" PRIxPTR " %" PRIxPTR "\n", (uintptr_t)caller, (uintptr_t)outer);
}
#endif

static const char *const names[] = {"held"};
static const bool references[] = {true};
static const adamic_field_kind kinds[] = {adamic_field_reference};
static const adamic_shape shape = {1, names, references, NULL, kinds};

// key is an immortal constant, as the emitter writes a record literal's key, retained at each write.
static adamic_string key = ADAMIC_STRING("k");

// node is a graph member with no region yet, adopted as record_graph_test.go's harness adopts its
// nodes, with no label: its one slot stays empty.
static adamic_object *node(void) {
 adamic_object *object = adamic_object_new(&shape);
 return adamic_graph_adopt(object, adamic_object_allocation_size(object));
}

int main(int count, char **arguments) {
 if (count != 2) { return 2; }
#ifdef ADAMIC_TRACE_RELEASE
 // main's run-time address, so the probe can take the load bias off every traced address.
 fprintf(stderr, "release-anchor %" PRIxPTR "\n", (uintptr_t)main);
#endif
 const char *mode = arguments[1];
 if (strcmp(mode, "node-only") == 0) {
  adamic_object *b = node();
  adamic_release(b);
 } else if (strcmp(mode, "counted-record") == 0 || strcmp(mode, "graph-record") == 0) {
  adamic_object *b = node();
  adamic_record *record = strcmp(mode, "graph-record") == 0 ? adamic_record_new_graph(adamic_field_reference) : adamic_record_new(adamic_field_reference);
  adamic_record_define(record, adamic_retain(&key), (adamic_value){.reference = adamic_retain(b)});
  adamic_release(record);
  adamic_release(b);
 } else if (strcmp(mode, "graph-record-empty") == 0) {
  adamic_record *record = adamic_record_new_graph(adamic_field_reference);
  adamic_release(record);
 } else if (strcmp(mode, "counted-record-empty") == 0) {
  adamic_record *record = adamic_record_new(adamic_field_reference);
  adamic_release(record);
 } else if (strcmp(mode, "graph-record-scalar-free") == 0 || strcmp(mode, "counted-record-plain-object") == 0) {
  // A plain object, never adopted, so no graph member: a graph table's take passes it by.
  adamic_object *plain = adamic_object_new(&shape);
  adamic_record *record = strcmp(mode, "graph-record-scalar-free") == 0 ? adamic_record_new_graph(adamic_field_reference) : adamic_record_new(adamic_field_reference);
  adamic_record_define(record, adamic_retain(&key), (adamic_value){.reference = adamic_retain(plain)});
  adamic_release(record);
  adamic_release(plain);
 } else {
  return 3;
 }
 return 0;
}
`

var recordReleaseProbeModes = []struct{ label, name string }{
	{"M0", "node-only"},
	{"M1", "counted-record"},
	{"M2", "graph-record"},
	{"M3", "graph-record-empty"},
	{"M4", "counted-record-empty"},
	{"M5", "graph-record-scalar-free"},
	{"M6", "counted-record-plain-object"},
}

// recordReleaseProbeHook is count.h's release hook, and what the trace build's copy says in its place.
const (
	recordReleaseProbeHookBefore = "#define ADAMIC_COUNT_RELEASE() atomic_fetch_add_explicit(&adamic_counted.releases, 1, memory_order_relaxed)\n"
	recordReleaseProbeHookAfter  = "void adamic_trace_release(void *caller, void *outer);\n#define ADAMIC_COUNT_RELEASE() (atomic_fetch_add_explicit(&adamic_counted.releases, 1, memory_order_relaxed), adamic_trace_release(__builtin_return_address(0), __builtin_return_address(1)))\n"
)

var (
	recordReleaseProbeCounts   = regexp.MustCompile(`(?m)^adamic: counts: .*$`)
	recordReleaseProbeReleases = regexp.MustCompile(`releases (\d+)`)
	recordReleaseProbeAnchor   = regexp.MustCompile(`(?m)^release-anchor ([0-9a-f]+)$`)
	recordReleaseProbeTrace    = regexp.MustCompile(`(?m)^release-from ([0-9a-f]+) ([0-9a-f]+)$`)
)

// recordReleaseProbeTraced builds the harness against a runtime whose count.h traces each release, as
// recordGraphMutants builds its mutants: every runtime .c compiled from a scratch copy, then linked
// directly with main.c. It adds -DADAMIC_TRACE_RELEASE, frame pointers (so the hook's second return
// address can be read), -Wno-frame-address (clang's -Wall may warn of a nonzero
// __builtin_return_address; -Wno-unknown-warning-option in case a clang has no such group), and
// -rdynamic at the link. The sanitized flags already carry -g.
func recordReleaseProbeTraced(t *testing.T) string {
	t.Helper()
	flags := append(slices.Clone(Flags(Options{Sanitize: true, Count: true})), "-DADAMIC_TRACE_RELEASE", "-fno-omit-frame-pointer", "-Wno-unknown-warning-option", "-Wno-frame-address")
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	within := filepath.Join(scratch, "runtime")
	if err := os.MkdirAll(within, 0o755); err != nil {
		t.Fatal(err)
	}
	var sources []string
	for _, file := range files {
		contents := file.contents
		if file.name == "count.h" {
			text := string(contents)
			if found := strings.Count(text, recordReleaseProbeHookBefore); found != 1 {
				t.Fatalf("count.h holds the release hook %d times, want once", found)
			}
			contents = []byte(strings.Replace(text, recordReleaseProbeHookBefore, recordReleaseProbeHookAfter, 1))
		}
		if err := os.WriteFile(filepath.Join(within, file.name), contents, 0o644); err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(file.name, ".c") {
			sources = append(sources, file.name)
		}
	}
	if err := eachInParallel(len(sources), func(index int) error {
		arguments := append(slices.Clone(flags), "-c", sources[index], "-o", strings.TrimSuffix(sources[index], ".c")+".o")
		if output, err := runtimeMutantClang(within, arguments...).CombinedOutput(); err != nil {
			return fmt.Errorf("compiling runtime %s: %w\n%s", sources[index], err, output)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scratch, "main.c"), []byte(recordReleaseProbeHarness), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(scratch, "traced")
	arguments := append(slices.Clone(flags), "-rdynamic", "-I", "runtime", "-o", binary, "main.c")
	for _, source := range sources {
		arguments = append(arguments, filepath.Join("runtime", strings.TrimSuffix(source, ".c")+".o"))
	}
	arguments = append(arguments, "-lm")
	if output, err := runtimeMutantClang(scratch, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("linking the traced harness: %v\n%s", err, output)
	}
	return binary
}

// recordReleaseProbeSymbolize names the function at each file address: addr2line -f -i -a's inline
// chain, innermost first, each frame then the function it was inlined into, joined by " < ", or with no addr2line on the path the ELF symbol table's
// function holding it (no inline frames). It says which it used.
func recordReleaseProbeSymbolize(t *testing.T, binary string, addresses []uint64) (map[uint64]string, string) {
	t.Helper()
	named := map[uint64]string{}
	if len(addresses) == 0 {
		return named, "none"
	}
	for _, tool := range []string{"addr2line", "llvm-addr2line"} {
		path, err := exec.LookPath(tool)
		if err != nil {
			continue
		}
		arguments := []string{"-f", "-i", "-a", "-e", binary}
		for _, address := range addresses {
			arguments = append(arguments, "0x"+strconv.FormatUint(address, 16))
		}
		output, err := exec.Command(path, arguments...).Output()
		if err != nil {
			t.Logf("%s failed: %v", tool, err)
			continue
		}
		// Each address's group starts with its "0x" line, then a function line and a location line
		// for each frame, innermost first.
		group := -1
		var frames []string
		var lines []string
		for _, line := range strings.Split(string(output), "\n") {
			if strings.TrimSpace(line) != "" {
				lines = append(lines, strings.TrimSpace(line))
			}
		}
		flush := func() {
			if group >= 0 && group < len(addresses) {
				named[addresses[group]] = strings.Join(frames, " < ")
			}
			frames = nil
		}
		for index := 0; index < len(lines); index++ {
			if strings.HasPrefix(lines[index], "0x") {
				flush()
				group++
				continue
			}
			frames = append(frames, lines[index])
			// The location line that follows each function line.
			index++
		}
		flush()
		if group+1 != len(addresses) {
			t.Fatalf("%s gave %d groups for %d addresses:\n%s", tool, group+1, len(addresses), output)
		}
		return named, tool + " -f -i -a"
	}
	file, err := elf.Open(binary)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	symbols, err := file.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	for _, address := range addresses {
		named[address] = "??"
		for _, symbol := range symbols {
			if elf.ST_TYPE(symbol.Info) == elf.STT_FUNC && symbol.Value <= address && address < symbol.Value+symbol.Size {
				named[address] = symbol.Name
				break
			}
		}
	}
	return named, "the ELF symbol table (no addr2line on the path)"
}

// recordReleaseProbeHistogram is one traced run's releases by the function that called adamic_release,
// and by that function with its own caller.
type recordReleaseProbeHistogram struct {
	callers, pairs map[string]int
	total          int
}

func recordReleaseProbeLines(histogram map[string]int) []string {
	keys := make([]string, 0, len(histogram))
	for key := range histogram {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("  %3d  %s", histogram[key], key))
	}
	return lines
}

func TestRecordReleaseProbe(t *testing.T) {
	if os.Getenv("RECORD_RELEASE_PROBE") != "1" {
		t.Skip("a probe: set RECORD_RELEASE_PROBE=1")
	}
	counted := filepath.Join(t.TempDir(), "counted")
	if err := Build(recordReleaseProbeHarness, counted, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	countedReleases := map[string]string{}
	for _, mode := range recordReleaseProbeModes {
		run := recordGraphExecute(t, counted, mode.name)
		line := recordReleaseProbeCounts.FindString(run.stderr)
		if run.exit != 0 || line == "" {
			t.Errorf("%s %s: exit %d, no counts line?\nstderr:\n%s", mode.label, mode.name, run.exit, run.stderr)
			continue
		}
		t.Logf("%s %s: %s", mode.label, mode.name, line)
		if releases := recordReleaseProbeReleases.FindStringSubmatch(line); releases != nil {
			countedReleases[mode.name] = releases[1]
		}
	}

	traced := recordReleaseProbeTraced(t)
	file, err := elf.Open(traced)
	if err != nil {
		t.Skipf("the trace's addresses are read against an ELF binary (Linux): %v", err)
	}
	symbols, err := file.Symbols()
	file.Close()
	if err != nil {
		t.Fatal(err)
	}
	var mainValue uint64
	for _, symbol := range symbols {
		if symbol.Name == "main" && elf.ST_TYPE(symbol.Info) == elf.STT_FUNC {
			mainValue = symbol.Value
		}
	}
	if mainValue == 0 {
		t.Fatal("no main in the traced binary's symbol table")
	}
	type traceRun struct {
		label, name string
		pairs       [][2]uint64
	}
	var runs []traceRun
	seen := map[uint64]bool{}
	var addresses []uint64
	for _, mode := range recordReleaseProbeModes {
		if mode.label != "M1" && mode.label != "M2" {
			continue
		}
		run := recordGraphExecute(t, traced, mode.name)
		if run.exit != 0 {
			t.Fatalf("traced %s: exit %d\nstderr:\n%s", mode.name, run.exit, run.stderr)
		}
		anchor := recordReleaseProbeAnchor.FindStringSubmatch(run.stderr)
		if anchor == nil {
			t.Fatalf("traced %s: no anchor line\nstderr:\n%s", mode.name, run.stderr)
		}
		anchorValue, err := strconv.ParseUint(anchor[1], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		// The run-time address less the load bias is the file's; less one more lands inside the call
		// instruction rather than on the one after it.
		fileAddress := func(text string) uint64 {
			value, err := strconv.ParseUint(text, 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			if value == 0 {
				return 0
			}
			return value - (anchorValue - mainValue) - 1
		}
		record := traceRun{label: mode.label, name: mode.name}
		for _, match := range recordReleaseProbeTrace.FindAllStringSubmatch(run.stderr, -1) {
			pair := [2]uint64{fileAddress(match[1]), fileAddress(match[2])}
			record.pairs = append(record.pairs, pair)
			for _, address := range pair {
				if !seen[address] {
					seen[address] = true
					addresses = append(addresses, address)
				}
			}
		}
		if line := recordReleaseProbeCounts.FindString(run.stderr); line != "" {
			t.Logf("traced %s %s: %s", mode.label, mode.name, line)
		}
		runs = append(runs, record)
	}
	named, symbolizer := recordReleaseProbeSymbolize(t, traced, addresses)
	t.Logf("symbolized with %s", symbolizer)
	histograms := map[string]recordReleaseProbeHistogram{}
	for _, run := range runs {
		histogram := recordReleaseProbeHistogram{callers: map[string]int{}, pairs: map[string]int{}}
		for _, pair := range run.pairs {
			caller, outer := named[pair[0]], named[pair[1]]
			histogram.callers[caller]++
			histogram.pairs[caller+"  called from  "+outer]++
			histogram.total++
		}
		histograms[run.name] = histogram
		t.Logf("%s %s: %d traced releases (the counted build's line says %s); by adamic_release's caller:\n%s", run.label, run.name, histogram.total, countedReleases[run.name], strings.Join(recordReleaseProbeLines(histogram.callers), "\n"))
		t.Logf("%s %s: by caller and its caller:\n%s", run.label, run.name, strings.Join(recordReleaseProbeLines(histogram.pairs), "\n"))
	}
	// What M2 counts that M1 doesn't, or less, by caller and its caller.
	difference := map[string]int{}
	for key, value := range histograms["graph-record"].pairs {
		difference[key] += value
	}
	for key, value := range histograms["counted-record"].pairs {
		difference[key] -= value
	}
	for key, value := range difference {
		if value == 0 {
			delete(difference, key)
		}
	}
	t.Logf("M2 less M1, by caller and its caller:\n%s", strings.Join(recordReleaseProbeLines(difference), "\n"))
}
