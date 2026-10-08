package native

import (
	"math"
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// Keep target-independent C runnable natively. These identifiers describe only
// actual intrinsic uses whose arguments cannot be held exactly by Preview 1.
// validateBuild reads them before creating build files or invoking clang.
func (e *emitter) nodeFSFileWASIMarker(call ir.NodeFSFile) {
	marker := ""
	switch call.Operation {
	case "open":
		if !e.wasiCreationMode(call.Arguments[1], call.Arguments[2]) {
			marker = "ADAMIC_WASI_FS_OPEN_MODE"
		}
	case "write_file", "write_buffer":
		if !e.wasiCreationMode(call.Arguments[2], call.Arguments[3]) {
			marker = "ADAMIC_WASI_FS_WRITE_FILE_MODE"
		}
	case "mkdir":
		if mode, ok := call.Arguments[2].(ir.NumberConstant); !ok || mode.Value != 0777 {
			marker = "ADAMIC_WASI_FS_MKDIR_MODE"
		}
	case "utimes":
		if !wasiWholeSeconds(call.Arguments[1]) || !wasiWholeSeconds(call.Arguments[2]) {
			marker = "ADAMIC_WASI_FS_UTIMES"
		}
	case "utimes_dates", "utimes_atime_date", "utimes_mtime_date":
		// Date objects can change or contain fractional seconds, negative times or
		// NaN. A numeric literal proof does not apply to an object's stored time.
		marker = "ADAMIC_WASI_FS_UTIMES"
	}
	if marker != "" {
		declaration := "#define " + marker + " 1"
		if !slices.Contains(e.declarations, declaration) {
			e.declarations = append(e.declarations, declaration)
		}
	}
}

func (e *emitter) wasiCreationMode(flag, mode ir.Expression) bool {
	if value, ok := mode.(ir.NumberConstant); ok && value.Value == 0666 {
		return true
	}
	// Node validates mode even when it is ignored; native and WASI retain that
	// validation. Opening an existing file without O_CREAT needs no mode syscall.
	if value, ok := flag.(ir.StringConstant); ok {
		switch e.program.Strings[value.Index] {
		case "r", "r+", "rs", "sr", "rs+", "sr+":
			return true
		}
	}
	return false
}

func wasiWholeSeconds(expression ir.Expression) bool {
	value, ok := expression.(ir.NumberConstant)
	return ok && value.Value >= 0 && value.Value <= 18446744073 && value.Value == math.Floor(value.Value)
}
