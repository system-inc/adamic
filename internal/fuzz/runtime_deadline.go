package fuzz

import (
	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	boundedrun.Register("runtime-library-clang-ar", func(args []string) (string, error) {
		options := native.Options{Sanitize: true}
		if len(args) > 1 && args[1] == "thread" {
			options = native.Options{ThreadSanitize: true}
		}
		return native.RuntimeLibrary(args[0], options)
	})
}

// Cold sanitized runtime builds fit inside the 114s observed cold setup; ten
// minutes leaves ample margin and bounds all inherited clang/ar descendants.
func boundedRuntimeLibrary(directory string, thread ...bool) (string, error) {
	arguments := []string{directory}
	if len(thread) > 0 && thread[0] {
		arguments = append(arguments, "thread")
	}
	return boundedrun.Operation(boundedrun.Build, "runtime-library-clang-ar", arguments...)
}
