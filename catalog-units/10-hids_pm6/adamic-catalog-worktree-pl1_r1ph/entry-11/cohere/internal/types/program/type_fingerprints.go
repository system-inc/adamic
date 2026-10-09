package program

import (
	"crypto/sha256"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// TypeFingerprints gives each project file a hash of everything its types can depend on, so a
// type-aware rule's findings for a file can be cached on the file's bytes plus this.
//
// A type-aware rule asks the checker about nodes in one file, and the checker's answers come from that
// file, every file it reaches through imports, and everything global. So the fingerprint covers:
//
//   - the file's import closure among project files, hashed per strongly connected component so import
//     cycles, which real trees have, are one unit, and so a file reached by 40% of the tree is hashed
//     once rather than once per importer;
//   - a global component folded into every fingerprint: every file in the program that is not a
//     project file (installed declarations), and every project file that can change other files' types
//     without being imported: declaration files, scripts, and anything that says `declare global` or
//     `declare module`.
//
// The global component is a superset on purpose. Including a file that turns out not to matter costs a
// miss when it changes; leaving out one that does would replay a type-aware finding computed against
// types that are gone, which is the stale verdict this cache must never produce. The lib files are not
// here because they are embedded in the binary, which is in the cache's key.
//
// Edges come from the program's own module resolution, normalised the way DependentClosure normalises
// them: compared raw, on a case-insensitive volume, they produced a graph with almost no edges, which
// reads as valid rather than as broken.
//
// Edges only join project files, so each member also carries where every one of its imports resolved,
// installed files included. A re-export that starts resolving to another installed declaration, both
// already in the program, moves no edge and no file in the global component: a package.json's "types"
// changing is enough. Without its resolutions in its hash, the importers of that re-exporter kept their
// key while the types they see changed (#9bjjk4a, found answering @system_cohere_lint).
func (g *Graph) TypeFingerprints() map[tspath.PathKey][sha256.Size]byte {
	global, edges, contents, resolutions := g.typeGraph()
	members := make(map[tspath.PathKey][sha256.Size]byte, len(resolutions))
	for path, resolved := range resolutions {
		members[path] = withResolutions(contents[path], resolved)
	}
	return fingerprintComponents(g.Anchor, g.ProjectFiles(), edges, members, global)
}

// typeGraphParts is typeGraph's answer.
type typeGraphParts struct {
	global      [sha256.Size]byte
	edges       map[tspath.PathKey][]tspath.PathKey
	contents    map[tspath.PathKey][sha256.Size]byte
	resolutions map[tspath.PathKey][sha256.Size]byte
}

// typeGraph is what every fingerprint is built over: the global component, hashing every file that can
// change other files' types without being imported, the import edges between project files, every file's
// content hash, computed once for all three, and each project file's resolutions.
//
// Computed once per graph. A walk keyed on both fingerprints asked for it twice, about 45ms each on ahra
// and all of it before the first worker starts (#zqsdzbq). The callers only read what it returns.
func (g *Graph) typeGraph() ([sha256.Size]byte, map[tspath.PathKey][]tspath.PathKey, map[tspath.PathKey][sha256.Size]byte, map[tspath.PathKey][sha256.Size]byte) {
	g.typeGraphOnce.Do(func() {
		parts := &g.typeGraphParts
		parts.global, parts.edges, parts.contents, parts.resolutions = g.computeTypeGraph()
	})
	return g.typeGraphParts.global, g.typeGraphParts.edges, g.typeGraphParts.contents, g.typeGraphParts.resolutions
}

func (g *Graph) computeTypeGraph() ([sha256.Size]byte, map[tspath.PathKey][]tspath.PathKey, map[tspath.PathKey][sha256.Size]byte, map[tspath.PathKey][sha256.Size]byte) {
	allFiles := g.Program.GetSourceFiles()
	projectFiles := g.ProjectFiles()

	contents, beyond := hashContentsInParallel(allFiles)

	isProject := make(map[tspath.PathKey]bool, len(projectFiles))
	for _, sourceFile := range projectFiles {
		isProject[sourceFile.PathKey()] = true
	}

	global := sha256.New()
	globalPaths := make([]tspath.PathKey, 0, len(allFiles))
	for _, sourceFile := range allFiles {
		if !isProject[sourceFile.PathKey()] || beyond[sourceFile.PathKey()] {
			globalPaths = append(globalPaths, sourceFile.PathKey())
		}
	}
	// By the anchored path, in sort and in hash, so every spelling of the root hashes alike (#547dhjz).
	stable := make(map[tspath.PathKey]string, len(globalPaths))
	for _, path := range globalPaths {
		stable[path] = g.Anchor.Stable(string(path))
	}
	sort.Slice(globalPaths, func(first, second int) bool { return stable[globalPaths[first]] < stable[globalPaths[second]] })
	for _, path := range globalPaths {
		sum := contents[path]
		global.Write([]byte(stable[path]))
		global.Write([]byte{0})
		global.Write(sum[:])
	}
	var globalSum [sha256.Size]byte
	copy(globalSum[:], global.Sum(nil))

	edges := make(map[tspath.PathKey][]tspath.PathKey, len(projectFiles))
	for fromPath, resolutions := range g.Program.GetResolvedModules() {
		if !isProject[fromPath] {
			continue
		}
		for _, resolution := range resolutions {
			if resolution == nil || resolution.ResolvedFileName == "" {
				continue
			}
			toward := g.pathFor(resolution.ResolvedFileName.AsString())
			if isProject[toward] && toward != fromPath {
				edges[fromPath] = append(edges[fromPath], toward)
			}
		}
	}
	return globalSum, edges, contents, g.resolutionSums(isProject)
}

// resolutionSums hashes, for each project file, every module specifier and type reference it wrote and the
// file each resolved to, or that it resolved to nothing. Sorted, so the order the compiler resolved in cannot
// move it.
func (g *Graph) resolutionSums(isProject map[tspath.PathKey]bool) map[tspath.PathKey][sha256.Size]byte {
	entries := make(map[tspath.PathKey][]string, len(isProject))
	for fromPath, resolutions := range g.Program.GetResolvedModules() {
		if !isProject[fromPath] {
			continue
		}
		for key, resolution := range resolutions {
			target := ""
			if resolution != nil {
				target = g.Anchor.Stable(resolution.ResolvedFileName.AsString())
			}
			entries[fromPath] = append(entries[fromPath], fmt.Sprintf("module\x00%s\x00%d\x00%s", key.Name, key.Mode, target))
		}
	}
	for fromPath, references := range g.Program.GetResolvedTypeReferenceDirectives() {
		if !isProject[fromPath] {
			continue
		}
		for key, reference := range references {
			target := ""
			if reference != nil {
				target = g.Anchor.Stable(reference.ResolvedFileName.AsString())
			}
			entries[fromPath] = append(entries[fromPath], fmt.Sprintf("type\x00%s\x00%d\x00%s", key.Name, key.Mode, target))
		}
	}
	sums := make(map[tspath.PathKey][sha256.Size]byte, len(isProject))
	for path := range isProject {
		lines := entries[path]
		sort.Strings(lines)
		hash := sha256.New()
		for _, line := range lines {
			hash.Write([]byte(line))
			hash.Write([]byte{0})
		}
		var sum [sha256.Size]byte
		copy(sum[:], hash.Sum(nil))
		sums[path] = sum
	}
	return sums
}

// withResolutions is a member's hash: its own content or shape, and where its imports resolved.
func withResolutions(own [sha256.Size]byte, resolutions [sha256.Size]byte) [sha256.Size]byte {
	return sha256.Sum256(append(own[:], resolutions[:]...))
}

// reachesBeyondItsImports reports whether a project file can change other files' types without being
// imported by them. Detected generously, by kind and by text, since a false positive costs a miss.
func reachesBeyondItsImports(sourceFile *ast.SourceFile) bool {
	if sourceFile.IsDeclarationFile || sourceFile.ExternalModuleIndicator == nil {
		return true
	}
	text := sourceFile.Text()
	return strings.Contains(text, "declare global") || strings.Contains(text, "declare module")
}

// hashContentsInParallel hashes every file's text. The texts are already in memory; this is the only
// per-run cost the fingerprints add over the walk.
//
// It also answers reachesBeyondItsImports for each file in the same pass, since that reads the same text, and
// asked one file at a time after the hashing it cost about 11ms on ahra on the path to the first walk worker.
func hashContentsInParallel(files []*ast.SourceFile) (map[tspath.PathKey][sha256.Size]byte, map[tspath.PathKey]bool) {
	sums := make([][sha256.Size]byte, len(files))
	reaches := make([]bool, len(files))
	workers := runtime.NumCPU()
	var waitGroup sync.WaitGroup
	for worker := range workers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for index := worker; index < len(files); index += workers {
				sums[index] = sha256.Sum256([]byte(files[index].Text()))
				reaches[index] = reachesBeyondItsImports(files[index])
			}
		}()
	}
	waitGroup.Wait()
	contents := make(map[tspath.PathKey][sha256.Size]byte, len(files))
	beyond := make(map[tspath.PathKey]bool, len(files))
	for index, sourceFile := range files {
		contents[sourceFile.PathKey()] = sums[index]
		if reaches[index] {
			beyond[sourceFile.PathKey()] = true
		}
	}
	return contents, beyond
}

