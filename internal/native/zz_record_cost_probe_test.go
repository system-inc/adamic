package native

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
)

// A probe, never a test that lands (#1vb0x5s, Runtime's condition 3): what it costs to make a record a
// graph-region member, at 100K and 1M records, against today's counted record and against Node. It
// drives the runtime alone, as record_graph_test.go's harness does, in two builds: a release build
// (-O2, no sanitizers, no counts) for time and resident memory, and the counted, sanitized build
// record_graph_test.go judges by, for allocations, frees, regions made and merges. It runs only when
// RECORD_COST_PROBE=1, and logs a table; only an unbalanced counted run fails it.
const recordCostHarness = `#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "graph_regions.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/resource.h>
#include <time.h>

static const char *const names[] = {"next"};
static const bool references[] = {true};
static const adamic_field_kind kinds[] = {adamic_field_reference};
static const adamic_shape shape = {1, names, references, NULL, kinds};
// key is "k" as the emitter writes a key it knows: an immortal literal, retained for each write.
static adamic_string key = ADAMIC_STRING("k");

static double seconds(void) {
 struct timespec now;
 clock_gettime(CLOCK_MONOTONIC, &now);
 return (double)now.tv_sec + (double)now.tv_nsec / 1e9;
}

// one makes a record holding a new node under "k" and hands the caller one reference to each. In a
// graph shape the node is adopted as record_graph_test.go's nodes are, the record is a graph record,
// and the node's field holds the record back, a cycle per record, as the closed row's store does: the
// field is empty, so there's nothing for a drop to let go of.
static adamic_record *one(bool graph, adamic_object **node) {
 adamic_object *made = adamic_object_new(&shape);
 if (graph) { made = adamic_graph_adopt(made, adamic_object_allocation_size(made)); }
 adamic_record *record = graph ? adamic_record_new_graph(adamic_field_reference) : adamic_record_new(adamic_field_reference);
 adamic_record_define(record, adamic_retain(&key), (adamic_value){.reference = adamic_retain(made)});
 if (graph) { made->slots[0].reference = adamic_graph_hold(made, record); }
 *node = made;
 return record;
}

int main(int count, char **arguments) {
 if (count != 3) { return 2; }
 const char *mode = arguments[1];
 char *end;
 unsigned long long parsed = strtoull(arguments[2], &end, 10);
 if (arguments[2][0] == '\0' || *end != '\0') { return 2; }
 size_t records = (size_t)parsed;
 bool graph, kept;
 if (strcmp(mode, "counted-kept") == 0) { graph = false; kept = true; }
 else if (strcmp(mode, "graph-kept") == 0) { graph = true; kept = true; }
 else if (strcmp(mode, "counted-dropped") == 0) { graph = false; kept = false; }
 else if (strcmp(mode, "graph-dropped") == 0) { graph = true; kept = false; }
 else { return 3; }
 // A kept record goes into one plain array that lives to the end, so each record carries an outside
 // reference; a dropped one is let go of in its own iteration, record first as the closed row does.
 adamic_array *kept_records = kept ? adamic_array_make(0, adamic_ir_Union) : NULL;
 double start = seconds();
 for (size_t index = 0; index < records; index++) {
  adamic_object *node;
  adamic_record *record = one(graph, &node);
  if (kept) { adamic_array_push(kept_records, (adamic_value){.reference = record}); } else { adamic_release(record); }
  adamic_release(node);
 }
 double made = seconds();
 adamic_release(kept_records);
 double freed = seconds();
 struct rusage usage;
 getrusage(RUSAGE_SELF, &usage);
 printf("mode=%s n=%zu make=%.6f free=%.6f peak_kb=%ld", mode, records, made - start, freed - made, usage.ru_maxrss);
#ifdef ADAMIC_COUNT
 // Every graph member is a heap allocation and a region is freed with its members, so allocations
 // equal to frees leave no region behind; the test also counts the regions' reports as they free.
 size_t allocations = (size_t)adamic_counted.allocations, frees = (size_t)adamic_counted.frees;
 printf(" allocations=%zu frees=%zu regions=%zu merges=%zu\n", allocations, frees, (size_t)adamic_counted.graph_regions, (size_t)adamic_counted.graph_merges);
 printf("%s\n", allocations == frees ? "PASS" : "FAIL");
#else
 printf("\n");
#endif
 return 0;
}
`

