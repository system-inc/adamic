package native

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"
)

const historicalCommonFlags = "-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -pthread"

// These literals pin order and bytes to the pre-LTO policy, independently of Flags.
func TestNonShippingFlagsStayIdentical(t *testing.T) {
	t.Parallel()
	for _, release := range []bool{false, true} {
		for _, row := range []struct {
			name    string
			options Options
			suffix  string
		}{
			{"counted", Options{Release: release, Count: true}, " -DADAMIC_COUNT -O2"},
			{"sanitized", Options{Release: release, Sanitize: true}, " -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"},
			{"oracle counts", Options{Release: release, Sanitize: true, Count: true}, " -DADAMIC_COUNT -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"},
			{"slab probe", Options{Release: release, Sanitize: true, Count: true, Slabs: true, cpu: "haswell"}, " -DADAMIC_COUNT -DADAMIC_SLABS -march=haswell -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"},
		} {
			want := strings.ReplaceAll(historicalCommonFlags+row.suffix, " ", "\x00")
			for _, got := range [][]string{Flags(row.options), LinkFlags(row.options)} {
				if strings.Join(got, "\x00") != want {
					t.Errorf("%s release=%v: got %q, want %q", row.name, release, got, want)
				}
			}
		}
	}
	want := strings.ReplaceAll(historicalCommonFlags+" -O2", " ", "\x00")
	for _, got := range [][]string{Flags(Options{}), LinkFlags(Options{})} {
		if strings.Join(got, "\x00") != want {
			t.Errorf("ordinary oracle/test flags changed: %q", got)
		}
	}
}

func TestReleaseFlagsReachLink(t *testing.T) {
	t.Parallel()
	options := Options{Release: true}
	compile, link := Flags(options), LinkFlags(options)
	if !slices.Contains(compile, "-flto=thin") {
		t.Fatal("shipped release is not ThinLTO")
	}
	if len(link) < len(compile) || !slices.Equal(compile, link[:len(compile)]) {
		t.Fatal("link lost or reordered compilation flags")
	}
	if goruntime.GOOS != "darwin" && !slices.Contains(link, "-fuse-ld=lld") {
		t.Fatal("release link is not using lld")
	}
}

func TestReleaseArithmeticIsNeverFused(t *testing.T) {
	t.Parallel()
	cpu, why, ok := fusingProcessor()
	if !ok {
		t.Skip(why)
	}
	binary := filepath.Join(t.TempDir(), "release")
	if err := Build(fusedHarness, binary, Options{Release: true, cpu: cpu}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != "0 0 0\n" {
		t.Fatalf("shipped arithmetic differs: %q", got)
	}
	// This changes only the actual program compilation/link command, not the runtime.
	// With the semantic option omitted, clang's default contraction must be observable.
	source := filepath.Join(t.TempDir(), "fused.c")
	if err := os.WriteFile(source, []byte(fusedHarness), 0o644); err != nil {
		t.Fatal(err)
	}
	library, err := RuntimeLibrary("", Options{Release: true, cpu: cpu})
	if err != nil {
		t.Fatal(err)
	}
	flags := slices.DeleteFunc(LinkFlags(Options{Release: true, cpu: cpu}), func(s string) bool { return s == "-ffp-contract=off" })
	flags = append(flags, "-o", binary, source)
	flags = append(flags, RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, err := exec.Command("clang", flags...).CombinedOutput(); err != nil {
		t.Fatalf("mutant compile: %v\n%s", err, output)
	}
	if got := runWithInput(t, "", binary); got == "0 0 0\n" {
		t.Fatal("link contraction-option mutant survived")
	}
	t.Log("link-only -ffp-contract=off omission caught by arithmetic output")
}

// Not parallel: real cold archive builds record clang through process-wide PATH.
func TestRuntimeCompilesEveryUnitUnfused(t *testing.T) {
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{}
	for _, file := range files {
		if strings.HasSuffix(file.name, ".c") {
			expected[file.name] = true
		}
	}
	directory := t.TempDir()
	log := filepath.Join(directory, "commands.jsonl")
	wrapper := `#!/usr/bin/env python3
import json, os, sys
args=sys.argv[1:]
if '-c' in args and os.path.basename(args[args.index('-c')+1]) == os.environ.get('ADAMIC_TEST_DROP_RUNTIME_CONTRACT'):
 args=[arg for arg in args if arg!='-ffp-contract=off']
with open(LOG, 'a') as log:
 log.write(json.dumps(args)+'\n')
os.execv(COMPILER, [COMPILER, *args])
`
	encode := func(s string) string { b, _ := json.Marshal(s); return string(b) }
	wrapper = strings.ReplaceAll(strings.ReplaceAll(wrapper, "LOG", encode(log)), "COMPILER", encode(compiler))
	if err := os.WriteFile(filepath.Join(directory, "clang"), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
	audit := func() error {
		data, err := os.ReadFile(log)
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			var args []string
			if err := json.Unmarshal([]byte(line), &args); err != nil {
				return err
			}
			index := slices.Index(args, "-c")
			if index < 0 {
				continue
			}
			name := filepath.Base(args[index+1])
			if !expected[name] || seen[name] {
				return fmt.Errorf("unexpected or duplicate runtime compilation: %s", name)
			}
			seen[name] = true
			contract := ""
			for _, arg := range args {
				if strings.HasPrefix(arg, "-ffp-contract=") {
					contract = arg
				}
			}
			if contract != "-ffp-contract=off" {
				return fmt.Errorf("runtime %s lacks effective -ffp-contract=off", name)
			}
		}
		if len(seen) != len(expected) {
			return fmt.Errorf("runtime command coverage: %d of %d", len(seen), len(expected))
		}
		return nil
	}
	for _, row := range []struct {
		name    string
		options Options
	}{
		{"shipped release", Options{Release: true}},
		{"ordinary oracle", Options{Sanitize: true}},
		{"oracle counts", Options{Sanitize: true, Count: true}},
	} {
		t.Run(row.name, func(t *testing.T) {
			// A fresh runtime cache: Linux reads XDG_CACHE_HOME, macOS $HOME/Library/Caches.
			t.Setenv("XDG_CACHE_HOME", t.TempDir())
			t.Setenv("HOME", t.TempDir())
			t.Setenv("ADAMIC_TEST_DROP_RUNTIME_CONTRACT", "")
			if err := os.WriteFile(log, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := RuntimeLibrary("", row.options); err != nil {
				t.Fatal(err)
			}
			if err := audit(); err != nil {
				t.Fatal(err)
			}
			t.Logf("all %d runtime compile commands include contraction protection, including dtoa.c and ieee754.c", len(expected))
		})
	}
	if !expected["dtoa.c"] || !expected["ieee754.c"] {
		t.Fatal("floating-point runtime units missing from audit")
	}
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ADAMIC_TEST_DROP_RUNTIME_CONTRACT", "dtoa.c")
	if err := os.WriteFile(log, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RuntimeLibrary("", Options{Release: true}); err != nil {
		t.Fatalf("mutant must compile successfully: %v", err)
	}
	if err := audit(); err == nil || err.Error() != "runtime dtoa.c lacks effective -ffp-contract=off" {
		t.Fatalf("one-runtime-file omission mutant not caught by contract audit: %v", err)
	}
	t.Log("actual dtoa.c-only flag omission caught by runtime compile-command audit")
}
