package main

import (
	"fmt"
	"path/filepath"
)

func loadDesignSystemThrough(program string, fileSystem *recording) DesignSystemResult {
	projectRoot := projectRootOf(program)
	if projectRoot == "" {
		return DesignSystemResult{Err: fmt.Errorf("could not determine the project root from the program")}
	}

	// Every question about the disk goes through the program's own filesystem rather than to `os`,
	// because that filesystem is where the run cache's input recorder sits. A stylesheet read behind
	// its back is an input the cache never signs, so editing an ignored or untracked stylesheet the
	// theme imports replayed the old verdict (#ym4v8bc). Asked through it, each candidate entry point
	// probed and missed is recorded absent, so creating one invalidates, and each probe of the
	// package walk is recorded the same way. It is reached through a rule.RecordingFS, which also keeps
	// the read set the findings cache keys the Tailwind rules on.
	fileExists := fileSystem.FileExists

	entryPoint := FindEntryPoint(projectRoot, fileExists)
	if entryPoint == "" {
		return DesignSystemResult{Err: fmt.Errorf("%w: looked under %s", ErrNoTailwindEntryPoint, projectRoot)}
	}

	packageRoot := findTailwindPackageRoot(filepath.Dir(entryPoint), fileExists)
	if packageRoot == "" {
		return DesignSystemResult{
			EntryPoint: entryPoint,
			Err: fmt.Errorf(
				"%s: no installed tailwindcss package found, so `@import \"tailwindcss\"` cannot resolve",
				entryPoint,
			),
		}
	}

	system, err := adamicLoad(adamicLoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: packageRoot,
	})
	if err != nil {
		return DesignSystemResult{EntryPoint: entryPoint, Err: err}
	}
	// The stylesheets themselves are read by the engine, which does not take a filesystem, so each
	// one in the `@import` graph is stated through the program's afterwards. That records it as a
	// present input with its signature, which is all the cache needs: it re-signs the file on the
	// next run and misses when the file moved.
	for _, stylesheet := range system.Stylesheets {
		fileSystem.Stat(stylesheet)
	}
	// The descriptor table is built here, on the counted path, rather than lazily on first use. A
	// second build path would not be visible to TestDesignSystemIsBuiltOncePerProgram, and an
	// invisible build path is exactly how the per-file rebuild came back the last time.
	return DesignSystemResult{System: system, Table: adamicTable(system), EntryPoint: entryPoint}
}
