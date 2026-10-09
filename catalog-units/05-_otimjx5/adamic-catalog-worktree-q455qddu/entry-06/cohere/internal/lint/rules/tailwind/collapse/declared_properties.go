// What a framework class declares, computed from its handle body rather than read from a table.
//
// This is the consumer-facing half of the emitter port. frameworkhandlers.go, frameworkmultihandlers.go
// and frameworkgaphandlers.go each hold one slice of the ported `handle` bodies; this file is the one
// entry point that dispatches across all three and answers the question `no-conflicting-classes`
// actually asks, which is not a reading but a list of CSS property names.
//
// # Why a reading was never the right source for this, and a declaration list is
//
// conflict_facts.go's own file comment worked through this and reached the right conclusion for the
// tools it had. `Table.Lookup` returns `Order`, indices into `PropertyOrder`, and inverting those to
// property names gets 1,355 of 1,530 corpus classes right and 75 wrong. The three ways it fails are
// all the same failure: a reading is a lossy projection of a declaration list.
//
//   - A `--tw-sort` override replaces the reading outright. `divide-y` reads `divide-y-width`, which
//     is not a CSS property at all, while the class emits four real ones.
//   - A property outside `PropertyOrder` contributes no position. `backdrop-blur-sm` declares
//     `-webkit-backdrop-filter` and `backdrop-filter`; only one is invertible.
//   - Two declarations can share a position, so a count cannot be recovered from an order.
//
// Every one of those is a fact the declaration list has and the reading has thrown away. So this file
// asks the emitters for the list itself and never goes near a reading, which is why it does not
// inherit any of the 75.
//
// # Custom properties are stripped here, and that is a different question from the ordering tables
//
// The ordering tables keep `--tw-*` custom properties because they are what separates `shadow-lg`
// from `ring-1` in the sort. This strips them, because two classes both setting `--tw-border-style`
// are not in conflict about anything an author sees. That is the same split the deleted conflict
// tables drew, kept rather than rediscovered: `OrderingPropertiesByRoot` and the conflict tables
// answered deliberately different questions from the same handle bodies, and now one emitter answers
// both by being asked differently.
//
// # The negation prefix
//
// Upstream registers `-m` as root `m` with `supportsNegative`, so the emitters are keyed on the
// undashed name. `descriptor.go`'s `frameworkReading` already strips the prefix for the same reason.
// Measured against the deleted table: of its 308 roots, 69 were negative spellings of a root the
// emitters answer, and stripping is what makes those 69 resolvable rather than absent.
package tailwind

import "strings"

// ValueResolution is what the caller learned by resolving a named value against the live theme.
//
// Two fields rather than a bool, because two different theme questions decide an arm and only one of
// them is about colour. Most of these roots split colour against not-colour, which `IsColor`
// answers. `font` splits on `--font` against `--font-weight`, neither of which is a colour and both
// of which are named values that infer as nothing, which is what `Namespace` carries.
//
// The zero value is a value that resolved through no theme namespace, which is the correct input for
// an arbitrary value and for a bare number.
type ValueResolution struct {
	// IsColor is whether the value resolved through a colour namespace.
	IsColor bool
	// Namespace is the theme namespace the value resolved through, such as `--font`. Empty when the
	// value resolved through none, or through one no emitter branches on.
	Namespace string
}

