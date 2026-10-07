// Census runs the real stage-0 gate. It never lowers a checker-rejected program.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

type run struct {
	Roots       []string `json:"roots"`
	Diagnostics []string `json:"diagnostics,omitempty"`
	Kind        string   `json:"kind"`
	Where       string   `json:"where,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	Fix         string   `json:"fix,omitempty"`
	Seconds     float64  `json:"seconds"`
}

func inspect(paths []string) run {
	started := time.Now()
	result := run{Roots: paths}
	program, err := load.Load(paths)
	if err == nil {
		_, err = lower.Lower(context.Background(), program)
		result.Kind = "accepted"
	}
	if err != nil {
		var checked *load.CheckError
		var refused *lower.Refused
		var notYet *lower.NotYet
		switch {
		case errors.As(err, &checked):
			result.Kind = "checker"
			result.Diagnostics = checked.Diagnostics
		case errors.As(err, &refused):
			result.Kind = "Refused"
			result.Where = refused.Where
			result.Reason = refused.What
			result.Fix = refused.Fix
		case errors.As(err, &notYet):
			result.Kind = "NotYet"
			result.Where = notYet.Where
			result.Reason = notYet.What
		default:
			result.Kind = "error"
			result.Reason = err.Error()
		}
	}
	result.Seconds = time.Since(started).Seconds()
	return result
}

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		panic("usage: census source-directory output.jsonl [--non-source-only]")
	}
	if len(os.Args) == 4 && os.Args[3] != "--non-source-only" {
		panic("unknown census option")
	}
	var paths []string
	var entries []string
	nonSourceOnly := len(os.Args) == 4 && os.Args[3] == "--non-source-only"
	err := filepath.WalkDir(os.Args[1], func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && (filepath.Ext(path) == ".ts" || filepath.Ext(path) == ".a") {
			paths = append(paths, path)
		}
		if !entry.IsDir() && (nonSourceOnly && filepath.Ext(path) != ".ts" && filepath.Ext(path) != ".a" || !nonSourceOnly) {
			entries = append(entries, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	sort.Strings(paths)
	sort.Strings(entries)
	output, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer output.Close()
	encoder := json.NewEncoder(output)
	for index, path := range entries {
		result := inspect([]string{path})
		if err := encoder.Encode(result); err != nil {
			panic(err)
		}
		fmt.Fprintf(os.Stderr, "%d/%d %s %s %.3fs\n", index+1, len(entries), path, result.Kind, result.Seconds)
	}
	if nonSourceOnly {
		return
	}
	result := inspect(paths)
	if err := encoder.Encode(result); err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "whole program: %d roots %s %.3fs\n", len(paths), result.Kind, result.Seconds)
}
