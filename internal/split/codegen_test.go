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
	"testing"
)

// An opt-in complete emission snapshot lets the base and hook be compared using
// the same source paths and harness, including fixtures deliberately refused.
func TestCodegenSnapshot(t *testing.T) {
	directory := os.Getenv("ADAMIC_SPLIT_SNAPSHOT")
	if directory == "" {
		t.Skip("set ADAMIC_SPLIT_SNAPSHOT to record every oracle fixture")
	}
	count := 0
	err := filepath.WalkDir("../oracle/testdata", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".a" {
			return nil
		}
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
				for extension, source := range map[string]string{".c": native.C(program), ".mjs": javascript.JavaScript(program)} {
					if err := os.WriteFile(output+extension, []byte(source), 0644); err != nil {
						return err
					}
				}
			}
		}
		if problem != nil {
			status = problem.Error()
		}
		if err := os.WriteFile(output+".status", []byte(status), 0644); err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d oracle source files snapshotted", count)
}

// Exercise the analysis over the same varied IR that backs the source oracle.
// This sweep is opt-in because it rebuilds every checked source program.
func TestAnalyzeOracleFixtures(t *testing.T) {
	if os.Getenv("ADAMIC_SPLIT_SWEEP") == "" {
		t.Skip("set ADAMIC_SPLIT_SWEEP for the complete oracle IR sweep")
	}
	count := 0
	err := filepath.WalkDir("../oracle/testdata", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".a" {
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
