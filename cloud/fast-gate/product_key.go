//go:build ignore

// This helper is compiled inside the candidate module, then removed from its tree.
// Key remains the candidate's implementation; this file only transports JSON.
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/system-inc/adamic/internal/buildcache"
)

func main() {
	var request struct {
		Root    string
		Recipes []buildcache.Inputs
	}
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		panic(err)
	}
	var keys, paths []string
	for _, recipe := range request.Recipes {
		key, err := buildcache.Key(request.Root, recipe)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		keys = append(keys, key)
		for _, name := range recipe.Files {
			err := filepath.WalkDir(filepath.Join(request.Root, name), func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() && entry.Name() == ".git" {
					return filepath.SkipDir
				}
				relative, err := filepath.Rel(request.Root, path)
				if err != nil {
					return err
				}
				paths = append(paths, filepath.ToSlash(relative))
				return nil
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}
	sort.Strings(paths)
	if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"keys": keys, "paths": paths}); err != nil {
		panic(err)
	}
}
