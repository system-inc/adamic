package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nodeFSFile(call ir.NodeFSFile) string {
	args := make([]string, len(call.Arguments))
	for i, arg := range call.Arguments {
		args[i] = e.value(arg)
	}
	switch call.Operation {
	case "host_join":
		return "adamicNodePath.join(" + strings.Join(args, ", ") + ")"
	case "host_tmpdir":
		return "adamicNodeOS.tmpdir()"
	case "host_cwd":
		return "adamicNodeProcess.cwd()"
	case "host_chdir":
		return "adamicNodeProcess.chdir(" + args[0] + ")"
	case "date_new":
		return "new Date(" + args[0] + ")"
	case "date_time":
		return "(" + args[0] + ").getTime()"
	case "is_file":
		return "(" + args[0] + ").isFile()"
	case "is_directory":
		return "(" + args[0] + ").isDirectory()"
	case "is_symbolic_link":
		return "(" + args[0] + ").isSymbolicLink()"
	case "read_buffer", "read_buffer_fd":
		return "adamicNodeFSFile.readFileSync(" + args[0] + ", {flag:" + args[1] + "})"
	case "read_file", "read_fd":
		return "adamicNodeFSFile.readFileSync(" + args[0] + ", {encoding:'utf8', flag:" + args[1] + "})"
	case "open":
		return "adamicNodeFSFile.openSync(" + args[0] + ", " + args[1] + ", " + args[2] + ")"
	case "read_sync":
		return "adamicNodeFSFile.readSync(" + args[0] + ", " + args[1] + ", " + args[2] + ", " + args[3] + ", " + args[4] + ")"
	case "write":
		return "adamicNodeFSFile.writeSync(" + args[0] + ", " + args[1] + ", " + args[2] + ", 'utf8')"
	case "close":
		return "adamicNodeFSFile.closeSync(" + args[0] + ")"
	case "write_file", "write_fd", "write_buffer", "write_buffer_fd":
		return "adamicNodeFSFile.writeFileSync(" + args[0] + ", " + args[1] + ", {flag:" + args[2] + ", mode:" + args[3] + ", flush:" + args[4] + ", encoding:'utf8'})"
	case "exists":
		return "adamicNodeFSFile.existsSync(" + args[0] + ")"
	case "stat":
		return "adamicNodeFSFile.statSync(" + args[0] + ", {throwIfNoEntry:" + args[1] + "})"
	case "mkdir":
		return "adamicNodeFSFile.mkdirSync(" + args[0] + ", {recursive:" + args[1] + ", mode:" + args[2] + "})"
	case "mkdtemp":
		return "adamicNodeFSFile.mkdtempSync(" + args[0] + ")"
	case "rm":
		return "adamicNodeFSFile.rmSync(" + args[0] + ", {recursive:" + args[1] + ", force:" + args[2] + "})"
	case "unlink":
		return "adamicNodeFSFile.unlinkSync(" + args[0] + ")"
	case "utimes", "utimes_dates", "utimes_atime_date", "utimes_mtime_date":
		return "adamicNodeFSFile.utimesSync(" + args[0] + ", " + args[1] + ", " + args[2] + ")"
	}
	panic("javascript: unknown fs file operation " + call.Operation)
}
