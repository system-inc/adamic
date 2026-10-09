package tailwind

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The violation half. Each case is a class written twice, on one of the three surfaces the
// configuration names.
func TestNoDuplicateClassesReportsRepeats(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
		wantIds  []string
	}{
		{
			name:     "jsx attribute",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex items-center flex" />;`,
			wantIds:  []string{"duplicateClass"},
		},
		{
			name:     "class attribute",
			fileName: "Component.tsx",
			source:   `const element = <div class="px-4 py-2 px-4" />;`,
			wantIds:  []string{"duplicateClass"},
		},
		{
			name:     "expression container",
			fileName: "Component.tsx",
			source:   `const element = <div className={"flex flex"} />;`,
			wantIds:  []string{"duplicateClass"},
		},
		{
			// The surface an attribute-only rule cannot see. On the real tree these hold 2,067
			// literals.
			name:     "callee argument",
			fileName: "Component.tsx",
			source:   `const merged = cn('px-4 py-2 px-4');`,
			wantIds:  []string{"duplicateClass"},
		},
		{
			name:     "callee with several arguments",
			fileName: "Component.tsx",
			source:   `const merged = cn('flex flex', 'gap-2 gap-2');`,
			wantIds:  []string{"duplicateClass", "duplicateClass"},
		},
		{
			// The third surface, 125 literals on the real tree.
			name:     "variable by name",
			fileName: "Styles.ts",
			source:   `const className = 'rounded-md rounded-md';`,
			wantIds:  []string{"duplicateClass"},
		},
		{
			// Upstream's getESCalleeName names a member call by its last property, so `theme.cn` is `cn`.
			name:     "method named like a callee",
			fileName: "Component.tsx",
			source:   `const value = theme.cn('flex flex');`,
			wantIds:  []string{"duplicateClass"},
		},
		{
			name:     "plural variable name",
			fileName: "Styles.ts",
			source:   `const classNames = 'p-2 p-2';`,
			wantIds:  []string{"duplicateClass"},
		},
		{
			// Two distinct classes each repeated: two findings, not four and not one.
			name:     "two different duplicates",
			fileName: "Component.tsx",
			source:   `const element = <div className="a b c a b" />;`,
			wantIds:  []string{"duplicateClass", "duplicateClass"},
		},
		{
			// A class repeated three times is still one finding for that class.
			name:     "same class three times",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex flex flex" />;`,
			wantIds:  []string{"duplicateClass"},
		},
		{
			// Irregular whitespace must not hide the repeat; `no-unnecessary-whitespace` owns the
			// spacing itself.
			name:     "extra whitespace between repeats",
			fileName: "Component.tsx",
			source:   "const element = <div className=\"text-sm   text-sm\" />;",
			wantIds:  []string{"duplicateClass"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDuplicateClasses, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The clean half, which is the one the old corpus never had.
//
// A rule with only violation fixtures has been shown to detect and never to discriminate, and the
// defect it misses is the expensive one: a rule that fires on correct code gets disabled, and a
// disabled rule enforces nothing.
func TestNoDuplicateClassesStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "distinct classes",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex items-center gap-2" />;`,
		},
		{
			// The pair that collapses is a different rule's finding. This one must not fire on it.
			name:     "collapsible but not duplicate",
			fileName: "Component.tsx",
			source:   `const element = <div className="px-4 py-4" />;`,
		},
		{
			name:     "single class",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex" />;`,
		},
		{
			name:     "empty string",
			fileName: "Component.tsx",
			source:   `const element = <div className="" />;`,
		},
		{
			name:     "only whitespace",
			fileName: "Component.tsx",
			source:   `const element = <div className="   " />;`,
		},
		{
			// A near-miss: same prefix, different class.
			name:     "similar but distinct",
			fileName: "Component.tsx",
			source:   `const element = <div className="border-l border-r" />;`,
		},
		{
			// An attribute nobody said carries classes.
			name:     "unrelated attribute",
			fileName: "Component.tsx",
			source:   `const element = <div title="flex flex" />;`,
		},
		{
			// A function nobody said combines classes.
			name:     "unrelated callee",
			fileName: "Component.tsx",
			source:   `const value = someOtherFunction('flex flex');`,
		},
		{
			// A variable whose name does not claim to hold classes.
			name:     "unrelated variable",
			fileName: "Styles.ts",
			source:   `const description = 'flex flex';`,
		},
		{
			// A member call is named by its last property and pathed by the whole chain, and neither
			// `join` nor `cn.join` is a callee, though the chain starts with one.
			name:     "method on an object named like a callee",
			fileName: "Component.tsx",
			source:   `const value = cn.join('flex flex');`,
		},
		{
			// A dynamic class is `no-concatenated-classes`'s finding, and the visible fragments here
			// hold no repeat.
			name:     "template hole with distinct statics",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex ${dynamic} gap-2`} />;",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDuplicateClasses, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoDuplicateClassesFixIsCorrect checks what the fix would actually write.
//
// A fix is a rewrite of the user's source that lands unattended, so asserting only that a finding
// was reported leaves the dangerous half untested. The two properties that matter: the surviving
// classes are the distinct ones, and they keep the order the author wrote them in, because
// reordering here would collide with `enforce-consistent-class-order`.
//
// Each repeat goes with the first byte of the separator before it and nothing more. The rest of a
// long separator is `no-unnecessary-whitespace`'s, so on its own this rule leaves it, and the two
// close the gap together in one pass (TestTailwindFixersComposeInOnePass).
func TestNoDuplicateClassesFixIsCorrect(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "removes the later repeat",
			source: `const element = <div className="flex items-center flex" />;`,
			want:   `const element = <div className="flex items-center" />;`,
		},
		{
			name:   "keeps original order",
			source: `const element = <div className="a b c a b" />;`,
			want:   `const element = <div className="a b c" />;`,
		},
		{
			name:   "collapses three into one",
			source: `const element = <div className="flex flex flex" />;`,
			want:   `const element = <div className="flex" />;`,
		},
		{
			name:   "leaves the excess of a long separator to the whitespace rule",
			source: "const element = <div className=\"text-sm   text-sm\" />;",
			want:   "const element = <div className=\"text-sm  \" />;",
		},
		{
			name:   "a repeat on its own line takes its newline with it",
			source: "const element = <div className=\"flex\n  items-center\n  flex\" />;",
			want:   "const element = <div className=\"flex\n  items-center  \" />;",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.want)
		})
	}
}

// TestTemplateHoleIsReportedButNotFixed covers the case where the rule knows enough to complain and
// not enough to repair.
//
// The visible fragments genuinely repeat, so the finding is real. Rewriting from those fragments
// would silently delete the interpolation, which is a fix that changes what the code means.
func TestTemplateHoleIsReportedButNotFixed(t *testing.T) {
	t.Parallel()
	// A template literal is not read as a class literal at all today, so the guard is asserted
	// directly against the reporting path rather than through the parser.
	literal := ClassLiteral{
		Text:   "flex ${dynamic} flex",
		Origin: ClassLiteralOriginAttribute,
	}

	var diagnostics []rule.Diagnostic
	ctx := rule.Context{
		Report: func(diagnostic rule.Diagnostic) { diagnostics = append(diagnostics, diagnostic) },
	}
	reportDuplicates(ctx, literal)

	if len(diagnostics) != 1 {
		t.Fatalf("expected one finding for the repeated visible class, got %d", len(diagnostics))
	}
	if len(diagnostics[0].Fixes) != 0 {
		t.Fatal("a literal with a template hole must not be rewritten, because the fix would drop the interpolation")
	}
}

// TestAttributeOnlyReadingLosesFindings is the known-dirty control.
//
// The tempting shape for this rule is to listen to JSX attributes and stop, which is what the
// obvious reading of "class names live in className" produces. It measures correct on most of the
// tree: attributes are 7,773 of the 9,965 class literals in the real corpus. It is also blind to
// 2,192 of them.
//
// So a deliberately narrowed reader runs against the same sources and is required to lose findings
// that the real rule catches. A guard that has never returned a positive has not been shown to work.
func TestAttributeOnlyReadingLosesFindings(t *testing.T) {
	t.Parallel()
	sourcesOnlyNonAttributeSurfacesCatch := []string{
		`const merged = cn('px-4 py-2 px-4');`,
		`const className = 'rounded-md rounded-md';`,
	}

	attributeOnly := NewClassLiteralReader(ClassLiteralSettings{
		AttributePatterns: DefaultClassLiteralSettings().AttributePatterns,
		// Deliberately empty: this is the mistake being reproduced.
		CalleeNamePatterns: nil,
		VariablePatterns:   nil,
	})

	lost := 0
	for _, source := range sourcesOnlyNonAttributeSurfacesCatch {
		// The real rule must find it.
		result := rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", source)
		if len(result.Diagnostics) == 0 {
			t.Fatalf("the real rule missed %q, which this control depends on it catching", source)
		}

		// The narrowed reader must not, which is what makes it dangerous.
		narrowed := rule_testing.Run(t, rule.Rule{
			Name: "attribute-only-control",
			Run: func(ctx rule.Context, _ any) rule.Listeners {
				listeners := rule.Listeners{}
				for _, kind := range ListenerKinds() {
					listeners[kind] = func(node *ast.Node) {
						for _, literal := range attributeOnly.ClassLiteralsIn(node) {
							reportDuplicates(ctx, literal)
						}
					}
				}
				return listeners
			},
		}, "Component.tsx", source)

		if len(narrowed.Diagnostics) == 0 {
			lost++
		}
	}

	if lost == 0 {
		t.Fatal("the attribute-only reader lost nothing, so this corpus does not exercise the callee " +
			"and variable surfaces and the passing tests above prove less than they appear to")
	}
	if lost != len(sourcesOnlyNonAttributeSurfacesCatch) {
		t.Errorf("expected the attribute-only reader to lose all %d, it lost %d",
			len(sourcesOnlyNonAttributeSurfacesCatch), lost)
	}
}

// A repeat inside a template's run is read, and fixed the way Prettier's Tailwind plugin fixes it.
//
// Until this, no-duplicate-classes read only string literals, so a repeat in a template run was
// invisible, and the order rule, which leaves a run holding a repeat unordered, left it unordered
// in silence. Each fixed case is quoted from what the plugin wrote for the same input. A glued
// fragment such as `px-` is half a class the hole completes, and is never compared.
func TestNoDuplicateClassesReadsTemplateRuns(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, source, want string
	}{
		{
			name:   "a repeat before a hole",
			source: "const merged = cn(`flex flex ${size}`);",
			want:   "const merged = cn(`flex ${size}`);",
		},
		{
			name:   "a repeat in a run that ends glued to a hole",
			source: "const merged = cn(`items-center flex flex px-${size} block`);",
			want:   "const merged = cn(`items-center flex px-${size} block`);",
		},
		{
			name:   "a repeat after a hole",
			source: "const element = <div className={`${size} gap-2 block gap-2`} />;",
			want:   "const element = <div className={`${size} gap-2 block`} />;",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.want)
		})
	}
}

// A repeat split across a hole is reported and not fixed.
//
// Both classes apply whatever the hole holds, so it is a real repeat. The plugin dedupes each run on
// its own and leaves `flex ${size} flex` as written, so a fix would be the one rewrite of a class
// string the plugin never makes. The finding is what keeps it from sitting there silently.
func TestNoDuplicateClassesReportsARepeatAcrossAHoleWithoutAFix(t *testing.T) {
	t.Parallel()
	result := rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", "const merged = cn(`flex ${size} flex`);")
	rule_testing.ExpectFindings(t, result, "duplicateClass")
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("a repeat across a hole must not be fixed, got %+v", result.Diagnostics[0].Fixes)
	}
}

// The template runs that hold no repeat, including glued fragments that look like one.
func TestNoDuplicateClassesTemplateRunsStaySilent(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const merged = cn(`flex ${size} block`);",
		"const merged = cn(`px-${a} px-${b}`);",
		"const merged = cn(`flex px-${a} flex-${b} block`);",
	} {
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", source))
	}
}