// recordCostModes are the harness's shapes in table order: whether each makes graph records, and
// whether it keeps them to the end.
var recordCostModes = []struct {
	name        string
	graph, kept bool
}{
	{"counted-kept", false, true},
	{"graph-kept", true, true},
	{"counted-dropped", false, false},
	{"graph-dropped", true, false},
}

var recordCostSizes = []int{100000, 1000000}

// recordCostMeasure is a release run's best of three: make and free seconds each the least of the
// three, and the least peak resident set, in KiB.
type recordCostMeasure struct {
	made, freed float64
	peak        int
}

// recordCostCount is one counted run: what the harness read, how many graph region reports the
// runtime wrote as its regions freed, and why the run is red, or "" when it balances.
type recordCostCount struct {
	allocations, frees, regions, merges, reports int
	red                                          string
}

// recordCostFields reads a run's measurement line, its "name=value" fields.
func recordCostFields(output string) (map[string]string, error) {
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "mode=") {
			continue
		}
		fields := map[string]string{}
		for _, field := range strings.Fields(line) {
			name, value, found := strings.Cut(field, "=")
			if !found {
				return nil, fmt.Errorf("unreadable field %q in %q", field, line)
			}
			fields[name] = value
		}
		return fields, nil
	}
	return nil, fmt.Errorf("no measurement line in %q", output)
}

func recordCostInteger(t *testing.T, fields map[string]string, name string) int {
	t.Helper()
	value, err := strconv.Atoi(fields[name])
	if err != nil {
		t.Fatalf("field %s: %v (fields %v)", name, err, fields)
	}
	return value
}

func recordCostSeconds(t *testing.T, fields map[string]string, name string) float64 {
	t.Helper()
	value, err := strconv.ParseFloat(fields[name], 64)
	if err != nil {
		t.Fatalf("field %s: %v (fields %v)", name, err, fields)
	}
	return value
}

// recordCostBest runs a command three times and keeps the best of each measure. native says the peak
// is getrusage's own, bytes on macOS (residentKibibytes); Node's maxRSS is KiB everywhere.
func recordCostBest(t *testing.T, native bool, arguments ...string) recordCostMeasure {
	t.Helper()
	var best recordCostMeasure
	for attempt := range 3 {
		// Fork from a small shell, as heap_test.go's peak does, so the child's high-water mark doesn't
		// inherit the Go test process's resident set.
		command := exec.Command("/bin/sh", append([]string{"-c", `"$@" & child=$!; wait "$child"`, "rss"}, arguments...)...)
		var stderr strings.Builder
		command.Stderr = &stderr
		output, err := command.Output()
		if err != nil {
			t.Fatalf("%s: %v\nstdout:\n%s\nstderr:\n%s", strings.Join(arguments, " "), err, output, stderr.String())
		}
		fields, err := recordCostFields(string(output))
		if err != nil {
			t.Fatalf("%s: %v", strings.Join(arguments, " "), err)
		}
		made, freed, peak := recordCostSeconds(t, fields, "make"), recordCostSeconds(t, fields, "free"), recordCostInteger(t, fields, "peak_kb")
		if native {
			peak = residentKibibytes(peak, goruntime.GOOS)
		}
		if attempt == 0 || made < best.made {
			best.made = made
		}
		if attempt == 0 || freed < best.freed {
			best.freed = freed
		}
		if attempt == 0 || peak < best.peak {
			best.peak = peak
		}
	}
	return best
}

