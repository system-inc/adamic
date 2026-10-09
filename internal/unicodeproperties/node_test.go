package unicodeproperties

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// nodeScanner is one Node process. It builds a string of every code point, split
// so that no surrogate pair is formed by accident (a lead followed by a trail
// would be one supplementary character), then matches \p{...} with the u flag.
// The self-check at the top is what a scanner that drops U+0000, or that joins
// the surrogate block into one pair, fails before any of our tables are trusted.
const nodeScanner = `
const chunks = [];
function bmp(from, to) {
  const parts = new Array(to - from + 1);
  for (let codePoint = from, index = 0; codePoint <= to; codePoint++, index++) {
    parts[index] = String.fromCharCode(codePoint);
  }
  return parts.join("");
}
chunks.push([bmp(0x0000, 0xD7FF), index => index]);
chunks.push([bmp(0xD800, 0xDBFF), index => 0xD800 + index]);
chunks.push([bmp(0xDC00, 0xDFFF), index => 0xDC00 + index]);
chunks.push([bmp(0xE000, 0xFFFF), index => 0xE000 + index]);
for (let plane = 1; plane <= 16; plane++) {
  const base = plane * 0x10000;
  const parts = new Array(0x10000);
  for (let index = 0; index < 0x10000; index++) parts[index] = String.fromCodePoint(base + index);
  chunks.push([parts.join(""), index => base + (index >> 1)]);
}

function rangesOf(expression) {
  const re = new RegExp("\\p{" + expression + "}", "gu");
  const ranges = [];
  for (const [text, toCodePoint] of chunks) {
    re.lastIndex = 0;
    let start = -1;
    let previous = -1;
    let match;
    while ((match = re.exec(text)) !== null) {
      const codePoint = toCodePoint(match.index);
      if (start < 0) {
        start = previous = codePoint;
      } else if (codePoint === previous + 1) {
        previous = codePoint;
      } else {
        ranges.push(start, previous);
        start = previous = codePoint;
      }
    }
    if (start >= 0) ranges.push(start, previous);
  }
  const coalesced = [];
  for (let index = 0; index < ranges.length; index += 2) {
    const start = ranges[index];
    const end = ranges[index + 1];
    if (coalesced.length > 0 && start === coalesced[coalesced.length - 1] + 1) {
      coalesced[coalesced.length - 1] = end;
    } else {
      coalesced.push(start, end);
    }
  }
  return coalesced;
}

function same(got, want) {
  if (got.length !== want.length) return false;
  for (let index = 0; index < got.length; index++) if (got[index] !== want[index]) return false;
  return true;
}
const ascii = rangesOf("ASCII");
if (!same(ascii, [0, 0x7f])) throw new Error("scanner ASCII " + ascii.join(","));
const any = rangesOf("Any");
if (!same(any, [0, 0x10ffff])) throw new Error("scanner Any " + any.slice(0, 8).join(",") + " n=" + any.length);
const surrogates = rangesOf("Cs");
if (!same(surrogates, [0xd800, 0xdfff])) throw new Error("scanner Cs " + surrogates.join(","));

const fs = require("node:fs");
const input = fs.readFileSync(0, "utf8").split("\n");
for (const expression of input) {
  if (expression === "") continue;
  let ranges;
  try {
    ranges = rangesOf(expression);
  } catch (error) {
    console.log("ERR " + expression);
    console.log(String(error && error.message ? error.message : error).split("\n")[0]);
    console.log(".");
    continue;
  }
  console.log("OK " + expression);
  for (let index = 0; index < ranges.length; index += 2) {
    console.log(ranges[index].toString(16) + " " + ranges[index + 1].toString(16));
  }
  console.log(".");
}
`

// nodeStrings checks properties of strings with the v flag. Each sequence is
// hex code points. A line "." ends a property. The answer is one line per
// sequence, "1" or "0".
const nodeStrings = `
const fs = require("node:fs");
const lines = fs.readFileSync(0, "utf8").split("\n");
let expression = "";
let re = null;
for (const line of lines) {
  if (line === "") continue;
  if (line === ".") {
    expression = "";
    re = null;
    continue;
  }
  if (expression === "") {
    expression = line;
    try {
      re = new RegExp("^\\p{" + expression + "}$", "v");
    } catch (error) {
      console.log("ERR " + String(error && error.message ? error.message : error).split("\n")[0]);
      re = null;
    }
    continue;
  }
  if (re === null) {
    console.log("0");
    continue;
  }
  const codePoints = line.split(" ").map(part => parseInt(part, 16));
  console.log(re.test(String.fromCodePoint(...codePoints)) ? "1" : "0");
}
`

