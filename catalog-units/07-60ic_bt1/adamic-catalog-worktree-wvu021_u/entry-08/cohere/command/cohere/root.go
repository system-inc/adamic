package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// projectMarker is the file whose directory is a project's root.
//
// The tsconfig alone, and not CohereSettings.json beside it, for two reasons. It is what every phase
// actually stands on: the type graph cannot be built without it, while the lint config is only
// loaded by the phases that lint, so `--types` is a complete run in a directory that has a tsconfig
// and no lint config. And it is the rule tsc itself uses when given no project, so `cohere` and
// `tsc` started in the same directory check the same program.
//
// Letting the lint config anchor the walk would be worse rather than merely different. On the ahra
// tree `projects/www-ahras-world` holds its own tsconfig and no lint config, and the root tsconfig
// excludes `projects/` as separate programs. Walking past the nearer tsconfig to find a lint config
// would check ahra's program from inside a directory that program excludes, and print a confident
// green about code the caller was not standing in. Anchoring on the tsconfig instead fails loudly
// there, naming the lint config it could not find, which is the answer a reader can act on.
const projectMarker = "tsconfig.json"

// swiftProjectMarker is the file whose directory is a Swift package's root, and the run there belongs
// to the Swift engine rather than to this binary's own phases.
//
// Discovery looks for both markers at each directory on the way up, and the nearest directory holding
// either wins, for the reason projectMarker gives: the nearer project is the one the caller is
// standing in. From inside `projects/ahraos-macos` that is its Package.swift, never ahra's tsconfig
// above it, which would check a program that excludes `projects/` and print a green about Swift code
// nothing looked at.
//
// For one engine's run, a directory holding both is TypeScript's. A run from the root checks both: discovery
// finds each marker as a project of its own, and each project's run names its engine (see discovery.go).
const swiftProjectMarker = "Package.swift"

// projectEngine is which engine checks the project at a root.
type projectEngine string

const (
	// engineTypeScript is this binary's own pipeline, over the program a tsconfig defines.
	engineTypeScript projectEngine = "TypeScript"

	// engineSwift is cohere-swift, which this binary runs and whose records it renders.
	engineSwift projectEngine = "Swift"
)

// projectLocation is where a run is anchored, and which directory each kind of path resolves against.
//
// Two directories, because there are two kinds of path and they mean different things. A path the
// project supplies by default (its tsconfig, its lint config) belongs to the project and resolves
// against its root. A path the caller typed (a file to check, a config named by flag) belongs to the
// shell it was typed in and resolves against the working directory. Collapsing the two is what made
// `cohere` usable only from the root: the defaults were resolved against wherever the caller stood.
type projectLocation struct {
	// Root is the project's directory: the program's current directory, where git is asked what
	// changed, and where the format walk starts.
	Root string

	// Engine is which engine checks the project at Root, decided by the marker that made it the root.
	Engine projectEngine

	// ConfigFileName is the absolute path of the tsconfig that defines the program, and empty for a
	// Swift package, which has none.
	ConfigFileName string

	// LintConfigFileName is the absolute path of the lint config.
	LintConfigFileName string

	// LintConfigFileNameGiven is whether `--lint-config` named it. A named file that is missing is an
	// error; the default one missing is zero config, the house stack (#bfxz13m).
	LintConfigFileNameGiven bool

	// ArgumentBase is the directory a path typed on the command line resolves against.
	ArgumentBase string

	// WorkingDirectory is where the process was started, kept so the report can say when the root
	// it checked is somewhere else.
	WorkingDirectory string
}

// locationRequest is what the command line said, before anything was resolved.
type locationRequest struct {
	// WorkingDirectory is the process's own, absolute.
	WorkingDirectory string

	// Directory is `--directory`, empty when not given.
	Directory string

	// Engine, when set, is the engine that checks Directory, overriding what its markers say: the run of a
	// discovered project names it, since a directory holding both markers is two projects (#f9nftxz).
	Engine projectEngine

	ConfigFileName      string
	ConfigFileNameGiven bool

	LintConfigFileName      string
	LintConfigFileNameGiven bool
}