// recordCostCounted runs the counted build once. A graph region writes a report line as it frees, a
// million of them at the larger size, so stderr is read as it comes and those lines are only counted;
// every other line is kept, the runtime's counts line last.
func recordCostCounted(t *testing.T, binary, mode string, records int, graph bool) recordCostCount {
	t.Helper()
	command := exec.Command(binary, mode, strconv.Itoa(records))
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout strings.Builder
	command.Stdout = &stdout
	pipe, err := command.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	var count recordCostCount
	var others []string
	scanner := bufio.NewScanner(pipe)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		if recordGraphReport.MatchString(scanner.Text()) {
			count.reports++
			continue
		}
		others = append(others, scanner.Text())
	}
	scanned := scanner.Err()
	// Wait closes the pipe, so whatever a failed scan left unread is drained first.
	_, _ = io.Copy(io.Discard, pipe)
	if err := command.Wait(); err != nil || scanned != nil {
		t.Fatalf("counted %s %d: %v %v\nstdout:\n%s\nstderr:\n%s", mode, records, err, scanned, stdout.String(), strings.Join(others, "\n"))
	}
	fields, err := recordCostFields(stdout.String())
	if err != nil {
		t.Fatalf("counted %s %d: %v", mode, records, err)
	}
	count.allocations = recordCostInteger(t, fields, "allocations")
	count.frees = recordCostInteger(t, fields, "frees")
	count.regions = recordCostInteger(t, fields, "regions")
	count.merges = recordCostInteger(t, fields, "merges")
	var reds []string
	if !strings.Contains(stdout.String(), "\nPASS\n") {
		reds = append(reds, fmt.Sprintf("the harness read allocations %d frees %d", count.allocations, count.frees))
	}
	// The runtime's own counts line, written as the process exits, is the last word: every heap value
	// freed or let go of with a statement region, as recordGraphRed holds a counted run to.
	last := ""
	if len(others) > 0 {
		last = others[len(others)-1] + "\n"
	}
	if counts := recordGraphCounts.FindStringSubmatch(last); counts == nil {
		reds = append(reds, "no counts line")
	} else {
		allocations, _ := strconv.Atoi(counts[1])
		frees, _ := strconv.Atoi(counts[2])
		statement, _ := strconv.Atoi(counts[3])
		if allocations != frees+statement {
			reds = append(reds, fmt.Sprintf("heap values leaked: %d (allocations %d, frees %d, in regions %d)", allocations-frees-statement, allocations, frees, statement))
		}
	}
	// Each graph record is one region, which reports once as it frees; a counted record makes no graph
	// member, so none report.
	want := 0
	if graph {
		want = records
	}
	if count.reports != want {
		reds = append(reds, fmt.Sprintf("%d graph regions freed, want %d", count.reports, want))
	}
	count.red = strings.Join(reds, "; ")
	return count
}

func recordCostRatio(numerator, denominator float64) string {
	if denominator == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.2fx", numerator/denominator)
}

func recordCostBytesPerRecord(peak, baseline, records int) float64 {
	return float64(peak-baseline) * 1024 / float64(records)
}

