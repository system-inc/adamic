package program

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/suppression"
)

// RuleAccount is what one rule did on one file.
//
// Declining is recorded as deliberately as reporting, because "the rule looked and had nothing to
// say" and "the rule never looked" are different facts that produce identical output everywhere
// else. That ambiguity kept four rules dead for months in the gate this tool replaces.
type RuleAccount struct {
	Name string

	// Listened is whether the rule accepted the file. A rule that declines returns no listeners,
	// which is the cheapest thing it can do and the thing most worth being able to see.
	Listened bool

	// DeclinedBecause names why a rule did not run when the configuration made that choice rather
	// than the rule. Empty when the rule itself declined, which is a different fact.
	DeclinedBecause string

	// NodeKinds is what the rule asked to hear about, sorted so two runs read the same.
	NodeKinds []string

	// NodesOffered is how many nodes the walk called it for.
	NodesOffered int

	// Duration is setup plus listeners for this one file.
	Duration time.Duration

	// Findings are what it reported and what a directive withheld, kept apart so a suppressed
	// finding is visible as suppressed rather than absent.
	Findings   []Finding
	Suppressed []Finding
}

// Finding is one diagnostic reduced to what a person reading an explanation needs.
type Finding struct {
	Line      int
	Column    int
	MessageId string
	Message   string
}

// Explanation is the full account of one file: every rule, what it did, and what it cost.
//
// The debugging counterpart to --timing. Timing says a rule is expensive across the tree; this says
// what it did on the file where it was expensive.
type Explanation struct {
	FileName string

	// FileIgnored is set when no rule ran because an ignorePattern excluded the file, with the
	// pattern responsible. A file skipped by configuration and a file with no findings are
	// otherwise identical output.
	FileIgnored bool
	IgnoredBy   string

	TotalDuration time.Duration

	// Accounts is every rule offered this file, listened or not, in registry order.
	Accounts []RuleAccount
}

// Explain runs every rule against one file and records what each did.
//
// A separate path from Walk rather than a flag on it. Walk is the hot path across thousands of
// files and must stay free of per-rule bookkeeping; this runs once, on one file, where the overhead
// is irrelevant and completeness matters more than speed.
func (g *Graph) Explain(ctx context.Context, sourceFile *ast.SourceFile, rules []rule.Rule) (*Explanation, error) {
	if sourceFile == nil {
		return nil, fmt.Errorf("no file to explain")
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("no rules to explain: none were given")
	}

	explanation := &Explanation{FileName: sourceFile.FileName().AsString()}

	var resolution configuration.Resolved
	if g.LintConfig != nil {
		resolution = g.LintConfig.Resolve(sourceFile.FileName().AsString())
		if resolution.Ignored {
			explanation.FileIgnored = true
			explanation.IgnoredBy = resolution.IgnoredBy
			return explanation, nil
		}
	}

	directives := suppression.Build(sourceFile.Text())

	fileChecker, release := g.CheckerForFile(ctx, sourceFile)
	defer release()

	fileCache := rule.NewFileCache()

	for _, subject := range rules {
		account := RuleAccount{Name: subject.Name}

		if g.LintConfig != nil {
			status, _ := resolution.StatusOf(subject.Name)
			if reason := declineReason(status); reason != "" {
				account.DeclinedBecause = reason
				explanation.Accounts = append(explanation.Accounts, account)
				continue
			}
		}

		options, err := g.RuleOptions.Decode(subject.Name, resolution.RawOptionsFor(subject.Name))
		if err != nil {
			account.DeclinedBecause = fmt.Sprintf("its options could not be read: %v", err)
			explanation.Accounts = append(explanation.Accounts, account)
			continue
		}

		// Findings are captured per rule rather than merged, which is the whole point: the reader
		// wants to know which rule said what, not the union.
		captured := &account
		report := func(diagnostic rule.Diagnostic) {
			// Byte offset rather than UTF-16 character: this is for a person reading a terminal and
			// jumping to a line, and the byte form is what the rest of the tool prints.
			line, column := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, diagnostic.Range.Pos())
			finding := Finding{
				Line:      line + 1,
				Column:    column + 1,
				MessageId: diagnostic.Message.Id,
				Message:   diagnostic.Message.Description,
			}
			if directives.Suppresses(subject.Name, diagnostic.Range.Pos()) {
				captured.Suppressed = append(captured.Suppressed, finding)
				return
			}
			captured.Findings = append(captured.Findings, finding)
		}

		start := time.Now()
		listeners := subject.Run(rule.Context{
			SourceFile:  sourceFile,
			Program:     rule.ViewProgram(g.Program, sourceFile, subject),
			TypeChecker: fileChecker,
			FileCache:   fileCache,
			Report:      report,
		}, options)

		if len(listeners) == 0 {
			// The rule looked at the file and declined it. Recorded rather than skipped, because a
			// rule that declines everything is otherwise indistinguishable from one finding nothing.
			account.Duration = time.Since(start)
			explanation.Accounts = append(explanation.Accounts, account)
			continue
		}

		account.Listened = true
		account.NodeKinds = kindNames(listeners)

		offered := 0
		merged := make([][]func(node *ast.Node), ast.KindCount)
		for kind, listener := range listeners {
			counting := listener
			merged[kind] = append(merged[kind], func(node *ast.Node) {
				offered++
				counting(node)
			})
		}

		walk(sourceFile.AsNode(), merged)
		account.NodesOffered = offered
		account.Duration = time.Since(start)

		explanation.Accounts = append(explanation.Accounts, account)
		explanation.TotalDuration += account.Duration
	}

	return explanation, nil
}

// declineReason turns a configuration status into the sentence a reader needs.
//
// Each phrasing says who decided, because "somebody turned this off here" and "nobody has said
// whether this should run" are different facts. Collapsing them describes a brand-new rule as
// though it had been deliberately excluded.
func declineReason(status configuration.Status) string {
	switch status {
	case configuration.StatusScopedOff:
		return "the config turns it off for this file"
	case configuration.StatusUnconfigured:
		return "it is not in the config, so nobody has said whether it should run"
	}
	return ""
}

// kindNames lists the node kinds a rule registered for, sorted for a stable reading order.
func kindNames(listeners rule.Listeners) []string {
	names := make([]string, 0, len(listeners))
	for kind := range listeners {
		names = append(names, kind.String())
	}
	sort.Strings(names)
	return names
}
