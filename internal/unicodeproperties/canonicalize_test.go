package unicodeproperties

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCanonicalizeExamples(t *testing.T) {
	t.Parallel()
	for _, pair := range []struct{ from, to rune }{
		{'A', 'a'}, {'a', 'a'}, {'Z', 'z'},
		{0x212A, 'k'},    // kelvin sign, simple fold to k
		{0x017F, 's'},    // long s
		{0x0131, 0x0131}, // dotless i: only a Turkic fold, and it runs the other way
		{0x0130, 0x0130}, // I with dot: Turkic and full folds only, so it stays
		{0x00DF, 0x00DF}, // sharp s has no simple fold
		{0x1E9E, 0x00DF}, // capital sharp s folds to sharp s
		{0x212B, 0x00E5}, // angstrom sign folds to a with ring
		{0x2126, 0x03C9}, // ohm sign folds to omega
		{0x00C5, 0x00E5},
		{0x1F80, 0x1F80},
		{0x1F88, 0x1F80}, // prosgegrammeni, simple fold; full fold would grow
		{0xFB05, 0xFB06},
		{0x03A3, 0x03C3}, {0x03C2, 0x03C3}, {0x03C3, 0x03C3},
		{0x10400, 0x10428}, // Deseret, a supplementary pair
	} {
		if got := CanonicalizeUnicode(pair.from); got != pair.to {
			t.Errorf("CanonicalizeUnicode(U+%04X) = U+%04X, want U+%04X", pair.from, got, pair.to)
		}
	}
	for _, pair := range []struct{ from, to uint16 }{
		{'a', 'A'}, {'A', 'A'},
		{0x212A, 0x212A}, // kelvin sign is its own uppercase in Unicode 17
		{0x017F, 0x017F}, // long s uppercases to S, which the ASCII filter rejects
		{0x0131, 0x0131}, // dotless i uppercases to I, same filter
		{0x0130, 0x0130},
		{0x00DF, 0x00DF}, // sharp s uppercases to SS, two units
		{0x1E9E, 0x1E9E},
		{0x212B, 0x212B}, // angstrom sign does not uppercase to A with ring
		{0x00E5, 0x00C5},
		{0x00C5, 0x00C5},
		{0x2126, 0x2126}, // ohm sign does not uppercase to omega
		{0x03C9, 0x03A9},
		{0x03C2, 0x03A3},
		{0x03C3, 0x03A3},
		{0x1F80, 0x1F80}, // uppercase is two code points, so it stays
		{0x1F88, 0x1F88},
		{0xFB05, 0xFB05},
		{0xFB06, 0xFB06},
		{0x00B5, 0x039C}, // micro sign uppercases to Greek mu
	} {
		if got := CanonicalizeLegacy(pair.from); got != pair.to {
			t.Errorf("CanonicalizeLegacy(U+%04X) = U+%04X, want U+%04X", pair.from, got, pair.to)
		}
	}

	if got, want := UnicodeEquivalents(0x212A), []rune{0x004B, 0x006B, 0x212A}; !sameRunes(got, want) {
		t.Errorf("UnicodeEquivalents(kelvin) = %s, want %s", formatRunes(got), formatRunes(want))
	}
	if got, want := UnicodeEquivalents('K'), []rune{0x004B, 0x006B, 0x212A}; !sameRunes(got, want) {
		t.Errorf("UnicodeEquivalents(K) = %s, want %s", formatRunes(got), formatRunes(want))
	}
	if got, want := LegacyEquivalents(0x212A), []uint16{0x212A}; !sameUnits(got, want) {
		t.Errorf("LegacyEquivalents(kelvin) = %s", formatUnits(got))
	}
	if got, want := LegacyEquivalents(0x017F), []uint16{0x017F}; !sameUnits(got, want) {
		t.Errorf("LegacyEquivalents(long s) = %s", formatUnits(got))
	}
	if got, want := LegacyEquivalents('S'), []uint16{'S', 's'}; !sameUnits(got, want) {
		t.Errorf("LegacyEquivalents(S) = %s", formatUnits(got))
	}
	if got, want := UnicodeEquivalents('S'), []rune{'S', 's', 0x017F}; !sameRunes(got, want) {
		t.Errorf("UnicodeEquivalents(S) = %s", formatRunes(got))
	}
	if CanonicalizeUnicode(-1) != -1 || CanonicalizeUnicode(0x110000) != 0x110000 {
		t.Error("CanonicalizeUnicode changed an argument that is not a code point")
	}
	if UnicodeEquivalents(-1) != nil || UnicodeEquivalents(0x110000) != nil {
		t.Error("UnicodeEquivalents of a non-code-point is not nil")
	}
	if got := UnicodeEquivalents(0x10FFFF); len(got) != 1 || got[0] != 0x10FFFF {
		t.Errorf("U+10FFFF equivalents %s", formatRunes(got))
	}
}

