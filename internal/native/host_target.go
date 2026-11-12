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
	"adamic_fs_file_watch":         {"fs.watchFile", "wasi preview1 has no stat-polling watcher provider"},
	"adamic_fs_file_unwatch":       {"fs.unwatchFile", "wasi preview1 has no stat-polling watcher provider"},
	"adamic_fs_file_host_watch":    {"fs.watch", "wasi preview1 has no file system event provider"},
	"adamic_fs_file_host_watch_on": {"FSWatcher.on", "wasi preview1 has no file system event provider"},
	"adamic_host_stdout_columns":   {"process.stdout.columns", "wasi preview1 reports no terminal size"},
	"adamic_node_argv":             {"process.argv", "WASI has no executable paths"},
	"adamic_node_filename":         {"__filename", "WASI has no executable paths"},
	"adamic_node_dirname":          {"__dirname", "WASI has no executable paths"},
	"adamic_fs_file_mkdtemp":       {"fs.mkdtempSync", "WASI has no temporary directory creation"},
	"adamic_node_pid":              {"process.pid", "WASI has no process identifiers"},
	"adamic_node_platform":         {"process.platform", "WASI has no Node host platform"},
	"adamic_node_memory_usage":     {"process.memoryUsage", "WASI has no allocator observations"},
	"adamic_node_set_timeout":      {"setTimeout", "wasi preview1 has no monotonic timer provider"},
	// A Stats' fields preview 1's filestat holds no answer for (ir.DynamicKeyedStatsFieldNames), each
	// read through its own function (ir.StatsFieldOperations): wasi-libc's stat zeroes them, or keeps
	// only the file type in mode, so the read is refused by name rather than answer that (#6xfkdwj).
	"adamic_fs_file_stats_mode":         {"fs.Stats.mode", "wasi preview1's filestat has no permission bits"},
	"adamic_fs_file_stats_uid":          {"fs.Stats.uid", "wasi preview1's filestat has no owner"},
	"adamic_fs_file_stats_gid":          {"fs.Stats.gid", "wasi preview1's filestat has no group"},
	"adamic_fs_file_stats_rdev":         {"fs.Stats.rdev", "wasi preview1's filestat has no device number"},
	"adamic_fs_file_stats_blksize":      {"fs.Stats.blksize", "wasi preview1's filestat has no block size"},
	"adamic_fs_file_stats_blocks":       {"fs.Stats.blocks", "wasi preview1's filestat has no block count"},
	"adamic_fs_file_stats_birthtime_ms": {"fs.Stats.birthtimeMs", "wasi preview1's filestat has no birth time"},
	"adamic_fs_file_stats_birthtime":    {"fs.Stats.birthtime", "wasi preview1's filestat has no birth time"},
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
