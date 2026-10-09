package program

import (
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// StatSnapshot holds what the run cache's check found at every path it statted: a file or a directory with its
// stat, or nothing there. A run the check could not replay has just statted every input it recorded, the files
// the last build read, the directories holding them and the paths it looked for and did not find, and the build
// that follows asks the disk the same questions again. Two consumers read the answers instead (#kdee854):
//
//   - the content pack validates a file it serves against the check's stat rather than taking one: about 18ms of
//     a graph phase's 32ms of serving on ahra, at 12 workers;
//   - the build's existence checks and stats are answered from it (Options.CheckedStats), most of all module
//     resolution's, which asks whether a path exists about 35,000 times on an ahra edit run and is told no for
//     most of them, two stats each beneath the recorder.
//
// It is sound because nothing the run reads can be newer than its clock. The run-cache clock starts before the
// check (startRunCacheClock), so a path that changed after its stat here, a file edited or a file created where
// the build was told nothing is, moved its own times or its directory's past the run's readSince, and either
// RecordRunCache refuses to record that run or the next run's check finds the absent path present. Every other
// cache keys on what the build actually saw. The run reports the tree as it stood at the check, and nothing
// replays it past a later change.
//
// The one writer that clock cannot see is the run itself. A graph rebuilt after the fix phase rewrote files must
// not read through a snapshot taken before the rewrite (distrustCheckStats in the command).
type StatSnapshot struct {
	// answers maps a path to its os.FileInfo, or to absentAnswer when the check found nothing there.
	answers sync.Map
}

// absentAnswer is a path the check found nothing at.
type absentAnswer struct{}

// NewStatSnapshot returns an empty snapshot for a check to fill.
func NewStatSnapshot() *StatSnapshot {
	return &StatSnapshot{}
}

// note keeps what a stat already taken said about path: its information, or that nothing is there. An error
// other than nothing being there, a permission error, notes nothing, so a build asks the disk itself.
func (s *StatSnapshot) note(path string, information os.FileInfo, err error) {
	switch {
	case err == nil:
		s.answers.Store(path, information)
	case os.IsNotExist(err):
		s.answers.Store(path, absentAnswer{})
	}
}

// answer is what the check found at path: present with its information, absent, or not statted at all. Any
// spelling the check did not stat, a trailing slash, a doubled separator or a dot segment, is not found here and
// goes to the disk.
func (s *StatSnapshot) answer(path string) (information os.FileInfo, present bool, known bool) {
	// The disk refuses a name with a trailing slash, the root excepted: since typed file paths (#64159) a stat
	// goes through io/fs, whose paths have none, where before the separator was removed first.
	if s == nil || len(path) > 1 && strings.HasSuffix(path, "/") {
		return nil, false, false
	}
	value, found := s.answers.Load(path)
	if !found {
		return nil, false, false
	}
	if information, isPresent := value.(os.FileInfo); isPresent {
		return information, true, true
	}
	return nil, false, true
}

// identityOf is the identity noted for path, if the check statted a file there.
func (s *StatSnapshot) identityOf(path string) (fileIdentity, bool) {
	information, present, _ := s.answer(path)
	if !present || information.IsDir() {
		return fileIdentity{}, false
	}
	return identityFromInfo(information)
}

// checkedStatsFS answers existence checks and stats from the run cache's check where it statted the path, and
// asks the disk beneath for everything else. It answers as the disk does (vfs's Common): a stat follows a
// symbolic link, a file exists when something that is not a directory is there, and a directory when a directory
// is. Listings, reads and realpaths pass through.
type checkedStatsFS struct {
	vfs.FS
	snapshot *StatSnapshot
	answered *atomic.Int64
}

func (f *checkedStatsFS) FileExists(path tspath.RootedFilePath) bool {
	if information, present, known := f.snapshot.answer(path.AsString()); known {
		f.answered.Add(1)
		return present && !information.IsDir()
	}
	return f.FS.FileExists(path)
}

func (f *checkedStatsFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	if information, present, known := f.snapshot.answer(path.AsString()); known {
		f.answered.Add(1)
		return present && information.IsDir()
	}
	return f.FS.DirectoryExists(path)
}

func (f *checkedStatsFS) Stat(path tspath.RootedPath) vfs.FileInfo {
	if information, present, known := f.snapshot.answer(path.AsString()); known {
		f.answered.Add(1)
		if !present {
			return nil
		}
		return information
	}
	return f.FS.Stat(path)
}
