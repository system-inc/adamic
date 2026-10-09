package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/types/program"
)

// Ownership is how a repository with many TypeScript programs gets one report with each file in it once
// (#wvgxtey). Discovery finds every tsconfig.json; this reads each one and decides three things.
//
//   - A solution root, a tsconfig that includes no file and references others (`files: []` plus
//     references, as `tsc --init` monorepos and Vite's template write it), is not a project: there is
//     nothing in it to check. The tsconfigs it references are, whatever they are named, so they join the
//     projects found, which is how Vite's tsconfig.app.json and tsconfig.node.json are checked.
//   - Each file belongs to the nearest tsconfig that includes it: of the projects whose tsconfig includes
//     it, the deepest whose directory holds the file; if none holds it, the one whose directory shares the
//     longest path with it; ties in discovery order. That is the editor's rule, since TypeScript's
//     language server opens a file under the nearest tsconfig.json up its directory chain that includes
//     it, so cohere and the editor judge a file by the same options. The other projects that include it
//     yield it: it stays in their programs, where their type questions still need it, and they neither
//     lint it nor report its diagnostics. A project left with no file of its own is not run.
//   - Formatting goes by directory: a project's format walk leaves out the directories of the projects
//     below it, whose own runs format them, so two runs never write one file.
//
// One TypeScript project reads nothing: there is nothing to share, and the run keeps the cost it had.

// projectYieldVariable names the file that tells a project's run what it yields. See projectYield.
const projectYieldVariable = "COHERE_PROJECT_YIELD"

// projectYield is what one project's run leaves to others, as its parent wrote it.
type projectYield struct {
	// Files are files the tsconfig includes that a nearer one owns, absolute.
	Files []string `json:"files"`

	// Directories are directories the run's format walk leaves out, absolute, each formatted by the run
	// of the project there. The project's own directory means the walk formats nothing: another tsconfig
	// in the same directory formats it.
	Directories []string `json:"directories"`
}

// ownership is what reading the TypeScript projects decided, for the report and for each child.
type ownership struct {
	// Solutions are the solution roots, by label: found, and not run.
	Solutions []string

	// Yielding are projects left with no file of their own, by label: found, and not run.
	Yielding []string

	// Yields are what each project's run leaves to others, by label. A project with nothing to leave has
	// no entry.
	Yields map[string]projectYield

	// SharedDirectory are projects whose directory an earlier project's run already keeps its cache in, by
	// label. They run with the cache off: one cache table per directory, and two programs over one table
	// would replay each other's records.
	SharedDirectory map[string]bool
}

// configPath is the absolute path of a TypeScript project's tsconfig.
func (found discovery) configPath(project discoveredProject) string {
	name := project.ConfigFile
	if name == "" {
		name = projectMarker
	}
	return filepath.Join(found.Root, filepath.FromSlash(project.Directory), name)
}

