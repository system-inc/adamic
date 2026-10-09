// Package housesets decides, from the program, which house sets a zero-config run applies where
// (#bfxz13m, #j6p9t1g).
//
// The evidence is the code, never package.json: a dependency the code never imports applies nothing.
// cohere:react fits a file that imports react or holds JSX, cohere:next a file that imports next or is
// one of Next's own files in a program that imports next, and cohere:tailwind every file when the
// project's root stylesheet is Tailwind's. Everything asked is something the run already holds: each
// file's resolved import list, the parse's JSX fact, and, for Tailwind, the stylesheet probes the
// Tailwind rules make anyway. There is no walk of the disk and no git.
package housesets

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
)

// FileSystem is what detection asks the disk, through the program's own filesystem, so the run cache
// records each probe as an input and a stylesheet created later invalidates the verdict.
type FileSystem interface {
	FileExists(path tspath.RootedFilePath) bool
	ReadFile(path tspath.RootedFilePath) (string, bool)
}

// Detect reads which house sets fit which of the project's files.
//
// projectRoot is the directory whose tsconfig the run uses: Next's file conventions and the Tailwind
// stylesheet are found relative to it. settings are the project's own, from the file in
// settingsDirectory, nil when it writes none: a Tailwind location they name is where the stylesheet is
// looked for, as the rules will look.
func Detect(files []*ast.SourceFile, projectRoot string, fileSystem FileSystem, settings map[string]json.RawMessage, settingsDirectory string) configuration.HouseDetection {
	detection := configuration.HouseDetection{ReactFiles: map[string]bool{}, NextFiles: map[string]bool{}}

	importsNext := map[string]bool{}
	for _, file := range files {
		react := false
		for _, specifier := range file.Imports() {
			switch module := specifier.Text(); {
			case isReactSpecifier(module):
				react = true
			case module == "next" || strings.HasPrefix(module, "next/"):
				importsNext[file.FileName().AsString()] = true
			}
		}
		// JSX can only be written in a .tsx or .jsx file, so only those are asked, and the question is
		// the parse's own subtree fact, which the file's nodes cache once computed.
		if !react && isJsxFile(file.FileName().AsString()) {
			react = file.AsNode().SubtreeFacts()&ast.SubtreeContainsJsx != 0
		}
		if react {
			detection.ReactFiles[file.FileName().AsString()] = true
		}
	}

	// Next's own files are Next's only in a program that is Next's: an app/ directory in a project that
	// never imports next is just a directory.
	if len(importsNext) > 0 {
		for _, file := range files {
			if importsNext[file.FileName().AsString()] || isNextConventionFile(projectRoot, file.FileName().AsString()) {
				detection.NextFiles[file.FileName().AsString()] = true
			}
		}
	}

	detection.TailwindEntryPoint, detection.TailwindSkipped = detectTailwind(projectRoot, fileSystem, settings, settingsDirectory)
	return detection
}

// isReactSpecifier reports whether a module specifier is react's: `react`, a path into it
// (`react/jsx-runtime`), or a `react-` package (`react-dom`, `react-dom/client`).
func isReactSpecifier(module string) bool {
	return module == "react" || strings.HasPrefix(module, "react/") || strings.HasPrefix(module, "react-")
}

func isJsxFile(fileName string) bool {
	extension := filepath.Ext(fileName)
	return extension == ".tsx" || extension == ".jsx"
}

// nextConventionDirectories are the router roots whose every file is part of Next's tree, at the
// project root or under src/.
var nextConventionDirectories = []string{"app/", "pages/", "src/app/", "src/pages/"}

// nextConventionFiles are the files Next reads by name from the project root or src/, without the
// extension.
var nextConventionFiles = map[string]bool{"middleware": true, "instrumentation": true, "next.config": true}

// isNextConventionFile reports whether a file is one Next reads by where it sits: anything under an
// app/ or pages/ router root, or middleware, instrumentation and next.config at the root or in src/.
func isNextConventionFile(projectRoot string, fileName string) bool {
	relative, err := filepath.Rel(projectRoot, fileName)
	if err != nil || strings.HasPrefix(relative, "..") {
		return false
	}
	relative = filepath.ToSlash(relative)
	for _, directory := range nextConventionDirectories {
		if strings.HasPrefix(relative, directory) {
			return true
		}
	}
	directory, base := filepath.Split(relative)
	if directory != "" && directory != "src/" {
		return false
	}
	return nextConventionFiles[strings.TrimSuffix(base, filepath.Ext(base))]
}

// tailwindImport matches a stylesheet's import of Tailwind 4 (`@import "tailwindcss"`) or a Tailwind 3
// directive (`@tailwind base`).
var tailwindImport = regexp.MustCompile(`@import\s+(url\()?["']tailwindcss[/"']|@tailwind\s`)

// tailwindConfigFiles are the config file names Tailwind 3 reads at the project root.
var tailwindConfigFiles = []string{
	"tailwind.config.ts", "tailwind.config.js", "tailwind.config.mjs", "tailwind.config.cjs", "tailwind.config.mts", "tailwind.config.cts",
}

// detectTailwind returns the root stylesheet when it makes the project Tailwind's, or else why not, as
// a clause for the run's first line.
//
// A project whose settings["better-tailwindcss"] name a location (entryPoint, tailwindConfig or cwd)
// is Tailwind's wherever that location resolves, by the search the rules make (#gj5nm6e): the project
// said where its stylesheet is, and upstream runs the rules there, on Tailwind's default theme when
// the stylesheet is not found. Without one, the stylesheet is probed for where the rules probe.
func detectTailwind(projectRoot string, fileSystem FileSystem, settings map[string]json.RawMessage, settingsDirectory string) (string, string) {
	fileExists := func(path string) bool { return fileSystem.FileExists(tspath.RootedFilePath(path)) }
	if location, named := tailwind.LocationInSettings(settings, settingsDirectory); named {
		entryPoint, _, err := tailwind.ConfiguredEntryPoint(projectRoot, location, fileExists)
		if err != nil {
			return "", "the location settings[\"better-tailwindcss\"] name does not resolve (" + err.Error() + "), so its rules are skipped"
		}
		return entryPoint, ""
	}
	entryPoint := tailwind.FindEntryPoint(projectRoot, fileExists)
	if entryPoint == "" {
		return "", "no Tailwind stylesheet at any of the places its rules look (app/globals.css and the others), so its rules are skipped"
	}
	if contents, read := fileSystem.ReadFile(tspath.RootedFilePath(entryPoint)); read && tailwindImport.MatchString(contents) {
		return entryPoint, ""
	}
	for _, name := range tailwindConfigFiles {
		if fileExists(filepath.Join(projectRoot, name)) {
			return entryPoint, ""
		}
	}
	relative, err := filepath.Rel(projectRoot, entryPoint)
	if err != nil {
		relative = entryPoint
	}
	return "", filepath.ToSlash(relative) + " does not import tailwindcss and there is no tailwind.config, so its rules are skipped"
}
