// find answers, for each product name key on its standard input, whether this tree's build cache holds the product
// (#sbxeb0b): Loom's tree builder asks it once for every product a test made last time, and skips the tests whose
// products are all found.
//
//	go build -o <directory>/find ./internal/buildcache/cmd/find
//	<directory>/find < name-keys
//
// Run it inside the adamic tree, with the environment its product tests run in (ADAMIC_BUILD_CACHE_DIR,
// ADAMIC_BUILD_OUTPUT_DIR, ADAMIC_BUILD_STORE, ADAMIC_TOOLS, PATH's go) but untraced, and build it with their GOFLAGS:
// every read set values the build settings of the binary that looks it up, so a finder built otherwise misses every
// product, naming that entry. Standard input is one name key per line (blank lines are skipped); standard output is one JSON line per name
// key, in the same order: {"nameKey", "outcome": "found"|"missing", "key", "name", "why"}, key and name only when
// found, why only when missing. A found product is in the cache under key, as a test's lookup leaves it; buildcache.Find
// says what found means. It exits 0 when it answered every name key, and 1 with a line on standard error when it
// couldn't (a line that isn't a name key, a tree it can't find).
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/system-inc/adamic/internal/buildcache"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: find < name-keys (one per line; one JSON line out for each)")
		os.Exit(2)
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "find:", err)
		os.Exit(1)
	}
}

func run() error {
	var nameKeys []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			nameKeys = append(nameKeys, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	findings, err := buildcache.Find(nameKeys)
	if err != nil {
		return err
	}
	output := bufio.NewWriter(os.Stdout)
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	for _, finding := range findings {
		if err := encoder.Encode(finding); err != nil {
			return err
		}
	}
	return output.Flush()
}
