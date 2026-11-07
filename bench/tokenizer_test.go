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

func tokenizerSource(test *testing.T) string {
	test.Helper()
	source, readError := os.ReadFile("tokenizer.ts")
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source)
}

func runTokenizer(test *testing.T, source string) string {
	test.Helper()
	command, cancel := boundedrun.Command(20*time.Second, "node", "--input-type=module-typescript")
	defer cancel()
	command.Env = environment()
	command.Stdin = strings.NewReader(source)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("tokenizer must run: %v\n%s", combinedOutputError, output)
	}
	return string(output)
}

type tokenizerCounts struct {
	length, lines, identifiers, numbers, strings, symbols int
}

func readTokenizerCounts(test *testing.T, output string) tokenizerCounts {
	test.Helper()
	match := regexp.MustCompile(
		`^(\d+) code units, (\d+) lines: (\d+) identifiers, (\d+) numbers summing to \d+, (\d+) strings, (\d+) symbols, longest \S*\n$`,
	).FindStringSubmatch(output)
	if match == nil {
		test.Fatalf("missing tokenizer counts: %q", output)
	}
	counts := tokenizerCounts{}
	for index, field := range []*int{
		&counts.length,
		&counts.lines,
		&counts.identifiers,
		&counts.numbers,
		&counts.strings,
		&counts.symbols,
	} {
		value, parseError := strconv.Atoi(match[index+1])
		if parseError != nil {
			test.Fatal(parseError)
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

func TestTokenizerInputMix(test *testing.T) {
	test.Parallel()
	source := tokenizerSource(test)
	output := runTokenizer(test, source)
	if again := runTokenizer(test, source); again != output {
		test.Fatalf("fixed seed is not deterministic:\n%s\n%s", output, again)
	}
	counts := readTokenizerCounts(test, output)
	if tokenizerIdentifierShareError := tokenizerIdentifierShare(counts); tokenizerIdentifierShareError != nil {
		test.Fatal(tokenizerIdentifierShareError)
	}
	total := counts.identifiers + counts.numbers + counts.strings + counts.symbols
	if total < 600000 || total > 650000 || counts.length < 2000000 || counts.lines < 10000 {
		test.Fatalf("input lost its size or whitespace mix: %+v", counts)
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
			test.Errorf("%s outside %d%% to %d%%: %d/%d", kind.name, kind.min, kind.max, kind.count, total)
		}
	}
	test.Log(strings.TrimSpace(output))
}

func TestTokenizerOverflowMutant(test *testing.T) {
	test.Parallel()
	source := tokenizerSource(test)
	const safe = "seed = (seed * 16807) % 2147483647;"
	if strings.Count(source, safe) != 1 {
		test.Fatal("missing unique generator mutant seam")
	}
	mutant := strings.Replace(source, safe, "seed = (seed * 1103515245 + 12345) % 2147483648;", 1)
	counts := readTokenizerCounts(test, runTokenizer(test, mutant))
	tokenizerIdentifierShareError := tokenizerIdentifierShare(counts)
	if tokenizerIdentifierShareError == nil {
		test.Fatalf("overflow mutant survived identifier floor: %+v", counts)
	}
	test.Logf("overflow mutant ran successfully and was caught: %v", tokenizerIdentifierShareError)
}
