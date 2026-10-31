package lowered

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
)

func TestProbeTraceOfResolution(t *testing.T) {
	t.Parallel()
	for _, paths := range corpus {
		files := map[string]string{}
		var entry string
		for index, file := range fromHere(paths) {
			absolute, _ := filepath.Abs(file)
			data, err := os.ReadFile(absolute)
			if err != nil {
				t.Fatal(err)
			}
			files[filepath.ToSlash(absolute)] = string(data)
			if index == 0 {
				entry = absolute
			}
		}
		disk := sealedOver(files)
		if _, err := load.LoadFrom([]string{entry}, buildcache.Relative, disk); err != nil {
			t.Errorf("%s: %v", paths[0], err)
		}
		t.Logf("%s: %d asked", paths[0], len(disk.trace()))
		for _, name := range disk.trace() {
			t.Logf("  %s", name)
		}
	}
}
