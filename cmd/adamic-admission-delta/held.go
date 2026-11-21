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
// while the backends' stderr is the row's, byte for byte. The row is itself the pin of the stop's words (Adamic on
// #d358hc1): the tests that hold these programs hold that they stop, and few hold what the stop says.
type heldStop struct {
	Path string `json:"path"`
	// Stderr writes the checkout the run judges as {root}, since a stop that names its program's file names it inside
	// the run's own scratch directory, which no row could match two days running.
	Stderr string `json:"stderr"`
	// Hatch names the hatch the program uses and the line it is written on, as <kind>[ (<detail>)], line <n>: <the
	// line's text>. A valid program's Node answer is correct and only a hatch may stop loudly (Adamic on #d358hc1), so
	// a row without one is refused, and a stop of a program with none is a disagreement.
	Hatch string `json:"hatch"`
	// Test names the test that holds the stop, as <repository path>_test.go: <TestName>, so a row can't be added
	// without one, and its guard (held_list_test.go) can find it.
	Test string `json:"test"`
}
type heldStopList struct {
	Comment string     `json:"_comment"`
	Stops   []heldStop `json:"stops"`
}

// heldStopTest is a row's test as the list names it: a Go test file by its repository path, in a package's directory
// and never at the root, then one test function in it.
var heldStopTest = regexp.MustCompile(`^((?:[A-Za-z0-9_.-]+/)+[A-Za-z0-9_.-]+_test\.go): (Test[A-Za-z0-9_]*)$`)

// heldStopHatch is a row's hatch as the list names it: its kind, an optional detail, then the line of the program the
// hatch is written on and that line's text, trimmed. A language hatch, the checker's own unsoundness that a ruling
// makes a checked stop, has no hatch in the source, so its detail cites that ruling as docs/rulings.md:<line>,
// "<the row's opening words>", and its line is the construct that stops.
var heldStopHatch = regexp.MustCompile(`^(cast|non-null|predicate|definite-assignment|relation-lie|index-write|index-read|computed-key|library-claim|language)(?: \(([^()\n]+)\))?, line ([1-9][0-9]*): ([^\n]*[^\s])$`)

// heldStopRuling is the ruling a language hatch's detail cites: a line of docs/rulings.md and the opening words of
// its row, since rows move as rulings land.
var heldStopRuling = regexp.MustCompile(`docs/rulings\.md:([1-9][0-9]*), "([^"\n]+)"`)

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
		if test := heldStopTest.FindStringSubmatch(stop.Test); test == nil || !repositoryPath(test[1]) {
			return list, fmt.Errorf("held stop needs the test that holds it, as <repository path>_test.go: <TestName>: %s", stop.Path)
		}
		hatch := heldStopHatch.FindStringSubmatch(stop.Hatch)
		if hatch == nil {
			return list, fmt.Errorf("held stop needs the hatch its program uses, as <kind>[ (<detail>)], line <n>: <text>: %s", stop.Path)
		}
		if hatch[1] == "language" && !heldStopRuling.MatchString(hatch[2]) {
			return list, fmt.Errorf("held stop's language hatch needs the ruling it stands on, as docs/rulings.md:<line>, \"<opening words>\": %s", stop.Path)
		}
	}
	return list, nil
}

// bothStop says the backends stopped alike where Node ran to the end: the JavaScript backend and native have the same
// exit, stdout and stderr, that exit is 70, and Node exits 0, none of the three with an execution error.
func bothStop(node, javascript, native observation) bool {
	return node.Error == "" && javascript.Error == "" && native.Error == "" && node.Exit == 0 && javascript.Exit == 70 && native.Exit == 70 && javascript.Stdout == native.Stdout && javascript.Stderr == native.Stderr
}

// holds says a run is a checked stop: a both-stop at a listed path where each backend's stderr is that row's exactly,
// once the row's {root} is the checkout the run judged, root. A listed path alone isn't enough, so a stop whose
// message changes, on either backend, disagrees until its row, and its test, are changed with it.
func (list heldStopList) holds(path, root string, node, javascript, native observation) bool {
	if !bothStop(node, javascript, native) {
		return false
	}
	for _, stop := range list.Stops {
		if stop.Path == path {
			stderr := strings.ReplaceAll(stop.Stderr, "{root}", root)
			return stderr == native.Stderr && stderr == javascript.Stderr
		}
	}
	return false
}
