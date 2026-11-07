package lower

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// process.platform is a host read (#t90fcv5): a plain read lowers to host_platform, recognized by its
// library member under any receiver spelling, and is never folded into a string constant.
func TestProcessPlatformReadIsAHostRead(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ name, source string }{
		{"default import", `import process from 'node:process'; console.log(process.platform);`},
		{"alias", `import host from 'node:process'; console.log(host.platform === 'linux' ? 'yes' : 'no');`},
	} {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			program, lowerSourceError := lowerSource(test, one.source)
			if lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
			if reads := strings.Count(fmt.Sprintf("%#v", program), `Operation:"host_platform"`); reads != 1 {
				test.Fatalf("want one host_platform read, got %d", reads)
			}
			for _, constant := range program.Strings {
				if constant == "linux" && !strings.Contains(one.source, "'linux'") {
					test.Fatalf("a platform name was folded into the program's strings: %q", program.Strings)
				}
			}
		})
	}
}

// A write and a delete never reach lowering: the checker refuses both, since platform is read-only.
// An element read and a destructured binding stop earlier, at the namespace object they read from.
func TestProcessPlatformRejectsNonReadUses(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ name, source, want string }{
		{"optional", `console.log(process?.platform);`, "node:process.Process.platform read through an optional chain"},
		{"element", `console.log(process['platform']);`, "a namespace object used as a value"},
		{"destructured", `const { platform } = process; console.log(platform);`, "a namespace object used as a value"},
	} {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, `import process from 'node:process'; `+one.source)
			var notYet *NotYet
			if !errors.As(lowerSourceError, &notYet) || !strings.Contains(notYet.What, one.want) {
				test.Fatalf("want platform NotYet %s, got %v", one.want, lowerSourceError)
			}
		})
	}
}
