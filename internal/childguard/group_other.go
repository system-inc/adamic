//go:build !linux

package childguard

// awaitExit reports false: off Linux the standard library doesn't use waitid's `WNOWAIT` either (darwin's waitid
// returns for a stopped process even with `WEXITED`, golang.org/issue/19314), so the leader's exit is seen only by
// reaping it, and processGroup.reap says what that gives up.
func awaitExit(int) bool {
	return false
}