func TestEquivalentsAreClosed(t *testing.T) {
	t.Parallel()
	if len(unicodeClass) != len(unicodeClassCanon) || len(legacyClass) != len(legacyClassCanon) {
		t.Fatalf("class tables disagree with their canon lists: unicode %d/%d legacy %d/%d",
			len(unicodeClass), len(unicodeClassCanon), len(legacyClass), len(legacyClassCanon))
	}
	for index := 1; index < len(unicodeFold); index++ {
		if unicodeFold[index][0] <= unicodeFold[index-1][0] {
			t.Fatalf("unicodeFold not sorted at %d", index)
		}
	}
	for index := 1; index < len(unicodeClassCanon); index++ {
		if unicodeClassCanon[index] <= unicodeClassCanon[index-1] {
			t.Fatalf("unicodeClassCanon not sorted at %d", index)
		}
	}
	for cp := rune(0); cp <= 0x10FFFF; cp++ {
		canon := CanonicalizeUnicode(cp)
		if CanonicalizeUnicode(canon) != canon {
			t.Fatalf("U+%04X -> U+%04X is not idempotent", cp, canon)
		}
		equivalents := UnicodeEquivalents(cp)
		if !containsRune(equivalents, cp) {
			t.Fatalf("U+%04X missing from %s", cp, formatRunes(equivalents))
		}
		var previous rune = -1
		for _, member := range equivalents {
			if member <= previous {
				t.Fatalf("equivalents of U+%04X not strictly sorted: %s", cp, formatRunes(equivalents))
			}
			previous = member
			if CanonicalizeUnicode(member) != canon {
				t.Fatalf("U+%04X is listed with U+%04X but folds to U+%04X", member, cp, CanonicalizeUnicode(member))
			}
		}
	}
	for cu := 0; cu <= 0xFFFF; cu++ {
		codeUnit := uint16(cu)
		canon := CanonicalizeLegacy(codeUnit)
		if CanonicalizeLegacy(canon) != canon {
			t.Fatalf("legacy U+%04X -> U+%04X is not idempotent", codeUnit, canon)
		}
		equivalents := LegacyEquivalents(codeUnit)
		if !containsUnit(equivalents, codeUnit) {
			t.Fatalf("legacy U+%04X missing from %s", codeUnit, formatUnits(equivalents))
		}
		var previous int = -1
		for _, member := range equivalents {
			if int(member) <= previous {
				t.Fatalf("legacy equivalents of U+%04X not strictly sorted", codeUnit)
			}
			previous = int(member)
			if CanonicalizeLegacy(member) != canon {
				t.Fatalf("legacy U+%04X is listed with U+%04X but maps to U+%04X", member, codeUnit, CanonicalizeLegacy(member))
			}
		}
	}
}

