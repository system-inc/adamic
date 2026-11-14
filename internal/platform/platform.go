package platform

import "fmt"

// MissingExitCode belongs only to Missing, including when wrapped.
// Admission delta shares this code with the compiler; diagnostic text is not a key.
const MissingExitCode = 5

// Missing says a program needs a platform the build cannot provide.
// Requirement names the SDK or tool that platform needs. It is not a program refusal.
type Missing struct {
	Platform    string
	Requirement string
}

func (e *Missing) Error() string {
	return fmt.Sprintf("%s platform missing: %s", e.Platform, e.Requirement)
}
