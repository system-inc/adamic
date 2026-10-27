package lower

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// process.platform is a host read (#t90fcv5): a plain read lowers to host_platform, recognized by its
// library member under any receiver spelling, and is never folded into a string constant.
func TestProcessPlatformReadIsAHostRead(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, source string }{
		{"default import", `import process from 'node:process'; console.log(process.platform);`},
		{"global", `console.log(process.platform);`},
		{"alias", `import host from 'node:process'; console.log(host.platform === 'linux' ? 'yes' : 'no');`},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, one.source)
			if err != nil {
				t.Fatal(err)
			}
			if reads := strings.Count(fmt.Sprintf("%#v", program), `Operation:"host_platform"`); reads != 1 {
				t.Fatalf("want one host_platform read, got %d", reads)
			}
			for _, constant := range program.Strings {
				if constant == "linux" && !strings.Contains(one.source, "'linux'") {
					t.Fatalf("a platform name was folded into the program's strings: %q", program.Strings)
				}
			}
		})
	}
}

func TestProcessPlatformRejectsNonReadUses(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, source, want string }{
		{"write", `process.platform = 'linux';`, "written"},
		{"delete", `delete process.platform;`, "deleted"},
		{"optional", `console.log(process?.platform);`, "read through an optional chain"},
		{"element", `console.log(process['platform']);`, "used as a value"},
		{"destructured", `const { platform } = process; console.log(platform);`, "used as a value"},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, `import process from 'node:process'; `+one.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "node:process.Process.platform "+one.want) {
				t.Fatalf("want platform NotYet %s, got %v", one.want, err)
			}
		})
	}
}