// Not parallel: each batch is its own Node process, and two of them at once
// were killed for memory on a 4 GB machine. The check is
// new RegExp('^\\uXXXX$', 'i').test(String.fromCharCode(Y)) for every code
// unit X and every code unit Y. The haystack is those 65536 units once each
// and every match is one unit wide, so a scan of pattern X is that test.
func TestCanonicalizeLegacyNode(t *testing.T) {
	valuesExamined, valueDisagreements := legacyValuesAgainstNode(t)
	fmt.Printf("legacy canonicalize values: %d code units against Node toUpperCase, disagreements %d\n", valuesExamined, len(valueDisagreements))
	for _, problem := range valueDisagreements {
		t.Error(problem)
	}

	const batchSize = 2048
	var (
		examined      int
		disagreements []string
		patterns      int
	)
	for start := 0; start <= 0xFFFF; start += batchSize {
		end := start + batchSize - 1
		if end > 0xFFFF {
			end = 0xFFFF
		}
		var input strings.Builder
		for cu := start; cu <= end; cu++ {
			equivalents := LegacyEquivalents(uint16(cu))
			fmt.Fprintf(&input, "%04X", cu)
			for _, member := range equivalents {
				fmt.Fprintf(&input, " %04X", member)
			}
			input.WriteByte('\n')
			patterns++
		}
		got, bad := runLegacyBatch(t, input.String())
		examined += got
		disagreements = append(disagreements, bad...)
	}
	fmt.Printf("legacy canonicalize: %d patterns x 65536 code units = %d, disagreements %d\n", patterns, examined, len(disagreements))
	if examined != patterns*65536 {
		t.Fatalf("examined %d code units, want %d", examined, patterns*65536)
	}
	reportDisagreements(t, disagreements)
}

// Top-level ranges are the gate's leaves. Strides use eight-line shards and
// four shards per range; cheaper class and dense scans keep sixteen-line shards.
// Every scan still visits every Unicode code point against Node.
// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange0(t *testing.T) { runUnicodeCanonicalizeRange(t, 0) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange1(t *testing.T) { runUnicodeCanonicalizeRange(t, 1) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange2(t *testing.T) { runUnicodeCanonicalizeRange(t, 2) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange3(t *testing.T) { runUnicodeCanonicalizeRange(t, 3) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange4(t *testing.T) { runUnicodeCanonicalizeRange(t, 4) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange5(t *testing.T) { runUnicodeCanonicalizeRange(t, 5) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange6(t *testing.T) { runUnicodeCanonicalizeRange(t, 6) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange7(t *testing.T) { runUnicodeCanonicalizeRange(t, 7) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange8(t *testing.T) { runUnicodeCanonicalizeRange(t, 8) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange9(t *testing.T) { runUnicodeCanonicalizeRange(t, 9) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange10(t *testing.T) { runUnicodeCanonicalizeRange(t, 10) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange11(t *testing.T) { runUnicodeCanonicalizeRange(t, 11) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange12(t *testing.T) { runUnicodeCanonicalizeRange(t, 12) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange13(t *testing.T) { runUnicodeCanonicalizeRange(t, 13) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange14(t *testing.T) { runUnicodeCanonicalizeRange(t, 14) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange15(t *testing.T) { runUnicodeCanonicalizeRange(t, 15) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange16(t *testing.T) { runUnicodeCanonicalizeRange(t, 16) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange17(t *testing.T) { runUnicodeCanonicalizeRange(t, 17) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange18(t *testing.T) { runUnicodeCanonicalizeRange(t, 18) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange19(t *testing.T) { runUnicodeCanonicalizeRange(t, 19) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange20(t *testing.T) { runUnicodeCanonicalizeRange(t, 20) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange21(t *testing.T) { runUnicodeCanonicalizeRange(t, 21) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange22(t *testing.T) { runUnicodeCanonicalizeRange(t, 22) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange23(t *testing.T) { runUnicodeCanonicalizeRange(t, 23) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange24(t *testing.T) { runUnicodeCanonicalizeRange(t, 24) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange25(t *testing.T) { runUnicodeCanonicalizeRange(t, 25) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange26(t *testing.T) { runUnicodeCanonicalizeRange(t, 26) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange27(t *testing.T) { runUnicodeCanonicalizeRange(t, 27) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange28(t *testing.T) { runUnicodeCanonicalizeRange(t, 28) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange29(t *testing.T) { runUnicodeCanonicalizeRange(t, 29) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange30(t *testing.T) { runUnicodeCanonicalizeRange(t, 30) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange31(t *testing.T) { runUnicodeCanonicalizeRange(t, 31) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange32(t *testing.T) { runUnicodeCanonicalizeRange(t, 32) }

