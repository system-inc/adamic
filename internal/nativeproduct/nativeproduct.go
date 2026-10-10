// Package nativeproduct builds C with native.Build as a keyed product, so a test reads a binary built ahead instead of
// running clang in its unit (#tze35sy, the build law). Any two tests building the same C with the same options share
// one product, whatever called them, so a TestProduct_ that lowers a program and builds it is what a tree build
// builds, and the test that builds the same program reads it.
package nativeproduct

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// Build is native.Build(source, <product>/program, options) as a product, and returns the binary's path. Its key is
// the source's content, everything about options that changes the binary (native.Flags, whether it is a request
// handler or a split build), the runtime every build links (internal/native, which embeds it) and clang's version.
func Build(t testing.TB, source string, options native.Options) string {
	t.Helper()
	directory := buildcache.Product(t, Inputs(source, options), func(directory string) error {
		return native.Build(source, filepath.Join(directory, "program"), options)
	})
	return filepath.Join(directory, "program")
}

// Inputs is Build's key for source built with options.
func Inputs(source string, options native.Options) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "native.Build",
		Files: []string{"go.mod", "internal/native"},
		Flags: append([]string{
			fmt.Sprintf("source sha256 %x", sha256.Sum256([]byte(source))),
			fmt.Sprintf("request=%t split=%t ADAMIC_NATIVE_SPLIT=%s", options.Request, options.Split, os.Getenv("ADAMIC_NATIVE_SPLIT")),
		}, native.Flags(options)...),
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}
}
