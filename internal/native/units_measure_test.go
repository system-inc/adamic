package native

import (
	"encoding/json"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/internal/tracked"
)

// Not parallel: paired compiler timings must not compete with other tests in this process.
func TestMeasureClangUnits(test *testing.T) {
	directory := os.Getenv("ADAMIC_CLANG_MEASURE")
	if directory == "" {
		test.Skip("set ADAMIC_CLANG_MEASURE to emitted C evidence directory")
	}
	cache := filepath.Join(directory, "cache")
	// The native caches live under the build cache, so this benchmark's scratch cache is one.
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	test.Setenv("ADAMIC_BUILD_OUTPUT_DIR", "") // its writes too: an inherited output directory would take them
	test.Setenv("ADAMIC_GATE_UNCACHED", "0")
	test.Setenv("ADAMIC_NATIVE_SPLIT", "0")
	selected := os.Getenv("ADAMIC_CLANG_PROGRAM")
	filename := "timings.jsonl"
	if selected != "" {
		filename = "timings-" + selected + ".jsonl"
	}
	output, createError := os.Create(filepath.Join(directory, filename))
	if createError != nil {
		test.Fatal(createError)
	}
	defer output.Close()
	encoder := json.NewEncoder(output)
	metadata := map[string]string{}
	// The tree's commit: git where it has .git, else the manifest a Loom runner's unpacked source carries.
	commit, headError := tracked.Head("../..")
	if headError != nil {
		test.Fatal(headError)
	}
	metadata["commit"] = commit
	// The processors this process may run on, as nproc reports them (both read the scheduler's affinity mask), without
	// a process the gate's tools would have to declare; go, clang and node are the tools the builds run, as toolchain
	// resolves them.
	metadata["nproc"] = strconv.Itoa(goruntime.NumCPU())
	for name, arguments := range map[string][]string{"go": {"version"}, "clang": {"--version"}, "node": {"--version"}} {
		data, outputError := toolchain.Command(name, arguments...).Output()
		if outputError != nil {
			test.Fatal(outputError)
		}
		metadata[name] = strings.Split(strings.TrimSpace(string(data)), "\n")[0]
	}
	cpuMax, readError := os.ReadFile("/sys/fs/cgroup/cpu.max")
	if readError != nil {
		test.Fatal(readError)
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
				library, err = splitTSGoRuntime(source, options)
			} else {
				library, err = RuntimeLibrary("", options)
			}
			if err != nil {
				test.Fatal(err)
			}
			filename := filepath.Join(directory, "main.c")
			if writeError := os.WriteFile(filename, []byte(source), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			args := sourceFlags(source, options)
			if trace {
				args = append(
					args,
					"-ftime-trace="+filepath.Join(directory, program+"-"+loop+".json"),
					"-ftime-trace-granularity=500",
				)
			}
			args = append(args, "-I", filepath.Dir(library), "-o", binary, filename)
			args = append(args, RuntimeLinkFlags(library)...)
			if program == "cohere-typeaware" {
				args = append(args, filepath.Join(directory, "tsgo.a"), "-lpthread", "-ldl")
			}
			args = append(args, "-lm")
			command = "clang " + strings.Join(args, " ")
			started = time.Now()
			if data, combinedOutputError := toolchain.Command(
				"clang",
				args...,
			).CombinedOutput(); combinedOutputError != nil {
				test.Fatalf("clang: %v\n%s", combinedOutputError, data)
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
				test.Fatal(err)
			}
		}
		elapsed := time.Since(started).Seconds()
		if elapsed > 300 {
			slow = true
		}
		loadAfter, _ := os.ReadFile("/proc/loadavg")
		row := map[string]any{
			"program":     program,
			"loop":        loop,
			"round":       round,
			"seconds":     elapsed,
			"flags":       Flags(options),
			"jobs":        options.Jobs,
			"split":       options.Split,
			"instrument":  command,
			"load_before": strings.TrimSpace(string(loadBefore)),
			"load_after":  strings.TrimSpace(string(loadAfter)),
			"build_flags": metadata,
			"cache":       "runtime warm; checker warm for typeaware; unit cache enabled, cold for full, warm for one-line/warm; no program result cache",
		}
		if encodeError := encoder.Encode(row); encodeError != nil {
			test.Fatal(encodeError)
		}
		test.Logf("%s %s round=%d jobs=%d %.6fs", program, loop, round, options.Jobs, elapsed)
	}

	for _, program := range []string{"lint-harness", "typescript-parser", "cohere-typeaware"} {
		if selected != "" && selected != program {
			continue
		}
		data, readError := os.ReadFile(filepath.Join(directory, program+".c"))
		if readError != nil {
			test.Fatal(readError)
		}
		source := string(data)
		header, units, splitCError := splitC(source)
		if splitCError != nil {
			test.Fatal(splitCError)
		}
		test.Logf(
			"%s lines=%d bytes=%d units=%d header=%d",
			program,
			strings.Count(source, "\n"),
			len(source),
			len(units),
			len(header),
		)
		changedData, readError := os.ReadFile(filepath.Join(directory, program+"-changed.c"))
		if readError != nil {
			test.Fatal(readError)
		}
		changed := string(changedData)
		changedHeader, changedUnits, splitCError := splitC(changed)
		if splitCError != nil {
			test.Fatal(splitCError)
		}
		if header != changedHeader || len(units) != len(changedUnits) {
			test.Fatal("module edit changed the shared header or unit count")
		}
		changedCount := 0
		for index := range units {
			if units[index] != changedUnits[index] {
				changedCount++
			}
		}
		test.Logf("%s one-line module edit changed %d of %d units", program, changedCount, len(units))
		if changedCount != 1 {
			test.Fatal("edit must invalidate exactly one translation unit")
		}

		for _, mode := range []struct {
			name    string
			options Options
		}{{"sanitized", Options{Sanitize: true}}, {"release", Options{}}, {"counted", Options{Count: true}}} {
			var err error
			if program == "cohere-typeaware" {
				_, err = splitTSGoRuntime(source, mode.options)
			} else {
				_, err = RuntimeLibrary("", mode.options)
			}
			if err != nil {
				test.Fatal(err)
			}
			slow = false
			run(program, mode.name+"-trace", 0, mode.options, source, true)
			for round := 1; round <= 3; round++ {
				run(program, mode.name+"-whole", round, mode.options, source, false)
				for _, jobs := range []int{1, 5} {
					// Only remove the named object-cache directory in this benchmark's scratch cache.
					if removeError := os.RemoveAll(filepath.Join(cache, "native-units")); removeError != nil {
						test.Fatal(removeError)
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
					test.Log("sample exceeded five minutes; keeping one paired round")
					break
				}
			}
		}
	}
}