// DeclaredPropertiesFor returns the CSS property names a framework class declares.
//
// The candidate rather than a root, because that is the whole point of computing this: `border-4` is
// a width and `border-red-500` is a colour, and the two declare different properties from one root.
// A table keyed on the root can hold only one of those answers, which is why the deleted one needed
// thirteen per-class overrides and a fifteen-entry colour table beside it.
//
// `resolution` is passed in rather than decided here, matching UtilityBranchFor and for the same
// reason: upstream those questions are `resolveThemeColor` and `theme.resolve` consulting the
// repository's own `@theme`, and answering them in this file would put one repository's tokens in a
// table describing Tailwind.
//
// Returns false where the class declares nothing a conflict can be about: a root no emitter answers,
// or one whose every declaration is a custom property. Both are declines rather than empty answers,
// because `reportConflicts` skips a class it cannot resolve and an empty property list would instead
// pair with everything.
func DeclaredPropertiesFor(candidate *ParsedCandidate, resolution ValueResolution) ([]string, bool) {
	if candidate == nil {
		return nil, false
	}

	// A static is keyed on its whole name and has no value to branch on, so it is answered from the
	// compiled bodies upstream registers rather than through an emitter.
	if candidate.Kind == ParsedCandidateKindStatic {
		declarations, isStatic := FrameworkStaticDeclarations[candidate.Root]
		if !isStatic {
			return nil, false
		}
		properties := make([]string, 0, len(declarations))
		for _, declaration := range declarations {
			properties = appendVisibleProperty(properties, declaration.Property)
		}
		return properties, len(properties) > 0
	}

	if candidate.Kind != ParsedCandidateKindFunctional {
		return nil, false
	}

	nodes := emitFunctionalRoot(candidate, resolution)
	if len(nodes) == 0 {
		return nil, false
	}
	properties := collectVisibleProperties(nil, nodes)
	return properties, len(properties) > 0
}

// collectVisibleProperties walks a declaration list, descending into nested rules.
//
// Six framework roots wrap their declarations in a rule, and their properties are as real as any
// other: `space-x-4` declares `margin-inline-start`, it just declares it on the element's children.
// Walking only the top level would decline all six, and `PropertySort` descends for the same reason,
// so this matches it rather than inventing a second traversal.
//
// Whether those properties can conflict with another class's is `selectorShapeOf`'s question and is
// answered beside this one. Both are read off the same nodes.
func collectVisibleProperties(properties []string, nodes []*Node) []string {
	for _, node := range nodes {
		switch node.Kind {
		case KindRule, KindAtRule:
			properties = collectVisibleProperties(properties, node.Nodes)

		case KindDeclaration:
			// An absent value is not a declaration the browser sees, and PropertySort skips it for
			// the same reason. `font-mono` emits `font-feature-settings` with an undefined value
			// unless the theme entry defined one, and counting it would make `font-mono` conflict
			// with a class that really does set feature settings.
			if !node.ValuePresent {
				continue
			}
			properties = appendVisibleProperty(properties, node.Property)
		}
	}
	return properties
}

// emitFunctionalRoot dispatches one candidate to whichever slice ported its handle body.
//
// The three tables partition the framework's functional roots and do not overlap, which
// TestEmitterSlicesDoNotOverlap holds. The order here is therefore not a precedence, and it is
// written most-specific-first anyway so that a future overlap is a visible choice rather than a
// silent one.
func emitFunctionalRoot(candidate *ParsedCandidate, resolution ValueResolution) []*Node {
	root := strings.TrimPrefix(candidate.Root, "-")

	if _, isGapRoot := gapEmitters[root]; isGapRoot {
		return EmitGapRoot(root, UtilityBranchFor(candidate, gapTypeListFor(root), resolution.IsColor, resolution.Namespace))
	}

	// The two value-independent slices take a resolved value. They emit one list per root, so the
	// value is a sentinel: every property name they produce is fixed at registration, and passing a
	// real resolved value would change nothing they return.
	resolved := ResolvedUtilityValue{Value: declaredPropertiesSentinel}

	if utility, known := FrameworkFunctionalUtilities[root]; known {
		// A `staticValues` entry is answered before the handle body runs and carries its own
		// property, which can differ from the root's. Recognised here so that branch is reachable
		// rather than always taking the ordinary path.
		if candidate.Value != nil && candidate.Value.Kind == ParsedValueKindNamed {
			for _, static := range utility.StaticValues {
				if static.Name == candidate.Value.Value {
					resolved = ResolvedUtilityValue{IsStaticValue: true, StaticValueName: static.Name}
					break
				}
			}
		}
		return utility.Emit(root, resolved)
	}
	if utility, known := FrameworkMultiDeclarationUtilities[root]; known {
		// A root-defined keyword is answered by the handler before the theme is consulted, which is
		// how `transition-none` emits one declaration where `transition-*` emits three. That is the
		// branch `ClassDeclaredProperties` recorded for exactly one class.
		literal := ""
		if candidate.Value != nil && candidate.Value.Kind == ParsedValueKindNamed {
			literal = candidate.Value.Value
		}
		return utility.Emit(root, literal, resolved)
	}
	return nil
}

