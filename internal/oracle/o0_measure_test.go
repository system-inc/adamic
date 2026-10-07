package oracle

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// TestO0Measure is opt-in instrumentation, not a replacement for the gate.
// Not parallel: paired timings must not compete with other tests in this process.
// ADAMIC_O0_BASELINE names an uncached go test -json ./internal/oracle log.
// ADAMIC_O0_OUTPUT names the JSONL observations; ADAMIC_O0_ALL measures every fixture.
func TestO0Measure(t *testing.T) {
	baseline := os.Getenv("ADAMIC_O0_BASELINE")
	if baseline == "" {
		t.Skip("set ADAMIC_O0_BASELINE to an uncached oracle JSON log")
	}
	file, err := os.Open(baseline)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	elapsed := map[string]float64{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 16<<20)
	for scanner.Scan() {
		var event struct {
			Action, Test string
			Elapsed      float64
		}
		if json.Unmarshal(scanner.Bytes(), &event) == nil && (event.Action == "pass" || event.Action == "fail") && strings.HasPrefix(event.Test, "TestNativeAgreesWithNode/") {
			elapsed[strings.TrimPrefix(event.Test, "TestNativeAgreesWithNode/")] = event.Elapsed
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, fixture := range fixtures {
		if fixture.lowers {
			paths = append(paths, fixture.path)
		}
	}
	sort.Strings(paths)
	ranked := append([]string{}, paths...)
	sort.Slice(ranked, func(i, j int) bool {
		if elapsed[ranked[i]] == elapsed[ranked[j]] {
			return ranked[i] < ranked[j]
		}
		return elapsed[ranked[i]] > elapsed[ranked[j]]
	})
	if len(ranked) < 50 || len(elapsed) < len(paths) {
		t.Fatalf("incomplete baseline: %d observations for %d fixtures", len(elapsed), len(paths))
	}
	selected := map[string]string{}
	for _, path := range ranked[:25] {
		selected[path] = "slow"
	}
	// Fixed stride through the sorted remainder gives exactly 25 distinct fixtures.
	var remainder []string
	for _, path := range paths {
		if selected[path] == "" {
			remainder = append(remainder, path)
		}
	}
	stride := len(remainder) / 25
	for i := 0; i < 25; i++ {
		selected[remainder[i*stride]] = "stride"
	}
	if os.Getenv("ADAMIC_O0_ALL") == "1" {
		for _, path := range paths {
			if selected[path] == "" {
				selected[path] = "population"
			}
		}
	}
	output, err := os.Create(os.Getenv("ADAMIC_O0_OUTPUT"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	encoder := json.NewEncoder(output)
	write := func(value any) {
		if err := encoder.Encode(value); err != nil {
			t.Fatal(err)
		}
	}
	metadata := map[string]string{}
	for name, command := range map[string][]string{"commit": {"git", "rev-parse", "HEAD"}, "nproc": {"nproc"}, "cpu.max": {"cat", "/sys/fs/cgroup/cpu.max"}, "go": {"go", "version"}, "clang": {"clang", "--version"}, "node": {"node", "--version"}, "load_before": {"cat", "/proc/loadavg"}} {
		bytes, err := exec.Command(command[0], command[1:]...).CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		metadata[name] = strings.TrimSpace(string(bytes))
	}
	metadata["mode"] = "uncached observations; fresh runtime artifacts once per exact flag set"
	write(metadata)
	type variant struct {
		name    string
		options native.Options
	}
	variants := []variant{{"sanitized", native.Options{Sanitize: true}}, {"release", native.Options{}}, {"counted", native.Options{Count: true}}}
	root := t.TempDir()
	runtimeDirectory := filepath.Join(repository, "internal/native/runtime")
	entries, err := os.ReadDir(runtimeDirectory)
	if err != nil {
		t.Fatal(err)
	}
	// All runtime bytes are snapshotted before any reuse. No disk or result cache is added.
	snapshot := map[string][]byte{}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".c") || strings.HasSuffix(entry.Name(), ".h") {
			bytes, err := os.ReadFile(filepath.Join(runtimeDirectory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			snapshot[entry.Name()] = bytes
		}
	}
	// Each artifact has one fixed variant and mode; there is no lookup cache.
	for _, variant := range variants {
		for _, mode := range []string{"current", "o0"} {
			flags := o0Flags(variant.options, mode)

			directory := filepath.Join(root, variant.name+"-"+mode)
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for name, bytes := range snapshot {
				if err := os.WriteFile(filepath.Join(directory, name), bytes, 0644); err != nil {
					t.Fatal(err)
				}
			}
			start := time.Now()
			var objects []string
			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), ".c") {
					continue
				}
				object := filepath.Join(directory, entry.Name()+".o")
				o0Compile(t, append(append([]string{}, flags...), "-c", filepath.Join(directory, entry.Name()), "-o", object))
				objects = append(objects, object)
			}
			library := filepath.Join(directory, "runtime.a")
			if bytes, err := exec.Command("ar", append([]string{"rcs", library}, objects...)...).CombinedOutput(); err != nil {
				t.Fatalf("ar: %v: %s", err, bytes)
			}

			write(map[string]any{"stage": "runtime", "variant": variant.name, "mode": mode, "seconds": time.Since(start).Seconds(), "flags": flags})
		}
	}
	repeats := 3
	if os.Getenv("ADAMIC_O0_ONCE") == "1" {
		repeats = 1
	}
	for _, path := range paths {
		if selected[path] == "" {
			continue
		}
		t.Run(path, func(t *testing.T) {
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			program, err := lowered(t, absolute)
			if err != nil {
				t.Fatal(err)
			}
			write(map[string]any{"fixture": path, "selection": selected[path], "stage": "lower", "seconds": time.Since(start).Seconds()})
			backend := filepath.Join(t.TempDir(), "backend.mjs")
			if err := os.WriteFile(backend, []byte(javascript.JavaScript(program)), 0644); err != nil {
				t.Fatal(err)
			}
			for loop := 0; loop < repeats; loop++ {
				modes := []string{"current", "o0"}
				if loop%2 == 1 {
					modes = []string{"o0", "current"}
				}
				for _, variant := range variants {
					if variant.name == "counted" && uncounted[path] {
						continue
					}
					observations := map[string]run{}
					for _, mode := range modes {
						directory := t.TempDir()
						flags := o0Flags(variant.options, mode)
						library := filepath.Join(root, variant.name+"-"+mode, "runtime.a")
						start := time.Now()
						source := native.C(program)
						emission := time.Since(start).Seconds()
						main := filepath.Join(directory, "main.c")
						object := filepath.Join(directory, "main.o")
						binary := filepath.Join(directory, "program")
						if err := os.WriteFile(main, []byte(source), 0644); err != nil {
							t.Fatal(err)
						}
						start = time.Now()
						o0Compile(t, append(append([]string{}, flags...), "-I", filepath.Dir(library), "-c", main, "-o", object))
						compile := time.Since(start).Seconds()
						start = time.Now()
						arguments := append(append([]string{}, flags...), object, "-o", binary)
						arguments = append(arguments, native.RuntimeLinkFlags(library)...)
						arguments = append(arguments, "-lm")
						o0Compile(t, arguments)
						link := time.Since(start).Seconds()
						environment := []string{}
						if variant.options.Sanitize {
							environment = append(environment, "ASAN_OPTIONS=detect_leaks=0")
						}
						name, args := binary, []string{}
						if variant.options.Count {
							name, args = pinnedStack(binary)
						}
						start = time.Now()
						result := executeWith(t, environment, name, args...)
						duration := time.Since(start).Seconds()
						observations[mode] = result
						write(map[string]any{"fixture": path, "selection": selected[path], "loop": loop, "variant": variant.name, "mode": mode, "flags": flags, "emit": emission, "clang": compile, "link": link, "run": duration, "hash": o0Hash(result), "stdout_hash": fmt.Sprintf("%x", sha256.Sum256(result.stdout)), "stderr_hash": fmt.Sprintf("%x", sha256.Sum256(result.stderr)), "exit": result.exitCode})
						// Node has no optimization mode, but repeat it alongside each pair to expose drift.
						for _, nodePath := range []string{absolute, backend} {
							start = time.Now()
							node := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), nodePath)
							stage := "node"
							if nodePath == backend {
								stage = "backend_node"
							}
							write(map[string]any{"fixture": path, "loop": loop, "variant": variant.name, "mode": mode, "stage": stage, "seconds": time.Since(start).Seconds(), "hash": o0Hash(node), "stdout_hash": fmt.Sprintf("%x", sha256.Sum256(node.stdout)), "stderr_hash": fmt.Sprintf("%x", sha256.Sum256(node.stderr)), "exit": node.exitCode})
							observations[mode+"/"+stage] = node
							if stage == "node" && node.exitCode == 0 && variant.options.Sanitize {
								start = time.Now()
								leak := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
								observations[mode+"/leak"] = leak
								write(map[string]any{"fixture": path, "loop": loop, "variant": variant.name, "mode": mode, "stage": "leak", "seconds": time.Since(start).Seconds(), "hash": o0Hash(leak), "stdout_hash": fmt.Sprintf("%x", sha256.Sum256(leak.stdout)), "stderr_hash": fmt.Sprintf("%x", sha256.Sum256(leak.stderr)), "exit": leak.exitCode})
							}
						}
					}
					for _, stage := range []string{"", "/node", "/backend_node", "/leak"} {
						before, beforePresent := observations["current"+stage]
						after, afterPresent := observations["o0"+stage]
						if beforePresent != afterPresent || (beforePresent && o0Hash(before) != o0Hash(after)) {
							write(map[string]any{"fixture": path, "variant": variant.name, "loop": loop, "stage": "difference", "execution": stage, "before": o0Diff(before), "after": o0Diff(after)})
							t.Logf("flag-dependent output: %s %s %s: %s", path, variant.name, stage, disagreement(before, after))
						}
					}
				}
			}
		})
	}
	bytes, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		t.Fatal(err)
	}
	write(map[string]string{"load_after": strings.TrimSpace(string(bytes))})
}