// Not parallel: the range owns a four-process Node wave; concurrent ranges oversubscribe it.
func TestCanonicalizeUnicodeNodeRange33(t *testing.T) { runUnicodeCanonicalizeRange(t, 33) }

func runUnicodeCanonicalizeRange(t *testing.T, index int) {
	t.Helper()
	group := unicodeNodeTopRanges[index]
	if t.Name() != group.name {
		t.Fatalf("top-level range %s dispatched as %s", t.Name(), group.name)
	}
	lines := unicodeCanonicalizeLines()
	shards := unicodeCanonicalizeShards(lines)
	if err := checkUnicodeTopRangeCoverage(lines, shards, unicodeNodeTopRanges); err != nil {
		t.Fatal(err)
	}
	for _, shard := range shards[group.start:group.end] {
		t.Run(shard.name(), func(t *testing.T) {
			t.Parallel()
			examined, disagreements, err := runUnicodeBatch(shard.input)
			if err != nil {
				t.Fatal(err)
			}
			want := (shard.end - shard.start) * 0x110000
			if examined != want {
				t.Fatalf("checked %d code points, want %d", examined, want)
			}
			t.Logf("%d complete scans, %d code points, disagreements %d", shard.end-shard.start, examined, len(disagreements))
			reportDisagreements(t, disagreements)
		})
	}
}

func unicodeCanonicalizeLines() []string {
	var lines []string
	for index, canon := range unicodeClassCanon {
		members := unicodeClass[index]
		var text strings.Builder
		fmt.Fprintf(&text, "%X", canon)
		for _, member := range members {
			fmt.Fprintf(&text, " %X", member)
		}
		body := text.String()
		lines = append(lines, "CLASS iu "+body, "CLASS iv "+body)
	}
	blocks, strides := unicodeSingletonGroups()
	for _, group := range blocks {
		lines = append(lines, "GROUP iu "+formatCodePoints(group))
	}
	for _, group := range strides {
		lines = append(lines, "GROUP iu "+formatCodePoints(group))
	}

	return lines
}

// Results are collected by input index, never completion order. Workers cannot call Fatal:
// they return errors, and the test reports the first failed batch in input order after joining.
type unicodeNodeResult struct {
	index         int
	examined      int
	disagreements []string
	err           error
}

func unicodeNodeBatches(lines []string, batchSize, workers int, run func(string) (int, []string, error)) ([]unicodeNodeResult, error) {
	if batchSize < 0 || workers < 0 {
		return nil, fmt.Errorf("negative Unicode batch size or worker count")
	}
	if len(lines) == 0 {
		return nil, nil
	}
	if batchSize == 0 {
		batchSize = len(lines)
	}
	limit := runtime.GOMAXPROCS(0)
	if workers == 0 || workers > limit {
		workers = limit
	}
	batchCount := (len(lines)-1)/batchSize + 1
	if workers > batchCount {
		workers = batchCount
	}
	jobs := make(chan int)
	completed := make(chan unicodeNodeResult, workers)
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range jobs {
				start := index * batchSize
				end := min(start+batchSize, len(lines))
				input := strings.Join(lines[start:end], "\n") + "\n"
				examined, disagreements, err := run(input)
				completed <- unicodeNodeResult{index, examined, disagreements, err}
			}
		}()
	}
	go func() {
		for index := 0; index < batchCount; index++ {
			jobs <- index
		}
		close(jobs)
		group.Wait()
		close(completed)
	}()
	results := make([]unicodeNodeResult, batchCount)
	runs := make([]int, batchCount)
	for result := range completed {
		runs[result.index]++
		results[result.index] = result
	}
	for index, result := range results {
		if runs[index] != 1 {
			return nil, fmt.Errorf("Unicode batch %d ran %d times, want 1", index, runs[index])
		}
		if result.err != nil {
			return nil, fmt.Errorf("Unicode batch %d: %w", index, result.err)
		}
		want := min(batchSize, len(lines)-index*batchSize) * 0x110000
		if result.examined != want {
			return nil, fmt.Errorf("Unicode batch %d examined %d code points, want %d", index, result.examined, want)
		}
	}
	return results, nil
}

