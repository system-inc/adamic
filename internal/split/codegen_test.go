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
func runCodegenSnapshot(test *testing.T, shard int) {
	directory := os.Getenv("ADAMIC_SPLIT_SNAPSHOT")
	if directory == "" {
		directory = test.TempDir()
	}
	count := 0
	for _, path := range compilingOracleFixtures(test) {
		if oracleFixtureShard(path) != shard {
			continue
		}
		snapshotOracleFixture(test, directory, path)
		count++
	}
	test.Logf("%d registered compiling oracle sources snapshotted", count)
}

func snapshotOracleFixture(test *testing.T, directory, path string) {
	test.Helper()
	err := func() error {
		relative, relativePathError := filepath.Rel("../oracle/testdata", path)
		if relativePathError != nil {
			return relativePathError
		}
		output := filepath.Join(directory, relative)
		if createDirectoryError := os.MkdirAll(filepath.Dir(output), 0755); createDirectoryError != nil {
			return createDirectoryError
		}
		loaded, problem := load.Load([]string{path})
		status := "ok"
		if problem == nil {
			sources, decodeJsonSourcesError := lower.DecodeJsonSources(context.Background(), loaded)
			descriptor := any(sources)
			if decodeJsonSourcesError != nil {
				descriptor = decodeJsonSourcesError.Error()
			}
			encoded, encodeError := json.Marshal(descriptor)
			if encodeError != nil {
				return encodeError
			}
			if writeError := os.WriteFile(output+".oracle.json", encoded, 0644); writeError != nil {
				return writeError
			}
			program, lowerError := lower.Lower(context.Background(), loaded)
			problem = lowerError
			if lowerError == nil {
				test.Logf("emitting %s", path)
				for extension, source := range map[string]string{
					".c":   native.C(program),
					".mjs": javascript.JavaScript(program),
				} {
					if writeError := os.WriteFile(output+extension, []byte(source), 0644); writeError != nil {
						return writeError
					}
				}
			}
		}
		if problem != nil {
			return problem
		}
		if writeError := os.WriteFile(output+".status", []byte(status), 0644); writeError != nil {
			return writeError
		}
		return nil
	}()
	if err != nil {
		test.Fatalf("%s: %v", path, err)
	}
}

// Exercise the analysis over the same varied IR that backs the source oracle: every .a and .ts program, the ones
// that don't load or lower skipped as they come. A .d.ts is no program, only declarations another file names
// (ambient_refused/reference_target.d.ts, reference_path.a's), so it isn't walked.
// Shard by fixture path so each unit lowers only its share of the corpus.
func runAnalyzeOracleFixtures(test *testing.T, shard int) {
	count := 0
	// Reads: internal/oracle/testdata/*
	walkError := filepath.WalkDir("../oracle/testdata", func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		source := filepath.Ext(path) == ".a" || (filepath.Ext(path) == ".ts" && !strings.HasSuffix(path, ".d.ts"))
		if entry.IsDir() || !source || oracleFixtureShard(path) != shard {
			return nil
		}
		loaded, loadError := load.Load([]string{path})
		if loadError != nil {
			return nil
		}
		program, lowerError := lower.Lower(context.Background(), loaded)
		if lowerError != nil {
			return nil
		}
		if decisions := split.Analyze(program); len(decisions) != len(program.Functions) {
			test.Fatalf("%s: missing function decisions", path)
		}
		count++
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	test.Logf("%d compilable oracle programs analyzed", count)
}