// resolveOwnership reads every TypeScript project's tsconfig, follows references, drops solution roots and
// projects left with nothing, and decides what each remaining project yields. found.Projects is rewritten
// to the projects that run.
func resolveOwnership(found *discovery) (ownership, error) {
	decided := ownership{Yields: map[string]projectYield{}, SharedDirectory: map[string]bool{}}
	typeScript := 0
	for _, project := range found.Projects {
		if project.Engine == engineTypeScript {
			typeScript++
		}
	}
	if typeScript == 0 {
		return decided, nil
	}
	if typeScript == 1 {
		// One program shares nothing, so only a reference can change what runs: a solution root, or a
		// tsconfig that points at one discovery's walk did not find. Reading a tsconfig is milliseconds, and
		// a byte search keeps ahra's run from paying even that.
		for _, project := range found.Projects {
			if project.Engine == engineTypeScript && !mentionsReferences(found.configPath(project)) {
				return decided, nil
			}
		}
	}

	reads := readProjectConfigs(*found, found.Projects)
	known := map[string]bool{}
	for _, project := range found.Projects {
		if project.Engine == engineTypeScript {
			known[found.configPath(project)] = true
		}
	}
	// The references, followed to the end: each one discovery did not find is a project of its own, in the
	// repository and outside what discovery never enters.
	for queue := found.Projects; len(queue) > 0; {
		var added []discoveredProject
		for _, project := range queue {
			read, isRead := reads[found.configPath(project)]
			if project.Engine != engineTypeScript || !isRead || read.err != nil {
				continue
			}
			for _, reference := range read.config.References {
				if known[reference] {
					continue
				}
				known[reference] = true
				referenced, inside := found.projectAt(reference)
				if !inside {
					continue
				}
				added = append(added, referenced)
			}
		}
		for configFile, read := range readProjectConfigs(*found, added) {
			reads[configFile] = read
		}
		found.Projects = append(found.Projects, added...)
		queue = added
	}
	found.sortProjects()

	// Solution roots: nothing to check, and their references are projects now.
	running := found.Projects[:0:0]
	for _, project := range found.Projects {
		if read, isRead := reads[found.configPath(project)]; project.Engine == engineTypeScript && isRead && read.err == nil &&
			len(read.config.FileNames) == 0 && len(read.config.References) > 0 {
			decided.Solutions = append(decided.Solutions, projectLabel(project))
			continue
		}
		running = append(running, project)
	}
	found.Projects = running

	// Each file to the nearest tsconfig that includes it.
	includers := map[string][]int{}
	for index, project := range found.Projects {
		read, isRead := reads[found.configPath(project)]
		if project.Engine != engineTypeScript || !isRead || read.err != nil {
			continue
		}
		for _, fileName := range read.config.FileNames {
			includers[fileName] = append(includers[fileName], index)
		}
	}
	yielded := make([][]string, len(found.Projects))
	for fileName, indexes := range includers {
		if len(indexes) < 2 {
			continue
		}
		owner := nearestIncluder(*found, fileName, indexes)
		for _, index := range indexes {
			if index != owner {
				yielded[index] = append(yielded[index], fileName)
			}
		}
	}

	running = found.Projects[:0:0]
	var runningYields [][]string
	for index, project := range found.Projects {
		read, isRead := reads[found.configPath(project)]
		if isRead && read.err == nil && len(read.config.FileNames) > 0 && len(yielded[index]) == len(read.config.FileNames) {
			decided.Yielding = append(decided.Yielding, projectLabel(project))
			continue
		}
		running = append(running, project)
		runningYields = append(runningYields, yielded[index])
	}
	found.Projects = running

	// What each running project yields, files and format directories, and which share a directory.
	cacheDirectories := map[string]bool{}
	for index, project := range found.Projects {
		if project.Engine != engineTypeScript {
			continue
		}
		yield := projectYield{Files: runningYields[index]}
		directory := filepath.Join(found.Root, filepath.FromSlash(project.Directory))
		if cacheDirectories[directory] {
			decided.SharedDirectory[projectLabel(project)] = true
			yield.Directories = []string{directory}
		} else {
			cacheDirectories[directory] = true
			for _, other := range found.Projects {
				if other.Engine == engineTypeScript && isBelow(other.Directory, project.Directory) {
					yield.Directories = append(yield.Directories, filepath.Join(found.Root, filepath.FromSlash(other.Directory)))
				}
			}
		}
		if len(yield.Files) > 0 || len(yield.Directories) > 0 {
			sort.Strings(yield.Files)
			sort.Strings(yield.Directories)
			decided.Yields[projectLabel(project)] = yield
		}
	}
	return decided, nil
}

// projectConfigRead is one tsconfig read, or why it could not be: a project whose tsconfig cannot be read
// still runs, and its own run reports what is wrong with it.
type projectConfigRead struct {
	config program.ProjectConfig
	err    error
}

// readProjectConfigs reads the TypeScript projects' tsconfigs, several at once, by absolute path.
func readProjectConfigs(found discovery, projects []discoveredProject) map[string]projectConfigRead {
	reads := map[string]projectConfigRead{}
	var lock sync.Mutex
	var group sync.WaitGroup
	reading := make(chan struct{}, discoveryParallelism)
	for _, project := range projects {
		if project.Engine != engineTypeScript {
			continue
		}
		configFile := found.configPath(project)
		group.Add(1)
		go func() {
			defer group.Done()
			reading <- struct{}{}
			config, err := program.ReadProjectConfig(configFile)
			<-reading
			lock.Lock()
			reads[configFile] = projectConfigRead{config: config, err: err}
			lock.Unlock()
		}()
	}
	group.Wait()
	return reads
}

// projectAt is the project a referenced tsconfig defines, and whether it lies in the repository discovery
// walked and outside what it never enters: a dependency, build or cache directory, a repository of its
// own, or a marker the root's ignorePatterns refused.
func (found discovery) projectAt(configFile string) (discoveredProject, bool) {
	relative, err := filepath.Rel(found.Root, configFile)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return discoveredProject{}, false
	}
	relative = filepath.ToSlash(relative)
	for _, segment := range strings.Split(path.Dir(relative), "/") {
		if neverDescended[segment] {
			return discoveredProject{}, false
		}
	}
	for _, repository := range found.NestedRepositories {
		if isBelow(path.Dir(relative), repository) || path.Dir(relative) == repository {
			return discoveredProject{}, false
		}
	}
	if slices.Contains(found.Refused, relative) {
		return discoveredProject{}, false
	}
	if !isRegularFile(configFile) {
		return discoveredProject{}, false
	}
	project := discoveredProject{Directory: path.Dir(relative), Engine: engineTypeScript}
	if name := path.Base(relative); name != projectMarker {
		project.ConfigFile = name
	}
	return project, true
}

