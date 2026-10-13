package native

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/system-inc/adamic/internal/toolchain"
)

// programCommand is exec.CommandContext for a program a test runs, in directory unless it is "". A bare name is a tool
// (node, clang): the program toolchain resolves it to, under toolchain's environment, with PWD set to directory as exec
// would set it for a command with no environment of its own. A path is a binary a test built, run as itself. A tool
// toolchain can't resolve fails as the command starts, with Path's error.
func programCommand(ctx context.Context, directory, name string, arguments ...string) *exec.Cmd {
	if strings.ContainsRune(name, '/') {
		command := exec.CommandContext(ctx, name, arguments...)
		command.Dir = directory
		return command
	}
	path, err := toolchain.Path(name)
	if err != nil {
		command := exec.CommandContext(ctx, name, arguments...)
		command.Err = err
		return command
	}
	command := exec.CommandContext(ctx, path, arguments...)
	command.Env = toolchain.Env()
	if directory != "" {
		absolute, err := filepath.Abs(directory)
		if err != nil {
			command.Err = err
			return command
		}
		command.Dir = directory
		command.Env = append(command.Env, "PWD="+absolute)
	}
	return command
}
