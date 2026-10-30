package native

import (
	"fmt"
	"strings"
)

// The emitter leaves this guard in C as well: standalone C compilation and the
// Build API must both fail closed, before a target archive or linker is needed.
func terminalSizeTarget(source string, options Options) error {
	if options.Target == "wasm32-wasi" && strings.Contains(source, `#error "wasi preview1 reports no terminal size"`) {
		return fmt.Errorf("native: wasi preview1 reports no terminal size")
	}
	return nil
}