func TestVersionMatchesNode(t *testing.T) {
	t.Parallel()
	version, unicode := nodeVersion(t)
	if !strings.HasPrefix(version, "v24.") {
		t.Fatalf("node is %s, want Node 24 (the oracle this package was built against)", version)
	}
	if unicode != NodeUnicodeVersion {
		t.Fatalf("Node reports Unicode %s, tables are %s (%s): rerun go generate", unicode, NodeUnicodeVersion, Version)
	}
}

func nodeVersion(t *testing.T) (string, string) {
	t.Helper()
	output, err := exec.Command("node", "--print", "process.version + \" \" + process.versions.unicode").Output()
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(string(output))
	if len(fields) != 2 {
		t.Fatalf("node version line %q", output)
	}
	return fields[0], fields[1]
}

func codePointExpressions() []string {
	expressions := make([]string, 0, len(binaryNames)+len(generalCategoryNames)*3+len(scriptNames)*4)
	for name := range binaryNames {
		expressions = append(expressions, name)
	}
	for name := range generalCategoryNames {
		expressions = append(expressions, name, "gc="+name, "General_Category="+name)
	}
	for name := range scriptNames {
		expressions = append(expressions, "sc="+name, "Script="+name)
	}
	for name := range scriptExtensionNames {
		expressions = append(expressions, "scx="+name, "Script_Extensions="+name)
	}
	sort.Strings(expressions)
	return expressions
}

// Not parallel: it runs Node over every code point of every property, and the
// two workers below already use both cores.
func TestNodeAgrees(t *testing.T) {
	expressions := codePointExpressions()
	const batchSize = 80
	var batches [][]string
	for start := 0; start < len(expressions); start += batchSize {
		end := start + batchSize
		if end > len(expressions) {
			end = len(expressions)
		}
		batches = append(batches, expressions[start:end])
	}

	type disagreement struct {
		expression string
		count      int
		sample     []uint32
		detail     string
	}
	var (
		mu       sync.Mutex
		problems []disagreement
		checked  int
	)
	jobs := make(chan []string)
	var workers sync.WaitGroup
	for worker := 0; worker < 2; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for batch := range jobs {
				got, err := runScanner(batch)
				mu.Lock()
				if err != nil {
					problems = append(problems, disagreement{detail: err.Error()})
					mu.Unlock()
					continue
				}
				checked += len(batch)
				for _, expression := range batch {
					property, ok := Lookup(expression, false)
					if !ok || property.Kind != KindCodePoints {
						problems = append(problems, disagreement{expression: expression, detail: "lookup failed"})
						continue
					}
					ranges, ok := got[expression]
					if !ok {
						problems = append(problems, disagreement{expression: expression, detail: "node produced no answer"})
						continue
					}
					if rangesEqual(property.Set.Ranges, ranges) {
						continue
					}
					count, sample := differences(property.Set.Ranges, ranges)
					problems = append(problems, disagreement{expression: expression, count: count, sample: sample})
				}
				mu.Unlock()
			}
		}()
	}
	for _, batch := range batches {
		jobs <- batch
	}
	close(jobs)
	workers.Wait()

	codePoints := len(expressions) * (0x10FFFF + 1)
	fmt.Printf("node code point check: %d expressions x %d code points = %d, disagreements %d\n",
		len(expressions), 0x10FFFF+1, codePoints, len(problems))
	if checked != len(expressions) && len(problems) == 0 {
		t.Fatalf("checked %d expressions, want %d", checked, len(expressions))
	}
	for _, problem := range problems {
		if problem.detail != "" && problem.expression == "" {
			t.Error(problem.detail)
			continue
		}
		if problem.detail != "" {
			t.Errorf("\\p{%s}: %s", problem.expression, problem.detail)
			continue
		}
		t.Errorf("\\p{%s}: %d code points disagree, first %s", problem.expression, problem.count, formatPoints(problem.sample))
	}
}

func rangesEqual(a, b []Range) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}

func differences(want, got []Range) (int, []uint32) {
	have := &Set{Ranges: want}
	other := &Set{Ranges: got}
	count := 0
	var sample []uint32
	for codePoint := rune(0); codePoint <= 0x10FFFF; codePoint++ {
		if have.Contains(codePoint) == other.Contains(codePoint) {
			continue
		}
		count++
		if len(sample) < 8 {
			sample = append(sample, uint32(codePoint))
		}
	}
	return count, sample
}

