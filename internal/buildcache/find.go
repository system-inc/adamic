package buildcache

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// Find is get's lookup without its build, for many name keys in one process (#sbxeb0b, cmd/find): Loom's tree builder
// asks it once for every product a test made last time and skips the tests whose products are all found. A product test
// is a process of its own, so each one valued its read sets afresh, every file, listing, tool report and product read
// again (tree 2 on Home, Oct 11: 1,462 finds took 74,199 s and their 8.4 million entries name 77,480 files). Find holds
// the tree's values for its whole run, as Settle does (the tree is fixed while a tree builds), and looks each product
// up once, however many read it.
//
// Its answer is get's, before get builds: a product its read sets name in this cache, after the store's read sets join
// the cache's as get joins them, is found under its key; anything else is missing, with the why get gives. A product
// only the store holds is missing too, with the store's keys in its why: get fetches it beside a description of its
// Inputs (<key>.inputs, which Loom's archive carries), and audits a share of fetches by rebuilding them, and only the
// test holds the Inputs and the build. So a found product is already where get leaves one, and Find places nothing but
// the read sets get would merge (and, for the tree builder, republishes what get would republish).
//
// A lookup that fails (a read set naming a product read in a cycle, or an unreadable read set file) is missing, why
// its error, since the test that builds it fails the same way. Find's own error is only one it couldn't answer at all.
func Find(nameKeys []string) ([]Finding, error) {
	for _, nameKey := range nameKeys {
		if !nameKeyPattern.MatchString(nameKey) {
			return nil, fmt.Errorf("%q is not a name key (64 lowercase hex digits)", nameKey)
		}
	}
	findings := make([]Finding, len(nameKeys))
	for index, nameKey := range nameKeys {
		findings[index] = Finding{NameKey: nameKey, Outcome: FindingMissing}
	}
	if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
		for index := range findings {
			findings[index].Why = "ADAMIC_BUILD_CACHE=off builds every product"
		}
		return findings, nil
	}
	root, err := repositoryRoot()
	if err != nil {
		return nil, err
	}
	cache, err := cacheDirectory()
	if err != nil {
		return nil, err
	}
	settledTree.Store(&sync.Map{})
	defer settledTree.Store(nil)
	looked, failed := lookAll(root, cache, nameKeys)
	// The store's read sets join the cache's for every product the cache's own don't find, as get joins them; then each
	// is looked up again, in a pass of its own, since a joined set can find a product an earlier answer read as missing.
	if storeAddress() != "" {
		joinFailed := make([]error, len(nameKeys))
		joined := parallelMap(nameKeys, func(index int, nameKey string) bool {
			if failed[index] != nil || looked[index].directory != "" {
				return false
			}
			joined, err := joinStoreReads(cache, nameKey, "")
			joinFailed[index] = err
			return joined
		})
		if slices.Contains(joined, true) {
			looked, failed = lookAll(root, cache, nameKeys)
		}
		for index, err := range joinFailed {
			if err != nil {
				failed[index] = err
			}
		}
	}
	for index, nameKey := range nameKeys {
		finding := &findings[index]
		switch {
		case failed[index] != nil:
			finding.Why = failed[index].Error()
		case looked[index].directory != "":
			file, _ := loadReads(cache, nameKey)
			finding.Outcome, finding.Key, finding.Name = FindingFound, looked[index].key, file.Name
		case len(looked[index].keys) > 0 && storeAddress() != "":
			keys := make([]string, len(looked[index].keys))
			for position, key := range looked[index].keys {
				keys[position] = key[:12]
			}
			finding.Why = fmt.Sprintf("%s; the store may hold it under %s, which its test fetches", looked[index].why, strings.Join(keys, ", "))
		default:
			finding.Why = looked[index].why
		}
	}
	// What get does on a hit beyond finding it: the tree builder republishes a product the store has lost, and its read
	// sets when they're old (refreshPublished, a no-op for everyone else). Network-bound, so many at once.
	parallelMap(findings, func(_ int, finding Finding) bool {
		// A test read set's marker is no product: never published (#sfqdxvy).
		if finding.Outcome == FindingFound && finding.Name != TestReadSetName {
			refreshPublished(cache, finding.NameKey, finding.Key, filepath.Join(cache, finding.Key), finding.Name)
		}
		return false
	})
	return findings, nil
}

// A Finding is one name key's answer: found, with the key its read sets give on this tree and the product's name, or
// missing, with why. It is cmd/find's JSON line.
type Finding struct {
	NameKey string `json:"nameKey"`
	Outcome string `json:"outcome"`
	Key     string `json:"key,omitempty"`
	Name    string `json:"name,omitempty"`
	Why     string `json:"why,omitempty"`
}

// A Finding's Outcome.
const (
	FindingFound   = "found"
	FindingMissing = "missing"
)

var nameKeyPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// lookAll looks every name key up in one pass: a product read by many is valued once, and each tree entry once while
// settledTree holds them.
func lookAll(root, cache string, nameKeys []string) ([]found, []error) {
	looked, failed := make([]found, len(nameKeys)), make([]error, len(nameKeys))
	known := productMemo{}
	for index, nameKey := range nameKeys {
		looked[index], failed[index] = productFor(root, cache, nameKey, 0, known)
	}
	return looked, failed
}

// joinStoreReads joins the store's read sets for nameKey to the cache's and says whether there were any. The file keeps
// the name it has, or takes name, or the store's, in that order.
func joinStoreReads(cache, nameKey, name string) (bool, error) {
	stored, sets := fetchReads(nameKey)
	if len(sets) == 0 {
		return false, nil
	}
	if file, err := loadReads(cache, nameKey); err == nil && file.Name != "" {
		name = file.Name
	}
	if name == "" {
		name = stored
	}
	return true, mergeReads(cache, nameKey, name, sets)
}

// findWorkers is how many store requests Find has in flight at once.
const findWorkers = 16

// parallelMap calls each on every item, findWorkers at a time, and returns what each returned, in order.
func parallelMap[T any](items []T, each func(index int, item T) bool) []bool {
	results := make([]bool, len(items))
	next := make(chan int)
	var group sync.WaitGroup
	for range min(findWorkers, len(items)) {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range next {
				results[index] = each(index, items[index])
			}
		}()
	}
	for index := range items {
		next <- index
	}
	close(next)
	group.Wait()
	return results
}
