package program

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// DependentClosureLimit is where a closure stops being a saving and starts being the whole tree.
//
// Measured over all 3,542 project files on the ahra tree: the mean one-hop set is 3.8 files and the
// mean full closure is 58.4, because 83% of files reach their entire closure within 25. For those
// the closure costs about what one hop costs, so there is nothing to trade.
//
// The tail is where the bound earns its place. `ClassName.ts` has 567 direct importers and closes at
// 1,477, which is 42% of the tree, and `useRouter.ts` grows 2, 70, 177, 209, 224, 459, 591, 655.
// Walking those is the whole-tree cost with extra bookkeeping, so the honest answer above the bound
// is to say the closure is too large and check everything.
//
// 500 rather than a rounder number because the distribution has a gap there: 5% of files close
// between 101 and 500 and a separate 5% close above it, so the bound falls between two populations
// rather than through the middle of one.
const DependentClosureLimit = 500

// DependentClosure is the seeds plus every project file that reaches them through imports.
//
// The second return is false when the closure exceeded DependentClosureLimit, and the caller is
// expected to check the whole tree rather than to use a truncated set. A partial closure reported as
// a complete one is the failure this whole scoping effort has to avoid: it would be fast, silent,
// and wrong in the direction that hides findings.
//
// Demonstrated before this existed, by changing one function's return type to `any`: the whole tree
// reported seven new findings, three in the edited file and four in a consumer two directories away,
// and a run scoped to the edited file alone saw only the three.
func (g *Graph) DependentClosure(seeds []*ast.SourceFile) ([]*ast.SourceFile, bool) {
	if len(seeds) == 0 {
		return nil, true
	}

	projectFiles := g.ProjectFiles()
	byPath := make(map[tspath.PathKey]*ast.SourceFile, len(projectFiles))
	for _, sourceFile := range projectFiles {
		byPath[sourceFile.PathKey()] = sourceFile
	}

	// Reversed once per call rather than cached, because a run builds one closure and the reversal is
	// a walk over a map the program already holds.
	//
	// Both sides go through toPath. The resolution map keys its rows by an already-normalised path
	// while a resolution names its target as an ordinary file name, and on a case-insensitive volume
	// those differ in case alone. Comparing them raw produced a graph with almost no edges, which read
	// as `every file reaches its whole closure in one hop` rather than as a broken map.
	importers := make(map[tspath.PathKey][]tspath.PathKey, len(projectFiles))
	for fromPath, resolutions := range g.Program.GetResolvedModules() {
		if _, ours := byPath[fromPath]; !ours {
			continue
		}
		for _, resolution := range resolutions {
			if resolution == nil || resolution.ResolvedFileName == "" {
				continue
			}
			toward := g.pathFor(resolution.ResolvedFileName.AsString())
			if _, ours := byPath[toward]; !ours {
				continue
			}
			importers[toward] = append(importers[toward], fromPath)
		}
	}

	seen := make(map[tspath.PathKey]struct{}, len(seeds))
	frontier := make([]tspath.PathKey, 0, len(seeds))
	for _, seed := range seeds {
		if _, already := seen[seed.PathKey()]; already {
			continue
		}
		seen[seed.PathKey()] = struct{}{}
		frontier = append(frontier, seed.PathKey())
	}

	for len(frontier) > 0 {
		next := make([]tspath.PathKey, 0, len(frontier))
		for _, current := range frontier {
			for _, importer := range importers[current] {
				if _, already := seen[importer]; already {
					continue
				}
				seen[importer] = struct{}{}
				if len(seen) > DependentClosureLimit {
					return nil, false
				}
				next = append(next, importer)
			}
		}
		frontier = next
	}

	// Emitted in ProjectFiles order rather than in discovery order, so a scoped run reports findings
	// in the sequence an unscoped run would and the two can be diffed directly.
	closure := make([]*ast.SourceFile, 0, len(seen))
	for _, sourceFile := range projectFiles {
		if _, reaches := seen[sourceFile.PathKey()]; reaches {
			closure = append(closure, sourceFile)
		}
	}
	return closure, true
}

// pathFor normalises a file name the way the program keys its own maps.
func (g *Graph) pathFor(fileName string) tspath.PathKey {
	return toPath(fileName, g.Anchor.Root(), g.Config.CaseSensitivity())
}
