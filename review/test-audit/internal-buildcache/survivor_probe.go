package main

import (
	"encoding/json"
	"fmt"
	cache "github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, err := os.MkdirTemp("/tmp/u015", "witness-")
	if err != nil {
		panic(err)
	}
	os.Setenv("ADAMIC_BUILD_CACHE_DIR", filepath.Join(root, "cache"))
	out := map[string]any{"mode": os.Args[1]}
	switch os.Args[1] {
	case "key":
		os.WriteFile(filepath.Join(root, "input"), []byte("fixed"), 0600)
		key, err := cache.Key(root, cache.Inputs{Name: "witness", Files: []string{"input"}, Flags: []string{"-flag"}, Toolchain: []string{"tool"}})
		out["key"], out["error"] = key, fmt.Sprint(err)
	case "outside":
		repo := filepath.Join(root, "repo")
		os.Mkdir(repo, 0700)
		os.WriteFile(filepath.Join(root, "escape"), []byte("owned fixture"), 0600)
		key, err := cache.Key(repo, cache.Inputs{Name: "witness", Files: []string{"../escape"}})
		out["key"], out["error"] = key, fmt.Sprint(err)
	case "lock":
		inputs := cache.Inputs{Name: "lock-witness"}
		first, err := cache.Get(inputs, func(string) error { return nil })
		if err != nil {
			panic(err)
		}
		os.Remove(first + ".lock")
		os.Mkdir(first+".lock", 0700)
		second, err := cache.Get(inputs, func(string) error { return nil })
		out["reused"], out["error"] = first == second, fmt.Sprint(err)
	case "describe":
		product, err := cache.Get(cache.Inputs{Name: "describe-witness", Toolchain: []string{"tool-witness"}}, func(string) error { return nil })
		if err != nil {
			panic(err)
		}
		data, err := os.ReadFile(product + ".inputs")
		out["description"], out["error"] = string(data), fmt.Sprint(err)
	case "tool":
		script := filepath.Join(root, "tool.sh")
		count := filepath.Join(root, "count")
		os.Setenv("ADAMIC_WITNESS_COUNT", count)
		os.WriteFile(script, []byte("#!/bin/sh\nprintf '.\\n' >> \"$ADAMIC_WITNESS_COUNT\"\nprintf 'fixed\\n'\n"), 0700)
		first := cache.Tool(script)
		second := cache.Tool(script)
		data, _ := os.ReadFile(count)
		out["same_report"], out["executions"] = first == second, len(strings.Fields(string(data)))
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
