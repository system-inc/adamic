package native

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Count actual queue drains, rather than relying on noisy timing: the serial
// allocator must skip them, then resume after a second heap thread frees slots.
func TestSerialSlabsSkipRemoteScans(t *testing.T) {
	source, err := os.ReadFile("testdata/heap_remote_scan.c")
	if err != nil {
		t.Fatal(err)
	}
	for _, variant := range []struct{ name, guard, want string }{
		{"fixed", "", "serial skip; remote reuse\n"},
		{"always scan mutant", "true", "serial scan\n"},
		{"never scan mutant", "false", "remote slots lost\n"},
	} {
		t.Run(variant.name, func(t *testing.T) {
			for _, options := range []Options{{}, {Sanitize: true, Slabs: true}} {
				binary := buildEdited(t, string(source), options, func(name, data string) string {
					if name != "heap.c" {
						return data
					}
					declaration := "static void drain_remote(chunk *each);"
					definition := "static void drain_remote(chunk *each) {"
					guard := "atomic_load_explicit(&thread_count, memory_order_relaxed) > 1"
					if strings.Count(data, declaration) != 1 || strings.Count(data, definition) != 1 || strings.Count(data, guard) != 1 {
						t.Fatal("heap scan anchors changed")
					}
					data = strings.Replace(data, declaration, "static size_t test_remote_drains;\nsize_t adamic_test_remote_drains(void) { return test_remote_drains; }\n"+declaration, 1)
					data = strings.Replace(data, definition, definition+"\n\ttest_remote_drains++;", 1)
					if variant.guard != "" {
						data = strings.Replace(data, guard, variant.guard, 1)
					}
					return data
				})
				output, err := exec.Command(binary).CombinedOutput()
				if string(output) != variant.want {
					t.Fatalf("sanitize %v: %v: %s", options.Sanitize, err, output)
				}
				if variant.guard == "" && err != nil {
					t.Fatal(err)
				}
				if variant.guard != "" && err == nil {
					t.Fatal("mutant passed")
				}
				t.Logf("sanitize %v: %s", options.Sanitize, strings.TrimSpace(string(output)))
			}
		})
	}
}
