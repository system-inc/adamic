package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/system-inc/cohere/internal/types/program"
	"github.com/system-inc/cohere/internal/types/sourcename"
)

// adamicHeld is the program's Adamic `.a` files, cleaned to the form the format scope holds names in
// (#6mhafvb). A `.a` file is in a program only when its tsconfig lists ".a" in "sourceExtensions", so this
// is the one place a run learns which `.a` files are source rather than static libraries.
func adamicHeld(graph *program.Graph) map[string]struct{} {
	held := map[string]struct{}{}
	if graph == nil {
		return held
	}
	for _, sourceFile := range graph.ProjectFiles() {
		if fileName := filepath.Clean(sourceFile.FileName().AsString()); sourcename.IsAdamic(fileName) {
			held[fileName] = struct{}{}
		}
	}
	return held
}

// adamicIncluded is adamicHeld for a run that builds no graph, `--format-only` (#m0dktbn): the `.a` files the
// tsconfig includes, read from the config alone. It is the program's own list less any `.a` file reached only
// by an import from outside the include, which Adamic's layout does not have: its code is listed where its
// tsconfig claims ".a". ReadProjectConfig cleans each name to the platform's separator, the form the walk's
// names and adamicHeld's are in, so a tsconfig's `/` names match a walk's `\` ones on Windows.
func adamicIncluded(configFileName string) (map[string]struct{}, error) {
	config, err := program.ReadProjectConfig(configFileName)
	if err != nil {
		return nil, err
	}
	held := map[string]struct{}{}
	for _, fileName := range config.FileNames {
		if sourcename.IsAdamic(fileName) {
			held[fileName] = struct{}{}
		}
	}
	return held, nil
}

// reportIgnoredAdamic names each `.a` file the program holds that git's ignore rules leave out of the format
// walk, and returns how many (#6mhafvb). Such a file is type-checked and linted, since the tsconfig's include
// does not consult .gitignore, but the walk never offers it to the formatter, so it would go unformatted in
// silence. `.a` is the static library's extension, and C and C++ ignore templates list `*.a`, which is how
// Adamic source lands there. Only `.a`: a `.ts` file git ignores is usually generated, and naming those would
// be noise.
func reportIgnoredAdamic(out io.Writer, root string, held map[string]struct{}) int {
	names := make([]string, 0, len(held))
	for fileName := range held {
		names = append(names, fileName)
	}
	sort.Strings(names)
	ignored := 0
	for _, fileName := range names {
		ignoredBy := ignoreSource(root, fileName)
		if ignoredBy == "" {
			continue
		}
		ignored++
		shown := fileName
		if relative, err := filepath.Rel(root, fileName); err == nil {
			shown = relative
		}
		fmt.Fprintf(out, "⚠ %s is in the program, but git's ignore rules leave it out of the format walk (%s): checked, not formatted\n",
			shown, ignoredBy)
	}
	return ignored
}
