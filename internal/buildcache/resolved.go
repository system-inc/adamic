package buildcache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// The resolved view (#7vd4sw3, under Loom's #v08721): a unit Loom keys resolved is handed its tree's products already
// answered, a map from each product's name key to the buildcache key the tree built it under, laid read-only where
// ADAMIC_BUILD_RESOLVED names, its presence the mode. Its lookups read that map and never a product's read sets, so a
// run reads no `.reads` file, and Loom values each product the run reached by its bytes (the map's sum), never by the
// keys its inputs spell. A product the map names is taken only when it's laid (existingPlace, whose stat of
// `<cache>/<key>` is the reach Loom's trace records, in the process that answers it). One it doesn't name, or a map
// that doesn't read, is a miss with its reason: the product is built as it would be today, and Loom refuses that
// run's record. buildcache never reads a sum; it's Loom's.

// resolvedVariable names the map a resolved unit's lookups read.
const resolvedVariable = "ADAMIC_BUILD_RESOLVED"

// resolvedVersion is the map's format, which Loom's runner writes (Reads, #v08721).
const resolvedVersion = 1

// A resolvedEntry is one product of the map: the buildcache key its tree built it under, and its archive's sum.
type resolvedEntry struct {
	Key string `json:"key"`
	Sum string `json:"sum"`
}

// A resolvedFile is the map as the runner writes it: {"version":1,"products":{"<name key>":{"key":…,"sum":…}}}.
type resolvedFile struct {
	Version  int                      `json:"version"`
	Products map[string]resolvedEntry `json:"products"`
}

// A resolvedMap is a map as this process read it: its products, or why it doesn't read.
type resolvedMap struct {
	products map[string]resolvedEntry
	err      error
}

// resolvedMaps holds each map this process read, by its path: the runner lays it before the unit starts and nothing
// writes it after, so a process reads it once. It's never handed to a child, which reads the map itself and stats
// every product it answers.
var resolvedMaps sync.Map

// resolvedProducts is the map ADAMIC_BUILD_RESOLVED names, and whether the variable is set: the mode. A map that
// doesn't read comes back as the error, with the mode on, so every lookup misses rather than reading the index.
func resolvedProducts() (map[string]resolvedEntry, bool, error) {
	path := os.Getenv(resolvedVariable)
	if path == "" {
		return nil, false, nil
	}
	if held, ok := resolvedMaps.Load(path); ok {
		read := held.(resolvedMap)
		return read.products, true, read.err
	}
	products, err := readResolved(path)
	resolvedMaps.Store(path, resolvedMap{products: products, err: err})
	return products, true, err
}

// readResolved reads the map at path strictly: an unknown field, another version, trailing bytes, or a name key or
// key that isn't 64 hex digits refuses the whole map, since a key is joined into a path under the cache.
func readResolved(path string) (map[string]resolvedEntry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var file resolvedFile
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("%s: %v", path, err)
	}
	if decoder.More() {
		return nil, fmt.Errorf("%s: more than one map", path)
	}
	if file.Version != resolvedVersion {
		return nil, fmt.Errorf("%s: version %d, not %d", path, file.Version, resolvedVersion)
	}
	for nameKey, entry := range file.Products {
		if !nameKeyPattern.MatchString(nameKey) || !nameKeyPattern.MatchString(entry.Key) {
			return nil, fmt.Errorf("%s: %q names %q, and both must be keys", path, nameKey, entry.Key)
		}
	}
	return file.Products, nil
}

// resolvedLookup is nameKey's product as the map answers it: the key it names, when that product is laid in the cache,
// or why not. It gives no keys for the store to try, so a miss builds, as a name key with no read set does today.
func resolvedLookup(cache, nameKey string, products map[string]resolvedEntry, err error) found {
	if err != nil {
		return found{why: fmt.Sprintf("the resolved map doesn't read: %v", err)}
	}
	entry, named := products[nameKey]
	if !named {
		return found{why: "the resolved map names no product for it"}
	}
	directory, laid := existingPlace(filepath.Join(cache, entry.Key))
	if !laid {
		return found{why: fmt.Sprintf("the resolved map's product %s isn't laid", entry.Key[:12])}
	}
	return found{key: entry.Key, directory: directory}
}
