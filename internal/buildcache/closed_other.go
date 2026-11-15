//go:build !linux || mips || mipsle || mips64 || mips64le

package buildcache

import "errors"

// closedProof proves nothing here: its reads are statfs's ST_RDONLY, /proc and faccessat2, Linux's, the last called by
// a number mips doesn't share (closed_linux.go), so an unpacked source's state is always refused.
func closedProof(root string) (string, error) {
	return "", errors.New("closure is proven only on Linux, by statfs's read-only flag or faccessat2")
}
