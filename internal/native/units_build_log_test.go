package native

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// buildLogLine is a line buildLog writes: kind, name, key12 or -, outcome, seconds.
var buildLogLine = regexp.MustCompile(`^(cCompile|cLink) (\S+) ([0-9a-f]{12}|-) (hit|miss|off) ([0-9]+\.[0-9]{3})$`)

// A split build writes one cCompile line per unit and one cLink line to ADAMIC_BUILD_LOG, each saying whether its
// object came from the cache and how long clang ran, and the log never reaches the binary: the same program built with
// the log and without it, cached and uncached, is the same bytes. A mutant that drops the outcome, or says hit where
// clang compiled, fails the outcomes read here; one that writes 0 seconds fails the seconds.
//
// Not parallel: t.Setenv changes ADAMIC_BUILD_LOG and ADAMIC_GATE_UNCACHED for the process.
func TestSplitBuildLogsEachUnitAndTheLink(t *testing.T) {
	// The program's names are this run's own, in units.h, which every unit includes, so every unit's preprocessed bytes
	// and so its key are new here and the first cached build compiles them all, whatever a shared cache already holds.
	nonce := strconv.FormatInt(time.Now().UnixNano(), 36)
	source := fmt.Sprintf(`#include "adamic.h"
#include <stdio.h>
static int adamic_state_%[1]s = 0;
static int adamic_next_%[1]s(void);
static int adamic_twice_%[1]s(void);
static int adamic_next_%[1]s(void) { return ++adamic_state_%[1]s; }
static int adamic_twice_%[1]s(void) { return adamic_next_%[1]s() + adamic_next_%[1]s(); }
int main(void) { printf("%%d\n", adamic_twice_%[1]s()); return 0; }
`, nonce)
	_, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	build := func(name, uncached string, logged bool) ([]byte, []string) {
		t.Helper()
		log := filepath.Join(directory, name+".log")
		if logged {
			t.Setenv("ADAMIC_BUILD_LOG", log)
		} else {
			t.Setenv("ADAMIC_BUILD_LOG", "")
		}
		t.Setenv("ADAMIC_GATE_UNCACHED", uncached)
		binary := filepath.Join(directory, name, "program")
		if err := os.MkdirAll(filepath.Dir(binary), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := Build(source, binary, Options{Split: true, Jobs: 2}); err != nil {
			t.Fatal(err)
		}
		contents, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		written, err := os.ReadFile(log)
		if !logged {
			if !os.IsNotExist(err) {
				t.Fatalf("%s: a build without ADAMIC_BUILD_LOG wrote %s: %v", name, log, err)
			}
			return contents, nil
		}
		if err != nil {
			t.Fatalf("%s: no build log: %v", name, err)
		}
		return contents, strings.Split(strings.TrimSuffix(string(written), "\n"), "\n")
	}
	// read checks one build's lines: a cCompile per unit with outcome, each unit once, then one cLink naming the binary.
	read := func(name string, lines []string, outcome string) {
		t.Helper()
		compiled := map[string]bool{}
		links := 0
		for _, line := range lines {
			match := buildLogLine.FindStringSubmatch(line)
			if match == nil {
				t.Fatalf("%s: %q isn't a build log line", name, line)
			}
			seconds, err := strconv.ParseFloat(match[5], 64)
			if err != nil {
				t.Fatal(err)
			}
			switch match[1] {
			case "cCompile":
				if match[4] != outcome || match[3] == "-" {
					t.Fatalf("%s: %q, want a keyed %s", name, line, outcome)
				}
				if compiled[match[2]+" "+match[3]] {
					t.Fatalf("%s: %s logged twice", name, match[2])
				}
				compiled[match[2]+" "+match[3]] = true
				// A hit's seconds are its preprocessing alone, which can round to nothing; a compile's can't.
				if outcome != "hit" && seconds <= 0 {
					t.Fatalf("%s: %q says clang took no time", name, line)
				}
			case "cLink":
				links++
				if match[2] != "program" || match[3] != "-" || match[4] != "off" || seconds <= 0 {
					t.Fatalf("%s: %q, want the program's link, unkeyed, off, with its seconds", name, line)
				}
			}
		}
		if len(compiled) != len(units) || links != 1 {
			t.Fatalf("%s: %d units and %d links logged, want %d and 1:\n%s", name, len(compiled), links, len(units), lines)
		}
	}
	quietUncached, _ := build("quiet-uncached", "1", false)
	loggedUncached, lines := build("logged-uncached", "1", true)
	read("logged-uncached", lines, "off")
	loggedMiss, lines := build("logged-miss", "0", true)
	read("logged-miss", lines, "miss")
	loggedHit, lines := build("logged-hit", "0", true)
	read("logged-hit", lines, "hit")
	quietHit, _ := build("quiet-hit", "0", false)
	for name, binary := range map[string][]byte{"logged-uncached": loggedUncached, "logged-miss": loggedMiss, "logged-hit": loggedHit, "quiet-hit": quietHit} {
		if !bytes.Equal(binary, quietUncached) {
			t.Fatalf("%s's binary differs from the one built uncached without a log", name)
		}
	}
	t.Logf("%d units: off, miss and hit each logged once a unit with one link, and five binaries the same bytes", len(units))
}

// A name holding whitespace is one field of one line: an output path with a tab or a newline in it can't split the line
// or forge a second. Mutant: replacing spaces alone, as record does.
//
// Not parallel: t.Setenv changes ADAMIC_BUILD_LOG for the process.
func TestBuildLogKeepsANameOnOneLine(t *testing.T) {
	log := filepath.Join(t.TempDir(), "build.log")
	t.Setenv("ADAMIC_BUILD_LOG", log)
	buildLog("cLink", "a b\tc\ncCompile forged 0123456789ab hit 9.000", "", "off", time.Second)
	written, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if want := "cLink a_b_c_cCompile_forged_0123456789ab_hit_9.000 - off 1.000\n"; string(written) != want {
		t.Fatalf("wrote %q, want %q", written, want)
	}
}
