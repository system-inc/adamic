// The live descriptor table: a Table built from the repository in front of us, outside tests.
//
// `Table.Lookup` answers the `{order, count}` half of the class-order sort, and until this file
// existed nothing outside a test could build the thing that answers it. The only two constructions
// in the tree assembled one from `testdata/descriptor_table.json` plus `descriptor_context.json`,
// both generated per repository and both deliberately untracked, so the shipped rules had a lookup
// and no table to call it on.
//
// # Why the table is not simply checked in
//
// descriptor.go's own comment on Table says it: `Namespaces` and `KeysByNamespace` are keyed on this
// repository's `@theme`, so a committed table would bake one repository's tokens into a file
// claiming to describe Tailwind, which is the bug this port exists to fix. That constraint is real
// and this file does not work around it. It splits the table along a seam that was measured rather
// than assumed, and composes the two halves at run time.
//
// # The seam, and the measurement that put it exactly here
//
// The observation already on the record is that generating against ~/Projects/ahra and
// www-connected-app produces byte-identical `roots` arrays. That is true and it is *not* evidence of
// framework invariance: those two repositories share the Structure submodule, so they share its
// `@theme` and its `@utility` blocks, and the agreement is a shared dependency rather than a
// property of Tailwind. Taking it as invariance would have shipped Structure's tokens as framework
// facts, which is the same class of bug in a new place.
//
// So a third design system was generated: a synthetic one importing `tailwindcss` and declaring only
// `--color-weird-*` and `--frobnicate-*`, sharing nothing with either repository. Against it, 258 of
// 301 shared roots differ. Every one of those 258 differs **only** in its three namespace-keyed
// maps; with those removed all three systems agree on every shared root, 0 differences. That is the
// seam, and it is why `baseDescriptors` can be checked in while the namespace axis cannot.
//
// The namespace axis then splits again, and this is what makes composing it live exact rather than
// approximate. Over 1,425 namespace buckets across the three systems, every bucket is one of exactly
// two things:
//
//   - **399 consumed by the root**, keyed on one of the 14 namespaces in FrameworkNamespaces, and
//     reading identically in all three systems. Printed into descriptor_base_table.go.
//   - **1,026 the root does not consume**, every one reading *exactly* its axis's `@none` bucket.
//     A namespace a root ignores produces no reading of its own: the value falls through the
//     bare-value precedence to the `@none` step.
//
// Zero buckets fell outside those two cases, and the generator re-runs that check on every
// invocation and refuses to print a table if any does. The consequence is the property this file
// needs: a repository can invent `--color-weird` and `border-weird-1` still reads correctly, because
// `border` does not consume `--color-weird`, so the bucket equals `@none`, and `@none` is already in
// the base table. Nothing has to know the token exists.
//
// # What is composed live, and why each piece has to be
//
//	Namespaces, KeysByNamespace   the repository's own `@theme`. The per-repository half by
//	                              definition, and the reason a table cannot be committed.
//	Statics                       the framework's, plus every static `@utility` the repository
//	                              declared. ahra declares `markdown-content` and `typing-dots` and
//	                              www-connected-app declares neither.
//	PerDeclaration                derived by probing the repository's own functional `@utility`
//	                              blocks. Both corpus repositories report the same 18 roots and they
//	                              report them because they share the submodule that declares them,
//	                              so deriving is the only answer right for a repository never seen.
//	Descriptors                   the base table's, plus a declining descriptor for each functional
//	                              `@utility` root the repository declared.
//	PropertyOrder                 `PropertyOrder` from property_order_table.go, measured identical
//	                              across all three systems. Shared rather than copied.
//
// # Cost
//
// Building a table is a few map copies over 301 roots and one PropertySort per static `@utility`
// block. It is folded into `loadDesignSystemForProgram` rather than given a second entry point, so
// it is covered by the same build-once counter `518fec5` asserts: adding to a counted path is cheap,
// and adding a second uncounted build path is how the per-file rebuild comes back.
package tailwind

