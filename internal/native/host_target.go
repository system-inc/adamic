package native

import (
	"fmt"
	"strings"
	"text/scanner"
)

// TargetRefused distinguishes an unsupported host member from a broken WASI build.
// Lowering is target independent; the build boundary knows the selected target.
type TargetRefused struct{ Member, Reason string }

func (r *TargetRefused) Error() string {
	return fmt.Sprintf("wasm32-wasi refuses %s: %s; use a native Node host target", r.Member, r.Reason)
}

var wasiHostRefusals = map[string]TargetRefused{
	"ADAMIC_WASI_FS_OPEN_MODE":       {"fs.openSync", "Preview 1 path_open has no mode parameter; creation mode must be proven 0666 or the flags must be proven not to create"},
	"ADAMIC_WASI_FS_WRITE_FILE_MODE": {"fs.writeFileSync", "Preview 1 path_open has no mode parameter; creation mode must be proven 0666 or the flags must be proven not to create"},
	"ADAMIC_WASI_FS_MKDIR_MODE":      {"fs.mkdirSync", "Preview 1 path_create_directory has no mode parameter; directory mode must be proven 0777"},
	"ADAMIC_WASI_FS_UTIMES":          {"fs.utimesSync", "Preview 1 timestamps are unsigned and Node WASI drops subsecond precision; both times must be proven nonnegative whole seconds within the timestamp range"},
	"adamic_node_argv":               {"process.argv", "WASI has no executable paths"},
	"adamic_fs_file_mkdtemp":         {"fs.mkdtempSync", "WASI has no temporary directory creation"},
	"adamic_node_pid":                {"process.pid", "WASI has no process identifiers"},
	"adamic_node_platform":           {"process.platform", "WASI has no Node host platform"},
	"adamic_node_columns":            {"process.stdout.columns", "WASI has no terminal size"},
	"adamic_node_memory_usage":       {"process.memoryUsage", "WASI has no allocator observations"},
}

func validateBuild(source string, options Options) error {
	if err := ValidateOptions(options); err != nil {
		return err
	}
	if options.Target != "wasm32-wasi" {
		return nil
	}
	// Examine identifiers, never string data or comments. Generated C names the
	// selected host entry even when the operation is in a closure or dead branch.
	var tokens scanner.Scanner
	tokens.Init(strings.NewReader(source))
	tokens.Error = func(*scanner.Scanner, string) {} // C permits escapes such as \? that Go does not.
	tokens.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanChars | scanner.ScanComments | scanner.SkipComments
	for token := tokens.Scan(); token != scanner.EOF; token = tokens.Scan() {
		if token == scanner.Ident {
			if refusal, exists := wasiHostRefusals[tokens.TokenText()]; exists {
				return &refusal
			}
		}
	}
	return nil
}
