package oracle

import (
	_ "embed"
	"encoding/json"
	"path/filepath"
	"strings"
)

// The October 8 ruling keeps assertion-bearing views probes in TypeScript.
// This explicit inventory resolves dynamic fixture names without redirecting
// the ordinary loader or the area's independent Adamic refusal fixtures.
//
//go:embed views_non_null_fixtures.json
var checkedViewNonNullFixtureInventory []byte

var checkedViewNonNullFixturePaths = func() map[string]bool {
	var paths []string
	if err := json.Unmarshal(checkedViewNonNullFixtureInventory, &paths); err != nil {
		panic(err)
	}
	result := make(map[string]bool, len(paths))
	for _, path := range paths {
		result[path] = true
	}
	return result
}()

func checkedViewFixturePath(path string) string {
	key := filepath.ToSlash(path)
	if filepath.IsAbs(path) || strings.HasPrefix(key, "../") {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return path
		}
		root, err := filepath.Abs(repository)
		if err != nil {
			return path
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			return path
		}
		key = filepath.ToSlash(relative)
	}
	if checkedViewNonNullFixturePaths[key] {
		return strings.TrimSuffix(path, ".a") + ".ts"
	}
	return path
}

// Bound original declarations must retain the TypeScript assertion contract
// when copied to a temporary fixture, rather than regain an Adamic extension.
func checkedViewFixtureCopyPath(path, source string) string {
	if filepath.Ext(checkedViewFixturePath(source)) == ".ts" && filepath.Ext(path) == ".a" {
		return strings.TrimSuffix(path, ".a") + ".ts"
	}
	return path
}
