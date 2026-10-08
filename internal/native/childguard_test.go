package native

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// Silent compilers and buffered drivers get the shared startup allowance;
// subsequent stalls count from the last output, rather than process creation.
var nativeGuardOptions = childguard.Options{}

// Run helpers in a separate test process so their Fatal/Error paths are checked
// as failures, including helpers which normally accept nonzero child exits.
func TestNativeChildGuard(t *testing.T) {
	if site := os.Getenv("ADAMIC_NATIVE_GUARD_FIXTURE"); site != "" {
		nativeGuardOptions = childguard.Options{FirstOutput: 5 * time.Second, Stall: time.Second, Ceiling: 20 * time.Second}
		script := "printf 'ready\\n'; exec sleep 60"
		if os.Getenv("ADAMIC_NATIVE_GUARD_PROGRESS") == "1" {
			script = "for i in 1 2 3 4 5 6; do printf 'progress\\n'; printf 'progress\\n' >&2; sleep 0.3; done"
		}
		switch site {
		case "streamLines":
			lines, wait := streamLines(t, []string{"sh", "-c", script})
			for lines.Scan() {
			}
			if err := lines.Err(); err != nil {
				t.Fatal(err)
			}
			wait()
		case "runWithInput":
			runWithInput(t, "", "sh", "-c", script)
		case "outcome":
			outcome(t, "sh", "-c", script)
		case "wasiCommand":
			wasiCommand(t, t.TempDir(), "sh", "-c", script)
		case "observeWASI":
			observeWASI(t, t.TempDir(), "sh", "-c", script)
		default:
			t.Fatalf("unknown guard fixture %q", site)
		}
		return
	}
	for _, site := range []string{"streamLines", "runWithInput", "outcome", "wasiCommand", "observeWASI"} {
		t.Run(site, func(t *testing.T) {
			for _, progress := range []bool{false, true} {
				command := exec.Command(os.Args[0], "-test.run=^TestNativeChildGuard$", "-test.timeout=30s")
				command.Env = append(os.Environ(), "ADAMIC_NATIVE_GUARD_FIXTURE="+site)
				if progress {
					command.Env = append(command.Env, "ADAMIC_NATIVE_GUARD_PROGRESS=1")
				}
				output, err := command.CombinedOutput()
				if progress {
					if err != nil {
						t.Fatalf("progress: %v\n%s", err, output)
					}
				} else if err == nil || !strings.Contains(string(output), "stalled: no output") {
					t.Fatalf("hang did not fail through guard: %v\n%s", err, output)
				}
			}
		})
	}
}
