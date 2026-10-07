package fuzz

import (
	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	boundedrun.Register("runtime-library-clang-ar", func(args []string) (string, error) {
		return native.RuntimeLibrary(args[0], native.Options{Sanitize: true})
	})
}

// Cold sanitized runtime builds fit inside the 114s observed cold setup; ten
// minutes leaves ample margin and bounds all inherited clang/ar descendants.
func boundedRuntimeLibrary(directory string) (string, error) {
	return boundedrun.Operation(boundedrun.Build, "runtime-library-clang-ar", directory)
}
