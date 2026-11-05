package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// runtimeNotYetStop is a run-time NotYet (#zhgxzjg, Adamic, Nov 2): an honest stop where a key or a
// receiver can't be known before the read. It's no type lie, so it never sits among the negative
// witnesses. Both backends stop with exit 70 and the entry's one allowlisted message after printing
// exactly what Node printed so far, and Node, running the source, goes on to the end.
type runtimeNotYetStop struct {
	Path        string `json:"path"`
	RulingTask  string `json:"ruling_task"`
	LiftingTask string `json:"lifting_task"`
	Stderr      string `json:"stderr"`
}
type runtimeNotYetStopList struct {
	Comment string                       `json:"_comment"`
	Stops   map[string]runtimeNotYetStop `json:"stops"`
}

// runtimeNotYetForms is the closed allowlist of run-time NotYet messages, each followed by
// " (reading 'key')", and the ruling that made each. The words are internal/ir's DynamicKeyedStops, the
// nine a ruled fixture reaches. The table's two compiler-bug stops never qualify, and nor does a form
// not here: a new one needs Adamic's ruling and its line here.
var runtimeNotYetForms = map[string]string{
	"a primitive receiver's properties aren't read yet":                     "#zhgxzjg",
	"a method isn't read by a run-time key yet":                             "#zhgxzjg",
	"a name beginning with # isn't read by a run-time key yet":              "#zhgxzjg",
	"__proto__ isn't read by a run-time key yet":                            "#0sps03r",
	"a class instance's constructor isn't read by a run-time key yet":       "#0sps03r",
	"this runtime value's properties aren't read by a run-time key yet":     "#0sps03r",
	"an Error property Node sets and the native Error lacks isn't read yet": "#zhgxzjg",
	"an array's methods aren't read by a run-time key yet":                  "#ys4naww",
	"an object literal's getter isn't read by a run-time key yet":           "#zhgxzjg",
}

// runtimeNotYetMessage is a stop's whole stderr: one line, its form, and the key it was reading.
var runtimeNotYetMessage = regexp.MustCompile(`^adamic: panic: (.+?) \(reading '([^\n]*)'\)\n$`)

// runtimeNotYetConversions is the second closed allowlist, of the keyless stops String and a template
// make of an unknown (lower/to_primitive.go, #09dvwny, Adamic, Nov 3): each what, said whole as
// "adamic: panic: converting <what> to a string is not yet lowered\n", and the ruling that made it. Only
// these exact words qualify, never a pattern over the shape: a fourth what needs Adamic's ruling and its
// line here.
var runtimeNotYetConversions = map[string]string{
	"an object whose own toString is a function this program can't name": "#09dvwny",
	"an array of only numbers, booleans, null or undefined":              "#09dvwny",
	"a class object": "#09dvwny",
}

// The keyless shape's two ends, around its what.
const (
	runtimeNotYetConversionPrefix = "adamic: panic: converting "
	runtimeNotYetConversionSuffix = " to a string is not yet lowered\n"
)

// runtimeNotYetKeyless is the third closed allowlist, of keyless stops that aren't a conversion: each
// message whole, said as "adamic: panic: <message>\n", and the ruling that made it. Object.keys of a
// runtime host value names no key it was reading and converts nothing, so it has neither the keyed shape
// nor the conversion sentence (#ey7aysy, Adamic, Nov 4, under #0sps03r). Only these exact words qualify,
// matched against the whole stderr: a new message needs Adamic's ruling and its line here.
var runtimeNotYetKeyless = map[string]string{
	"this runtime value's own keys aren't enumerated yet": "#0sps03r",
}

// runtimeNotYetKeylessPrefix is what every keyless message is said after.
const runtimeNotYetKeylessPrefix = "adamic: panic: "

// taskReference is a task named as the lists name one, #id.
var taskReference = regexp.MustCompile(`^#[0-9a-z]+$`)

