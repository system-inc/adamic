package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeProcessTTYRefusals(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ expression, reason string }{
		{
			"process.stderr.columns ?? -1",
			"node:tty.WriteStream.columns on another stream (fix: read process.stdout.columns)",
		},
		// stdout's isTTY is main's binding (library_node_stdout_istty_test.go); stderr's stays refused there.
		{
			"process.stderr.isTTY ?? false",
			"node:process.Process.stderr.isTTY is not bound (fix: read process.stdout.isTTY)",
		},
	} {
		for _, expression := range []string{one.expression, strings.Split(one.expression, " ?? ")[0]} {
			for _, extension := range []string{".a", ".ts"} {
				test.Run(expression+extension, func(test *testing.T) {
					_, lowerSourceAsError := lowerSourceAs(
						test,
						"import 'node:process'; console.log("+expression+");",
						extension,
					)
					var refused *NotYet
					if !errors.As(lowerSourceAsError, &refused) || refused.What != one.reason {
						test.Fatalf("want exact guarded TTY refusal %q, got %v", one.reason, lowerSourceAsError)
					}
				})
			}
		}
	}

}