// fingerprintComponents runs Tarjan's algorithm over the project's import graph. Tarjan emits each
// strongly connected component after every component it reaches, so a component's hash can include its
// dependencies' finished hashes in one pass. A component's hash covers the global component, its
// members' anchored paths and contents in order, and its dependencies' hashes in order; every member gets it.
func fingerprintComponents(anchor PathAnchor, projectFiles []*ast.SourceFile, edges map[tspath.PathKey][]tspath.PathKey,
	contents map[tspath.PathKey][sha256.Size]byte, global [sha256.Size]byte) map[tspath.PathKey][sha256.Size]byte {

	index := 0
	indices := make(map[tspath.PathKey]int, len(projectFiles))
	lowlinks := make(map[tspath.PathKey]int, len(projectFiles))
	onStack := make(map[tspath.PathKey]bool, len(projectFiles))
	stack := []tspath.PathKey{}
	componentOf := make(map[tspath.PathKey]int, len(projectFiles))
	componentSums := [][sha256.Size]byte{}
	fingerprints := make(map[tspath.PathKey][sha256.Size]byte, len(projectFiles))

	var connect func(path tspath.PathKey)
	connect = func(path tspath.PathKey) {
		indices[path], lowlinks[path] = index, index
		index++
		stack = append(stack, path)
		onStack[path] = true

		for _, toward := range edges[path] {
			if _, visited := indices[toward]; !visited {
				connect(toward)
				lowlinks[path] = min(lowlinks[path], lowlinks[toward])
			} else if onStack[toward] {
				lowlinks[path] = min(lowlinks[path], indices[toward])
			}
		}
		if lowlinks[path] != indices[path] {
			return
		}

		members := []tspath.PathKey{}
		for {
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[top] = false
			members = append(members, top)
			if top == path {
				break
			}
		}
		// By the anchored path, in sort and in hash, so every spelling of the root hashes alike (#547dhjz).
		stableMembers := make(map[tspath.PathKey]string, len(members))
		for _, member := range members {
			stableMembers[member] = anchor.Stable(string(member))
		}
		sort.Slice(members, func(first, second int) bool { return stableMembers[members[first]] < stableMembers[members[second]] })

		component := len(componentSums)
		dependencies := map[int]bool{}
		for _, member := range members {
			componentOf[member] = component
		}
		for _, member := range members {
			for _, toward := range edges[member] {
				if other := componentOf[toward]; other != component {
					dependencies[other] = true
				}
			}
		}

		hash := sha256.New()
		hash.Write(global[:])
		for _, member := range members {
			sum := contents[member]
			hash.Write([]byte(stableMembers[member]))
			hash.Write([]byte{0})
			hash.Write(sum[:])
		}
		dependencySums := make([]string, 0, len(dependencies))
		for other := range dependencies {
			dependencySums = append(dependencySums, string(componentSums[other][:]))
		}
		sort.Strings(dependencySums)
		for _, dependency := range dependencySums {
			hash.Write([]byte(dependency))
		}
		var sum [sha256.Size]byte
		copy(sum[:], hash.Sum(nil))
		componentSums = append(componentSums, sum)
		for _, member := range members {
			fingerprints[member] = sum
		}
	}

	for _, sourceFile := range projectFiles {
		if _, visited := indices[sourceFile.PathKey()]; !visited {
			connect(sourceFile.PathKey())
		}
	}
	return fingerprints
}