// runtimeNotYetForm is the allowlisted form the stop message says, with the ruling that made it, or
// false: a keyed form of runtimeNotYetForms, a keyless conversion of runtimeNotYetConversions, or a
// whole keyless message of runtimeNotYetKeyless, the message whole in both keyless cases. A compiler bug
// said out loud never qualifies, even if its words were listed.
func runtimeNotYetForm(stderr string) (string, string, bool) {
	if strings.Contains(stderr, "compiler bug") {
		return "", "", false
	}
	if message := runtimeNotYetMessage.FindStringSubmatch(stderr); message != nil {
		ruling, allowed := runtimeNotYetForms[message[1]]
		return message[1], ruling, allowed
	}
	for what, ruling := range runtimeNotYetConversions {
		if stderr == runtimeNotYetConversionPrefix+what+runtimeNotYetConversionSuffix {
			return what, ruling, true
		}
	}
	for message, ruling := range runtimeNotYetKeyless {
		if stderr == runtimeNotYetKeylessPrefix+message+"\n" {
			return message, ruling, true
		}
	}
	return "", "", false
}
func readRuntimeNotYetStops(path string) (runtimeNotYetStopList, error) {
	var list runtimeNotYetStopList
	data, err := os.ReadFile(path)
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
		return list, fmt.Errorf("run-time NotYet list has trailing JSON")
	}
	if !strings.Contains(list.Comment, "Adding an entry is a ruling") || !strings.Contains(list.Comment, "@system_adamic") || list.Stops == nil {
		return list, fmt.Errorf("run-time NotYet list needs its ruling header and stops object")
	}
	for hash, stop := range list.Stops {
		if !validContentHash(hash) {
			return list, fmt.Errorf("invalid run-time NotYet content hash %q", hash)
		}
		if stop.Path == "" || filepath.IsAbs(stop.Path) || filepath.ToSlash(filepath.Clean(stop.Path)) != stop.Path || stop.Path == ".." || strings.HasPrefix(stop.Path, "../") {
			return list, fmt.Errorf("run-time NotYet needs its program's repository path: %s", hash)
		}
		if !taskReference.MatchString(stop.RulingTask) || !taskReference.MatchString(stop.LiftingTask) || stop.LiftingTask == stop.RulingTask {
			return list, fmt.Errorf("run-time NotYet needs its ruling task and the task that lifts it: %s", hash)
		}
		_, ruling, allowed := runtimeNotYetForm(stop.Stderr)
		if !allowed {
			return list, fmt.Errorf("run-time NotYet stop message is not a ruled form: %s", hash)
		}
		if stop.RulingTask != ruling {
			return list, fmt.Errorf("run-time NotYet ruling task %s is not its form's ruling %s: %s", stop.RulingTask, ruling, hash)
		}
	}
	return list, nil
}

// acceptsRuntimeNotYet says a run is the listed stop: Node runs the source to the end, exit 0, printing
// more than the backends did, and each backend prints exactly Node's first bytes, then stops with exit 70
// and the entry's whole message.
func acceptsRuntimeNotYet(stop runtimeNotYetStop, node, javascript, native observation) bool {
	if node.Error != "" || node.Exit != 0 {
		return false
	}
	for _, backend := range []observation{javascript, native} {
		if backend.Error != "" || backend.Exit != 70 || backend.Stderr != stop.Stderr {
			return false
		}
	}
	return javascript.Stdout == native.Stdout && strings.HasPrefix(node.Stdout, native.Stdout) && len(node.Stdout) > len(native.Stdout)
}

// runtimeNotYetRelated schedules a listed program, or an edit at a listed path, for a run, as
// negativeRelated does. The exemption still comes only from contentHash.
func runtimeNotYetRelated(path string, source []byte, list runtimeNotYetStopList) bool {
	if _, listed := list.Stops[contentHash(source)]; listed {
		return true
	}
	for _, stop := range list.Stops {
		if path == stop.Path {
			return true
		}
	}
	return false
}