// Not parallel: temporarily controls GOMAXPROCS to prove the worker cap and output ordering.
func TestUnicodeNodeBatchOrderAndLimit(t *testing.T) {
	previous := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previous)
	firstStarted := make(chan struct{})
	secondFinished := make(chan struct{})
	var mutex sync.Mutex
	active, peak := 0, 0
	run := func(input string) (int, []string, error) {
		mutex.Lock()
		active++
		peak = max(peak, active)
		mutex.Unlock()
		switch input {
		case "0\n":
			close(firstStarted)
			<-secondFinished
		case "1\n":
			<-firstStarted
		case "2\n":
			// This worker has already sent batch 1, so batch 0 completes later.
			close(secondFinished)
		}
		mutex.Lock()
		active--
		mutex.Unlock()
		return 0x110000, []string{strings.TrimSuffix(input, "\n")}, nil
	}
	results, err := unicodeNodeBatches([]string{"0", "1", "2", "3", "4", "5"}, 1, 20, run)
	if err != nil {
		t.Fatal(err)
	}
	if peak != 2 {
		t.Fatalf("peak workers %d, want GOMAXPROCS=2", peak)
	}
	for index, result := range results {
		if result.index != index || len(result.disagreements) != 1 || result.disagreements[0] != strconv.Itoa(index) {
			t.Fatalf("batch %d reported out of order: %+v", index, result)
		}
	}
}

func unicodeSingletonGroups() (blocks, strides [][]uint32) {
	nontrivial := make(map[uint32]struct{}, len(unicodeClass)*2)
	for _, class := range unicodeClass {
		for _, member := range class {
			nontrivial[member] = struct{}{}
		}
	}
	singletons := make([]uint32, 0, 0x110000-len(nontrivial))
	for cp := uint32(0); cp <= 0x10FFFF; cp++ {
		if _, hot := nontrivial[cp]; !hot {
			singletons = append(singletons, cp)
		}
	}
	const group = 1024
	for start := 0; start < len(singletons); start += group {
		end := start + group
		if end > len(singletons) {
			end = len(singletons)
		}
		block := make([]uint32, end-start)
		copy(block, singletons[start:end])
		blocks = append(blocks, block)
	}
	strides = make([][]uint32, group)
	for index, cp := range singletons {
		strides[index%group] = append(strides[index%group], cp)
	}
	dense := strides[:0]
	for _, stride := range strides {
		if len(stride) > 0 {
			dense = append(dense, stride)
		}
	}
	return blocks, dense
}

func formatCodePoints(points []uint32) string {
	var text strings.Builder
	for index, point := range points {
		if index > 0 {
			text.WriteByte(' ')
		}
		fmt.Fprintf(&text, "%X", point)
	}
	return text.String()
}

const legacyValueScript = `
const lines = [];
for (let codeUnit = 0; codeUnit <= 0xFFFF; codeUnit++) {
  const upper = String.fromCharCode(codeUnit).toUpperCase();
  let canon = codeUnit;
  if (upper.length === 1) {
    const mapped = upper.charCodeAt(0);
    if (!(codeUnit >= 128 && mapped < 128)) canon = mapped;
  }
  if (canon !== codeUnit) lines.push(codeUnit.toString(16) + " " + canon.toString(16));
}
console.log(lines.join("\n"));
console.log(".");
`

