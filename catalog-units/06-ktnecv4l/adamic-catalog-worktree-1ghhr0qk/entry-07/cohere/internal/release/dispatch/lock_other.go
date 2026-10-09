//go:build !unix

package dispatch

import (
	"fmt"
	"runtime"
)

// lockExclusive refuses on a platform without flock.
//
// Two runs building the committed Swift engine at once would check out different commits in one
// worktree under each other. Refusing is loud; racing would build an engine from a tree neither run
// asked for, and nothing would say so.
func lockExclusive(path string) (func(), error) {
	return nil, fmt.Errorf("building the Swift engine from a commit needs a file lock, and none is implemented for %s, so nothing was built (%s)",
		runtime.GOOS, path)
}