import "sort"

// NewTable builds this design system's descriptor table.
//
// The framework half comes from descriptor_base_table.go and the repository half from the live
// system, per the split in the file comment. The base half is shared rather than deep-copied:
// readings are immutable and the descriptors are only ever read, so a table per run costs the
// per-root map headers and nothing else.
//
// A nil system is a nil table rather than a bare framework one. A caller that lost its design system
// must decline, and handing it a table describing Tailwind-with-no-repository would answer every
// class confidently and be wrong on every repository token, which is the confident-green failure
// cohere exists to remove.
func NewTable(system *LoadedDesignSystem) *Table {
	if system == nil {
		return nil
	}

	table := &Table{
		TailwindVersion: system.TailwindVersion,
		Descriptors:     make(map[string]*Descriptor, len(baseDescriptors)+len(system.utilityRoots)),
		Statics:         make(map[string]Reading, len(FrameworkStaticDeclarations)+len(system.staticUtilityNodes)),
		PropertyOrder:   PropertyOrder,
		theme:           system.theme,
	}

	for root, descriptor := range baseDescriptors {
		table.Descriptors[root] = descriptor
	}
	// The framework statics come from their compiled declarations rather than from a table of
	// readings, so this path and the repository path below are the same PropertySort walk.
	//
	// `baseStatics` used to be read here. It held 869 readings against the 890 registrations
	// FrameworkStaticDeclarations carries, agreed with all 869, and was missing 21 that
	// `getClassList()` does not advertise: the deprecated-but-registered utilities such as
	// `bg-gradient-to-r`, `break-words` and `max-w-screen`. Two of those are written in the ahra tree
	// today, so the table was not merely incomplete, it was answering the wrong reading for classes
	// in front of it.
	for name := range FrameworkStaticDeclarations {
		if reading, found := FrameworkStaticReading(name); found {
			table.Statics[name] = reading
		}
	}

	table.addThemeNamespaces(system.theme)
	table.addRepositoryStatics(system)
	table.addRepositoryFunctionalRoots(system)

	return table
}

// addThemeNamespaces fills Namespaces and KeysByNamespace from the repository's own theme.
//
// This is the per-repository half of the table and the reason no table can be committed. It mirrors
// `internal/lint/rules/tailwind/tools/generate_descriptor_table/context.mjs`, which asks the engine the same question the
// same way, because the two must agree: the readings in the base table were measured under the
// namespace set that function produces, and a namespace set derived differently would make the table
// answer a different question than the lookup asks.
func (table *Table) addThemeNamespaces(theme *Theme) {
	// Every candidate prefix of every key, offered to the theme's own membership test rather than
	// split on a guessed boundary. A key is `--<namespace>-<name>` and the boundary is ambiguous
	// from the string alone: `--color-red-500` could split three ways, so which prefixes are real
	// namespaces is the theme's answer and not this function's.
	candidatePrefixes := map[string]bool{}
	for _, entry := range theme.Entries() {
		segments := splitThemeKey(entry.Key)
		for length := 1; length <= len(segments)-1; length++ {
			candidatePrefixes["--"+joinSegments(segments[:length])] = true
		}
	}

	table.KeysByNamespace = make(map[string]map[string]bool, len(candidatePrefixes))
	namespaces := make([]string, 0, len(candidatePrefixes))
	for prefix := range candidatePrefixes {
		keys := theme.KeysInNamespaces([]string{prefix})
		if len(keys) == 0 {
			continue
		}
		set := make(map[string]bool, len(keys))
		for _, key := range keys {
			set[key] = true
		}
		table.KeysByNamespace[prefix] = set
		namespaces = append(namespaces, prefix)
	}

	// Longest first, then lexical. The order is the engine's namespace precedence and part of the
	// contract rather than a presentation choice: `--drop-shadow` must be tested before `--shadow`,
	// or `drop-shadow-lg` resolves through the wrong namespace. It is also the order context.mjs
	// sorts in, which is the order the base table's readings were measured under.
	sort.Slice(namespaces, func(left, right int) bool {
		if len(namespaces[left]) != len(namespaces[right]) {
			return len(namespaces[left]) > len(namespaces[right])
		}
		return namespaces[left] < namespaces[right]
	})
	table.Namespaces = namespaces
}

