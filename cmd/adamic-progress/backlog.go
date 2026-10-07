package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// patchIDs streams diffs rather than retaining potentially large branch patches.
// Stable patch IDs ignore commit identity and whitespace, like git cherry.
// Merge and empty commits have no standalone patch and are excluded.
func (r repository) patchIDs(revisions []string) (map[string]bool, error) {
	args := []string{"-C", r.root, "log", "--no-walk=unsorted", "--no-merges", "--pretty=medium", "-p", "--no-ext-diff", "--no-textconv", "--no-renames", "--full-index", "--binary"}
	args = append(args, revisions...)
	producer := exec.CommandContext(r.ctx, "git", args...)
	var producerError, consumerError bytes.Buffer
	producer.Stderr = &producerError
	pipe, err := producer.StdoutPipe()
	if err != nil {
		return nil, err
	}
	consumer := exec.CommandContext(r.ctx, "git", "patch-id", "--stable")
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
	ids := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid git patch-id row %q", line)
		}
		ids[fields[0]] = true
	}
	return ids, nil
}

// patchCandidates groups commits by changed paths. Equal patch IDs require
// identical paths, so this avoids hashing unrelated main diffs without losing
// rebase matches. Filenames are Git-quoted consistently, including newlines.
func (r repository) patchCandidates(revisions []string) (map[string][]string, error) {
	args := []string{"log", "--no-merges", "--root", "--name-only", "--no-renames", "--no-ext-diff", "--no-textconv", "--format=%x00%H%x00"}
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
func (r repository) patchBacklog(pending []string) (int, error) {
	if len(pending) == 0 {
		return 0, nil
	}
	revisions := append([]string{}, pending...)
	revisions = append(revisions, "--not", r.ref)
	type candidateResult struct {
		groups map[string][]string
		err    error
	}
	mainCandidates := make(chan candidateResult, 1)
	mainRepo := r
	mainRepo.cache = nil
	go func() {
		groups, err := mainRepo.patchCandidates([]string{r.ref})
		mainCandidates <- candidateResult{groups, err}
	}()
	waitingGroups, err := r.patchCandidates(revisions)
	mainCandidate := <-mainCandidates
	if err != nil {
		return 0, err
	}
	if mainCandidate.err != nil {
		return 0, mainCandidate.err
	}
	mainGroups := mainCandidate.groups
	var mainCommits, waitingCommits []string
	for paths, commits := range waitingGroups {
		waitingCommits = append(waitingCommits, commits...)
		mainCommits = append(mainCommits, mainGroups[paths]...)
	}
	if len(waitingCommits) == 0 {
		return 0, nil
	}
	type result struct {
		ids map[string]bool
		err error
	}
	mainResult := make(chan result, 1)
	go func() {
		if len(mainCommits) == 0 {
			mainResult <- result{ids: map[string]bool{}}
			return
		}
		ids, err := r.patchIDs(mainCommits)
		mainResult <- result{ids: ids, err: err}
	}()
	waiting, err := r.patchIDs(waitingCommits)
	main := <-mainResult
	if err != nil {
		return 0, err
	}
	if main.err != nil {
		return 0, main.err
	}
	landed := main.ids
	count := 0
	for patch := range waiting {
		if !landed[patch] {
			count++
		}
	}
	return count, nil
}
