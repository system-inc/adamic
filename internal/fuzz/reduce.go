package fuzz

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

// Signature is how a program fails, as precisely as reducing it has to keep: the kind of failure and
// the text that names it.
//
//   - panic: native or the JavaScript backend panicked with a line containing Text, where Node
//     didn't.
//   - mismatch: the first output Node and an Adamic backend differ on, as the oracle compares them
//     (stdout line by line, then the exit code, then stderr), rendered as a line containing Text.
//   - refusal: the compiler's NotYet or Refused diagnostic contains Text, its location left out.
//   - crash: the compiler itself panicked, and its first panic line contains Text.
//
// An exact signature, the one derived from a program, matches only its whole line.
type Signature struct {
	Kind  string
	Text  string
	Exact bool
}

// SignatureKinds are the kinds a signature can have.
var SignatureKinds = []string{"panic", "mismatch", "refusal", "crash"}

// ParseSignature reads kind:text.
func ParseSignature(written string) (Signature, error) {
	kind, text, found := strings.Cut(written, ":")
	if !found || text == "" {
		return Signature{}, fmt.Errorf("a signature is <kind>:<text>, with kind one of %s", strings.Join(SignatureKinds, ", "))
	}
	for _, known := range SignatureKinds {
		if kind == known {
			return Signature{Kind: kind, Text: text}, nil
		}
	}
	return Signature{}, fmt.Errorf("no signature kind %q; the kinds are %s", kind, strings.Join(SignatureKinds, ", "))
}

func (s Signature) String() string {
	return s.Kind + ":" + s.Text
}

// matches says whether an observed line is this signature's.
func (s Signature) matches(line string) bool {
	if line == "" {
		return false
	}
	if s.Exact {
		return line == s.Text
	}
	return strings.Contains(line, s.Text)
}

// Observation is what one program came to, read for every kind of signature at once.
type Observation struct {
	Verdict Verdict
	// Checker is the checker's diagnostics by code, sorted: a candidate has to check as the original
	// did.
	Checker []string
	// Lines is each kind's line, "" where the program didn't fail that way.
	Lines map[string]string
}

// Has says whether the observation shows signature, with the verdict it was first seen with.
func (o Observation) Has(signature Signature) bool {
	return signature.matches(o.Lines[signature.Kind])
}

// Observe runs a program in directory as the kind of signature needs: the compiler alone for a
// refusal or a crash, all three ways for a panic or a mismatch. kind "" runs it all three ways.
func (c *Checkout) Observe(source string, name string, directory string, kind string) Observation {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return Observation{Verdict: Finding, Lines: map[string]string{}}
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		return Observation{Verdict: Finding, Lines: map[string]string{}}
	}
	lowered := execute(directory, nil, 30*time.Second, c.adamic, "c", path)
	observation := Observation{Lines: map[string]string{}, Checker: checkerCodes(string(lowered.Stderr))}
	if lowered.ExitCode != 0 || lowered.TimedOut {
		observation.Verdict = compilerRefusal(lowered).Verdict
		observation.Lines["crash"] = crashLine(string(lowered.Stderr))
		observation.Lines["refusal"] = refusalLine(string(lowered.Stderr))
		return observation
	}
	if kind == "refusal" || kind == "crash" {
		observation.Verdict = Agreed
		return observation
	}
	outcome := c.TryFile(path, directory)
	observation.Verdict = outcome.Verdict
	observation.Lines["panic"] = panicLine(outcome)
	observation.Lines["mismatch"] = mismatchLine(outcome)
	return observation
}

var checkerCode = regexp.MustCompile(`error (TS[0-9]+)`)

func checkerCodes(stderr string) []string {
	var codes []string
	for _, match := range checkerCode.FindAllStringSubmatch(stderr, -1) {
		codes = append(codes, match[1])
	}
	sort.Strings(codes)
	return codes
}

