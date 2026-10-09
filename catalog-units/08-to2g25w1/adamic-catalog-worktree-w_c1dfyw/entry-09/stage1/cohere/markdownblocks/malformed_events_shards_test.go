package markdownblocks

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

const testMdastMalformedEventsShards = 3

// The error modes are a pinned corpus, one case per contiguous shard.
func malformedEventsEnumeration(t *testing.T) [][]string {
	t.Helper()
	cases := []string{"unclosed", "not-open", "mismatch"}
	shards := make([][]string, testMdastMalformedEventsShards)
	seen := make(map[string]int)
	for index, name := range cases {
		shard := index * testMdastMalformedEventsShards / len(cases)
		shards[shard] = append(shards[shard], name)
	}
	for _, names := range shards {
		if len(names) == 0 {
			t.Fatal("empty malformed-events shard")
		}
		for _, name := range names {
			seen[name]++
		}
	}
	for _, name := range cases {
		if seen[name] != 1 {
			t.Fatalf("case %s occurs %d times", name, seen[name])
		}
	}
	if len(seen) != len(cases) {
		t.Fatal("malformed-events shard union differs from enumeration")
	}
	// A planted disagreement must route to exactly one shard. The opt-in
	// actual run changes that case's expected bytes and exercises the oracle check.
	caught := 0
	for _, names := range shards {
		for _, name := range names {
			if name == "not-open" {
				caught++
			}
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement belongs to %d shards", caught)
	}
	t.Logf("shard union: %d cases, each exactly once; %d shards", len(seen), len(shards))
	return shards
}

func malformedEventsShardName(index int) string { return fmt.Sprintf("shard-%03d", index) }

func malformedEventsExpected(name string, truth []byte) []byte {
	expected := []byte("adamic: panic: " + string(truth))
	if os.Getenv("ADAMIC_MALFORMED_EVENTS_PLANT_FAILURE") == "1" && name == "not-open" {
		expected = append(expected, []byte("planted disagreement\n")...)
	}
	return expected
}

func malformedEventsSetupTimer(t *testing.T) func() {
	started := time.Now()
	return func() { t.Logf("TestMdastMalformedEvents (setup): %.3fs", time.Since(started).Seconds()) }
}

// Cache the lowered backend products together, so checking/lowering and code
// generation happen once before any shard starts. No IR serialization is needed.
func malformedEventsProducts(t *testing.T, main string) (string, string) {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:      "markdownblocks-malformed-events-lowered",
		Files:     []string{"stage1/cohere/markdownblocks", "internal", "go.mod", "cohere"},
		Toolchain: []string{runtime.Version()},
	}
	loweredDir := buildcache.Product(t, inputs, func(dir string) error {
		program, err := loweredResult(main)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	options := native.Options{Sanitize: true}
	inputs.Name = "markdownblocks-malformed-events-native"
	inputs.Flags = append(native.Flags(options), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	for _, name := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED"} {
		inputs.Flags = append(inputs.Flags, name+"="+os.Getenv(name))
	}
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	nativeDir := buildcache.Product(t, inputs, func(dir string) error {
		source, err := os.ReadFile(filepath.Join(loweredDir, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "port"), options)
	})
	return filepath.Join(nativeDir, "port"), filepath.Join(loweredDir, "program.mjs")
}

// Each top-level shard is visible to go test -list and the test-only gate.
func TestMdastMalformedEvents_000(t *testing.T) { TestMdastMalformedEvents(t) }
func TestMdastMalformedEvents_001(t *testing.T) { TestMdastMalformedEvents(t) }
func TestMdastMalformedEvents_002(t *testing.T) { TestMdastMalformedEvents(t) }
func TestMdastMalformedEventsUnion(t *testing.T) {
	t.Parallel()
	shards := malformedEventsEnumeration(t)
	if len(shards) != testMdastMalformedEventsShards {
		t.Fatal("shard count")
	}
	// Keep the declaration list checked against the pinned enumeration.
	declarations := []string{"TestMdastMalformedEvents_000", "TestMdastMalformedEvents_001", "TestMdastMalformedEvents_002"}
	if len(declarations) != len(shards) {
		t.Fatal("top-level shard declaration count")
	}
}

func malformedEventsSelectedCases(t *testing.T) []string {
	if t.Name() == "TestMdastMalformedEvents" {
		return nil
	}
	shards := malformedEventsEnumeration(t)
	for index, cases := range shards {
		if t.Name() == fmt.Sprintf("TestMdastMalformedEvents_%03d", index) {
			return cases
		}
	}
	t.Fatalf("unknown malformed-events shard %s", t.Name())
	return nil
}

type malformedEventsProductSet struct{ goBinary, main, fork, nativeBinary, javascriptPath string }

var malformedEventsOnce sync.Once
var malformedEventsShared malformedEventsProductSet

func malformedEventsSharedSetup(t *testing.T, build func() malformedEventsProductSet) malformedEventsProductSet {
	t.Helper()
	malformedEventsOnce.Do(func() { malformedEventsShared = build() })
	if malformedEventsShared.goBinary == "" {
		t.Fatal("malformed-events shared setup did not complete")
	}
	return malformedEventsShared
}
func malformedEventsBuildDirectory(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp(artifactDirectory, "malformed-events-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
