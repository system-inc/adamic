//go:build !windows

package replace

// holdsTheDestination is false off Windows, where nothing an open file can do refuses a rename.
func holdsTheDestination(error) bool {
	return false
}
