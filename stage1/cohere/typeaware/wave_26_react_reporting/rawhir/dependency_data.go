package high_level_intermediate_representation

type DependencyPathEntry struct {
	Property string
	Optional bool
}

// ReactiveScopeDependency is one input to a reactive scope: an access path rooted at a value.
//
// The root is an `IdentifierId` and the path may be empty, which is the ordinary case for a bare
// value used directly. `Reactive` is carried from the `Place` the dependency was read through, and
// is upstream's `reactive` field; it is not consulted by this pass and exists for the pruning pass
// that would consume it (`prune_non_reactive_dependencies.rs:230` retains on exactly this).
type ReactiveScopeDependency struct {
	Identifier IdentifierId
	Reactive   bool
	Path       []DependencyPathEntry
}

func equalPaths(a, b []DependencyPathEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index].Property != b[index].Property || a[index].Optional != b[index].Optional {
			return false
		}
	}
	return true
}

func manualMemoRootsEqual(inferred, source ManualMemoRoot) bool {
	if inferred.IsGlobal != source.IsGlobal {
		return false
	}
	if inferred.IsGlobal {
		return inferred.Name == source.Name
	}
	return inferred.Place.Identifier == source.Place.Identifier
}
