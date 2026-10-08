package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegionAllocationAndMutants(t *testing.T) {
	source, err := os.ReadFile("testdata/region_alloc.c")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, old, changed, want string
		options                  Options
	}{
		{name: "release", options: Options{Release: true}},
		{name: "counted", options: Options{Count: true}},
		{name: "sanitized", options: Options{Count: true, Sanitize: true}},
		{name: "skip_cursor", old: "region->next = next + size;", changed: "region->next = next;", want: "invalid regional cursor extent", options: Options{Count: true}},
		{name: "skip_block_extent", old: "block->used = (size_t)(region->next - block->bytes);", changed: "/* mutant: lost completed extent */", want: "completed block extent was lost", options: Options{Count: true}},
		{name: "skip_alignment", old: "return (size + ALIGN - 1) & ~(size_t)(ALIGN - 1);", changed: "return size;", want: "unaligned regional object", options: Options{Count: true}},
	}
	for _, given := range cases {
		t.Run(given.name, func(t *testing.T) {
			directory := t.TempDir()
			files, err := readRuntime(runtime, "runtime")
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				data := string(file.contents)
				if given.old != "" && file.name == "region.c" {
					if strings.Count(data, given.old) != 1 {
						t.Fatal("mutant lost its unique anchor")
					}
					data = strings.Replace(data, given.old, given.changed, 1)
				}
				if err := os.WriteFile(filepath.Join(directory, file.name), []byte(data), 0644); err != nil {
					t.Fatal(err)
				}
			}
			library, err := RuntimeLibrary(directory, given.options)
			if err != nil {
				t.Fatal(err)
			}
			main := filepath.Join(directory, "main.c")
			if err := os.WriteFile(main, source, 0644); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(directory, "probe")
			args := append(LinkFlags(given.options), "-I", filepath.Dir(library), main, "-o", binary)
			args = append(args, RuntimeLinkFlags(library)...)
			args = append(args, "-lm")
			if out, err := exec.Command("clang", args...).CombinedOutput(); err != nil {
				t.Fatalf("must compile: %v %s", err, out)
			}
			command := exec.Command(binary)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1:malloc_fill_byte=239", "UBSAN_OPTIONS=halt_on_error=1")
			out, err := command.CombinedOutput()
			if given.old != "" {
				if err == nil || !strings.Contains(string(out), given.want) {
					t.Fatalf("mutant survived or failed incorrectly: %v %s", err, out)
				}
				t.Logf("mutant caught: %s", out)
			} else if err != nil || !strings.Contains(string(out), "regions clean\n") {
				t.Fatalf("control: %v %s", err, out)
			}
		})
	}
}
