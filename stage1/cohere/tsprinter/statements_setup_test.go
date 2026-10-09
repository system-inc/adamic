package tsprinter

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// Shared products are prepared eagerly before m.Run, never by a first shard.
// The subprocess owns the top-level setup test and its independent 90s limit.
// No setup work is included in a shard's test clock or case-command deadline.
type statementSharedProducts struct {
	Cases, Want, Specs, Port, CompilerHash, LoweredDir string
	Binary, Release, Library, LibraryHash              string
	Elapsed                                            float64
}

var statementSharedReady struct {
	products statementSharedProducts
	err      error
	skip     string
}

// This registry holds only per-leaf deadlines, never lazy build state.
var statementCaseContexts sync.Map

// Not parallel: TestMain prepares immutable statement inputs before m.Run.
func TestMain(m *testing.M) {
	flag.Parse()
	if flag.Lookup("test.list").Value.String() == "" && os.Getenv("ADAMIC_STATEMENTS_SETUP_CHILD") != "1" && !statementShardProofEnabled() && statementSetupSelected(flag.Lookup("test.run").Value.String()) {
		if os.Getenv("ADAMIC_TYPESCRIPT_SOURCE") == "" || os.Getenv("ADAMIC_TS_PRETTIER") == "" {
			statementSharedReady.skip = "set ADAMIC_TYPESCRIPT_SOURCE and ADAMIC_TS_PRETTIER for statement inputs"
		} else {
			directory, err := os.MkdirTemp("", "adamic-statements-shared-")
			if err != nil {
				statementSharedReady.err = err
			} else {
				defer os.RemoveAll(directory)
				statementPrepareBeforeTests(filepath.Join(directory, "ready.json"))
			}
		}
	}
	m.Run()
}

func statementSetupSelected(pattern string) bool {
	// A slash begins a subtest selector; these units have only top-level names.
	expression, err := regexp.Compile(strings.SplitN(pattern, "/", 2)[0])
	if err != nil {
		return false
	} // testing reports invalid run patterns itself.
	if expression.MatchString("TestStatementsAgainstGoAndPrettier_Setup") || expression.MatchString("TestStatementsAgainstGoAndPrettierUnion") {
		return true
	}
	for number := 0; number < testStatementsAgainstGoAndPrettierShards; number++ {
		if expression.MatchString(fmt.Sprintf("TestStatementsAgainstGoAndPrettier_%03d", number)) {
			return true
		}
	}
	return false
}

func statementPrepareBeforeTests(manifest string) {
	executable, err := os.Executable()
	if err != nil {
		statementSharedReady.err = err
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := statementContextCommand(ctx, executable, "-test.run=^TestStatementsAgainstGoAndPrettier_Setup$", "-test.timeout=90s", "-test.parallel=1", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_STATEMENTS_SETUP_CHILD=1", "ADAMIC_STATEMENTS_SETUP_MANIFEST="+manifest)
	output, err := command.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			err = fmt.Errorf("cooked: shared setup exceeded 90s: %w", ctx.Err())
		}
		statementSharedReady.err = fmt.Errorf("statement shared setup: %w\n%s", err, output)
		return
	}
	data, err := os.ReadFile(manifest)
	if err == nil {
		err = json.Unmarshal(data, &statementSharedReady.products)
	}
	statementSharedReady.err = err
}

func statementReadyShared(t *testing.T) statementSharedProducts {
	t.Helper()
	if statementSharedReady.skip != "" {
		t.Skip(statementSharedReady.skip)
	}
	if statementSharedReady.err != nil {
		t.Fatal(statementSharedReady.err)
	}
	if statementSharedReady.products.Binary == "" {
		t.Fatal("shared setup was not prepared before tests")
	}
	return statementSharedReady.products
}

func TestStatementsAgainstGoAndPrettier_Setup(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_STATEMENTS_SETUP_CHILD") == "1" {
		products := statementPrepareCommon(t)
		data, err := json.Marshal(products)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(os.Getenv("ADAMIC_STATEMENTS_SETUP_MANIFEST"), data, 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	products := statementReadyShared(t)
	t.Logf("shared setup subprocess: %.3fs; all common products ready before shards", products.Elapsed)
}

func TestStatementsSetupSelection(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{"", "TestStatementsAgainstGoAndPrettier", "^TestStatementsAgainstGoAndPrettier_005$", "^TestStatementsAgainstGoAndPrettier_Setup$", "^TestStatementsAgainstGoAndPrettierUnion$"} {
		if !statementSetupSelected(pattern) {
			t.Fatalf("no eager setup for %q", pattern)
		}
	}
	for _, pattern := range []string{"^TestStatementsShardDisagreement$", "^TestNumberConstructor$", "["} {
		if statementSetupSelected(pattern) {
			t.Fatalf("unrelated tests start statement setup: %q", pattern)
		}
	}
}
