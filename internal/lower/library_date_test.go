package lower

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDateRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{"const date = new Date(0); const view: {valueOf: () => number} = date; console.log(`${view.valueOf()}`);", "Date internal slot"},
		{"const get = Date.prototype.getTime; console.log(`${get()}`);", "method read as a value"},
		{"console.log(`${Date.prototype.getTime.call({})}`);", "Date internal slot"},
		{"console.log(`${Date.prototype.toJSON.call({toISOString: (): string => 'ISO'})}`);", "generic receiver"},
		{"const time = Date.now();", "wall clock is nondeterministic"},
		{"const clock = Date.now;", "wall clock is nondeterministic"},
		{"const date = new Date();", "wall clock is nondeterministic"},
		{"const text = Date();", "wall clock is nondeterministic"},
	} {
		_, err := lowerSource(t, probe.source)
		if err == nil || !strings.Contains(err.Error(), probe.reason) {
			t.Fatalf("%s: want %q, got %v", probe.source, probe.reason, err)
		}
	}
}

func TestDateLocalTZRequirement(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_DATE_TZ_PROBE") == "1" {
		for _, source := range []string{
			"const date = new Date(1970,0); console.log(`${date.getHours()}`);",
			"function parse(text: string): number {return Date.parse(text);} console.log(`${parse('1970')}`);",
			"function parse(text: string): number {return new Date(text).getTime();} console.log(`${parse('1970')}`);",
		} {
			_, err := lowerSource(t, source)
			if err == nil || !strings.Contains(err.Error(), "without TZ=UTC") {
				t.Fatalf("want an explicit UTC requirement, got %v", err)
			}
		}
		return
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, "-test.run=^TestDateLocalTZRequirement$")
	command.Env = append(os.Environ(), "TZ=America/New_York", "ADAMIC_DATE_TZ_PROBE=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
}
