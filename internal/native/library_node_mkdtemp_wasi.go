package native

import "strings"

// WASIMkdtempRefusal identifies this target capability gap before C compilation.
// A target oracle can record it as a refusal rather than a failed executable.
type WASIMkdtempRefusal struct{}

func (*WASIMkdtempRefusal) Error() string {
	return "native: WASI preview1 refuses node:fs.mkdtempSync: path_create_directory cannot enforce private directory mode 0700, and path_open cannot create a directory"
}

// Preview 1 can atomically create a directory, but cannot set its permissions.
// Node's mkdtemp requires 0700 regardless of the host's broader umask. Refuse
// before invoking clang instead of silently creating a public directory.
func validateNodeMkdtempTarget(source string, options Options) error {
	if options.Target == "wasm32-wasi" && strings.Contains(source, "#define ADAMIC_NODE_MKDTEMP 1\n") {
		return &WASIMkdtempRefusal{}
	}
	return nil
}
