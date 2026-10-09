//go:build !linux

package childguard

import (
	"fmt"
	"runtime"
	"time"
)

// Portable Go has no process-group/reaped-child CPU accounting API. A Darwin
// libproc or ps -S adapter needs platform-specific accounting and validation
// unavailable on this Linux-only gate box. Until that adapter is validated,
// retain Ceiling and never substitute wall silence for missing CPU accounting.
func processGroupCPU(group int) (time.Duration, error) {
	return 0, fmt.Errorf("process-group CPU accounting unavailable on %s", runtime.GOOS)
}