func legacyValuesAgainstNode(t *testing.T) (int, []string) {
	t.Helper()
	output := runNode(t, legacyValueScript, "", 2*time.Minute)
	got := map[uint16]uint16{}
	for _, line := range strings.Split(strings.TrimSuffix(output, "\n.\n"), "\n") {
		if line == "" || line == "." {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("value line %q", line)
		}
		from := parseHex16(t, fields[0])
		to := parseHex16(t, fields[1])
		got[from] = to
	}
	var problems []string
	for _, mapping := range legacyFold {
		want := mapping[1]
		if node, ok := got[mapping[0]]; !ok || node != want {
			problems = append(problems, fmt.Sprintf("legacy value U+%04X: tables U+%04X, node %s", mapping[0], want, hexOrMissing(got, mapping[0])))
		}
		delete(got, mapping[0])
	}
	for from, to := range got {
		problems = append(problems, fmt.Sprintf("legacy value U+%04X: tables say itself, node says U+%04X", from, to))
	}
	sort.Strings(problems)
	return 65536, problems
}

func hexOrMissing(got map[uint16]uint16, codeUnit uint16) string {
	to, ok := got[codeUnit]
	if !ok {
		return "itself"
	}
	return fmt.Sprintf("U+%04X", to)
}

const legacyScanScript = `
function same(got, want) {
  if (got.length !== want.length) return false;
  for (let index = 0; index < got.length; index++) if (got[index] !== want[index]) return false;
  return true;
}
const parts = new Array(0x10000);
for (let codeUnit = 0; codeUnit <= 0xFFFF; codeUnit++) parts[codeUnit] = String.fromCharCode(codeUnit);
const all = parts.join("");
function scan(codeUnit) {
  const re = new RegExp("\\u" + codeUnit.toString(16).padStart(4, "0"), "gi");
  const found = [];
  let match;
  while ((match = re.exec(all)) !== null) {
    if (match[0].length !== 1) throw new Error("legacy match width " + match[0].length + " at " + match.index);
    found.push(match.index);
    if (re.lastIndex === match.index) re.lastIndex++;
  }
  return found;
}
function expect(codeUnit, want, label) {
  const got = scan(codeUnit);
  if (!same(got, want)) throw new Error(label + " got " + got.map(n => n.toString(16)).join(","));
}
expect(0x53, [0x53, 0x73], "S");
expect(0x73, [0x53, 0x73], "s");
expect(0x17F, [0x17F], "long s");
expect(0x131, [0x131], "dotless i");
expect(0x49, [0x49, 0x69], "I");
expect(0x69, [0x49, 0x69], "i");
expect(0x212A, [0x212A], "kelvin");
expect(0x212B, [0x212B], "angstrom");
expect(0xC5, [0xC5, 0xE5], "A ring");
expect(0x1F80, [0x1F80], "ypogegrammeni");
expect(0x1F88, [0x1F88], "prosgegrammeni");
expect(0xB5, [0xB5, 0x39C, 0x3BC], "micro");

const fs = require("node:fs");
const lines = fs.readFileSync(0, "utf8").split("\n");
let checked = 0;
for (const line of lines) {
  if (line === "") continue;
  const fields = line.split(" ");
  const pattern = parseInt(fields[0], 16);
  const want = fields.slice(1).map(part => parseInt(part, 16));
  const got = scan(pattern);
  checked++;
  if (!same(got, want)) {
    const extra = [];
    const missing = [];
    const have = new Set(got);
    const need = new Set(want);
    for (const codeUnit of got) if (!need.has(codeUnit)) extra.push(codeUnit.toString(16));
    for (const codeUnit of want) if (!have.has(codeUnit)) missing.push(codeUnit.toString(16));
    console.log("BAD " + fields[0] + " EXTRA " + extra.join(",") + " MISSING " + missing.join(","));
  }
}
console.log("TOTAL " + checked);
`