// locateProject decides the project root and resolves every path the run uses against it.
//
// An explicit flag always wins over discovery, in this order:
//
//   - `--directory` is a `cd`: it is the root, and everything relative resolves against it, typed or
//     not, which is what the flag has always meant.
//   - `--tsconfig` names the project, so the root is the directory that holds it. The name itself was
//     typed, so it resolves against the working directory.
//   - Otherwise the root is the nearest directory at or above the working directory that holds a
//     tsconfig.json or a Package.swift. See projectMarker for why the tsconfig and not the lint
//     config, and swiftProjectMarker for how the two markers share the walk.
//
// The engine follows the marker. `--tsconfig` names a TypeScript program, so it is always TypeScript.
// `--directory` is a `cd`, so the directory it names is read for markers the way discovery reads each
// directory it passes.
//
// Discovery does not stop at a repository boundary, and that is deliberate: a submodule such as
// `libraries/structure` holds no tsconfig of its own and is part of the program above it, so
// stopping at its `.git` would refuse the most common place someone stands inside a project.
func locateProject(request locationRequest) (projectLocation, error) {
	base := request.WorkingDirectory
	if request.Directory != "" {
		base = absoluteFrom(request.WorkingDirectory, request.Directory)
	}
	location := projectLocation{ArgumentBase: base, WorkingDirectory: request.WorkingDirectory}

	switch {
	case request.Directory != "":
		location.Root = base
		location.ConfigFileName = absoluteFrom(base, request.ConfigFileName)
		// A directory with neither marker stays TypeScript, so the graph builder fails naming the
		// tsconfig it could not find, as it did before Swift existed.
		location.Engine = engineTypeScript
		if engine, found := engineAt(base); found {
			location.Engine = engine
		}
		if request.Engine != "" {
			location.Engine = request.Engine
		}

	case request.ConfigFileNameGiven:
		location.ConfigFileName = absoluteFrom(base, request.ConfigFileName)
		location.Root = filepath.Dir(location.ConfigFileName)
		location.Engine = engineTypeScript

	default:
		root, engine, err := findProjectRoot(base)
		if err != nil {
			return projectLocation{}, err
		}
		location.Root = root
		location.Engine = engine
		if engine == engineTypeScript {
			location.ConfigFileName = filepath.Join(root, request.ConfigFileName)
		}
	}

	location.LintConfigFileNameGiven = request.LintConfigFileNameGiven
	if request.LintConfigFileNameGiven {
		location.LintConfigFileName = absoluteFrom(base, request.LintConfigFileName)
	} else {
		location.LintConfigFileName = absoluteFrom(location.Root, request.LintConfigFileName)
	}
	return location, nil
}

// findProjectRoot walks up from a directory to the nearest one holding a tsconfig.json or a
// Package.swift, and says which engine that marker belongs to.
//
// Finding nothing is an error that names where the walk began, never a fallback to the starting
// directory. A fallback would hand the graph builder a tsconfig path that does not exist, and the
// error that came back would name a file in the wrong directory, which sends a reader looking for a
// typo rather than for the fact that they are standing outside any project.
func findProjectRoot(start string) (string, projectEngine, error) {
	directory := filepath.Clean(start)
	for {
		if engine, found := engineAt(directory); found {
			return directory, engine, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", "", fmt.Errorf(
				"no %s or %s in %s or any directory above it, so there is no project here to check: "+
					"run from inside one, or name it with -tsconfig or -directory",
				projectMarker, swiftProjectMarker, start,
			)
		}
		directory = parent
	}
}

// engineAt says which engine owns a directory by the marker it holds, TypeScript first. See
// swiftProjectMarker for why a directory holding both is TypeScript's.
func engineAt(directory string) (projectEngine, bool) {
	if isRegularFile(filepath.Join(directory, projectMarker)) {
		return engineTypeScript, true
	}
	if isRegularFile(filepath.Join(directory, swiftProjectMarker)) {
		return engineSwift, true
	}
	return "", false
}

// isRegularFile reports whether a path names something that is not a directory. A directory named
// `tsconfig.json` is not a project, and treating it as one would build a graph from nothing.
func isRegularFile(path string) bool {
	information, err := os.Stat(path)
	return err == nil && !information.IsDir()
}

// absoluteFrom resolves a path against a directory unless it is already absolute.
func absoluteFrom(directory string, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(directory, path)
}

// rootNote is the sentence the report carries when the root is not where the run started, and empty
// when it is.
//
// Silent in the ordinary case, because a line on every run is one people learn to skip. When the two
// differ it is the most important context in the output: a green run from `modules/tasks` that
// checked the whole of ahra and a green run that checked only `modules/tasks` would otherwise print
// the same verdict, and only one of them is what the caller probably assumed.
func (l projectLocation) rootNote() string {
	if filepath.Clean(l.Root) == filepath.Clean(l.WorkingDirectory) {
		return ""
	}
	return fmt.Sprintf("checked the project at %s (the run started in %s)", l.Root, l.WorkingDirectory)
}
