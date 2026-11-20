package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// Shards partition non-diff inputs by path, keeping every changed input whole.
// Partial runs have their own verdict so they cannot stand in for a full gate.
func selectCorpora(corpora []corpus, filter, shard string) ([]corpus, bool, error) {
	index, count, err := parseShard(shard)
	if err != nil {
		return nil, false, err
	}
	known := map[string]bool{}
	for _, c := range corpora {
		known[c.Name] = true
	}
	wanted := map[string]bool{}
	if filter != "" {
		for _, name := range strings.Split(filter, ",") {
			if !known[name] {
				return nil, false, fmt.Errorf("unknown corpus filter %q", name)
			}
			wanted[name] = true
		}
	}
	selected := []corpus{}
	for _, c := range corpora {
		if c.Name == "diff" {
			selected = append(selected, c)
			continue
		}
		if filter != "" && !wanted[c.Name] {
			continue
		}
		part := corpus{Name: c.Name, Programs: []program{}}
		for _, p := range c.Programs {
			digest := sha256.Sum256([]byte(p.Path))
			if int(binary.LittleEndian.Uint64(digest[:8])%uint64(count)) == index-1 {
				part.Programs = append(part.Programs, p)
			}
		}
		selected = append(selected, part)
	}
	return selected, filter == "" && count == 1, nil
}

// parseShard reads a shard as i/N, 1 <= i <= N, and an empty one as the whole, 1/1.
func parseShard(shard string) (int, int, error) {
	if shard == "" {
		return 1, 1, nil
	}
	parts := strings.Split(shard, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("shard must be i/N, with 1 <= i <= N")
	}
	index, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid shard %q", shard)
	}
	count, err := strconv.Atoi(parts[1])
	if err != nil || count < 1 || index < 1 || index > count {
		return 0, 0, fmt.Errorf("invalid shard %q", shard)
	}
	return index, count, nil
}

// shardByIndex keeps, of corpora in their order, each path's first listing whose position in that order, counting
// every path once, is index-1 modulo count. The agreement mode's shards are then disjoint by path and cover the
// manifest together, and a shard's programs can be read off the manifest by hand. Every corpus stays, emptied or not.
func shardByIndex(corpora []corpus, index, count int) []corpus {
	kept := []corpus{}
	seen := map[string]bool{}
	position := 0
	for _, c := range corpora {
		part := corpus{Name: c.Name, Programs: []program{}}
		for _, p := range c.Programs {
			if seen[p.Path] {
				continue
			}
			seen[p.Path] = true
			if position%count == index-1 {
				part.Programs = append(part.Programs, p)
			}
			position++
		}
		kept = append(kept, part)
	}
	return kept
}