// declaredPropertiesSentinel is the value handed to the two value-independent slices.
//
// Their emitters carry it into every declaration they build and nothing here reads it back: this
// file collects property names. Named rather than spelled inline so it is visibly the same
// not-a-CSS-value the emitter tests use.
const declaredPropertiesSentinel = "zzsentinel"

// gapTypeListFor returns the ordered data types a gap root infers against.
//
// The order is per-root and load-bearing. `InferDataType` returns the first match, so `bg` listing
// `position` before `length` makes `bg-[3px]` a position rather than a length, and re-sorting the
// list silently changes what the class is said to declare. It is read off the checked-in descriptor
// row rather than restated here, so there is one copy of the ordering and it is the measured one.
//
// A root with no descriptor row yields nil, which makes `InferDataType` match nothing and the branch
// fall to its default arm. That is the correct reading of "this root does not discriminate on type".
func gapTypeListFor(root string) []DataType {
	descriptor, hasRow := baseDescriptors[root]
	if !hasRow {
		return nil
	}
	return descriptor.TypeList
}

// appendVisibleProperty adds a property name unless it is invisible to a conflict comparison.
//
// Two exclusions, and they are the ones the deleted conflict tables applied when they were
// generated. A custom property is not something two classes can visibly conflict about. An empty
// name is not a declaration at all; it appears when a node carries structure rather than a property.
//
// Deduplicated because a handle body can declare one property twice, as the shadow family's fallback
// arm does, and a conflict comparison asks which properties are shared rather than how many times.
func appendVisibleProperty(properties []string, property string) []string {
	if property == "" || strings.HasPrefix(property, "--") {
		return properties
	}
	for _, existing := range properties {
		if existing == property {
			return properties
		}
	}
	return append(properties, property)
}

// ComposesFor reports whether a root's utilities layer rather than overwrite each other.
//
// `shadow-lg` and `shadow-xl` do not conflict: each writes its own `--tw-shadow` and both write a
// `box-shadow` naming every shadow variable, so the two `box-shadow` declarations are byte identical
// and the utilities stack. `px-4` and `px-8` write their value straight into `padding-inline` and
// the second wins. That difference is what this answers.
//
// # How, and why it is two emissions rather than a table
//
// Emit the root twice with two different values and compare the declarations an author can see. A
// `--tw-*` custom property is excluded because it is the root's private contribution and is meant to
// differ: that difference is the mechanism, not a conflict. What decides is whether the shared
// property came out the same both times.
//
// This is upstream's own definition rather than a proxy for it. The generator that produced the
// deleted `ComposingRoots` compiled two values through the real engine and compared declaration
// text, and this is the same comparison with the same inputs.
//
// # The limit, and it is structural rather than a gap to fill later
//
// Only the two value-independent slices can answer this. The 57 value-partitioned roots in
// frameworkgaphandlers.go take a `UtilityBranch`, which carries the value's shape and never the
// value, so two values of a gap root emit identical text whether the root composes or not. Measured:
// 27 of 27 composing gap roots compare identical, and so do 25 of 30 non-composing ones, which is a
// measurement of nothing. Answering them needs the branch to carry a value, and that is a change to
// how the gap wave is called rather than a bigger table here.
//
// So the second return is whether the question was answerable, and a caller that gets false must
// keep whatever answer it had. `ok` is false for a gap root, for an unported root, and for a root
// whose every declaration is a custom property.
func ComposesFor(root string) (composes bool, ok bool) {
	stripped := strings.TrimPrefix(root, "-")

	first := visibleDeclarationText(emitRootWithValue(stripped, composesProbeFirst))
	second := visibleDeclarationText(emitRootWithValue(stripped, composesProbeSecond))
	if first == "" && second == "" {
		return false, false
	}
	return first == second, true
}

