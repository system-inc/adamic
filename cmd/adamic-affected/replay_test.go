package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

var replayTrace = flag.String("affected-replay-trace", "", "real trace for interleaved observer timing")
var replayRoot = flag.String("affected-replay-root", "", "fixed main checkout for trace replay")

func TestObservationReplayTimings(t *testing.T) {
	// Not parallel: before and after share one immutable checkout and one box.
	if *replayTrace == "" || *replayRoot == "" {
		t.Skip("explicit replay input required")
	}
	file, err := os.Open(*replayTrace)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 16<<20)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	file.Close()
	root := *replayRoot
	metadata := map[string]string{"cached": "uncached", "instrument": "go test -count=1 -timeout 30m ./cmd/adamic-affected -run ^TestObservationReplayTimings$ -args -affected-replay-trace " + *replayTrace + " -affected-replay-root " + root}
	for key, args := range map[string][]string{"commit": {"git", "rev-parse", "HEAD"}, "nproc": {"nproc"}, "go": {"go", "version"}, "clang": {"clang", "--version"}, "node": {"node", "--version"}} {
		value, err := command(root, args[0], args[1:]...)
		if err != nil {
			t.Fatal(err)
		}
		metadata[key] = strings.TrimSpace(string(value))
		if key == "clang" {
			metadata[key] = strings.Split(metadata[key], "\n")[0]
		}
	}
	quota, err := os.ReadFile("/sys/fs/cgroup/cpu.max")
	if err != nil {
		t.Fatal(err)
	}
	metadata["cpu.max"] = strings.TrimSpace(string(quota))
	var expected *closure
	for trial := 0; trial < 3; trial++ {
		for _, kind := range []string{"before", "after"} {
			load, err := os.ReadFile("/proc/loadavg")
			if err != nil {
				t.Fatal(err)
			}
			metadata["load_before"] = strings.TrimSpace(string(load))
			value := closure{Observed: map[string]string{}}
			started := time.Now()
			for _, line := range lines {
				if kind == "before" {
					observeLine(root, root, line, &value)
				} else {
					collectLine(root, root, line, &value)
				}
			}
			if kind == "after" {
				for key := range value.Observed {
					path := key
					if !filepath.IsAbs(path) {
						path = filepath.Join(root, path)
					}
					if err := add(root, path, value.Observed); err != nil {
						t.Fatal(err)
					}
				}
			}
			seconds := time.Since(started).Seconds()
			load, err = os.ReadFile("/proc/loadavg")
			if err != nil {
				t.Fatal(err)
			}
			metadata["load_after"] = strings.TrimSpace(string(load))
			if expected == nil {
				expected = &value
			} else if !reflect.DeepEqual(expected.Observed, value.Observed) || !reflect.DeepEqual(expected.Uncertain, value.Uncertain) {
				t.Fatal("replay changed closure or uncertainty")
			}
			row := struct {
				Trial   int
				Kind    string
				Seconds float64
				Flags   map[string]string
			}{trial + 1, kind, seconds, metadata}
			encoded, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(string(encoded))
		}
	}
}
