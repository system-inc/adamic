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
