package lowered

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// A program's resolution looks for paths outside its sources: on Oct 31, every directory's package.json from each
// source's directory up to the root (a package scope's type decides a file's module kind) and, beside an import, the
// names it might mean (values.a.ts, values.d.a). None of them is held by the sealed disk, so each answered absent; a
// bare import would look in node_modules too. The trace keeps each path by where it is from the sources, so it names
// no machine and two checkouts store the same trace:
//
//   - "at <path>" is a path under the sources' directory, by its path from there;
//   - "above <path>" is a path under some directory above it, by its path from that directory, and stands for that
//     path under every directory above the sources', up to the root: resolution climbed them all, at a depth each
//     checkout has its own of.
//
// A listing of a directory is refused rather than traced: on disk it holds more than the sources, and no absence keys
// it.

// resolutionOf is the trace of the paths asked, by base, the sources' directory.
func resolutionOf(base string, asked []string) ([]string, error) {
	base = filepath.ToSlash(base)
	lines := map[string]bool{}
	for _, name := range asked {
		if strings.HasSuffix(name, "/") {
			return nil, fmt.Errorf("lowered: the checker lists %s, which on disk may hold more than the program's sources: the program is refused for the cache", name)
		}
		if rest, under := strings.CutPrefix(name, base+"/"); under {
			lines["at "+rest] = true
			continue
		}
		for above := path.Dir(base); ; above = path.Dir(above) {
			prefix := strings.TrimSuffix(above, "/") + "/"
			if rest, under := strings.CutPrefix(name, prefix); under {
				lines["above "+rest] = true
				break
			}
			if above == "/" {
				return nil, fmt.Errorf("lowered: the checker asked after %s, which is no path under the root", name)
			}
		}
	}
	trace := make([]string, 0, len(lines))
	for line := range lines {
		trace = append(trace, line)
	}
	sort.Strings(trace)
	return trace, nil
}

// absent is nil when every path trace names, beside base, is missing from disk: each absence the sealed load assumed is
// this disk's too, so a fresh load here lowers the same program. A path that's there refuses the program.
func absent(base string, trace []string) error {
	for _, line := range trace {
		kind, rest, _ := strings.Cut(line, " ")
		var places []string
		switch kind {
		case "at":
			places = []string{filepath.Join(base, filepath.FromSlash(rest))}
		case "above":
			for above := filepath.Dir(base); ; above = filepath.Dir(above) {
				places = append(places, filepath.Join(above, filepath.FromSlash(rest)))
				if above == filepath.Dir(above) {
					break
				}
			}
		default:
			return fmt.Errorf("lowered: the trace's line %q is neither at nor above", line)
		}
		for _, place := range places {
			if _, err := os.Lstat(place); err == nil {
				return fmt.Errorf("lowered: the program's resolution looks for %s, which is on disk, and its lowering would read it: the program is refused for the cache", place)
			} else if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("lowered: the program's resolution looks for %s, which can't be told absent (%v): the program is refused for the cache", place, err)
			}
		}
	}
	return nil
}