func runLegacyBatch(t *testing.T, input string) (int, []string) {
	t.Helper()
	output := runNode(t, legacyScanScript, input, 3*time.Minute)
	return parseBatch(t, output, 65536)
}

const unicodeScanScript = `
const chunks = [];
function bmp(from, to) {
  const parts = new Array(to - from + 1);
  for (let codePoint = from, index = 0; codePoint <= to; codePoint++, index++) {
    parts[index] = String.fromCharCode(codePoint);
  }
  return [parts.join(""), index => from + index, 1];
}
chunks.push(bmp(0x0000, 0xD7FF));
chunks.push(bmp(0xD800, 0xDBFF));
chunks.push(bmp(0xDC00, 0xDFFF));
chunks.push(bmp(0xE000, 0xFFFF));
for (let plane = 1; plane <= 16; plane++) {
  const base = plane * 0x10000;
  const parts = new Array(0x10000);
  for (let index = 0; index < 0x10000; index++) parts[index] = String.fromCodePoint(base + index);
  chunks.push([parts.join(""), index => base + (index >> 1), 2]);
}
function same(got, want) {
  if (got.length !== want.length) return false;
  for (let index = 0; index < got.length; index++) if (got[index] !== want[index]) return false;
  return true;
}
function matchesOf(pattern, flags) {
  const re = new RegExp(pattern, flags);
  const found = [];
  for (const [text, toCodePoint, width] of chunks) {
    re.lastIndex = 0;
    let match;
    while ((match = re.exec(text)) !== null) {
      if (match[0].length !== width) throw new Error("width " + match[0].length + " want " + width + " for /" + pattern + "/" + flags + " at " + match.index);
      if (width === 2 && (match.index & 1) !== 0) throw new Error("odd index for " + pattern);
      found.push(toCodePoint(match.index));
      if (re.lastIndex === match.index) re.lastIndex++;
    }
  }
  found.sort((a, b) => a - b);
  for (let index = 1; index < found.length; index++) {
    if (found[index] === found[index - 1]) throw new Error("duplicate for " + pattern);
  }
  return found;
}
function expect(pattern, flags, want, label) {
  const got = matchesOf(pattern, flags);
  if (!same(got, want)) throw new Error(label + " got " + got.map(n => n.toString(16)).join(",") + " want " + want.map(n => n.toString(16)).join(","));
}
expect("\\u{41}", "giu", [0x41, 0x61], "A iu");
expect("\\u{212A}", "giu", [0x4B, 0x6B, 0x212A], "kelvin iu");
expect("\\u{212A}", "giv", [0x4B, 0x6B, 0x212A], "kelvin iv");
expect("\\u{10400}", "giu", [0x10400, 0x10428], "deseret iu");
expect("[\\u{2D}\\u{5C}\\u{5D}\\u{A0}\\u{D800}\\u{10000}]", "giu", [0x2D, 0x5C, 0x5D, 0xA0, 0xD800, 0x10000], "class iu");
expect("[\\u{A0}\\u{D800}\\u{10000}]", "giv", [0xA0, 0xD800, 0x10000], "class iv");

const fs = require("node:fs");
const lines = fs.readFileSync(0, "utf8").split("\n");
let checked = 0;
for (const line of lines) {
  if (line === "") continue;
  const fields = line.split(" ");
  const kind = fields[0];
  const flag = fields[1];
  const flags = flag === "iu" ? "giu" : flag === "iv" ? "giv" : "";
  if (flags === "") throw new Error("flag " + flag);
  const numbers = fields.slice(2).map(part => parseInt(part, 16));
  let pattern;
  let want;
  if (kind === "CLASS") {
    pattern = "\\u{" + numbers[0].toString(16) + "}";
    want = numbers.slice(1);
  } else if (kind === "GROUP") {
    pattern = "[" + numbers.map(codePoint => "\\u{" + codePoint.toString(16) + "}").join("") + "]";
    want = numbers;
  } else {
    throw new Error("kind " + kind);
  }
  want = want.slice().sort((a, b) => a - b);
  const got = matchesOf(pattern, flags);
  checked++;
  if (!same(got, want)) {
    const extra = [];
    const missing = [];
    const have = new Set(got);
    const need = new Set(want);
    for (const codePoint of got) if (!need.has(codePoint)) extra.push(codePoint.toString(16));
    for (const codePoint of want) if (!have.has(codePoint)) missing.push(codePoint.toString(16));
    const head = kind === "CLASS" ? numbers[0].toString(16) : "n" + numbers.length;
    console.log("BAD " + kind + " " + flag + " " + head + " EXTRA " + extra.join(",") + " MISSING " + missing.join(","));
  }
}
console.log("TOTAL " + checked);
`

