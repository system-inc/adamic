package load

import (
	_ "embed"
	"strings"
)

// Node 24 builtinModules plus the builtins which require the node: prefix.
// A host not implemented yet is still a builtin, never a user package.
func NodeBuiltin(specifier string) (string, bool) {
	name := strings.TrimPrefix(specifier, "node:")
	switch name {
	case "_http_agent", "_http_client", "_http_common", "_http_incoming", "_http_outgoing", "_http_server", "_stream_duplex", "_stream_passthrough", "_stream_readable", "_stream_transform", "_stream_wrap", "_stream_writable", "_tls_common", "_tls_wrap", "assert", "assert/strict", "async_hooks", "buffer", "child_process", "cluster", "console", "constants", "crypto", "dgram", "diagnostics_channel", "dns", "dns/promises", "domain", "events", "fs", "fs/promises", "http", "http2", "https", "inspector", "inspector/promises", "module", "net", "os", "path", "path/posix", "path/win32", "perf_hooks", "process", "punycode", "querystring", "readline", "readline/promises", "repl", "stream", "stream/consumers", "stream/promises", "stream/web", "string_decoder", "sys", "timers", "timers/promises", "tls", "trace_events", "tty", "url", "util", "util/types", "v8", "vm", "wasi", "worker_threads", "zlib":
		return "node:" + name, true
	case "test", "test/reporters", "sea", "sqlite":
		return "node:" + name, strings.HasPrefix(specifier, "node:")
	}
	return "", false
}

const nodeRequirePath = preludeDirectory + "/node_require.d.ts"

// @types/node's NodeRequire returns any. Refine the literal overload to the
// host's declared module type so its members carry the same declaration symbols
// as an import. The fallback is unknown: refused calls never introduce any.
//
//go:embed node_require.d.ts
var nodeRequirePrelude string
