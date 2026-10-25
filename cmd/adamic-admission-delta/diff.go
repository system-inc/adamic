package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Changed programs come from Git, including paths outside the generated corpora.
// Disabling rename detection treats a renamed program as an added head input.
// Every changed .a is a program. A changed .ts is one exactly when the manifest
// lists it (programs, every path of head's corpora), since most .ts files in the
// tree are modules a program imports. Language's rule on #sa5ptyj, so the tool
// keeps no globs of its own. The rest are returned as unmatched, never guessed.
func changedPrograms(root, base, head string, programs map[string]bool) ([]program, []string, error) {
	listing := execute(root, time.Minute, "git", "diff", "--name-only", "--diff-filter=AM", "--no-renames", "-z", base, head, "--", "*.a", "*.ts", ":(exclude)*.d.ts")
	if listing.Exit != 0 || listing.Error != "" {
		return nil, nil, fmt.Errorf("diff corpus: %s %s", listing.Error, listing.Stderr)
	}
	changed := []program{}
	unmatched := []string{}
	for _, path := range strings.Split(listing.Stdout, "\x00") {
		if path == "" {
			continue
		}
		if strings.HasSuffix(path, ".ts") {
			unmatched = append(unmatched, path)
			continue
		}
		blob, err := git(root, "rev-parse", head+":"+path)
		if err != nil {
			return nil, nil, err
		}
		changed = append(changed, program{Path: path, Blob: blob})
	}
	sort.Slice(changed, func(i, j int) bool { return changed[i].Path < changed[j].Path })
	sort.Strings(unmatched)
	return changed, unmatched, nil
}
