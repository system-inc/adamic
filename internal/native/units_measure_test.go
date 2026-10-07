package native

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: paired compiler timings must not compete with other tests in this process.
func TestMeasureClangUnits(t *testing.T) {
	directory := os.Getenv("ADAMIC_CLANG_MEASURE")
	if directory == "" {
		t.Skip("set ADAMIC_CLANG_MEASURE to emitted C evidence directory")
	}
	cache := filepath.Join(directory, "cache")
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	t.Setenv("ADAMIC_NATIVE_SPLIT", "0")
	selected := os.Getenv("ADAMIC_CLANG_PROGRAM")
	filename := "timings.jsonl"
	if selected != "" {
		filename = "timings-" + selected + ".jsonl"
	}
	output, err := os.Create(filepath.Join(directory, filename))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	encoder := json.NewEncoder(output)
	metadata := map[string]string{}
	for name, arguments := range map[string][]string{"commit": {"git", "rev-parse", "HEAD"}, "nproc": {"nproc"}, "go": {"go", "version"}, "clang": {"clang", "--version"}, "node": {"node", "--version"}} {
		data, err := exec.Command(arguments[0], arguments[1:]...).Output()
		if err != nil {
			t.Fatal(err)
		}
		metadata[name] = strings.Split(strings.TrimSpace(string(data)), "\n")[0]
	}
	cpuMax, err := os.ReadFile("/sys/fs/cgroup/cpu.max")
	if err != nil {
		t.Fatal(err)
	}
	metadata["cpu.max"] = strings.TrimSpace(string(cpuMax))
	slow := false
	run := func(program, loop string, round int, options Options, source string, trace bool) {
		loadBefore, _ := os.ReadFile("/proc/loadavg")
		started := time.Now()
		binary := filepath.Join(directory, "program")
		command := "native.Build"
		if trace || (program == "cohere-typeaware" && !options.Split) {
			var library string
			var err error
			if program == "cohere-typeaware" {
				library, err = splitTSGoRuntime(options)
			} else {
				library, err = RuntimeLibrary("", options)
			}
			if err != nil {
				t.Fatal(err)
			}
			filename := filepath.Join(directory, "main.c")
			if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			args := Flags(options)
			if trace {
				args = append(args, "-ftime-trace="+filepath.Join(directory, program+"-"+loop+".json"), "-ftime-trace-granularity=500")
			}
			args = append(args, "-I", filepath.Dir(library), "-o", binary, filename)
			args = append(args, RuntimeLinkFlags(library)...)
			if program == "cohere-typeaware" {
				args = append(args, filepath.Join(directory, "tsgo.a"), "-lpthread", "-ldl")
			}
			args = append(args, "-lm")
			command = "clang " + strings.Join(args, " ")
			started = time.Now()
			if data, err := exec.Command("clang", args...).CombinedOutput(); err != nil {
				t.Fatalf("clang: %v\n%s", err, data)
			}
		} else {
			var err error
			if program == "cohere-typeaware" {
				err = BuildSplitTSGo(source, binary, filepath.Join(directory, "tsgo.a"), options)
				command = "native.BuildSplitTSGo"
			} else {
				err = Build(source, binary, options)
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		elapsed := time.Since(started).Seconds()
		if elapsed > 300 {
			slow = true
		}
		loadAfter, _ := os.ReadFile("/proc/loadavg")
		row := map[string]any{"program": program, "loop": loop, "round": round, "seconds": elapsed, "flags": Flags(options), "jobs": options.Jobs, "split": options.Split, "instrument": command, "load_before": strings.TrimSpace(string(loadBefore)), "load_after": strings.TrimSpace(string(loadAfter)), "build_flags": metadata, "cache": "runtime warm; checker warm for typeaware; unit cache enabled, cold for full, warm for one-line/warm; no program result cache"}
		if err := encoder.Encode(row); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s %s round=%d jobs=%d %.6fs", program, loop, round, options.Jobs, elapsed)
	}

	for _, program := range []string{"lint-harness", "typescript-parser", "cohere-typeaware"} {
		if selected != "" && selected != program {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, program+".c"))
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		header, units, err := splitC(source)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s lines=%d bytes=%d units=%d header=%d", program, strings.Count(source, "\n"), len(source), len(units), len(header))
		changedData, err := os.ReadFile(filepath.Join(directory, program+"-changed.c"))
		if err != nil {
			t.Fatal(err)
		}
		changed := string(changedData)
		changedHeader, changedUnits, err := splitC(changed)
		if err != nil {
			t.Fatal(err)
		}
		if header != changedHeader || len(units) != len(changedUnits) {
			t.Fatal("module edit changed the shared header or unit count")
		}
		changedCount := 0
		for index := range units {
			if units[index] != changedUnits[index] {
				changedCount++
			}
		}
		t.Logf("%s one-line module edit changed %d of %d units", program, changedCount, len(units))
		if changedCount != 1 {
			t.Fatal("edit must invalidate exactly one translation unit")
		}

		for _, mode := range []struct {
			name    string
			options Options
		}{{"sanitized", Options{Sanitize: true}}, {"release", Options{}}, {"counted", Options{Count: true}}} {
			var err error
			if program == "cohere-typeaware" {
				_, err = splitTSGoRuntime(mode.options)
			} else {
				_, err = RuntimeLibrary("", mode.options)
			}
			if err != nil {
				t.Fatal(err)
			}
			slow = false
			run(program, mode.name+"-trace", 0, mode.options, source, true)
			for round := 1; round <= 3; round++ {
				run(program, mode.name+"-whole", round, mode.options, source, false)
				for _, jobs := range []int{1, 5} {
					// Only remove the named object-cache directory in this benchmark's scratch cache.
					if err := os.RemoveAll(filepath.Join(cache, "adamic", "units")); err != nil {
						t.Fatal(err)
					}
					options := mode.options
					options.Split = true
					options.Jobs = jobs
					run(program, mode.name+"-full", round, options, source, false)
					before := mode.options
					before.Jobs = jobs
					run(program, mode.name+"-whole-one-line", round, before, changed, false)
					run(program, mode.name+"-one-line", round, options, changed, false)
					run(program, mode.name+"-warm", round, options, changed, false)
				}
				if slow {
					t.Log("sample exceeded five minutes; keeping one paired round")
					break
				}
			}
		}
	}
}