// addRepositoryStatics adds a reading for every static `@utility` block the repository declared.
//
// A static utility takes no value, so its reading is a constant rather than a model: PropertySort
// over the block body, which is exactly what the engine does. The repository wins a collision with
// the framework's table, matching upstream, where `@utility` registration runs after the framework's
// and `Utilities.set` overwrites.
func (table *Table) addRepositoryStatics(system *LoadedDesignSystem) {
	for root, nodes := range system.staticUtilityNodes {
		sorted := PropertySort(nodes)
		table.Statics[root] = Reading{Order: sorted.Order, Count: sorted.Count}
	}
}

// addRepositoryFunctionalRoots gives every functional `@utility` root a descriptor that declines.
//
// A repository's functional block is answered by the `@utility` evaluator, not by this table, so the
// descriptor's only job is to say the root is real and this table is not what answers it. That
// distinction is the whole point of adding them rather than leaving them absent: an absent root
// makes Lookup return false for "no descriptor", which is indistinguishable from a root nobody can
// write, while present-and-declining is what lets a caller escalate to the evaluator.
//
// Derived from the repository rather than carried in the base table. Both corpus repositories report
// the same roots because they share the submodule that declares them, and a repository this port has
// never seen gets its own.
//
// # Why every one of them declines, and not only the per-declaration ones
//
// `context.mjs` splits these two ways. It marks a root per-declaration only when probing shows its
// arity changing with the value — 18 roots on this repository, all `<animation-root>-*` — and gives
// every other repository root a measured descriptor row like any framework root.
//
// That split cannot be reproduced here, and reproducing half of it would be worse than not trying. A
// measured row needs the engine: the extractor probes each root with 527 value shapes and reads the
// result out of `compileAstNodes`. Nothing in shipped Go can do that, which is the same reason the
// framework's rows are generated rather than computed. So a live table's only honest answer for a
// repository root is to decline and let the evaluator, which *can* compile the block, answer it.
//
// Measured, because the cost of declining is a real number rather than a shrug: over the fixture
// corpus this moves 180 cases from answered to declined. Every one is then recoverable — 58 are read
// back by the evaluator, exactly matching the engine, and the remaining 122 had no engine reading to
// begin with, so they were never scorable. Declining here costs nothing a caller cannot recover, and
// a wrong row would cost a wrong sort with no way to notice.
// TestLiveTableDeclinesOnlyWhereTheEvaluatorAnswers holds that, so it cannot decay into a claim.
func (table *Table) addRepositoryFunctionalRoots(system *LoadedDesignSystem) {
	for root, kinds := range system.utilityRoots {
		if !kinds[UtilityKindFunctional] {
			continue
		}
		table.Descriptors[root] = &Descriptor{Root: root, PerDeclaration: true}
	}
}

// splitThemeKey splits a `--a-b-c` key into its segments, without the leading dashes.
//
// Written out rather than `strings.Split(key[2:], "-")` so a key that is not a custom property at
// all yields nothing instead of a segment list built from the wrong offset.
func splitThemeKey(key string) []string {
	if len(key) < 2 || key[0] != '-' || key[1] != '-' {
		return nil
	}
	var segments []string
	current := 2
	for index := 2; index <= len(key); index++ {
		if index == len(key) || key[index] == '-' {
			segments = append(segments, key[current:index])
			current = index + 1
		}
	}
	return segments
}

// joinSegments rejoins key segments with the dash that split them.
func joinSegments(segments []string) string {
	joined := ""
	for index, segment := range segments {
		if index > 0 {
			joined += "-"
		}
		joined += segment
	}
	return joined
}
