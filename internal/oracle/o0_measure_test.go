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
	"sync"
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

// TestO0MeasureWall replays the complete fixture population with four workers.
// It measures a correctness lane with sanitized builds at O1 and release/count
// builds at O0. The separate O2 lane is additional work and is not timed here.
// Not parallel: each lane must finish before the next paired lane starts.
func TestO0MeasureWall(t *testing.T) {
	if os.Getenv("ADAMIC_O0_WALL_OUTPUT") == "" {
		t.Skip("set ADAMIC_O0_WALL_OUTPUT for the full fixture wall experiment")
	}
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	output, err := os.Create(os.Getenv("ADAMIC_O0_WALL_OUTPUT"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	var outputLock sync.Mutex
	encoder := json.NewEncoder(output)
	write := func(value any) {
		outputLock.Lock()
		defer outputLock.Unlock()
		if err := encoder.Encode(value); err != nil {
			t.Fatal(err)
		}
	}
	metadata := map[string]string{"mode": "uncached observations; warm runtime artifacts; sanitized O1 in both lanes"}
	for name, command := range map[string][]string{"commit": {"git", "rev-parse", "HEAD"}, "nproc": {"nproc"}, "cpu.max": {"cat", "/sys/fs/cgroup/cpu.max"}, "go": {"go", "version"}, "clang": {"clang", "--version"}, "node": {"node", "--version"}, "load_before": {"cat", "/proc/loadavg"}} {
		bytes, err := exec.Command(command[0], command[1:]...).CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		metadata[name] = strings.TrimSpace(string(bytes))
	}
	write(metadata)
	type variant struct {
		name    string
		options native.Options
	}
	variants := []variant{{"sanitized", native.Options{Sanitize: true}}, {"release", native.Options{}}, {"counted", native.Options{Count: true}}}
	libraries := map[string]string{}
	for _, variant := range variants {
		library, err := native.RuntimeLibrary("", variant.options)
		if err != nil {
			t.Fatal(err)
		}
		libraries[variant.name+"/current"] = library
		if variant.options.Sanitize {
			libraries[variant.name+"/o0"] = library
			continue
		}
		// A fresh private artifact, never a result cache. The fixed flag set, source
		// snapshot and compiler are held unchanged for the lifetime of this test.
		libraries[variant.name+"/o0"] = o0WallRuntime(t, o0Flags(variant.options, "o0"))
	}
	var observationsLock sync.Mutex
	observations := map[string]run{}
	for loop := 0; loop < 3; loop++ {
		modes := []string{"current", "o0"}
		if loop%2 == 1 {
			modes = []string{"o0", "current"}
		}
		for _, mode := range modes {
			start := time.Now()
			t.Run(fmt.Sprintf("%d/%s", loop, mode), func(t *testing.T) {
				for _, fixture := range fixtures {
					if !fixture.lowers {
						continue
					}
					t.Run(fixture.path, func(t *testing.T) {
						t.Parallel()
						path, err := filepath.Abs(filepath.Join(repository, fixture.path))
						if err != nil {
							t.Fatal(err)
						}
						lowerStart := time.Now()
						program, err := lowered(t, path)
						if err != nil {
							t.Fatal(err)
						}
						write(map[string]any{"stage": "lower", "fixture": fixture.path, "loop": loop, "mode": mode, "seconds": time.Since(lowerStart).Seconds()})
						observe := func(variant string, result run, seconds float64) {
							key := fmt.Sprintf("%d/%s/%s", loop, fixture.path, variant)
							observationsLock.Lock()
							defer observationsLock.Unlock()
							observations[key+"/"+mode] = result
							before, beforePresent := observations[key+"/current"]
							after, afterPresent := observations[key+"/o0"]
							write(map[string]any{"fixture": fixture.path, "loop": loop, "mode": mode, "variant": variant, "stage": "execution", "seconds": seconds, "hash": o0Hash(result), "stdout_hash": fmt.Sprintf("%x", sha256.Sum256(result.stdout)), "stderr_hash": fmt.Sprintf("%x", sha256.Sum256(result.stderr)), "exit": result.exitCode})
							if beforePresent && afterPresent && o0Hash(before) != o0Hash(after) {
								write(map[string]any{"stage": "difference", "fixture": fixture.path, "loop": loop, "variant": variant, "before": o0Diff(before), "after": o0Diff(after)})
							}
						}
						nodeStart := time.Now()
						oracle := onNode(t, path)
						observe("node", oracle, time.Since(nodeStart).Seconds())
						nodeStart = time.Now()
						backend := onJavaScriptBackend(t, program)
						observe("backend_node", backend, time.Since(nodeStart).Seconds())
						var sanitized run
						var sanitizedBinary string
						for _, variant := range variants {
							if variant.options.Count && uncounted[fixture.path] {
								continue
							}
							flagsMode := mode
							if variant.options.Sanitize {
								flagsMode = "current"
							}
							flags := o0Flags(variant.options, flagsMode)
							library := libraries[variant.name+"/"+mode]
							directory := t.TempDir()
							main := filepath.Join(directory, "main.c")
							binary := filepath.Join(directory, "program")
							emissionStart := time.Now()
							c := native.C(program)
							emission := time.Since(emissionStart).Seconds()
							if err := os.WriteFile(main, []byte(c), 0644); err != nil {
								t.Fatal(err)
							}
							object := filepath.Join(directory, "main.o")
							compileStart := time.Now()
							o0Compile(t, append(append([]string{}, flags...), "-I", filepath.Dir(library), "-c", main, "-o", object))
							compile := time.Since(compileStart).Seconds()
							linkStart := time.Now()
							args := append(append([]string{}, flags...), object, "-o", binary)
							args = append(args, native.RuntimeLinkFlags(library)...)
							args = append(args, "-lm")
							o0Compile(t, args)
							write(map[string]any{"stage": "build", "fixture": fixture.path, "variant": variant.name, "loop": loop, "mode": mode, "flags": flags, "emit": emission, "clang": compile, "link": time.Since(linkStart).Seconds()})
							name, runArgs := binary, []string{}
							environment := []string{}
							if variant.options.Sanitize {
								environment = append(environment, "ASAN_OPTIONS=detect_leaks=0")
							}
							if variant.options.Count {
								name, runArgs = pinnedStack(binary)
							}
							runStart := time.Now()
							result := executeWith(t, environment, name, runArgs...)
							observe(variant.name, result, time.Since(runStart).Seconds())
							if variant.options.Sanitize {
								sanitized, sanitizedBinary = result, binary
							}
							if variant.name == "release" && disagreement(sanitized, result) != "" {
								t.Errorf("release differs from sanitized: %s", disagreement(sanitized, result))
							}
						}
						expected := oracle
						if fixture.checked {
							expected = backend
							if sanitized.exitCode != 70 || oracle.exitCode == 70 {
								t.Error("inserted check did not fire as expected")
							}
						}
						if difference := disagreement(expected, sanitized); difference != "" {
							t.Errorf("sanitized differs from Node: %s", difference)
						}
						if !fixture.checked && disagreement(oracle, backend) != "" {
							t.Error("backend differs from source Node")
						}
						if !fixture.checked && oracle.exitCode == 0 {
							leakStart := time.Now()
							leak := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitizedBinary)
							observe("leak", leak, time.Since(leakStart).Seconds())
							if leak.exitCode != 0 {
								t.Errorf("leak check: exit %d, %s", leak.exitCode, leak.stderr)
							}
						}
					})
				}
			})
			write(map[string]any{"stage": "wall", "loop": loop, "mode": mode, "seconds": time.Since(start).Seconds()})
		}
	}
	bytes, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		t.Fatal(err)
	}
	write(map[string]string{"load_after": strings.TrimSpace(string(bytes))})
}

func o0WallRuntime(t *testing.T, flags []string) string {
	t.Helper()
	directory := t.TempDir()
	source := filepath.Join(repository, "internal/native/runtime")
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h") {
			continue
		}
		bytes, err := os.ReadFile(filepath.Join(source, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), bytes, 0644); err != nil {
			t.Fatal(err)
		}
	}
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
	return library
}
