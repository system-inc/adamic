// Comparing the dependencies a developer wrote against the ones the compiler inferred.
//
// This is React's `compareDeps` and its `CompareDependencyResult`
// (`Validation/ValidatePreservedManualMemoization.ts` at `react_conformance.UpstreamSha`), the third
// firing condition of `preserve-manual-memoization` and the one that owns eleven of the thirteen
// goldens the two scope conditions leave silent.
//
// # What it decides
//
// A developer wrote `useMemo(fn, [props.a.b])`. The compiler infers what the callback actually
// reads. If the inferred dependency is `props.a` where the source said `props.a.b`, the value now
// recomputes whenever anything under `props.a` changes -- more often than the developer asked for.
// The reverse is equally wrong in the other direction.
//
// # The asymmetry, which is upstream's and is easy to get backwards
//
// An inferred path longer than the source path is fine: reading `props.a.b` when the source declared
// `props.a` recomputes no more often than promised, because any change to `props.a.b` is a change to
// `props.a`. An inferred path shorter than the source path is not fine, for the mirror reason.
//
// So the test is not equality. It is "the inferred path is at least as precise as the source path",
// with one exception below.
//
// # Refs are the exception, and the reason is one line of algebra
//
// A ref is not an immutable value: `ref_prev === ref_new` does not imply
// `ref_prev.current === ref_new.current`. So a partial match through a `current` access proves
// nothing, and the longer-is-fine rule is withdrawn wherever either path reads `current`.
package high_level_intermediate_representation

// CompareDependencyResult is how an inferred dependency relates to a written one.
//
// The constant order is upstream's and is load-bearing: `MergeCompareDependencyResults` takes the
// maximum, so a dependency compared against several source entries reports the most specific
// disagreement rather than the first one found.
type CompareDependencyResult uint8

const (
	// CompareDependencyOk means the inferred dependency is at least as precise as the source one.
	CompareDependencyOk CompareDependencyResult = iota
	// CompareDependencyRootDifference means the two name different values entirely.
	CompareDependencyRootDifference
	// CompareDependencyPathDifference means the paths diverge, or the inferred one is less precise.
	CompareDependencyPathDifference
	// CompareDependencySubpath means the source path is a strict prefix of the inferred one in a way
	// that is not acceptable.
	CompareDependencySubpath
	// CompareDependencyRefAccessDifference means a partial match ran through a `current` access,
	// where a prefix match proves nothing because a ref is not immutable.
	CompareDependencyRefAccessDifference
)

func (result CompareDependencyResult) String() string {
	switch result {
	case CompareDependencyOk:
		return "ok"
	case CompareDependencyRootDifference:
		return "the inferred and source dependencies name different values"
	case CompareDependencyPathDifference:
		return "the inferred dependency reads a different path than the source declared"
	case CompareDependencySubpath:
		return "the source dependency is a prefix of the inferred one"
	case CompareDependencyRefAccessDifference:
		return "a partial match through a ref access, which proves nothing"
	default:
		return "<unknown compare-dependency result>"
	}
}

// MergeCompareDependencyResults returns the more specific of two disagreements.
//
// Upstream's `merge`, a max over the constant order. A dependency is compared against every source
// entry and only reported if none matched; the reported reason is the strongest disagreement seen,
// so "different root" does not mask a path difference that was also present.
func MergeCompareDependencyResults(first, second CompareDependencyResult) CompareDependencyResult {
	if first > second {
		return first
	}
	return second
}

// CompareManualMemoDependencies reports how an inferred dependency relates to a written one.
//
// Upstream's `compareDeps`. Returns `CompareDependencyOk` when the inferred dependency is at least
// as precise as the source one, and otherwise the most specific reason it is not.
func CompareManualMemoDependencies(inferred, source ManualMemoDependency) CompareDependencyResult {
	if !manualMemoRootsEqual(inferred.Root, source.Root) {
		return CompareDependencyRootDifference
	}

	// Walk the shared prefix. A property mismatch means the paths diverge; an optionality mismatch
	// is reported immediately rather than noted, because an optional inferred read where the source
	// was non-optional is less precise regardless of what the rest of the path does.
	isSubpath := true
	shared := len(inferred.Path)
	if len(source.Path) < shared {
		shared = len(source.Path)
	}
	for index := 0; index < shared; index++ {
		if inferred.Path[index].Property != source.Path[index].Property {
			isSubpath = false
			break
		}
		if inferred.Path[index].Optional != source.Path[index].Optional {
			return CompareDependencyPathDifference
		}
	}

	// Equal length is an exact match. Longer inferred is acceptable -- reading deeper recomputes no
	// more often than promised -- unless the inferred path passes through a ref's `current`, where
	// a prefix match proves nothing.
	if isSubpath && (len(source.Path) == len(inferred.Path) ||
		(len(inferred.Path) >= len(source.Path) && !pathReadsRefCurrent(inferred.Path))) {
		return CompareDependencyOk
	}
	if !isSubpath {
		return CompareDependencyPathDifference
	}
	if pathReadsRefCurrent(source.Path) || pathReadsRefCurrent(inferred.Path) {
		return CompareDependencyRefAccessDifference
	}
	return CompareDependencySubpath
}

// manualMemoRootsEqual reports whether two dependency roots name the same value.
//
// A global is compared by name and a local by identifier, and the two kinds never match each other.
// That asymmetry is upstream's: a global has no identifier in this function to compare against.
func manualMemoRootsEqual(inferred, source ManualMemoRoot) bool {
	if inferred.IsGlobal != source.IsGlobal {
		return false
	}
	if inferred.IsGlobal {
		return inferred.Name == source.Name
	}
	return inferred.Place.Identifier == source.Place.Identifier
}

// pathReadsRefCurrent reports whether a path passes through a `current` access.
//
// The whole of upstream's ref exception. Named rather than inlined because it is asked three times
// in `CompareManualMemoDependencies` and the reason is the same each time: a ref is not immutable,
// so `ref_prev === ref_new` does not imply `ref_prev.current === ref_new.current`, and a prefix
// match through one proves nothing.
func pathReadsRefCurrent(path []DependencyPathEntry) bool {
	for _, entry := range path {
		if entry.Property == "current" {
			return true
		}
	}
	return false
}