func o0Compile(t *testing.T, flags []string) {
	t.Helper()
	if bytes, err := bounded(t, "clang", flags...).CombinedOutput(); err != nil {
		t.Fatalf("clang %q: %v\n%s", flags, err, bytes)
	}
}
func o0Flags(options native.Options, mode string) []string {
	flags := native.Flags(options)
	if mode == "o0" {
		for i, flag := range flags {
			if flag == "-O1" || flag == "-O2" {
				flags[i] = "-O0"
			}
		}
	}
	return flags
}
func o0Hash(value run) string {
	bytes, _ := json.Marshal(o0Diff(value))
	return fmt.Sprintf("%x", sha256.Sum256(bytes))
}
func o0Diff(value run) map[string]any {
	return map[string]any{"stdout": value.stdout, "stderr": value.stderr, "exit": value.exitCode}
}

// Mutants isolate every observable component and boundary ambiguity.
func TestO0HashCatchesMutants(t *testing.T) {
	original := run{stdout: []byte("ab"), stderr: []byte("c"), exitCode: 0}
	mutants := []run{{stdout: []byte("aB"), stderr: []byte("c")}, {stdout: []byte("ab"), stderr: []byte("C")}, {stdout: []byte("ab"), stderr: []byte("c"), exitCode: 70}, {stdout: []byte("a"), stderr: []byte("bc")}}
	for index, mutant := range mutants {
		if o0Hash(original) == o0Hash(mutant) {
			t.Fatalf("mutant %d escaped", index)
		}
	}
}
