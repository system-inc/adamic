package tsgo_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Not parallel: this coverage proof changes ADAMIC_TEST_SHARD with test.Setenv.
func TestBridgeUnitsCoverEveryPiece(test *testing.T) {
	if len(bridgeCases) != 36 || len(bridgeBindings) != 36 {
		test.Fatalf("bridge coverage count: %d cases, %d bindings, want 36", len(bridgeCases), len(bridgeBindings))
	}
	core := []string{
		"abi",
		"input-length",
		"unlinked-build",
		"unlinked-c",
		"unlinked-js",
		"output-length",
		"stale-handle",
		"linkage",
		"output-free",
		"region",
		"region-ownership",
	}
	helpers := []string{
		"bridgeABI",
		"bridgeInputLength",
		"bridgeUnlinkedBuild",
		"bridgeUnlinkedC",
		"bridgeUnlinkedJavaScript",
		"bridgeOutputLength",
		"bridgeStale",
		"bridgeLinkage",
		"bridgeOutputFree",
		"bridgeRegion",
		"bridgeRegionOwnership",
	}
	files := []string{"sample.ts", "checker.ts", "parser.ts", "types.ts", "utilities.ts"}
	pieces := []string{"oracle", "timing-1", "timing-2", "timing-3", "wrong-position"}
	seen := map[string]bool{}
	nameOf := func(fn any) string {
		name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
		return name[strings.LastIndex(name, ".")+1:]
	}
	for index, unit := range bridgeCases {
		if seen[unit.name] || nameOf(bridgeBindings[index]) != unit.name {
			test.Fatalf("bridge binding %d: %s", index, unit.name)
		}
		seen[unit.name] = true
		var piece, file, helper string
		round := 0
		if index < len(core) {
			piece = core[index]
			helper = helpers[index]
		} else {
			at := index - len(core)
			file = files[at/5]
			piece = pieces[at%5]
			helper = "bridgeTiming"
			round = at % 5
			if at%5 == 0 {
				helper = "bridgeOracle"
				round = 0
			}
			if at%5 == 4 {
				helper = "bridgeWrongPosition"
				round = 0
			}
		}
		if unit.piece != piece || unit.file != file || unit.round != round || nameOf(unit.check) != helper {
			test.Fatalf("bridge piece %d changed: %+v, want %s/%s/%d/%s", index, unit, piece, file, round, helper)
		}
	}
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	queries := 0
	for _, file := range bridgeRoots(repository) {
		text, readError := os.ReadFile(file)
		if readError != nil {
			test.Fatal(readError)
		}
		positions := bridgePositions(text)
		if len(positions) != min(400, len(text)) {
			test.Fatalf("query coverage count: %s: %d, want %d", file, len(positions), min(400, len(text)))
		}
		for index, position := range positions {
			if position != index*len(text)/len(positions) {
				test.Fatalf("query coverage position %d: %s", index, file)
			}
		}
		queries += len(positions)
	}
	active := 0
	for _, unit := range bridgeCases {
		if bridgeFileActive(unit.file) {
			active++
		}
	}
	want := 16
	if os.Getenv("ADAMIC_TSGO_CORPUS") != "" {
		want = 31
	}
	if active != want {
		test.Fatalf("active bridge coverage count: %d, want %d", active, want)
	}
	for count := 1; count <= 40; count++ {
		owned := make([]int, len(bridgeCases))
		for shard := 0; shard < count; shard++ {
			test.Setenv("ADAMIC_TEST_SHARD", fmt.Sprintf("%d/%d", shard, count))
			for index := range bridgeCases {
				if bridgeCaseActive(index) {
					owned[index]++
				}
			}
		}
		for index, unit := range bridgeCases {
			expected := 0
			if bridgeFileActive(unit.file) {
				expected = 1
			}
			if owned[index] != expected {
				test.Fatalf(
					"shard coverage %d/%d: %s owns %d, want %d",
					index,
					count,
					unit.name,
					owned[index],
					expected,
				)
			}
		}
	}
	test.Logf(
		"coverage: %d active pieces, %d query positions, 5 query analyses; each piece has one owner for 1..40 shards",
		active,
		queries,
	)
}

// Not parallel: this cache proof changes its cache directory with test.Setenv. Its products are its own: one-line files in
// a cache of its own, never the run's, so no TestProduct_ builds them and its census stays out of the run's.
func TestBridgeProductCacheIsVerified(test *testing.T) {
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", test.TempDir())
	test.Setenv("ADAMIC_BUILD_OUTPUT_DIR", "") // its writes too: an inherited output directory would take them
	test.Setenv("ADAMIC_BUILD_CACHE", "")
	test.Setenv("ADAMIC_BUILD_LOG", filepath.Join(test.TempDir(), "builds.log"))
	builds := 0
	build := func(directory string) error {
		builds++
		hashes := map[string]string{}
		for _, name := range bridgeProductNames {
			data := []byte(name)
			if writeError := os.WriteFile(filepath.Join(directory, name), data, 0o600); writeError != nil {
				return writeError
			}
			hashes[name] = fmt.Sprintf("%x", sha256.Sum256(data))
		}
		data, encodeError := json.Marshal(hashes)
		if encodeError != nil {
			return encodeError
		}
		return os.WriteFile(filepath.Join(directory, "products.json"), data, 0o600)
	}
	first, bridgeProductGetError := bridgeProductGet("reuse", build)
	if bridgeProductGetError != nil {
		test.Fatal(bridgeProductGetError)
	}
	second, bridgeProductGetError := bridgeProductGet("reuse", build)
	if bridgeProductGetError != nil || first != second || builds != 1 {
		test.Fatalf("product was rebuilt: %d builds, %s, %s, %v", builds, first, second, bridgeProductGetError)
	}
	if writeError := os.WriteFile(filepath.Join(first, "api"), []byte("corrupt"), 0o600); writeError != nil {
		test.Fatal(writeError)
	}
	if _, bridgeProductGetError := bridgeProductGet("reuse", build); bridgeProductGetError == nil {
		test.Fatal("corrupt product was accepted")
	}
	directory, bridgeProductGetError := bridgeProductGet("count", build)
	if bridgeProductGetError != nil {
		test.Fatal(bridgeProductGetError)
	}
	data, readError := os.ReadFile(filepath.Join(directory, "products.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var hashes map[string]string
	if decodeError := json.Unmarshal(data, &hashes); decodeError != nil {
		test.Fatal(decodeError)
	}
	hashes["extra"] = hashes["api"]
	data, encodeError := json.Marshal(hashes)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(filepath.Join(directory, "products.json"), data, 0o600); writeError != nil {
		test.Fatal(writeError)
	}
	if _, bridgeProductGetError := bridgeProductGet("count", build); bridgeProductGetError == nil {
		test.Fatal("wrong product count was accepted")
	}
	test.Log("same hash built once; corrupt bytes and extra product refused")
}
