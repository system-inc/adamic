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
	"adamic_host_argv":                {"process.argv", "WASI has no executable paths"},
	"adamic_host_exec_path":           {"process.execPath", "WASI has no executable paths"},
	"adamic_host_executing_file_path": {"executable identity", "WASI has no executable paths"},
	"adamic_host_memory_usage":        {"process.memoryUsage", "WASI has no allocator observations"},
	"adamic_node_argv":                {"process.argv", "WASI has no executable paths"},
	"adamic_fs_file_mkdtemp":          {"fs.mkdtempSync", "WASI has no temporary directory creation"},
	"adamic_node_pid":                 {"process.pid", "WASI has no process identifiers"},
	"adamic_node_platform":            {"process.platform", "WASI has no Node host platform"},
	"adamic_node_columns":             {"process.stdout.columns", "WASI has no terminal size"},
	"adamic_node_memory_usage":        {"process.memoryUsage", "WASI has no allocator observations"},
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
