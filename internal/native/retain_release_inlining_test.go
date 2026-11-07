package native

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

func TestRetainReleaseInliningBoundary(test *testing.T) {
	test.Parallel()
	for _, profile := range []struct {
		name    string
		options Options
	}{
		{"ordinary", Options{}}, {"address", Options{Sanitize: true}}, {"thread", Options{ThreadSanitize: true}},
	} {
		test.Run(profile.name, func(test *testing.T) {
			directory := test.TempDir()
			files, flags, compiler, _, err := runtimeInputs("", profile.options, nil)
			if err != nil {
				test.Fatal(err)
			}
			for _, file := range files {
				if strings.HasSuffix(file.name, ".h") {
					if err := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0644); err != nil {
						test.Fatal(err)
					}
				}
			}
			source := `#include "adamic.h"
void *retain_probe(void *value) { return adamic_retain(value); }
void release_probe(void *value) { adamic_release(value); }
`
			if err := os.WriteFile(filepath.Join(directory, "probe.c"), []byte(source), 0644); err != nil {
				test.Fatal(err)
			}
			compile := func(source, output string, extra ...string) {
				test.Helper()
				arguments := append(
					append([]string{}, flags...),
					"-I",
					directory,
					"-S",
					"-emit-llvm",
					source,
					"-o",
					output,
				)
				arguments = append(arguments, extra...)
				command, release := boundedrun.Command(time.Minute, compiler, arguments...)
				defer release()
				command.Dir = directory
				if output, err := command.CombinedOutput(); err != nil {
					test.Fatalf("clang: %v\n%s", err, output)
				}
			}
			compile("probe.c", "probe.ll")
			data, err := os.ReadFile(filepath.Join(directory, "probe.ll"))
			if err != nil {
				test.Fatal(err)
			}
			ir := string(data)
			for _, name := range []string{"adamic_retain", "adamic_release"} {
				if profile.options.Sanitize {
					definition := regexp.MustCompile(
						`(?m)^define [^\n]*@` + name + `\([^\n]*#([0-9]+)`,
					).FindStringSubmatch(ir)
					if definition == nil {
						test.Fatalf("sanitized %s must remain out of line", name)
					}
					attributes := regexp.MustCompile(
						`(?m)^attributes #` + definition[1] + ` = \{([^}]*)\}`,
					).FindStringSubmatch(ir)
					if attributes == nil || !strings.Contains(attributes[1], "noinline") {
						test.Fatalf("sanitized %s lacks noinline", name)
					}
				} else if strings.Contains(ir, "@"+name+"(") {
					test.Fatalf("%s %s was not inlined", profile.name, name)
				}
			}
			if err := os.WriteFile(
				filepath.Join(directory, "unused.c"),
				[]byte("#include \"adamic.h\"\n"),
				0644,
			); err != nil {
				test.Fatal(err)
			}
			compile("unused.c", "unused.ll", "-Wunused-function", "-Werror")
		})
	}
}
