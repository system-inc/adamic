package program

import (
	"crypto/sha256"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// DesignSystemKey is what the design-system rules' findings depend on beyond a file's own bytes: every path
// the design system read, present or absent, with what was there. A Tailwind rule's answer on an unchanged
// file can move when a stylesheet does, when a new `@utility` or theme token lands, or when a candidate entry
// point that was missing appears, and each of those is a path in this set (#35nqkwc).
//
// It is checked rather than rebuilt: the next run re-hashes these paths, which costs a few stat calls and a
// few small reads, and replays the design-system rules only if every one still holds. Building the design
// system to compare it would cost the run the replay exists to skip.
type DesignSystemKey struct {
	Reads []DesignSystemKeyRead

	// Fingerprint covers every read, so entries can say which design system their findings were produced
	// under. Never zero for a real key: zero is how an entry says its findings were produced under none.
	Fingerprint [sha256.Size]byte
}

// DesignSystemKeyRead is one path and what was there: nothing, a directory with a listing, or a file with
// bytes. Hash covers the listing or the bytes, and where a symbolic link led. Path is anchored (PathAnchor), so
// the fingerprint over it is one for every spelling of the project root (#547dhjz).
type DesignSystemKeyRead struct {
	Path      string
	Present   bool
	Directory bool
	Hash      [sha256.Size]byte
}

// designSystemKeyFromReads keys the reads a load made. It declines, returning false, when any present path
// moved between the load reading it and this hashing it: its bytes now might not be the bytes the rules saw,
// and a key over the wrong bytes replays a finding against a stylesheet nobody read.
func designSystemKeyFromReads(reads []rule.FileRead, anchor PathAnchor) (*DesignSystemKey, bool) {
	key := &DesignSystemKey{Reads: make([]DesignSystemKeyRead, 0, len(reads))}
	for _, read := range reads {
		if read.Present {
			info, err := os.Stat(read.Path)
			if err != nil || info.Size() != read.Size || info.ModTime().UnixNano() != read.ModifiedNanoseconds {
				return nil, false
			}
		}
		current, err := readDesignSystemPath(read.Path)
		if err != nil || current.Present != read.Present {
			return nil, false
		}
		current.Path = anchor.Stable(read.Path)
		key.Reads = append(key.Reads, current)
	}
	key.Fingerprint = designSystemFingerprint(key.Reads)
	return key, true
}

// emptyDesignSystemKey is the key of a run whose rules never loaded the design system. Its findings depend on
// no stylesheet, so any later state of the stylesheets leaves them true.
func emptyDesignSystemKey() *DesignSystemKey {
	return &DesignSystemKey{Fingerprint: designSystemFingerprint(nil)}
}

// stillHolds reports whether every path is as recorded, each read where anchor spells it. Anything unreadable
// is a no: a key that cannot be checked is a miss, never a hit.
func (key *DesignSystemKey) stillHolds(anchor PathAnchor) bool {
	if key == nil || key.Fingerprint == ([sha256.Size]byte{}) || key.Fingerprint != designSystemFingerprint(key.Reads) {
		return false
	}
	for _, recorded := range key.Reads {
		current, err := readDesignSystemPath(anchor.Spelled(recorded.Path))
		current.Path = recorded.Path
		if err != nil || current != recorded {
			return false
		}
	}
	return true
}

// readDesignSystemPath is what is at path now. Absence is anything that is not there at all: a directory
// where a file was looked for and missed counts as present, and so as a change, which costs a re-run rather
// than risking a stale replay.
func readDesignSystemPath(path string) (DesignSystemKeyRead, error) {
	read := DesignSystemKeyRead{Path: path}
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return read, nil
	}
	if err != nil {
		return read, err
	}
	read.Present = true
	hash := sha256.New()
	if info.Mode()&fs.ModeSymlink != 0 {
		target, err := filepath.EvalSymlinks(path)
		if err != nil {
			return read, err
		}
		hash.Write([]byte("link\x00" + target + "\x00"))
		if info, err = os.Stat(path); err != nil {
			return read, err
		}
	}
	if info.IsDir() {
		read.Directory = true
		entries, err := os.ReadDir(path)
		if err != nil {
			return read, err
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			kind := "f"
			if entry.IsDir() {
				kind = "d"
			}
			names = append(names, kind+entry.Name())
		}
		sort.Strings(names)
		for _, name := range names {
			hash.Write([]byte(name + "\x00"))
		}
	} else {
		contents, err := os.ReadFile(path)
		if err != nil {
			return read, err
		}
		hash.Write(contents)
	}
	copy(read.Hash[:], hash.Sum(nil))
	return read, nil
}

func designSystemFingerprint(reads []DesignSystemKeyRead) [sha256.Size]byte {
	hash := sha256.New()
	hash.Write([]byte("design system key 1\x00"))
	for _, read := range reads {
		hash.Write([]byte(read.Path + "\x00"))
		flags := byte(0)
		if read.Present {
			flags |= 1
		}
		if read.Directory {
			flags |= 2
		}
		hash.Write([]byte{flags})
		hash.Write(read.Hash[:])
	}
	var fingerprint [sha256.Size]byte
	copy(fingerprint[:], hash.Sum(nil))
	return fingerprint
}
