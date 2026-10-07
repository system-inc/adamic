// Command adamic-stage1 builds only the approved stage 1 parse and lint drivers.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func main() {
	driver := flag.String("driver", "", "parse or lint")
	output := flag.String("o", "", "binary output")
	emit := flag.String("emit", "", "also save the exact emitted C")
	policy := flag.String("policy", "profile", "profile, thin, o2 or generate")
	manifest := flag.String("write-manifest", "", "write metadata for a regenerated text profile instead of building")
	training := flag.String("training-hash", "", "committed training manifest hash")
	flag.Parse()
	roots := map[string]string{"parse": "stage1/cohere/parse/parse.a", "lint": "stage1/cohere/lint/main.ts"}
	sourcePath, ok := roots[*driver]
	if !ok {
		fail(fmt.Errorf("driver must be parse or lint"))
	}
	program, err := load.Load([]string{sourcePath})
	if err != nil {
		fail(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		fail(err)
	}
	source := native.C(lowered)
	if *emit != "" {
		if err = os.WriteFile(*emit, []byte(source), 0o644); err != nil {
			fail(err)
		}
	}
	if *manifest != "" {
		data, err := os.ReadFile(*manifest)
		if err != nil {
			fail(err)
		}
		record, err := native.NewStage1ProfileManifest(source, data, *training)
		if err != nil {
			fail(err)
		}
		body, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			fail(err)
		}
		if err = os.WriteFile(filepath.Join(filepath.Dir(*manifest), "manifest.json"), append(body, '\n'), 0o644); err != nil {
			fail(err)
		}
		return
	}
	if *output == "" {
		if *emit != "" {
			return
		}
		fail(fmt.Errorf("-o is required"))
	}
	options := native.Options{Release: true}
	switch *policy {
	case "profile":
		options.Profile = filepath.Join(filepath.Dir(sourcePath), "profiles", runtime.GOOS+"-"+runtime.GOARCH, "profile.txt")
	case "thin":
	case "o2":
		options.Release = false
	case "generate":
		options.ProfileGenerate = true
	default:
		fail(fmt.Errorf("unknown build policy %q", *policy))
	}
	if err = native.Build(source, *output, options); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
