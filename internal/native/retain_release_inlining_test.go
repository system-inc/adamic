package native

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/native_runtime"
)

func TestRetainReleaseInliningBoundary(t *testing.T) {
	t.Parallel()
	for _, profile := range []struct {
		name    string
		options native_runtime.Options
	}{
		{"ordinary", native_runtime.Options{}}, {"address", native_runtime.Options{Sanitize: true}}, {"thread", native_runtime.Options{ThreadSanitize: true}},
	} {
		t.Run(profile.name, func(t *testing.T) {
			directory := t.TempDir()
			files, flags, compiler, _, err := native_runtime.RuntimeInputs("", profile.options, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				if strings.HasSuffix(file.Name, ".h") {
					if err := os.WriteFile(filepath.Join(directory, file.Name), file.Contents, 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
			source := `#include "adamic.h"
void *retain_probe(void *value) { return adamic_retain(value); }
void release_probe(void *value) { adamic_release(value); }
`
			if err := os.WriteFile(filepath.Join(directory, "probe.c"), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			compile := func(source, output string, extra ...string) {
				t.Helper()
				arguments := append(append([]string{}, flags...), "-I", directory, "-S", "-emit-llvm", source, "-o", output)
				arguments = append(arguments, extra...)
				command, release := boundedrun.Command(time.Minute, compiler, arguments...)
				defer release()
				command.Dir = directory
				if output, err := command.CombinedOutput(); err != nil {
					t.Fatalf("clang: %v\n%s", err, output)
				}
			}
			compile("probe.c", "probe.ll")
			data, err := os.ReadFile(filepath.Join(directory, "probe.ll"))
			if err != nil {
				t.Fatal(err)
			}
			ir := string(data)
			for _, name := range []string{"adamic_retain", "adamic_release"} {
				if profile.options.Sanitize {
					definition := regexp.MustCompile(`(?m)^define [^\n]*@` + name + `\([^\n]*#([0-9]+)`).FindStringSubmatch(ir)
					if definition == nil {
						t.Fatalf("sanitized %s must remain out of line", name)
					}
					attributes := regexp.MustCompile(`(?m)^attributes #` + definition[1] + ` = \{([^}]*)\}`).FindStringSubmatch(ir)
					if attributes == nil || !strings.Contains(attributes[1], "noinline") {
						t.Fatalf("sanitized %s lacks noinline", name)
					}
				} else if strings.Contains(ir, "@"+name+"(") {
					t.Fatalf("%s %s was not inlined", profile.name, name)
				}
			}
			if err := os.WriteFile(filepath.Join(directory, "unused.c"), []byte("#include \"adamic.h\"\n"), 0644); err != nil {
				t.Fatal(err)
			}
			compile("unused.c", "unused.ll", "-Wunused-function", "-Werror")
		})
	}
}
