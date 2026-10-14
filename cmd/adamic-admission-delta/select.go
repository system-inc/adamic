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
	index, count := 1, 1
	if shard != "" {
		parts := strings.Split(shard, "/")
		if len(parts) != 2 {
			return nil, false, fmt.Errorf("shard must be i/N, with 1 <= i <= N")
		}
		var err error
		index, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, false, fmt.Errorf("invalid shard %q", shard)
		}
		count, err = strconv.Atoi(parts[1])
		if err != nil || count < 1 || index < 1 || index > count {
			return nil, false, fmt.Errorf("invalid shard %q", shard)
		}
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