func formatPoints(points []uint32) string {
	parts := make([]string, len(points))
	for index, point := range points {
		parts[index] = fmt.Sprintf("U+%04X", point)
	}
	return strings.Join(parts, " ")
}

func runScanner(expressions []string) (map[string][]Range, error) {
	command := exec.Command("node", "--eval", nodeScanner)
	command.Stdin = strings.NewReader(strings.Join(expressions, "\n") + "\n")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	// Each completed property already produces output, so it renews the stall guard.
	if err := childguard.Run(command, childguard.Options{}); err != nil {
		return nil, fmt.Errorf("node scanner: %v\n%s", err, stderr.String())
	}
	return parseRanges(stdout.String())
}

func parseRanges(output string) (map[string][]Range, error) {
	parsed := map[string][]Range{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var expression string
	var ranges []Range
	open := false
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "OK "):
			if open {
				return nil, fmt.Errorf("new expression %q before %q closed", line, expression)
			}
			expression = strings.TrimPrefix(line, "OK ")
			ranges = nil
			open = true
		case strings.HasPrefix(line, "ERR "):
			return nil, fmt.Errorf("node rejected %s", strings.TrimPrefix(line, "ERR "))
		case line == ".":
			if !open {
				return nil, fmt.Errorf("closing a property that was not open")
			}
			parsed[expression] = ranges
			open = false
		default:
			fields := strings.Fields(line)
			if len(fields) != 2 {
				return nil, fmt.Errorf("range line %q", line)
			}
			start, err := strconv.ParseUint(fields[0], 16, 32)
			if err != nil {
				return nil, err
			}
			end, err := strconv.ParseUint(fields[1], 16, 32)
			if err != nil {
				return nil, err
			}
			ranges = append(ranges, Range{Start: uint32(start), End: uint32(end)})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if open {
		return nil, fmt.Errorf("expression %s was not closed", expression)
	}
	return parsed, nil
}

func TestNodeStringProperties(t *testing.T) {
	t.Parallel()
	var input strings.Builder
	total := 0
	for _, name := range []string{
		"Basic_Emoji",
		"Emoji_Keycap_Sequence",
		"RGI_Emoji_Modifier_Sequence",
		"RGI_Emoji_Flag_Sequence",
		"RGI_Emoji_Tag_Sequence",
		"RGI_Emoji_ZWJ_Sequence",
		"RGI_Emoji",
	} {
		property, ok := Lookup(name, true)
		if !ok {
			t.Fatalf("Lookup %s", name)
		}
		fmt.Fprintf(&input, "%s\n", name)
		for _, sequence := range property.Sequences {
			for index, codePoint := range []rune(sequence) {
				if index > 0 {
					input.WriteByte(' ')
				}
				fmt.Fprintf(&input, "%x", codePoint)
			}
			input.WriteByte('\n')
			total++
		}
		input.WriteString(".\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--eval", nodeStrings)
	command.Stdin = strings.NewReader(input.String())
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node strings: %v\n%s", err, stderr.String())
	}
	answers := strings.Fields(stdout.String())
	fmt.Printf("node string-property check: %d sequences, disagreements %d\n", total, countNot(answers, "1"))
	if len(answers) != total {
		t.Fatalf("node answered %d lines, want %d\n%s", len(answers), total, stdout.String()[:smaller(400, stdout.Len())])
	}
	var disagreed int
	cursor := 0
	for _, name := range []string{
		"Basic_Emoji", "Emoji_Keycap_Sequence", "RGI_Emoji_Modifier_Sequence",
		"RGI_Emoji_Flag_Sequence", "RGI_Emoji_Tag_Sequence", "RGI_Emoji_ZWJ_Sequence", "RGI_Emoji",
	} {
		property, _ := Lookup(name, true)
		for _, sequence := range property.Sequences {
			if answers[cursor] != "1" {
				disagreed++
				if disagreed <= 12 {
					t.Errorf("\\p{%s} does not match %q (node said %s)", name, sequence, answers[cursor])
				}
			}
			cursor++
		}
	}
	if disagreed > 12 {
		t.Errorf("%d sequences disagreed", disagreed)
	}
}

func countNot(answers []string, want string) int {
	count := 0
	for _, answer := range answers {
		if answer != want {
			count++
		}
	}
	return count
}

func smaller(a, b int) int {
	if a < b {
		return a
	}
	return b
}
