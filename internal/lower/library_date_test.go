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
		{"const time = Date.now();", "wall clock is nondeterministic"},
		{"const clock = Date.now;", "wall clock is nondeterministic"},
		{"const date = new Date();", "wall clock is nondeterministic"},
		{"const text = Date();", "wall clock is nondeterministic"},
		{"const date = new Date(0); console.log(date.toJSON());", "invalid dates return null"},
		{"console.log(`${Date.parse('Dec 17 1995')}`);", "legacy parsing is not built"},
		{"function parse(text:string):number {return Date.parse(text);} console.log(`${parse('1970')}`);", "constant ISO"},
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
		_, err := lowerSource(t, "const date = new Date(1970,0); console.log(`${date.getHours()}`);")
		if err == nil || !strings.Contains(err.Error(), "without TZ=UTC") {
			t.Fatalf("want an explicit UTC requirement, got %v", err)
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
