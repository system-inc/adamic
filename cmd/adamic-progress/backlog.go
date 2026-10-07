package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// patchIDs streams diffs rather than retaining potentially large branch patches.
// Stable patch IDs ignore commit identity and whitespace, like git cherry.
// Merge and empty commits have no standalone patch and are excluded.
const patchBatchSize = 16

// Each entry maps one immutable commit SHA to its patch ID (empty for no patch).
func (r repository) patchIDs(revisions []string) (map[string]string, error) {
	if len(revisions) > patchBatchSize {
		return nil, fmt.Errorf("patch batch exceeds %d commits", patchBatchSize)
	}
	args := []string{"-c", "core.quotePath=true", "-c", "diff.algorithm=myers", "log", "--no-walk=unsorted", "--no-merges", "--pretty=medium", "-p", "--no-ext-diff", "--no-textconv", "--no-renames", "--full-index", "--binary", "--src-prefix=a/", "--dst-prefix=b/", "--no-color", "--ignore-submodules=none", "--submodule=short"}
	args = append(args, revisions...)
	producer := r.gitCommand(args...)
	var producerError, consumerError bytes.Buffer
	producer.Stderr = &producerError
	pipe, err := producer.StdoutPipe()
	if err != nil {
		return nil, err
	}
	consumer := r.gitCommand("patch-id", "--stable")
	consumer.Stdin = pipe
	consumer.Stderr = &consumerError
	if err = producer.Start(); err != nil {
		return nil, err
	}
	output, readError := consumer.Output()
	if readError != nil {
		_ = producer.Process.Kill()
	}
	writeError := producer.Wait()
	if readError != nil {
		return nil, fmt.Errorf("git patch-id: %w: %s", readError, consumerError.String())
	}
	if writeError != nil {
		return nil, fmt.Errorf("git log patches: %w: %s", writeError, producerError.String())
	}
	ids := map[string]string{}
	for _, commit := range revisions {
		ids[commit] = ""
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || !validPatchObjectID(fields[0]) || !validPatchObjectID(fields[1]) {
			return nil, fmt.Errorf("invalid git patch-id row %q", line)
		}
		if _, ok := ids[fields[1]]; !ok {
			return nil, fmt.Errorf("unexpected patch commit %s", fields[1])
		}
		ids[fields[1]] = fields[0]
	}
	return ids, nil
}

// patchCandidates groups commits by changed paths. Equal patch IDs require
// identical paths, so this avoids hashing unrelated main diffs without losing
// rebase matches. Filenames are Git-quoted consistently, including newlines.
func (r repository) patchCandidates(revisions []string) (map[string][]string, error) {
	args := []string{"-c", "core.quotePath=true", "log", "--no-merges", "--root", "--name-only", "--no-renames", "--no-ext-diff", "--no-textconv", "--format=%x00%H%x00", "--ignore-submodules=none"}
	args = append(args, revisions...)
	output, err := r.git(args...)
	if err != nil {
		return nil, err
	}
	groups := map[string][]string{}
	segments := strings.Split(string(output), "\x00")
	for i := 1; i+1 < len(segments); i += 2 {
		commit := segments[i]
		var paths []string
		for _, line := range strings.Split(segments[i+1], "\n") {
			if line == "" {
				continue
			}
			// patch-id strips ASCII whitespace, including in diff path headers.
			line = strings.Map(func(c rune) rune {
				if strings.ContainsRune(" \t\r\n\v\f", c) {
					return -1
				}
				return c
			}, line)
			paths = append(paths, line)
		}
		if len(paths) > 0 {
			sort.Strings(paths)
			key := strings.Join(paths, "\n")
			groups[key] = append(groups[key], commit)
		}
	}
	return groups, nil
}

type patchCache struct {
	Version string            `json:"version"`
	Commits map[string]string `json:"commits"`
}

const patchCacheVersion = "stable-myers-no-renames-binary-v1"

type backlogResult struct {
	Count     int     `json:"-"`
	Hashed    int     `json:"patch_ids_hashed"`
	Cached    int     `json:"patch_ids_cached"`
	Batches   int     `json:"patch_batches"`
	CacheFile string  `json:"cache_file"`
	Seconds   float64 `json:"wall_seconds"`
}

