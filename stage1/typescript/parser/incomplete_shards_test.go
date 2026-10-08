package parser

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A child owns a small manifest, never the whole compiler corpus.
const incompleteShardDeadline = 2 * time.Minute

// Case headers come from both existing manifest runners. Require exactly the
// scheduled sequence before comparing bodies, including empty-file cases.
func incompleteFrames(output []byte, names []string) ([][]byte, error) {
	frames := [][]byte{}
	bodyStart := -1
	for start := 0; start < len(output); {
		end := bytes.IndexByte(output[start:], '\n')
		if end < 0 {
			return nil, fmt.Errorf("unterminated output after %d cases", len(frames))
		}
		end += start
		line := output[start:end]
		if bytes.HasPrefix(line, []byte("case ")) {
			if bodyStart >= 0 {
				frames = append(frames, output[bodyStart:start])
			}
			index, err := strconv.Atoi(string(line[5:]))
			expected := len(frames)
			if expected >= len(names) {
				return nil, fmt.Errorf("extra case %q after %s", line, names[len(names)-1])
			}
			if err != nil || index != expected {
				return nil, fmt.Errorf("missing case %s: expected case %d, got %q (missing, extra, or duplicated case)", names[expected], expected, line)
			}
			bodyStart = end + 1
		} else if bodyStart < 0 {
			return nil, fmt.Errorf("missing case %s: output before its header", names[0])
		}
		start = end + 1
	}
	if bodyStart >= 0 {
		frames = append(frames, output[bodyStart:])
	}
	if len(frames) != len(names) {
		return nil, fmt.Errorf("missing case %s: returned %d cases, expected %d", names[len(frames)], len(frames), len(names))
	}
	return frames, nil
}

func checkIncompleteFrames(output []byte, names []string) error {
	_, err := incompleteFrames(output, names)
	return err
}

func compareIncompleteFrames(want, got []byte, names []string) error {
	oracle, err := incompleteFrames(want, names)
	if err != nil {
		return fmt.Errorf("oracle: %w", err)
	}
	port, err := incompleteFrames(got, names)
	if err != nil {
		return err
	}
	var differences []error
	for index := range oracle {
		if !bytes.Equal(oracle[index], port[index]) {
			differences = append(differences, fmt.Errorf("case %s: %s", names[index], difference(port[index], oracle[index])))
		}
	}
	return errors.Join(differences...)
}

func TestIncompleteShardCoverage(t *testing.T) {
	t.Parallel()
	names := []string{"testdata/recovery/corePublic.ts-cut-37.ts.txt", "testdata/recovery/ts.moduleSpecifiers.ts-cut-1.ts.txt"}
	healthy := []byte("case 0\nfile\ncase 1\nfile\n")
	if err := compareIncompleteFrames(healthy, healthy, names); err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, output, caseName string }{
		{"one-byte", "case 0\ngile\ncase 1\nfile\n", names[0]},
		{"dropped-first", "case 1\nfile\n", names[0]},
		{"dropped-last", "case 0\nfile\n", names[1]},
		{"extra", "case 0\nfile\ncase 1\nfile\ncase 2\nfile\n", names[1]},
		{"duplicate", "case 0\nfile\ncase 0\nfile\n", names[1]},
	} {
		t.Run(probe.name, func(t *testing.T) {
			err := compareIncompleteFrames(healthy, []byte(probe.output), names)
			if err == nil || !strings.Contains(err.Error(), probe.caseName) {
				t.Fatalf("coverage check missed named failure: %v", err)
			}
			t.Log(err)
		})
	}
	t.Run("all-differences", func(t *testing.T) {
		err := compareIncompleteFrames(healthy, []byte("case 0\ngile\ncase 1\ngile\n"), names)
		for _, name := range names {
			if err == nil || !strings.Contains(err.Error(), name) {
				t.Fatalf("comparison stopped before named case %s: %v", name, err)
			}
		}
		t.Log(err)
	})

}
