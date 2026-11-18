package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/appleimports"
	"github.com/system-inc/adamic/internal/platform"
)

// A platform stop has two independent keys: our parse of the revision's modules,
// and the compiler's typed stop. Neither a diagnostic's words nor a path grants it.
type platformJudge struct {
	root, base, head string
	revisions        map[string]*revisionImports
}

type revisionImports struct {
	root, revision string
	files          map[string]bool
	graph          *appleimports.Graph
}

func (j *platformJudge) revision(sha string) (*revisionImports, error) {
	if r := j.revisions[sha]; r != nil {
		return r, nil
	}
	listing := execute(j.root, time.Minute, "git", "ls-tree", "-r", "--name-only", "-z", sha)
	if listing.Exit != 0 || listing.Error != "" {
		return nil, fmt.Errorf("darwin revision %s: %s %s", sha, listing.Error, listing.Stderr)
	}
	r := &revisionImports{root: j.root, revision: sha, files: map[string]bool{}}
	for _, file := range strings.Split(listing.Stdout, "\x00") {
		r.files[file] = true
	}
	r.graph = appleimports.New(func(file string) bool { return r.files[file] }, func(file string) (string, error) {
		source := execute(r.root, time.Minute, "git", "show", r.revision+":"+file)
		if source.Exit != 0 || source.Error != "" {
			return "", fmt.Errorf("darwin source %s:%s: %s %s", r.revision, file, source.Error, source.Stderr)
		}
		return source.Stdout, nil
	})
	j.revisions[sha] = r
	return r, nil
}

// apple says whether file's import graph at this revision reaches an apple/ module (internal/appleimports).
func (r *revisionImports) apple(file string) (bool, error) {
	return r.graph.Reaches(file)
}

func platformStopped(o observation) bool {
	return o.Exit == platform.MissingExitCode && o.Error == ""
}

// stop makes the whole row unjudged, including observations already recorded.
// Only a judged base can move into darwin. The lazy native build is needed
// only when head stopped natively and the base file already reaches apple/.
func (j *platformJudge) stop(record *entry, o observation, baseBuild func() observation) (bool, error) {
	if !platformStopped(o) {
		return false, nil
	}
	head, err := j.revision(j.head)
	if err != nil {
		return false, err
	}
	apple, err := head.apple(record.Path)
	if err != nil {
		return false, err
	}
	record.AppleImport = apple
	if !apple {
		record.Class = "platform-stop-without-apple"
		record.PlatformReason = "exit 5 without a reachable apple/ import at head"
		return true, nil
	}
	base, err := j.revision(j.base)
	if err != nil {
		return false, err
	}
	if base.files[record.Path] {
		apple, err := base.apple(record.Path)
		if err != nil {
			return false, err
		}
		if apple {
			baseStopped := platformStopped(record.Base)
			if baseBuild != nil {
				observed := baseBuild()
				record.BaseNativeCompile = &observed
				baseStopped = platformStopped(observed)
			}
			if !baseStopped {
				record.Class = "moved-into-darwin"
				record.PlatformReason = "a reachable apple/ import was judged at base"
				return true, nil
			}
		}
	}
	record.Class = "darwin"
	record.VoidReason = "darwin: unjudged here; judged by the darwin witnesses"
	return true, nil
}

// A program leaving the platform stop is judged normally. Its recorded base
// observation stays intact; a typed base stop with the first key is not a crash.
func (j *platformJudge) leaving(record *entry) error {
	if !platformStopped(record.Base) {
		return nil
	}
	base, err := j.revision(j.base)
	if err != nil {
		return err
	}
	apple, err := base.apple(record.Path)
	if err != nil {
		return err
	}
	if apple {
		switch compileClass(record.Head) {
		case "accepted":
			record.Class = "newly-accepted"
		case "refused":
			record.Class = "refused-by-both"
		}
	}
	return nil
}
