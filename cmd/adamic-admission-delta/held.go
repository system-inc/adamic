package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// heldStop is one row of cloud/admission-corpus/held-stops.json: a program whose both-stop a test already holds, so
// the agreement mode counts it as a checked stop rather than a disagreement, but only at that exact path and only
// while the backends' stderr is the row's, byte for byte.
type heldStop struct {
	Path   string `json:"path"`
	Stderr string `json:"stderr"`
	// Test names the test that holds the stop, as <file>_test.go: <TestName>, so a row can't be added without one.
	Test string `json:"test"`
}
type heldStopList struct {
	Comment string     `json:"_comment"`
	Stops   []heldStop `json:"stops"`
}

// heldStopTest is a row's test as the list names it: a Go test file, then one test function in it.
var heldStopTest = regexp.MustCompile(`^[A-Za-z0-9_]+_test\.go: Test[A-Za-z0-9_]*$`)

// readHeldStops reads the held-stop list. A missing file is an empty list, so a checkout from before the list existed
// judges every both-stop as a disagreement. Each row names its path once, its exact stderr and the test that holds
// it: the tool refuses the list otherwise.
func readHeldStops(path string) (heldStopList, error) {
	list := heldStopList{Stops: []heldStop{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return list, nil
	}
	if err != nil {
		return list, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&list); err != nil {
		return list, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return list, fmt.Errorf("held-stop list has trailing JSON")
	}
	if !strings.Contains(list.Comment, "Adding a row is a ruling") || list.Stops == nil {
		return list, fmt.Errorf("held-stop list needs its ruling header and stops array")
	}
	paths := map[string]bool{}
	for index, stop := range list.Stops {
		if !repositoryPath(stop.Path) {
			return list, fmt.Errorf("held stop row %d needs its program's repository path", index+1)
		}
		if paths[stop.Path] {
			return list, fmt.Errorf("held stop listed twice: %s", stop.Path)
		}
		paths[stop.Path] = true
		if stop.Stderr == "" {
			return list, fmt.Errorf("held stop needs the backends' exact stderr: %s", stop.Path)
		}
		// holds is judged before the memory-unsafe test, so a row whose stderr is a sanitizer's report would excuse a
		// memory error as a checked stop. A held stop is a hatch's own message, never one of those (Compiler on #ceffqvn).
		for _, report := range append([]string{"LeakSanitizer", "runtime error:", "DEADLYSIGNAL"}, sanitizerNames...) {
			if strings.Contains(stop.Stderr, report) {
				return list, fmt.Errorf("held stop's stderr is a sanitizer's report or a signal, which no row may hold: %s", stop.Path)
			}
		}
		if !heldStopTest.MatchString(stop.Test) {
			return list, fmt.Errorf("held stop needs the test that holds it, as <file>_test.go: <TestName>: %s", stop.Path)
		}
	}
	return list, nil
}

// bothStop says the backends stopped alike where Node ran to the end: the JavaScript backend and native have the same
// exit, stdout and stderr, that exit is 70, and Node exits 0, none of the three with an execution error.
func bothStop(node, javascript, native observation) bool {
	return node.Error == "" && javascript.Error == "" && native.Error == "" && node.Exit == 0 && javascript.Exit == 70 && native.Exit == 70 && javascript.Stdout == native.Stdout && javascript.Stderr == native.Stderr
}

// holds says a run is a checked stop: a both-stop at a listed path where each backend's stderr is that row's exactly.
// A listed path alone isn't enough, so a stop whose message changes, on either backend, disagrees until its row, and
// its test, are changed with it.
func (list heldStopList) holds(path string, node, javascript, native observation) bool {
	if !bothStop(node, javascript, native) {
		return false
	}
	for _, stop := range list.Stops {
		if stop.Path == path {
			return stop.Stderr == native.Stderr && stop.Stderr == javascript.Stderr
		}
	}
	return false
}