// nearestIncluder is which of the projects that include a file owns it. See Ownership.
func nearestIncluder(found discovery, fileName string, indexes []int) int {
	relative, err := filepath.Rel(found.Root, fileName)
	if err != nil {
		return indexes[0]
	}
	fileDirectory := path.Dir(filepath.ToSlash(relative))
	owner, ownerHolds, ownerShared := indexes[0], false, -1
	for _, index := range indexes {
		directory := found.Projects[index].Directory
		holds := directory == "." || fileDirectory == directory || isBelow(fileDirectory, directory)
		shared := sharedSegments(fileDirectory, directory)
		switch {
		case holds && !ownerHolds, holds == ownerHolds && shared > ownerShared:
			owner, ownerHolds, ownerShared = index, holds, shared
		}
	}
	return owner
}

// isBelow reports whether one slash directory, relative to the root, is strictly inside another.
func isBelow(directory string, ancestor string) bool {
	if directory == ancestor || directory == "." {
		return false
	}
	return ancestor == "." || strings.HasPrefix(directory, ancestor+"/")
}

// sharedSegments counts the leading path segments two slash directories share, the root counting as none.
func sharedSegments(left string, right string) int {
	if left == "." || right == "." {
		return 0
	}
	leftSegments, rightSegments := strings.Split(left, "/"), strings.Split(right, "/")
	shared := 0
	for shared < len(leftSegments) && shared < len(rightSegments) && leftSegments[shared] == rightSegments[shared] {
		shared++
	}
	return shared
}

// mentionsReferences reports whether a tsconfig's text could hold project references: a byte search, so a
// false positive costs one read and a false negative is impossible.
func mentionsReferences(configFile string) bool {
	contents, err := os.ReadFile(configFile)
	return err != nil || bytes.Contains(contents, []byte(`"references"`))
}

// writeProjectYield writes what one project yields to a file of its own for its run to read, and returns
// the path, or "" when it yields nothing.
func writeProjectYield(yield projectYield, directory string) (string, error) {
	if len(yield.Files) == 0 && len(yield.Directories) == 0 {
		return "", nil
	}
	encoded, err := json.Marshal(yield)
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp(directory, "yield-*.json")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.Write(encoded); err != nil {
		return "", err
	}
	return file.Name(), nil
}

// activeProjectYield is what this run yields, read once from the file its parent named, and empty for any
// run that is not one project of several.
var activeProjectYield = sync.OnceValues(func() (projectYield, error) {
	yieldFile := os.Getenv(projectYieldVariable)
	if yieldFile == "" {
		return projectYield{}, nil
	}
	contents, err := os.ReadFile(yieldFile)
	if err != nil {
		return projectYield{}, fmt.Errorf("reading what this project yields to the others: %w", err)
	}
	var yield projectYield
	if err := json.Unmarshal(contents, &yield); err != nil {
		return projectYield{}, fmt.Errorf("reading what this project yields to the others, %s: %w", yieldFile, err)
	}
	return yield, nil
})

// yieldedFiles is the run's yielded files as the set program.Options.Yielded takes, nil when it yields none.
func (yield projectYield) yieldedFiles() map[string]struct{} {
	if len(yield.Files) == 0 {
		return nil
	}
	files := make(map[string]struct{}, len(yield.Files))
	for _, fileName := range yield.Files {
		files[filepath.Clean(fileName)] = struct{}{}
	}
	return files
}

// nestedProjectsLayer is the format walk's count of files it left to the projects below it.
const nestedProjectsLayer = "nearer projects"

// withoutYieldedDirectories is a format walk less the files under the directories this run yields, each the
// run of the project there formats. A walk the run yields nothing from is returned as it is.
func withoutYieldedDirectories(enumeration formatfiles.Enumeration) formatfiles.Enumeration {
	yield, err := activeProjectYield()
	if err != nil || len(yield.Directories) == 0 {
		return enumeration
	}
	kept := enumeration.Files[:0:0]
	for _, fileName := range enumeration.Files {
		if yieldsFile(yield.Directories, fileName) {
			enumeration.IgnoredByLayer[nestedProjectsLayer]++
			continue
		}
		kept = append(kept, fileName)
	}
	enumeration.Files = kept
	keptAdamic := enumeration.Adamic[:0:0]
	for _, fileName := range enumeration.Adamic {
		if yieldsFile(yield.Directories, fileName) {
			enumeration.IgnoredByLayer[nestedProjectsLayer]++
			continue
		}
		keptAdamic = append(keptAdamic, fileName)
	}
	enumeration.Adamic = keptAdamic
	return enumeration
}

// yieldsFile reports whether a file lies in one of the yielded directories, or is one of them.
func yieldsFile(directories []string, fileName string) bool {
	fileName = filepath.Clean(fileName)
	for _, directory := range directories {
		if strings.HasPrefix(fileName, directory+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// activeProjectYieldFingerprint names what this run yields, for the run cache's key: empty when it yields
// nothing, so a run that is not one of several keeps the key it had.
func activeProjectYieldFingerprint() string {
	yield, err := activeProjectYield()
	if err != nil || (len(yield.Files) == 0 && len(yield.Directories) == 0) {
		return ""
	}
	encoded, _ := json.Marshal(yield)
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}
