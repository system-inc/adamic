package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeProcessTTYRefusals(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ expression, reason string }{
		{"process.stdout.columns ?? -1", "Node caches the terminal width and refreshes it on SIGWINCH from the event loop; adamic/host-tty"},
		// stdout's isTTY is main's binding (library_node_stdout_istty_test.go); stderr's stays refused there.
		{"process.stderr.isTTY ?? false", "node:process.Process.stderr.isTTY is not bound (fix: read process.stdout.isTTY)"},
	} {
		for _, expression := range []string{one.expression, strings.Split(one.expression, " ?? ")[0]} {
			for _, extension := range []string{".a", ".ts"} {
				t.Run(expression+extension, func(t *testing.T) {
					_, err := lowerSourceAs(t, "import 'node:process'; console.log("+expression+");", extension)
					var refused *NotYet
					if !errors.As(err, &refused) || refused.What != one.reason {
						t.Fatalf("want exact guarded TTY refusal %q, got %v", one.reason, err)
					}
				})
			}
		}
	}

}
