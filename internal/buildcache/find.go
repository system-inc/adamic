package buildcache

import (
	"fmt"
	"os"
	"regexp"
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
// Its answer is get's, before get builds: a product its read sets name in this cache is found under its key; anything
// else is missing, with the why get gives, and a name key the cache holds no read sets for is missing without a word to
// the store, which holds products, never read sets. A product only the store holds is missing too, with the store's
// keys in its why: get fetches it beside a description of its Inputs (<key>.inputs, which Loom's archive carries), and
// audits a share of fetches by rebuilding them, and only the test holds the Inputs and the build. So a found product is
// already where get leaves one, and Find places nothing.
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
