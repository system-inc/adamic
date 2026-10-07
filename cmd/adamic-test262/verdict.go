package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Outcome is one test's result. pass is both sides succeeding with the same output, or, for a
// negative runtime test, both sides failing the way the frontmatter requires. fail is a
// disagreement, including a native failure where Node succeeded. refused is stage 0 or the checker
// declining the program. not-typescript is a checker refusal independently confirmed
// by stock tsc with the same diagnostic code. crashed is a signal, a sanitizer, a timeout, or the compiler panicking.
type outcomeKind string

const (
	outcomeNotTypescript outcomeKind = "not-typescript"
	outcomePass          outcomeKind = "pass"
	outcomeFail          outcomeKind = "fail"
	outcomeRefused       outcomeKind = "refused"
	outcomeCrashed       outcomeKind = "crashed"
	outcomeSkipped       outcomeKind = "skipped"
	outcomeUnrun         outcomeKind = "unrun"
)

// execution is one run's observable result, the same shape the corpus fixtures record.
type execution struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Exit     int    `json:"exit"`
	Signal   string `json:"signal,omitempty"`
	TimedOut bool   `json:"timedOut,omitempty"`
}

// verdictInput is everything decide needs, so a fixture can stand in for a real run.
type verdictInput struct {
	NegativePhase string    `json:"negativePhase,omitempty"`
	NegativeType  string    `json:"negativeType,omitempty"`
	Node          execution `json:"node"`
	Native        execution `json:"native"`
}

type verdict struct {
	Kind   outcomeKind
	Reason string
}

// decide compares Node and the native binary. A native failure where Node succeeded is a fail,
// never a pass: the meter is how often Adamic matches Node, and counting a native failure as a
// pass would hide every miscompile. A negative runtime expectation is not optional either. Node
// throwing the named constructor, and native also exiting non-zero with the same stdout, is a pass;
// treating that as an ordinary success (both must exit 0) marks a correct negative test as a fail.
func decide(input verdictInput) verdict {
	if crashedExecution(input.Native) || crashedExecution(input.Node) {
		return verdict{Kind: outcomeCrashed, Reason: crashReason(input)}
	}
	if input.NegativePhase == "runtime" {
		return decideNegative(input)
	}
	if input.Node.Exit == 0 && input.Native.Exit == 0 && input.Node.Stdout == input.Native.Stdout && input.Node.Stderr == input.Native.Stderr {
		return verdict{Kind: outcomePass}
	}
	return verdict{Kind: outcomeFail, Reason: disagreeReason(input)}
}

func decideNegative(input verdictInput) verdict {
	nodeThrew := input.Node.Exit != 0 && errorNamed(input.Node.Stderr, input.NegativeType)
	nativeFailed := input.Native.Exit != 0
	if nodeThrew && nativeFailed && input.Node.Stdout == input.Native.Stdout {
		return verdict{Kind: outcomePass}
	}
	reason := "negative " + input.NegativeType + " not observed"
	if input.Node.Exit == 0 {
		reason = "negative " + input.NegativeType + ": node exited 0"
	} else if !nodeThrew {
		reason = "negative " + input.NegativeType + ": node did not throw it"
	} else if input.Native.Exit == 0 {
		reason = "negative " + input.NegativeType + ": native exited 0"
	} else if input.Node.Stdout != input.Native.Stdout {
		reason = "negative " + input.NegativeType + ": stdout disagrees"
	}
	return verdict{Kind: outcomeFail, Reason: reason}
}

func errorNamed(stderr string, name string) bool {
	if name == "" {
		return false
	}
	return regexp.MustCompile(`(?:^|\b)` + regexp.QuoteMeta(name) + `\b`).MatchString(stderr)
}

func crashedExecution(run execution) bool {
	if run.TimedOut || run.Signal != "" {
		return true
	}
	text := run.Stderr
	return strings.Contains(text, "AddressSanitizer") || strings.Contains(text, "UndefinedBehaviorSanitizer") || strings.Contains(text, "runtime error:")
}

func crashReason(input verdictInput) string {
	switch {
	case input.Native.TimedOut:
		return "native timed out"
	case input.Node.TimedOut:
		return "node timed out"
	case input.Native.Signal != "":
		return "native " + input.Native.Signal
	case input.Node.Signal != "":
		return "node " + input.Node.Signal
	case crashedExecution(input.Native):
		return "native sanitizer"
	default:
		return "node sanitizer"
	}
}

func disagreeReason(input verdictInput) string {
	if input.Node.Exit == 0 && input.Native.Exit != 0 {
		return "native failed where node passed (node 0, native " + strconv.Itoa(input.Native.Exit) + ")"
	}
	if input.Node.Exit != input.Native.Exit {
		return "exit " + strconv.Itoa(input.Node.Exit) + " vs " + strconv.Itoa(input.Native.Exit)
	}
	if input.Node.Stdout != input.Native.Stdout {
		return "stdout disagrees"
	}
	return "stderr disagrees"
}