// The two probe values handed to the emitters, which are values no design system resolves to.
//
// Their content does not matter and their difference does. Named rather than spelled inline so it is
// visible that the comparison turns on two values being distinct rather than on what they are.
const (
	composesProbeFirst  = "zzprobeone"
	composesProbeSecond = "zzprobetwo"
)

// emitRootWithValue emits a root's declarations for one concrete value.
//
// Separate from emitFunctionalRoot because that one hands every emitter the same sentinel: it
// collects property names, where the value is noise. Composition is the one consumer that reads the
// value back, so it needs two emissions that actually differ.
func emitRootWithValue(root string, value string) []*Node {
	// A gap root branches on the value's shape, so it is asked through a branch carrying that value
	// rather than through a resolved value. The branch says an ordinary length was written, which is
	// the arm these roots take for the values a conflict comparison is about: `shadow-sm` against
	// `shadow-lg`, not `shadow-red-500`, which is a different arm and a different question.
	if _, isGapRoot := gapEmitters[root]; isGapRoot {
		return EmitGapRoot(root, UtilityBranch{HasValue: true, DataType: DataTypeLength, Value: value})
	}

	resolved := ResolvedUtilityValue{Value: value}
	if utility, known := FrameworkFunctionalUtilities[root]; known {
		return utility.Emit(root, resolved)
	}
	if utility, known := FrameworkMultiDeclarationUtilities[root]; known {
		return utility.Emit(root, "", resolved)
	}
	return nil
}

// visibleDeclarationText renders the declarations two classes could visibly conflict about.
//
// Custom properties are excluded for the same reason appendVisibleProperty excludes them, and an
// absent value is skipped for the same reason PropertySort skips it: it is not a declaration the
// browser sees.
func visibleDeclarationText(nodes []*Node) string {
	var builder strings.Builder
	for _, node := range nodes {
		if node.Kind != KindDeclaration || !node.ValuePresent {
			continue
		}
		if strings.HasPrefix(node.Property, "--") {
			continue
		}
		builder.WriteString(node.Property)
		builder.WriteString(":")
		builder.WriteString(node.Value)
		builder.WriteString(";")
	}
	return builder.String()
}

// SelectorShapeOfNodes returns the selector a root's declarations land under.
//
// `.CLASS` when they land on the element itself, which is every root but six. The others wrap their
// declarations in a nested rule upstream, and this reads that wrapper off the emitted nodes rather
// than off a table keyed by root name.
//
// # Why this replaces a table rather than restating it
//
// `RootSelectorShapes` held eight entries generated by compiling each root and reading the selector
// back. That is a fact about the handle bodies, and the handle bodies are ported, so the fact is
// available here. The table was measured invariant across an independent design system and its own
// comment said the measurement was necessary and insufficient, because a repository `@utility` block
// can declare a nested selector no corpus happens to write. Reading the wrapper removes the question:
// there is nothing keyed on a name to be missing an entry.
//
// The substitution to `.CLASS` happens here. Upstream writes `&`, which is what the emitters carry,
// and every consumer of this wants the shape spelled the way the generator spelled it.
func SelectorShapeOfNodes(nodes []*Node) string {
	for _, node := range nodes {
		if node.Kind != KindRule {
			continue
		}
		return strings.ReplaceAll(node.Selector, "&", ".CLASS")
	}
	return ".CLASS"
}

// SelectorShapeForRoot returns the selector shape a framework root's utilities emit under.
//
// The root rather than a candidate, because a wrapper is a property of the handle body and none of
// the six branches on a value. Returns `.CLASS` for a root no emitter answers, which is the same
// answer an unwrapped root gives and the correct one for a repository `@utility` block, since
// `@utility` takes a declaration list rather than a selector.
func SelectorShapeForRoot(root string) string {
	return SelectorShapeOfNodes(emitRootWithValue(strings.TrimPrefix(root, "-"), declaredPropertiesSentinel))
}
