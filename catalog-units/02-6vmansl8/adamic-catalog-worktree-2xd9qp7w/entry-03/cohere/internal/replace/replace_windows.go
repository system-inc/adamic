package replace

import (
	"errors"
	"syscall"
)

// errorSharingViolation is ERROR_SHARING_VIOLATION, which syscall does not name.
const errorSharingViolation = syscall.Errno(32)

// holdsTheDestination reports whether Windows refused a rename because another process has the destination
// open. It answers access denied for that too, and also for a read-only destination, which a retry cannot
// change: that one fails after the bound, as loudly as before.
func holdsTheDestination(err error) bool {
	return errors.Is(err, syscall.ERROR_ACCESS_DENIED) || errors.Is(err, errorSharingViolation)
}
