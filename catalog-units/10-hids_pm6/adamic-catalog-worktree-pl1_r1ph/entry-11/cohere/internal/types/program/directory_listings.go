package program

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// DirectoryListings are directory listings a run has already read, for a build to serve its own listings
// from (Options.Listings): discovery reads every directory the repository keeps, and the tsconfig's include
// enumeration then read 2,399 of the same directories again, one at a time, 68ms of a cold ahra run
// (#bjv0tg4, from cache's trace on #g3046x5).
//
// A listing is the entries as os.ReadDir returned them, and a build reads it the way the real disk answers
// GetAccessibleEntries: regular files and directories as they are, a symbolic link by what it points to
// (asked of the disk then, as the real disk does), and anything else left out. A directory nobody listed is
// read from the disk. The listings are as fresh as the walk that read them, a moment before the build, which
// is the same promise the build's own memoizing layer makes for the listings it reads itself.
//
// Safe for concurrent use: discovery adds from its parallel walk, and a build reads afterwards.
type DirectoryListings struct {
	mutex   sync.RWMutex
	entries map[string][]os.DirEntry
}

// NewDirectoryListings holds no listing yet.
func NewDirectoryListings() *DirectoryListings {
	return &DirectoryListings{entries: map[string][]os.DirEntry{}}
}

// Add keeps one directory's listing, by its absolute path.
func (l *DirectoryListings) Add(directory string, entries []os.DirEntry) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.entries[filepath.ToSlash(filepath.Clean(directory))] = entries
}

// Len is how many directories are held.
func (l *DirectoryListings) Len() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return len(l.entries)
}

// Listing is a directory's entries as os.ReadDir returned them, sorted by name, and whether it was listed.
// The directory is an absolute path, slash or platform separated. An entry's Type() needs no stat, except a
// symbolic link's target; entry.Info() stats again. A directory that was not listed is not empty: read it
// from disk. The format walk reads discovery's listings through this (#g3046x5), as the build does through
// listingFS.
func (l *DirectoryListings) Listing(directory string) ([]os.DirEntry, bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	entries, listed := l.entries[path.Clean(filepath.ToSlash(directory))]
	return entries, listed
}

// listingFS serves GetAccessibleEntries from the listings and everything else from the disk beneath.
type listingFS struct {
	vfs.FS
	listings *DirectoryListings
}

func (l *listingFS) GetAccessibleEntries(directory tspath.RootedDirectoryPath) vfs.Entries {
	entries, listed := l.listings.Listing(directory.AsString())
	if !listed {
		return l.FS.GetAccessibleEntries(directory)
	}
	// What the real disk returns for the same entries (vfs/internal Common.GetAccessibleEntries).
	result := vfs.Entries{Symlinks: map[string]struct{}{}}
	add := func(name string, mode fs.FileMode, isLink bool) bool {
		switch {
		case mode.IsDir():
			result.Directories = append(result.Directories, name)
		case mode.IsRegular():
			result.Files = append(result.Files, name)
		default:
			return false
		}
		if isLink {
			result.Symlinks[name] = struct{}{}
		}
		return true
	}
	for _, entry := range entries {
		if add(entry.Name(), entry.Type(), false) {
			continue
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			if information := l.FS.Stat(tspath.RootedPath(directory.AsString() + "/" + entry.Name())); information != nil {
				add(entry.Name(), information.Mode(), true)
			}
		}
	}
	return result
}
