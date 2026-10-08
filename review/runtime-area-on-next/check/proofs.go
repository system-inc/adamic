package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func buildWithRuntime(code, directory, binary string, options native.Options) error {
	library, err := native.RuntimeLibraryForSource(directory, code, options)
	if err != nil {
		return err
	}
	path := binary + ".c"
	if err := os.WriteFile(path, []byte(code), 0644); err != nil {
		return err
	}
	args := append(native.LinkFlags(options), "-I", filepath.Dir(library), path, "-o", binary)
	args = append(args, native.RuntimeLinkFlags(library)...)
	args = append(args, "-lm")
	out, err := exec.Command("clang", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("clang: %v %s", err, out)
	}
	return nil
}
func changedRuntime(name, file, old, replacement string) string {
	directory := filepath.Join(evidence, name+"-runtime")
	if err := os.MkdirAll(directory, 0755); err != nil {
		panic(err)
	}
	files, err := filepath.Glob("internal/native/runtime/*")
	if err != nil {
		panic(err)
	}
	changed := false
	for _, path := range files {
		if filepath.Ext(path) != ".c" && filepath.Ext(path) != ".h" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		if filepath.Base(path) == file {
			if strings.Count(string(data), old) != 1 {
				panic("mutant seam must be unique: " + name)
			}
			data = []byte(strings.Replace(string(data), old, replacement, 1))
			changed = true
		}
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), data, 0644); err != nil {
			panic(err)
		}
	}
	if !changed {
		panic("no changed file")
	}
	return directory
}
func proofCode(name string) string {
	path := filepath.Join("review/runtime-area-on-next", name+".a")
	program, err := load.Load([]string{path})
	if err != nil {
		panic(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		panic(err)
	}
	return native.C(lowered)
}
func prove() {
	failed := 0
	mutants := []struct {
		name, program, file, old, changed, catch string
		options                                  native.Options
	}{
		{"canonical-identity", "async_identity", "closure.c", "if (held->code == code) {", "if (false && held->code == code) {", "stdout", native.Options{Sanitize: true}},
		{"async-owner-kind", "async_identity", "closure.c", "case adamic_kind_async_frame: return &((adamic_async_frame *)owner)->functions;", "case adamic_kind_async_frame: return &((adamic_environment *)owner)->functions;", "", native.Options{Sanitize: true}},
		{"parent-slot-owner", "inherited_statics", "object.c", "unsigned char actual = adamic_object_field_types(owner)[adamic_slot_index(owner, slot)];", "unsigned char actual = adamic_object_field_types(object)[adamic_slot_index(object, slot)];", "", native.Options{Sanitize: true}},
		{"cached-readiness-index", "parallel_shapes", "object.c", "!adamic_object_initialized(object)[adamic_slot_index(object, slot)]", "!adamic_object_initialized(object)[__atomic_load_n(&cache->packed, __ATOMIC_RELAXED) >> 48]", "field read failed:", native.Options{ThreadSanitize: true}},
		{"cached-type-index", "parallel_shapes", "object.c", "unsigned char actual = adamic_object_field_types(owner)[adamic_slot_index(owner, slot)];", "unsigned char actual = adamic_object_field_types(owner)[__atomic_load_n(&cache->packed, __ATOMIC_RELAXED) >> 48];", "field read failed:", native.Options{ThreadSanitize: true}},
	}
	for _, m := range mutants {
		code := proofCode(m.program)
		directory := changedRuntime(m.name, m.file, m.old, m.changed)
		binary := filepath.Join(evidence, m.name)
		if err := buildWithRuntime(code, directory, binary, m.options); err != nil {
			panic(err)
		}
		threads := []string{"4"}
		if m.options.ThreadSanitize {
			threads = []string{"4", "16"}
		}
		for _, thread := range threads {
			caught := false
			for attempt := 1; attempt <= 3; attempt++ {
				r := execute([]string{"ADAMIC_THREADS=" + thread, "ASAN_OPTIONS=detect_leaks=0", "TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1", "ADAMIC_TSAN_PERTURB=1"}, binary)
				save(fmt.Sprintf("%s-%s-%d", m.name, thread, attempt), r)
				sanitizer := strings.Contains(string(r.Stderr), "ERROR: AddressSanitizer") || strings.Contains(string(r.Stderr), "runtime error:")
				caught = r.ExitCode != 0 && (sanitizer || m.catch != "" && strings.Contains(string(r.Stderr), m.catch))
				if m.catch == "stdout" {
					absolute, err := filepath.Abs(filepath.Join("review/runtime-area-on-next", m.program+".a"))
					if err != nil {
						panic(err)
					}
					want := execute(nil, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", absolute)
					caught = r.ExitCode == 0 && len(r.Stderr) == 0 && want.ExitCode == 0 && !bytes.Equal(want.Stdout, r.Stdout)
					report, failure := leakcheck.Check(leakcheck.Program{
						C: code, Sanitized: binary, Counted: binary + "-count",
						BuildCounted: func(code, output string) error {
							return buildWithRuntime(code, directory, output, native.Options{Count: true})
						},
						Execute: func(extra []string, name string, args ...string) leakcheck.Run {
							run := execute(extra, name, args...)
							save(m.name+"-leaks", run)
							return run
						},
					})
					caught = caught && failure == nil && report == ""

				}
				if caught {
					fmt.Printf("CAUGHT %s threads=%s attempt=%d exit=%d\n", m.name, thread, attempt, r.ExitCode)
					break
				}
			}
			if !caught {
				fmt.Printf("MISSED %s threads=%s\n", m.name, thread)
				failed++
			}
		}
	}
	code := proofCode("region_partial")
	pattern := regexp.MustCompile(`adamic_object_size\((adamic_temporary_\d+)->shape->count\)`)
	if len(pattern.FindAllString(code, -1)) != 1 {
		panic("graph adoption seam is not unique")
	}
	code = pattern.ReplaceAllString(code, `(adamic_object_size($1->shape->count) - 2 * $1->shape->count)`)
	binary := filepath.Join(evidence, "short-graph-adoption")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		panic(err)
	}
	r := execute([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	save("short-graph-adoption", r)
	if r.ExitCode != 0 && strings.Contains(string(r.Stderr), "ERROR: AddressSanitizer: heap-buffer-overflow") {
		fmt.Println("CAUGHT short-graph-adoption ASan heap-buffer-overflow")
	} else {
		fmt.Printf("MISSED short-graph-adoption %s\n", r.Stderr)
		failed++
	}
	code = proofCode("async_identity")
	// End the global reference but omit precisely its release. Output is unchanged.
	pattern = regexp.MustCompile(`(adamic_closure \* (adamic_temporary_\d+) = adamic_global_0_kept;\n\s*adamic_global_0_kept = NULL;\n\s*)adamic_release\((adamic_temporary_\d+)\);`)
	matches := pattern.FindAllStringSubmatchIndex(code, -1)
	if len(matches) != 2 {
		panic("expected initial and final global reset")
	}
	match := matches[1]
	code = code[:match[0]] + code[match[2]:match[3]] + "/* mutant: missing final kept closure release */" + code[match[1]:]
	sanitized := filepath.Join(evidence, "missing-kept-release-sanitize")
	counted := filepath.Join(evidence, "missing-kept-release-count")
	if err := native.Build(code, sanitized, native.Options{Sanitize: true}); err != nil {
		panic(err)
	}
	if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
		panic(err)
	}
	want := execute(nil, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", "review/runtime-area-on-next/async_identity.a")
	clean := execute([]string{"ASAN_OPTIONS=detect_leaks=0"}, sanitized)
	save("missing-kept-release-output", clean)
	countedRun := execute(nil, counted)
	save("missing-kept-release-count", countedRun)
	report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: sanitized, Counted: counted, Execute: func(env []string, name string, args ...string) leakcheck.Run {
		run := execute(env, name, args...)
		save("missing-kept-release-leaks", run)
		return run
	}})
	if err == nil && report != "" && agrees(want, clean, false) && leakcheck.Unbalanced(countedRun) != "" {
		fmt.Printf("CAUGHT missing-kept-release LeakSanitizer and %s\n", leakcheck.Unbalanced(countedRun))
	} else {
		fmt.Printf("MISSED missing-kept-release err=%v output=%t report=%s\n", err, agrees(want, clean, false), report)
		failed++
	}
	fmt.Printf("PROOFS missed=%d\n", failed)
	if failed != 0 {
		os.Exit(1)
	}
}

const asyncVoidSource = "async function done(): Promise<void> { console.log('done'); }\nvoid done();\n"

func proveRefusal() {
	path := filepath.Join(evidence, "async_void.a")
	if err := os.WriteFile(path, []byte(asyncVoidSource), 0644); err != nil {
		panic(err)
	}
	want := execute(nil, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", path)
	save("async-void-node", want)
	if want.ExitCode != 0 || string(want.Stdout) != "done\n" || len(want.Stderr) != 0 {
		panic("source Node witness changed")
	}
	program, err := load.Load([]string{path})
	if err != nil {
		panic(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "unawaited async task through the void operator") {
		fmt.Printf("FAIL void refusal: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS void refusal: %v\n", err)
}
