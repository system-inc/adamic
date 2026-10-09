package tailwind

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// sharedReaderRules are the rules that read class surfaces through ClassLiteralSurfaces.ReaderFor.
var sharedReaderRules = []rule.Rule{
	EnforceCanonicalClasses, EnforceConsistentClassOrder, EnforceConsistentImportantPosition,
	EnforceConsistentVariableSyntax, EnforceConsistentVariantOrder, EnforceShorthandClasses,
	NoConcatenatedClasses, NoConflictingClasses, NoDeprecatedClasses, NoDuplicateClasses,
	NoUnknownClasses, NoUnnecessaryWhitespace,
}

// sharedReaderWalkSource plants a finding for every one of those rules, on all three surfaces and in
// the value positions the reader enters, so each rule reads nodes another rule already read.
const sharedReaderWalkSource = `declare function cn(...values: unknown[]): string;
declare const size: string;
declare const open: boolean;
export const className = open ? 'items-center flex' : 'flex-shrink-0';
export const merged = cn('px-4 py-2 px-4', ` + "`px-${size}`" + `, open && 'ps-4 pe-4');
export const elements = [
	<div className="px-4 py-4" />,
	<div className="flex block" />,
	<div className="flx items-center" />,
	<div className="text-[var(--brand)]" />,
	<div className="hover:sm:flex" />,
	<div className="!text-red-500" />,
	<div className="flex  items-center" />,
	<div className={` + "`flex  ${size}  gap-2 gap-2`" + `} />,
];
`

// TestSharedReaderFindsWhatEachRuleFindsAlone runs the twelve rules through the real walk, where they
// share one file cache and so one reader, and requires exactly the findings each rule makes when it
// runs alone with a cache of its own. A reading that one rule spoiled for the next, or a memo that
// answered one surface with another's, shows up here as a finding gained or lost.
//
// Every rule must report at least once, so the comparison cannot pass by everyone declining: the
// design-system rules decline without a stylesheet, so the fixture links the installed package in as
// a repository has it.
func TestSharedReaderFindsWhatEachRuleFindsAlone(t *testing.T) {
	t.Parallel()
	packageRoot := classOrderFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so there is no design system to ask")
	}

	directory := t.TempDir()
	for name, contents := range map[string]string{
		"Component.tsx":                 sharedReaderWalkSource,
		classOrderFixtureStylesheetPath: classOrderFixtureStylesheet,
		"tsconfig.json":                 `{"compilerOptions": {"jsx": "preserve", "strict": true, "noEmit": true}, "include": ["*.tsx"]}`,
	} {
		path := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating the directory for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	if err := os.MkdirAll(filepath.Join(directory, "node_modules"), 0o755); err != nil {
		t.Fatalf("creating node_modules: %v", err)
	}
	if err := os.Symlink(packageRoot, filepath.Join(directory, "node_modules", "tailwindcss")); err != nil {
		t.Fatalf("linking the installed tailwindcss: %v", err)
	}

	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building the program: %v", err)
	}
	files := graph.ProjectFiles()

	findings := func(rules []rule.Rule) []string {
		result, err := graph.Walk(context.Background(), files, rules)
		if err != nil {
			t.Fatalf("walking: %v", err)
		}
		var lines []string
		for _, diagnostic := range result.Diagnostics {
			lines = append(lines, fmt.Sprintf("%s %d-%d %s %q %v", diagnostic.RuleName, diagnostic.Range.Pos(),
				diagnostic.Range.End(), diagnostic.Message.Id, diagnostic.Message.Description, diagnostic.Fixes))
		}
		slices.Sort(lines)
		return lines
	}

	shared := findings(sharedReaderRules)
	var alone []string
	for _, subject := range sharedReaderRules {
		alone = append(alone, findings([]rule.Rule{subject})...)
	}
	slices.Sort(alone)

	for _, subject := range sharedReaderRules {
		reported := slices.ContainsFunc(alone, func(line string) bool { return strings.HasPrefix(line, subject.Name+" ") })
		if !reported {
			t.Errorf("%s reported nothing on the fixture, so the comparison cannot speak for it", subject.Name)
		}
	}
	if !slices.Equal(shared, alone) {
		t.Fatalf("the rules sharing a reader found something other than what each finds alone:\n  shared:\n    %s\n  alone:\n    %s",
			strings.Join(shared, "\n    "), strings.Join(alone, "\n    "))
	}
}