func runUnicodeBatch(input string) (int, []string, error) {
	// Preserve the original deadline for 80-line batches. Larger measurement batches
	// need proportionally more time, bounded by the gate's thirty-minute deadline.
	batches := max(1, (strings.Count(input, "\n")+79)/80)
	timeout := min(30*time.Minute, time.Duration(batches)*4*time.Minute)
	output, err := runNodeOutput(unicodeScanScript, input, timeout)
	if err != nil {
		return 0, nil, err
	}
	return parseBatchOutput(output, 0x110000)
}

func parseBatch(t *testing.T, output string, perPattern int) (int, []string) {
	t.Helper()
	examined, problems, err := parseBatchOutput(output, perPattern)
	if err != nil {
		t.Fatal(err)
	}
	return examined, problems
}

func parseBatchOutput(output string, perPattern int) (int, []string, error) {
	var problems []string
	checked := -1
	scanner := bufio.NewScanner(strings.NewReader(output))
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "TOTAL "):
			value, err := strconv.Atoi(strings.TrimPrefix(line, "TOTAL "))
			if err != nil {
				return 0, nil, fmt.Errorf("total %q", line)
			}
			checked = value
		case strings.HasPrefix(line, "BAD "):
			problems = append(problems, line)
		default:
			return 0, nil, fmt.Errorf("node line %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, nil, err
	}
	if checked < 0 {
		return 0, nil, fmt.Errorf("node did not report a total\n%s", output)
	}
	return checked * perPattern, problems, nil
}

func runNode(t *testing.T, script, input string, timeout time.Duration) string {
	t.Helper()
	output, err := runNodeOutput(script, input, timeout)
	if err != nil {
		t.Fatal(err)
	}
	return output
}

func runNodeOutput(script, input string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--eval", script)
	command.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		tail := stdout.String()
		if len(tail) > 800 {
			tail = tail[len(tail)-800:]
		}
		return "", fmt.Errorf("node: %v\n%s\n%s", err, stderr.String(), tail)
	}
	return stdout.String(), nil
}

func reportDisagreements(t *testing.T, disagreements []string) {
	t.Helper()
	const limit = 80
	for index, problem := range disagreements {
		if index == limit {
			t.Errorf("and %d more disagreements", len(disagreements)-limit)
			return
		}
		t.Error(problem)
	}
}

func parseHex16(t *testing.T, text string) uint16 {
	t.Helper()
	value, err := strconv.ParseUint(text, 16, 16)
	if err != nil {
		t.Fatalf("hex %q: %v", text, err)
	}
	return uint16(value)
}

func sameRunes(got, want []rune) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func sameUnits(got, want []uint16) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func containsRune(list []rune, value rune) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func containsUnit(list []uint16, value uint16) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func formatRunes(list []rune) string {
	parts := make([]string, len(list))
	for index, value := range list {
		parts[index] = fmt.Sprintf("U+%04X", value)
	}
	return strings.Join(parts, " ")
}

func formatUnits(list []uint16) string {
	parts := make([]string, len(list))
	for index, value := range list {
		parts[index] = fmt.Sprintf("U+%04X", value)
	}
	return strings.Join(parts, " ")
}
