package split_test

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/split"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Snapshots cover the oracle/testdata .a and .ts fixtures main registers as compiling.
// The directory also holds support modules and unregistered compiler witnesses
// (including known emitter bugs); they are not compiling oracle entry points.
// The analysis family below still sweeps every .a and .ts. The optional snapshot
// directory exports artifacts and never controls whether tests run.
func runCodegenSnapshot(t *testing.T, shard int) {
	directory := os.Getenv("ADAMIC_SPLIT_SNAPSHOT")
	if directory == "" {
		directory = t.TempDir()
	}
	count := 0
	for _, path := range compilingOracleFixtures(t) {
		if oracleFixtureShard(path) != shard {
			continue
		}
		snapshotOracleFixture(t, directory, path)
		count++
	}
	t.Logf("%d registered compiling oracle sources snapshotted", count)
}

func snapshotOracleFixture(t *testing.T, directory, path string) {
	t.Helper()
	err := func() error {
		relative, err := filepath.Rel("../oracle/testdata", path)
		if err != nil {
			return err
		}
		output := filepath.Join(directory, relative)
		if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
			return err
		}
		loaded, problem := load.Load([]string{path})
		status := "ok"
		if problem == nil {
			sources, err := lower.DecodeJsonSources(context.Background(), loaded)
			descriptor := any(sources)
			if err != nil {
				descriptor = err.Error()
			}
			encoded, err := json.Marshal(descriptor)
			if err != nil {
				return err
			}
			if err := os.WriteFile(output+".oracle.json", encoded, 0644); err != nil {
				return err
			}
			program, err := lower.Lower(context.Background(), loaded)
			problem = err
			if err == nil {
				t.Logf("emitting %s", path)
				for extension, source := range map[string]string{".c": native.C(program), ".mjs": javascript.JavaScript(program)} {
					if err := os.WriteFile(output+extension, []byte(source), 0644); err != nil {
						return err
					}
				}
			}
		}
		if problem != nil {
			return problem
		}
		if err := os.WriteFile(output+".status", []byte(status), 0644); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

// Exercise the analysis over the same varied IR that backs the source oracle: every .a and .ts program, the ones
// that don't load or lower skipped as they come. A .d.ts is no program, only declarations another file names
// (ambient_refused/reference_target.d.ts, reference_path.a's), so it isn't walked.
// Shard by fixture path so each unit lowers only its share of the corpus.
func runAnalyzeOracleFixtures(t *testing.T, shard int) {
	count := 0
	// Reads: internal/oracle/testdata/*
	err := filepath.WalkDir("../oracle/testdata", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		source := filepath.Ext(path) == ".a" || (filepath.Ext(path) == ".ts" && !strings.HasSuffix(path, ".d.ts"))
		if entry.IsDir() || !source || oracleFixtureShard(path) != shard {
			return nil
		}
		loaded, err := load.Load([]string{path})
		if err != nil {
			return nil
		}
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			return nil
		}
		if decisions := split.Analyze(program); len(decisions) != len(program.Functions) {
			t.Fatalf("%s: missing function decisions", path)
		}
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d compilable oracle programs analyzed", count)
}
