package native

import (
	"fmt"
	"os"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The checked-in C header is generated from the same constant the JavaScript runtime uses.
func TestJSONErrorStopHeaderIsCurrent(t *testing.T) {
	t.Parallel()
	want := fmt.Sprintf("// json_error_stop.h: the Error write stop from ir.JSONErrorWriteStop.\n// TestJSONErrorStopHeaderIsCurrent holds this generated spelling to its source.\n#ifndef ADAMIC_JSON_ERROR_STOP_H\n#define ADAMIC_JSON_ERROR_STOP_H\n\n#define ADAMIC_JSON_ERROR_WRITE_STOP %s\n\n#endif\n", cString(ir.JSONErrorWriteStop))
	got, err := os.ReadFile("../native_runtime/c/json_error_stop.h")
	if err != nil || string(got) != want {
		t.Fatalf("Error stop header differs from ir.JSONErrorWriteStop: %v", err)
	}
}
