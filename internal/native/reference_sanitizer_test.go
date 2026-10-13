package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Inspect optimized code rather than a wall-time threshold: sanitizer probes in
// every caller make large products expensive to compile on a busy gate machine.
func TestSanitizedReferenceHelpersStayOutlined(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.name, ".h") {
			if err := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	source := filepath.Join(directory, "probe.c")
	if err := os.WriteFile(source, []byte(`#include "adamic.h"
void *probe_retain(void *value) { return adamic_retain(value); }
void probe_release(void *value) { adamic_release(value); }
`), 0644); err != nil {
		t.Fatal(err)
	}
	for _, sanitize := range []bool{false, true} {
		arguments := append(Flags(Options{Sanitize: sanitize}), "-S", "-emit-llvm", "-o", "-", source)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		command := toolchain.Command("clang", arguments...)
		bounded := exec.CommandContext(ctx, command.Path, arguments...)
		bounded.Env = command.Env
		output, err := bounded.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("clang sanitize=%t: %v %s", sanitize, err, output)
		}
		text := string(output)
		for _, name := range []string{"retain", "release"} {
			start := strings.Index(text, "@probe_"+name+"(")
			if start < 0 {
				t.Fatalf("missing %s probe", name)
			}
			end := strings.Index(text[start:], "\n}")
			if end < 0 {
				t.Fatalf("missing %s probe body", name)
			}
			body := text[start : start+end]
			outlined := strings.Contains(body, "@adamic_"+name+"(")
			if outlined != sanitize {
				t.Fatalf("%s sanitize=%t: outlined=%t; %s", name, sanitize, outlined, body)
			}
			if sanitize {
				helper := regexp.MustCompile(`(?s)define [^\n]*@adamic_` + name + `\([^\n]*\).*?\n}`).FindString(text)
				if !strings.Contains(helper, "@__asan_report_") {
					t.Fatalf("%s helper lost heap sanitizer probes: %s", name, helper)
				}
			}
			if sanitize && strings.Contains(body, "@__asan_report_") {
				t.Fatalf("%s repeats heap sanitizer probes in caller", name)
			}
		}
	}
}

// Preprocessing must also work when the compiler has no __has_feature builtin.
// GCC 14 simulates older GCC with -U; -w suppresses the intentional builtin-removal
// warning, while preprocessing errors and the flag-value assertions remain fatal.
func TestAddressSanitizerGuardPreprocessesWithoutHasFeature(t *testing.T) {
	t.Parallel()
	source := filepath.Join(t.TempDir(), "guard.c")
	if err := os.WriteFile(source, []byte(`#include "adamic.h"
#ifdef ADAMIC_ADDRESS_SANITIZER
int adamic_address_probe = 1;
#else
int adamic_address_probe = 0;
#endif
`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct {
		name, compiler string
		flags          []string
		enabled        bool
	}{
		{"clang-plain", "clang", nil, false},
		{"clang-asan", "clang", []string{"-fsanitize=address"}, true},
		{"gcc-no-feature", "gcc", []string{"-U__has_feature", "-w"}, false},
		{"gcc-asan-no-feature", "gcc", []string{"-U__has_feature", "-w", "-fsanitize=address"}, true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			args := append([]string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-E", "-P", "-I", "runtime", source}, probe.flags...)
			command := toolchain.Command(probe.compiler, args...)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			bounded := exec.CommandContext(ctx, command.Path, args...)
			bounded.Env = command.Env
			var diagnostic strings.Builder
			bounded.Stderr = &diagnostic
			output, err := bounded.Output()
			if err != nil {
				t.Fatalf("preprocess: %v\n%s", err, &diagnostic)
			}
			want := "int adamic_address_probe = 0;"
			if probe.enabled {
				want = "int adamic_address_probe = 1;"
			}
			if !strings.Contains(string(output), want) {
				t.Fatalf("missing %q in preprocessed guard", want)
			}
		})
	}
}