func (r repository) patchBacklog(pending []string) (int, error) {
	result, err := r.cachedPatchBacklog(pending)
	return result.Count, err
}

func savePatchCache(file string, cache patchCache) error {
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(file), "patch-ids-*.tmp")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if _, err = temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(name, file)
}

func (r repository) cachedPatchBacklog(pending []string) (result backlogResult, err error) {
	start := time.Now()
	defer func() { result.Seconds = time.Since(start).Seconds() }()
	budget := r.backlogBudget
	if budget <= 0 {
		budget = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.ctx, budget)
	defer cancel()
	r.ctx = ctx
	defer func() {
		if ctx.Err() != nil {
			err = fmt.Errorf("backlog deadline/cancellation after %s: %w", budget, ctx.Err())
		}
	}()
	if len(pending) == 0 {
		return result, nil
	}
	revisions := append(append([]string{}, pending...), "--not", r.ref)
	waitingGroups, err := r.patchCandidates(revisions)
	if err != nil {
		return result, err
	}
	mainGroups, err := r.patchCandidates([]string{r.ref})
	if err != nil {
		return result, err
	}
	waitingCommits, mainCommits := map[string]bool{}, map[string]bool{}
	for paths, commits := range waitingGroups {
		for _, commit := range commits {
			waitingCommits[commit] = true
		}
		for _, commit := range mainGroups[paths] {
			mainCommits[commit] = true
		}
	}
	file, err := r.git("rev-parse", "--git-path", "adamic-progress/patch-ids-v1.json")
	if err != nil {
		return result, err
	}
	result.CacheFile = strings.TrimSpace(string(file))
	if !filepath.IsAbs(result.CacheFile) {
		result.CacheFile = filepath.Join(r.root, result.CacheFile)
	}
	result.CacheFile, err = filepath.Abs(result.CacheFile)
	if err != nil {
		return result, err
	}
	cache := patchCache{Version: patchCacheVersion, Commits: map[string]string{}}
	data, err := os.ReadFile(result.CacheFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return result, fmt.Errorf("patch cache: %w", err)
	}
	if err == nil {
		if err = json.Unmarshal(data, &cache); err != nil {
			return result, fmt.Errorf("patch cache: %w", err)
		}
		if cache.Version != patchCacheVersion || cache.Commits == nil {
			return result, errors.New("unsupported patch cache format")
		}
		for commit, id := range cache.Commits {
			if !validPatchObjectID(commit) || (id != "" && !validPatchObjectID(id)) {
				return result, errors.New("invalid patch cache object ID")
			}
		}
	}
	required := map[string]bool{}
	for commit := range waitingCommits {
		required[commit] = true
	}
	for commit := range mainCommits {
		required[commit] = true
	}
	var missing []string
	for commit := range required {
		if _, ok := cache.Commits[commit]; ok {
			result.Cached++
		} else {
			missing = append(missing, commit)
		}
	}
	sort.Strings(missing)
	for offset := 0; offset < len(missing); offset += patchBatchSize {
		if err = ctx.Err(); err != nil {
			return result, err
		}
		end := offset + patchBatchSize
		if end > len(missing) {
			end = len(missing)
		}
		ids, err := r.patchIDs(missing[offset:end])
		if err != nil {
			return result, err
		}
		for commit, id := range ids {
			cache.Commits[commit] = id
		}
		// Checkpoint only completed batches, using an atomic rename. Concurrent
		// invocations can lose cache entries, but never publish partial JSON.
		if err = savePatchCache(result.CacheFile, cache); err != nil {
			return result, fmt.Errorf("patch cache write: %w", err)
		}
		result.Hashed += end - offset
		result.Batches++
	}
	landed := map[string]bool{}
	for commit := range mainCommits {
		if id := cache.Commits[commit]; id != "" {
			landed[id] = true
		}
	}
	waiting := map[string]bool{}
	for commit := range waitingCommits {
		if id := cache.Commits[commit]; id != "" && !landed[id] {
			waiting[id] = true
		}
	}
	result.Count = len(waiting)
	return result, nil
}

func validPatchObjectID(id string) bool {
	if len(id) != 40 && len(id) != 64 {
		return false
	}
	for _, c := range id {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
