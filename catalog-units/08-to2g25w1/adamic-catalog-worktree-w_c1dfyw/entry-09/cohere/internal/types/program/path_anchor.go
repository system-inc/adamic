package program

import (
	"path/filepath"
	"strings"
)

// PathAnchor names paths relative to a project root, so that what the caches store and hash does not depend
// on how the root was spelled (#547dhjz).
//
// One directory can have several absolute spellings: on macOS /tmp is a symbolic link to /private/tmp, and a
// project under a linked directory is reached through the link from one shell and through its target from
// another. Every cache entry and fingerprint used to hold the spelled path, and the cache directory, inside the
// project, is the same directory under every spelling. So a run from the other spelling missed everything
// and then rewrote the files the first spelling's next run would read: alternating, every run checked every
// file, and nothing said why.
//
// A path under the root, reached through either the spelling this run was given or the root's resolved
// target, is stored as "root:" and its path below the root, slash separated. Any other path is stored as it
// is. The compiler names files by a key that is lowercased on a case-insensitive file system, so both roots
// are matched in that form too. Reading back, a stored path is joined to the root as this run spells it, so
// everything in memory, and everything printed, stays in the caller's spelling.
type PathAnchor struct {
	spelled  string
	resolved string

	// prefixes are the roots a path below one starts with, each ending in its separator: both roots, as
	// spelled and in the compiler's lowercased slash form. Computed once, since fingerprints ask per file.
	prefixes []string
}

// stablePrefix marks a path stored relative to the root.
const stablePrefix = "root:"

// NewPathAnchor anchors paths at root, as this run spells it. A root whose target cannot be resolved is
// anchored at its spelling alone, which is what every run did before.
func NewPathAnchor(root string) PathAnchor {
	anchor := PathAnchor{spelled: filepath.Clean(root)}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		anchor.resolved = filepath.Clean(resolved)
	}
	seen := map[string]bool{}
	for _, candidate := range []string{anchor.spelled, anchor.resolved} {
		if candidate == "" {
			continue
		}
		for _, form := range []string{candidate, filepath.ToSlash(candidate), strings.ToLower(filepath.ToSlash(candidate))} {
			for _, separator := range []string{string(filepath.Separator), "/"} {
				prefix := strings.TrimSuffix(form, separator) + separator
				if !seen[prefix] {
					seen[prefix] = true
					anchor.prefixes = append(anchor.prefixes, prefix)
				}
			}
		}
	}
	return anchor
}

// Root is the root as this run spells it.
func (anchor PathAnchor) Root() string { return anchor.spelled }

// Resolved is the root's target with every symbolic link resolved, or its spelling when that failed. A key
// that must be the same for every spelling of one directory is built from this.
func (anchor PathAnchor) Resolved() string {
	if anchor.resolved == "" {
		return anchor.spelled
	}
	return anchor.resolved
}

// Stable is path as the caches store and hash it: below the root, "root:" and its slash path from there, by
// either spelling of the root in either case; anywhere else, unchanged. The zero anchor changes nothing.
func (anchor PathAnchor) Stable(path string) string {
	for _, prefix := range anchor.prefixes {
		if below, found := strings.CutPrefix(path, prefix); found {
			return stablePrefix + filepath.ToSlash(below)
		}
		if path == prefix[:len(prefix)-1] {
			return stablePrefix
		}
	}
	return path
}

// Spelled is a stored path as this run spells it: one below the root joined to the root as given, and any
// other as it was stored.
func (anchor PathAnchor) Spelled(stored string) string {
	below, found := strings.CutPrefix(stored, stablePrefix)
	if !found {
		return stored
	}
	return filepath.Join(anchor.spelled, filepath.FromSlash(below))
}