// crashLine is the compiler's first panic line, the Go runtime's "panic: ...".
func crashLine(stderr string) string {
	for line := range strings.SplitSeq(stderr, "\n") {
		if strings.HasPrefix(line, "panic: ") {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// diagnosticLocation is where a diagnostic says it is, which reducing moves.
var diagnosticLocation = regexp.MustCompile(`^(?:adamic: )?[^ ]+:[0-9]+:[0-9]+: `)

// refusalLine is a NotYet or Refused diagnostic without its location.
func refusalLine(stderr string) string {
	for line := range strings.SplitSeq(stderr, "\n") {
		if strings.Contains(line, "Adamic 0.1 refuses") || (strings.Contains(line, "can't lower") && strings.Contains(line, "yet")) {
			return diagnosticLocation.ReplaceAllString(line, "")
		}
	}
	return ""
}

// panicLine is the panic line native wrote, or the JavaScript backend's, where Node's run differs.
func panicLine(outcome Outcome) string {
	if outcome.Verdict != Finding {
		return ""
	}
	for _, run := range []Run{outcome.Native, outcome.Backend} {
		for line := range strings.SplitSeq(string(run.Stderr), "\n") {
			if strings.HasPrefix(line, "adamic: panic: ") {
				return line
			}
		}
	}
	return ""
}

// mismatchLine is the first thing an Adamic backend's run differs from Node's on, without where in
// the output it was: native's if it differs, otherwise the JavaScript backend's.
func mismatchLine(outcome Outcome) string {
	if outcome.Verdict != Finding {
		return ""
	}
	for _, way := range []struct {
		name string
		run  Run
	}{{"native", outcome.Native}, {"javascript backend", outcome.Backend}} {
		if line := firstDifference(outcome.Node, way.run, way.name); line != "" {
			return line
		}
	}
	return ""
}

func firstDifference(node Run, other Run, name string) string {
	expected, actual := strings.Split(string(node.Stdout), "\n"), strings.Split(string(other.Stdout), "\n")
	for index := range max(len(expected), len(actual)) {
		var want, got string
		if index < len(expected) {
			want = expected[index]
		}
		if index < len(actual) {
			got = actual[index]
		}
		if index >= len(expected) || index >= len(actual) || want != got {
			return fmt.Sprintf("%s stdout: node %q, %s %q", name, want, name, got)
		}
	}
	switch {
	case other.TimedOut:
		return name + " never finished"
	case node.ExitCode != other.ExitCode:
		return fmt.Sprintf("%s exit: node %d, %s %d", name, node.ExitCode, name, other.ExitCode)
	case string(node.Stderr) != string(other.Stderr):
		return fmt.Sprintf("%s stderr: node %q, %s %q", name, firstLines(string(node.Stderr), 1), name, firstLines(string(other.Stderr), 1))
	}
	return ""
}

// Derive is the signature of a program's first failure, read from how it was observed.
func Derive(observation Observation) (Signature, error) {
	switch {
	case observation.Lines["crash"] != "":
		return Signature{Kind: "crash", Text: observation.Lines["crash"], Exact: true}, nil
	case observation.Lines["refusal"] != "":
		return Signature{Kind: "refusal", Text: observation.Lines["refusal"], Exact: true}, nil
	case len(observation.Checker) > 0:
		return Signature{}, fmt.Errorf("the checker refuses it (%s): it isn't Adamic, so there is nothing to reduce", strings.Join(observation.Checker, ", "))
	case observation.Lines["panic"] != "":
		return Signature{Kind: "panic", Text: observation.Lines["panic"], Exact: true}, nil
	case observation.Lines["mismatch"] != "":
		return Signature{Kind: "mismatch", Text: observation.Lines["mismatch"], Exact: true}, nil
	case observation.Verdict == Finding:
		return Signature{}, errors.New("it fails in a way no signature names (a sanitizer report, a leak, or C clang refused)")
	}
	return Signature{}, fmt.Errorf("it doesn't fail: %s", observation.Verdict)
}

// Reduction is what reducing a program came to.
type Reduction struct {
	Source    string
	Signature Signature
	Tried     int
}

// Reduce makes a program as small as it can while it still fails with signature, checks as it did,
// and comes to the same verdict: a candidate that fails some other way, or no longer fails, is never
// kept, so the reduction can't drift to a different failure. observe runs a candidate; slot names
// which of the workers runs it, so each can have a directory of its own.
func Reduce(path string, source string, signature Signature, original Observation, observe func(source string, slot int) Observation, workers int, budget int) Reduction {
	slots := make(chan int, workers)
	for slot := range workers {
		slots <- slot
	}
	var tried atomic.Int64
	keep := func(candidate Reducible) bool {
		tree := candidate.(*sourceTree)
		if !tree.parse().clean {
			return false
		}
		slot := <-slots
		defer func() { slots <- slot }()
		observation := observe(tree.Source(), slot)
		tried.Add(1)
		return keeps(signature, original, observation)
	}
	reduced := reduce(newSourceTree(path, source), keep, workers, budget)
	// What's left of the lines the removed code and comments stood on, held to the same test.
	result := reduced.Source()
	if tidy := withoutBlankLines(result); tidy != result && keep(newSourceTree(path, tidy)) {
		result = tidy
	}
	return Reduction{Source: result, Signature: signature, Tried: int(tried.Load())}
}

// withoutBlankLines is a text without its lines of nothing but spaces.
func withoutBlankLines(text string) string {
	var kept []string
	for line := range strings.SplitSeq(text, "\n") {
		if strings.TrimSpace(line) != "" {
			kept = append(kept, line)
		}
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.Join(kept, "\n") + "\n"
}

// keeps is the test every candidate passes to be kept: the same verdict, the same checker
// diagnostics, and the signature.
func keeps(signature Signature, original Observation, candidate Observation) bool {
	return candidate.Verdict == original.Verdict && strings.Join(candidate.Checker, ",") == strings.Join(original.Checker, ",") && candidate.Has(signature)
}