func TestRecordCostProbe(t *testing.T) {
	if os.Getenv("RECORD_COST_PROBE") != "1" {
		t.Skip("a measurement probe: RECORD_COST_PROBE=1 runs it")
	}
	directory := t.TempDir()
	release := filepath.Join(directory, "release")
	if err := Build(recordCostHarness, release, Options{}); err != nil {
		t.Fatal(err)
	}
	counted := filepath.Join(directory, "counted")
	if err := Build(recordCostHarness, counted, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	t.Logf("release flags: %s", strings.Join(Flags(Options{}), " "))
	node, err := exec.LookPath("node")
	if err != nil {
		t.Logf("node isn't on PATH, so there are no Node rows: %v", err)
		node = ""
	}
	script := filepath.Join("testdata", "record_cost", "records.mjs")
	nodeModes := []string{"counted-kept", "graph-kept"}

	// The N=0 runs are each shape's resident floor: a record's bytes are the peak's growth over it.
	baselines := map[string]int{}
	for _, mode := range recordCostModes {
		baselines[mode.name] = recordCostBest(t, true, release, mode.name, "0").peak
	}
	nodeBaselines := map[string]int{}
	if node != "" {
		for _, mode := range nodeModes {
			nodeBaselines[mode] = recordCostBest(t, false, node, "--expose-gc", script, mode, "0").peak
		}
	}

	var table, verdicts, comparisons strings.Builder
	fmt.Fprintf(&table, "\n%-7s %-16s %8s %10s %10s %9s %13s %15s %14s\n", "engine", "mode", "N", "make s", "free s", "peak MB", "bytes/record", "regions/record", "merges/record")
	for _, records := range recordCostSizes {
		measures := map[string]recordCostMeasure{}
		for _, mode := range recordCostModes {
			measure := recordCostBest(t, true, release, mode.name, strconv.Itoa(records))
			measures[mode.name] = measure
			count := recordCostCounted(t, counted, mode.name, records, mode.graph)
			regions, merges := "-", "-"
			if mode.graph {
				regions = fmt.Sprintf("%.3f", float64(count.regions)/float64(records))
				merges = fmt.Sprintf("%.3f", float64(count.merges)/float64(records))
			}
			fmt.Fprintf(&table, "%-7s %-16s %8d %10.4f %10.4f %9.1f %13.1f %15s %14s\n", "adamic", mode.name, records, measure.made, measure.freed, float64(measure.peak)/1024, recordCostBytesPerRecord(measure.peak, baselines[mode.name], records), regions, merges)
			if count.red == "" {
				fmt.Fprintf(&verdicts, "PASS %s N=%d: allocations %d frees %d, regions made %d merged %d, %d freed\n", mode.name, records, count.allocations, count.frees, count.regions, count.merges, count.reports)
			} else {
				fmt.Fprintf(&verdicts, "FAIL %s N=%d: %s\n", mode.name, records, count.red)
				t.Errorf("counted %s N=%d doesn't balance: %s", mode.name, records, count.red)
			}
		}
		nodeMeasures := map[string]recordCostMeasure{}
		if node != "" {
			for _, mode := range nodeModes {
				measure := recordCostBest(t, false, node, "--expose-gc", script, mode, strconv.Itoa(records))
				nodeMeasures[mode] = measure
				fmt.Fprintf(&table, "%-7s %-16s %8d %10.4f %10.4f %9.1f %13.1f %15s %14s\n", "node", mode, records, measure.made, measure.freed, float64(measure.peak)/1024, recordCostBytesPerRecord(measure.peak, nodeBaselines[mode], records), "-", "-")
			}
		}
		graph, plain := measures["graph-kept"], measures["counted-kept"]
		fmt.Fprintf(&comparisons, "N=%d: graph-kept against counted-kept: make %s, free %s, peak %s (bytes/record %s)", records,
			recordCostRatio(graph.made, plain.made), recordCostRatio(graph.freed, plain.freed), recordCostRatio(float64(graph.peak), float64(plain.peak)),
			recordCostRatio(recordCostBytesPerRecord(graph.peak, baselines["graph-kept"], records), recordCostBytesPerRecord(plain.peak, baselines["counted-kept"], records)))
		if nodeGraph, found := nodeMeasures["graph-kept"]; found {
			fmt.Fprintf(&comparisons, "; against Node's graph-kept: make %s, free %s, peak %s (bytes/record %s)\n",
				recordCostRatio(graph.made, nodeGraph.made), recordCostRatio(graph.freed, nodeGraph.freed), recordCostRatio(float64(graph.peak), float64(nodeGraph.peak)),
				recordCostRatio(recordCostBytesPerRecord(graph.peak, baselines["graph-kept"], records), recordCostBytesPerRecord(nodeGraph.peak, nodeBaselines["graph-kept"], records)))
		} else {
			comparisons.WriteString("; no Node run\n")
		}
	}
	t.Log(table.String())
	t.Logf("counted build balance:\n%s", verdicts.String())
	t.Logf("comparisons:\n%s", comparisons.String())
}
