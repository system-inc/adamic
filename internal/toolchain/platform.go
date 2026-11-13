package toolchain

import "fmt"

// PlatformMissingExitCode belongs only to PlatformMissing, including when wrapped.
// Admission delta shares this code with the compiler; diagnostic text is not a key.
const PlatformMissingExitCode = 5

// PlatformMissing says a program needs a platform the build cannot provide.
// Missing names the SDK or tool that platform needs. It is not a program refusal.
type PlatformMissing struct {
	Platform string
	Missing  string
}

func (e *PlatformMissing) Error() string {
	return fmt.Sprintf("%s platform missing: %s", e.Platform, e.Missing)
}
