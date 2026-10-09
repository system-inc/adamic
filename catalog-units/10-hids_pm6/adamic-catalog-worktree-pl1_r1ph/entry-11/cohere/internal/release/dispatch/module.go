package dispatch

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// FindModuleDirectory walks up from this executable and from the working directory looking for the
// cohere module.
//
// It returns an error naming what it looked for rather than guessing. A resolver that gives up and
// returns a bare command name is exactly the bug this tool was built to stop shipping.
//
// Shared by the launcher, which builds cohere from this module, and by cohere itself, which builds
// the Swift engine from the `swift/` directory beside it. Both ask the same question and must get the
// same answer, so there is one walk.
func FindModuleDirectory() (string, error) {
	candidates := []string{}

	if executable, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(executable); err == nil {
			executable = resolved
		}
		candidates = append(candidates, filepath.Dir(executable))
	}
	if workingDirectory, err := os.Getwd(); err == nil {
		candidates = append(candidates, workingDirectory)
	}

	for _, candidate := range candidates {
		if directory, found := walkUpForModule(candidate); found {
			return directory, nil
		}
	}

	return "", errors.New("could not find the cohere module: no go.mod declaring github.com/system-inc/cohere in any parent of this binary or the working directory")
}

// walkUpForModule climbs toward the filesystem root looking for the cohere module's go.mod.
func walkUpForModule(start string) (string, bool) {
	directory := start
	for {
		contents, err := os.ReadFile(filepath.Join(directory, "go.mod"))
		if err == nil && isCohereModule(contents) {
			return directory, true
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", false
		}
		directory = parent
	}
}

// isCohereModule reports whether a go.mod declares this module.
//
// The module path is matched rather than just the presence of a go.mod, so that a shim sitting
// inside some other Go project does not mistake that project for its own module.
func isCohereModule(contents []byte) bool {
	for line := range strings.SplitSeq(string(contents), "\n") {
		if strings.TrimSpace(line) == "module github.com/system-inc/cohere" {
			return true
		}
	}
	return false
}
