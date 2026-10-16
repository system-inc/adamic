package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Changed programs come from Git, including paths outside the generated corpora.
// Disabling rename detection treats a renamed program as an added head input.
func changedPrograms(root, base, head string) ([]program, error) {
	listing := execute(root, time.Minute, "git", "diff", "--name-only", "--diff-filter=AM", "--no-renames", "-z", base, head, "--", "*.a")
	if listing.Exit != 0 || listing.Error != "" {
		return nil, fmt.Errorf("diff corpus: %s %s", listing.Error, listing.Stderr)
	}
	programs := []program{}
	for _, path := range strings.Split(listing.Stdout, "\x00") {
		if path == "" {
			continue
		}
		blob, err := git(root, "rev-parse", head+":"+path)
		if err != nil {
			return nil, err
		}
		programs = append(programs, program{Path: path, Blob: blob})
	}
	sort.Slice(programs, func(i, j int) bool { return programs[i].Path < programs[j].Path })
	return programs, nil
}
