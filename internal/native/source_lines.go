package native

import (
	"fmt"
	"strings"
)

// The tag keeps a reset recognizable after a unit has been cut and assembled.
const sourceLineReset = `#line 1 "main.c" /* adamic source reset */`

func resolveSourceLines(source, file string) string {
	if !strings.Contains(source, "/* adamic source reset */") {
		return source
	}
	// Resets must not resolve generated C beside the entry source comp_dir.
	if !strings.HasPrefix(file, "/") {
		file = "/adamic-build/" + file
	}
	lines := strings.Split(source, "\n")
	for index, line := range lines {
		if strings.HasSuffix(line, "/* adamic source reset */") {
			lines[index] = fmt.Sprintf("#line %d %s /* adamic source reset */", index+2, cLiteral(file))
		}
	}
	return strings.Join(lines, "\n")
}
