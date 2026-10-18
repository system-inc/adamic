package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

func tokenizerSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("tokenizer.ts")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func runTokenizer(t *testing.T, source string) string {
	t.Helper()
	command, cancel := boundedrun.Command(20*time.Second, "node", "--input-type=module-typescript")
	defer cancel()
	command.Env = environment()
	command.Stdin = strings.NewReader(source)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("tokenizer must run: %v\n%s", err, output)
	}
	return string(output)
}

type tokenizerCounts struct {
	length, lines, identifiers, numbers, strings, symbols int
}

func readTokenizerCounts(t *testing.T, output string) tokenizerCounts {
	t.Helper()
	match := regexp.MustCompile(`^(\d+) code units, (\d+) lines: (\d+) identifiers, (\d+) numbers summing to \d+, (\d+) strings, (\d+) symbols, longest \S*\n$`).FindStringSubmatch(output)
	if match == nil {
		t.Fatalf("missing tokenizer counts: %q", output)
	}
	counts := tokenizerCounts{}
	for index, field := range []*int{&counts.length, &counts.lines, &counts.identifiers, &counts.numbers, &counts.strings, &counts.symbols} {
		value, err := strconv.Atoi(match[index+1])
		if err != nil {
			t.Fatal(err)
		}
		*field = value
	}
	return counts
}

// Check scanned tokens, not generator choices: words inside quoted strings must
// not inflate the identifier share. The intended share is about 37%, with room
// for punctuation tokens containing more than one code unit.
func tokenizerIdentifierShare(counts tokenizerCounts) error {
	total := counts.identifiers + counts.numbers + counts.strings + counts.symbols
	if total == 0 || counts.identifiers*4 < total {
		return fmt.Errorf("identifier share below 25%%: %d/%d", counts.identifiers, total)
	}
	return nil
}

func TestTokenizerInputMix(t *testing.T) {
	t.Parallel()
	source := tokenizerSource(t)
	output := runTokenizer(t, source)
	if again := runTokenizer(t, source); again != output {
		t.Fatalf("fixed seed is not deterministic:\n%s\n%s", output, again)
	}
	counts := readTokenizerCounts(t, output)
	if err := tokenizerIdentifierShare(counts); err != nil {
		t.Fatal(err)
	}
	total := counts.identifiers + counts.numbers + counts.strings + counts.symbols
	if total < 600000 || total > 650000 || counts.length < 2000000 || counts.lines < 10000 {
		t.Fatalf("input lost its size or whitespace mix: %+v", counts)
	}
	for _, kind := range []struct {
		name     string
		count    int
		min, max int
	}{
		{"identifiers", counts.identifiers, 25, 50},
		{"numbers", counts.numbers, 15, 35},
		{"strings", counts.strings, 5, 20},
		{"symbols", counts.symbols, 15, 40},
	} {
		if kind.count*100 < total*kind.min || kind.count*100 > total*kind.max {
			t.Errorf("%s outside %d%% to %d%%: %d/%d", kind.name, kind.min, kind.max, kind.count, total)
		}
	}
	t.Log(strings.TrimSpace(output))
}

func TestTokenizerOverflowMutant(t *testing.T) {
	t.Parallel()
	source := tokenizerSource(t)
	const safe = "seed = (seed * 16807) % 2147483647;"
	if strings.Count(source, safe) != 1 {
		t.Fatal("missing unique generator mutant seam")
	}
	mutant := strings.Replace(source, safe, "seed = (seed * 1103515245 + 12345) % 2147483648;", 1)
	counts := readTokenizerCounts(t, runTokenizer(t, mutant))
	err := tokenizerIdentifierShare(counts)
	if err == nil {
		t.Fatalf("overflow mutant survived identifier floor: %+v", counts)
	}
	t.Logf("overflow mutant ran successfully and was caught: %v", err)
}
