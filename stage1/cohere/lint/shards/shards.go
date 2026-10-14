// Package shards runs the stage 1 lint driver as N processes over one manifest and merges their output
// into exactly what one process prints.
//
// It is the interim parallelism of #tj6d455: Adamic has no way yet for a program to run copies of itself,
// and parallelMap arrives with concurrency on Adamic main. Each shard is the same native binary with
// `--shard <index>/<count>`; it runs every count-th row and numbers its cases as the whole manifest does,
// so merging is putting the case blocks back in case order. A file is linted, and fixed, whole inside one
// process, so a fix pass never crosses a shard.
package shards

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

// Run lints manifest with binary in count processes and returns the merged standard output. With
// countOnly, each shard prints its findings count and the result is their sum, as one process prints it.
func Run(binary, manifest string, count int, countOnly bool) ([]byte, error) {
	if count < 1 {
		return nil, fmt.Errorf("shard count %d, want at least 1", count)
	}
	text, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}
	rows := Cases(text)
	outputs := make([][]byte, count)
	failures := make([]error, count)
	var group sync.WaitGroup
	for index := 0; index < count; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			arguments := []string{"--manifest", manifest, "--shard", fmt.Sprintf("%d/%d", index, count)}
			if countOnly {
				arguments = append(arguments, "--count")
			}
			var stdout, stderr bytes.Buffer
			command := exec.Command(binary, arguments...)
			command.Stdout = &stdout
			command.Stderr = &stderr
			if err := command.Run(); err != nil {
				failures[index] = fmt.Errorf("shard %d/%d: %v: %s", index, count, err, stderr.Bytes())
				return
			}
			outputs[index] = stdout.Bytes()
		}(index)
	}
	group.Wait()
	for _, failure := range failures {
		if failure != nil {
			return nil, failure
		}
	}
	if countOnly {
		total := 0
		for index, output := range outputs {
			value, err := strconv.Atoi(string(bytes.TrimSpace(output)))
			if err != nil {
				return nil, fmt.Errorf("shard %d/%d printed %q, not a count", index, count, output)
			}
			total += value
		}
		return []byte(strconv.Itoa(total) + "\n"), nil
	}
	return Merge(outputs, rows)
}

var programPrefix = []byte("program ")

// Cases is how many cases manifest holds, counted as main.ts numbers them: every non-empty line is a case
// except a `program <tsconfig>` line, which opens the checker's program for the rows after it and prints no
// case of its own. Counting that line too made every sharded typed run expect one case more than any shard
// printed, so Merge refused it (#3mecm8z).
func Cases(manifest []byte) int {
	cases := 0
	for _, row := range bytes.Split(manifest, []byte("\n")) {
		if len(row) > 0 && !bytes.HasPrefix(row, programPrefix) {
			cases++
		}
	}
	return cases
}

var casePrefix = []byte("case ")

// Merge puts the shards' case blocks back in case order. A block starts at a line `case <n>` and runs to
// the next one; nothing a case prints starts a line that way, since findings, ranges, edits and the fixed
// text are printed under their own prefixes with newlines escaped. Every case number from 0 must appear
// exactly once, and exactly rows of them, the manifest's cases (Cases): a shard that lost or repeated a
// row fails here instead of shifting the output, including one that lost the last rows, which leaves no gap.
func Merge(outputs [][]byte, rows int) ([]byte, error) {
	blocks := map[int][]byte{}
	for shard, output := range outputs {
		if len(output) == 0 {
			continue
		}
		if !bytes.HasPrefix(output, casePrefix) {
			return nil, fmt.Errorf("shard %d output does not start with a case line", shard)
		}
		start := 0
		for start < len(output) {
			lineEnd := bytes.IndexByte(output[start:], '\n')
			if lineEnd < 0 {
				return nil, fmt.Errorf("shard %d: unterminated case line", shard)
			}
			number, err := strconv.Atoi(string(output[start+len(casePrefix) : start+lineEnd]))
			if err != nil {
				return nil, fmt.Errorf("shard %d: malformed case line %q", shard, output[start:start+lineEnd])
			}
			end := len(output)
			for search := start + lineEnd + 1; search < len(output); {
				next := bytes.IndexByte(output[search:], '\n')
				if bytes.HasPrefix(output[search:], casePrefix) && isCaseLine(output[search:]) {
					end = search
					break
				}
				if next < 0 {
					break
				}
				search += next + 1
			}
			if _, seen := blocks[number]; seen {
				return nil, fmt.Errorf("case %d printed twice", number)
			}
			blocks[number] = output[start:end]
			start = end
		}
	}
	if len(blocks) != rows {
		for number := 0; number < rows; number++ {
			if _, ok := blocks[number]; !ok {
				return nil, fmt.Errorf("case %d missing from every shard", number)
			}
		}
		return nil, fmt.Errorf("%d cases printed, want the manifest's %d", len(blocks), rows)
	}
	var merged bytes.Buffer
	for number := 0; number < rows; number++ {
		block, ok := blocks[number]
		if !ok {
			return nil, fmt.Errorf("case %d missing from every shard", number)
		}
		merged.Write(block)
	}
	return merged.Bytes(), nil
}

// isCaseLine is whether text starts with a whole `case <digits>` line.
func isCaseLine(text []byte) bool {
	rest := text[len(casePrefix):]
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	return digits > 0 && digits < len(rest) && rest[digits] == '\n'
}
