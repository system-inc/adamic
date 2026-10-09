package program_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/types/program"
)

// A path the compiler asked about with a typed question that answered no is still present if it is on
// disk in another form, and only a path genuinely not there is recorded absent.
//
// Module resolution asks whether `./Graph.css` is a directory, the answer is no, and the file is right
// there. Recording that no as absence made an untouched ahra tree miss on a four-month-old stylesheet,
// so the cache could never hit. Built through the real Build so the probe reaches the recorder the way
// the compiler sends it, not the way a test imagines it does.
func TestInputRecorderRecordsAFileProbedAsADirectoryAsPresent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(path, contents string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "tsconfig.json"), `{"compilerOptions":{"strict":true,"noEmit":true,"module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write(filepath.Join(root, "source", "index.ts"), "import './style.css';\nimport { a } from './missing';\nexport const b = a;\n")
	write(filepath.Join(root, "source", "style.css"), "body {}\n")

	recorder := program.NewInputRecorder()
	if _, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json"), CurrentDirectory: root, Inputs: recorder}); err != nil {
		t.Fatalf("build: %v", err)
	}
	present, absent, probed := recorder.Inputs()

	style := filepath.Join(root, "source", "style.css")
	if slices.Contains(absent, style) {
		t.Errorf("%s is on disk and was recorded absent, so every check would report it appearing", style)
	}
	for _, path := range absent {
		if _, err := os.Stat(path); err == nil {
			t.Errorf("%s is on disk and was recorded absent", path)
		}
	}

	// The control: the recorder does record absence, or the loop above passes on an empty list.
	if len(absent) == 0 {
		t.Fatal("nothing was recorded absent, though ./missing was looked for and does not exist: the check above proved nothing")
	}
	if !slices.Contains(present, filepath.Join(root, "source", "index.ts")) {
		t.Error("the source file the program read was not recorded present")
	}

	// And the whole recording replays as a hit on the untouched tree, which is the property all of
	// this serves.
	cache, err := program.RecordRunCache("k", present, nil, absent, probed, nil, 0, time.Time{})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := cache.Check("k"); err != nil {
		t.Fatalf("the untouched tree missed: %v", err)
	}
}

// A directory the build only asked whether it exists is recorded for that alone (#q51f02a). With no
// package.json at the project's root, module resolution asks it of every directory up to /, and on a busy
// machine /tmp changes many times a second: checked at full signature, those cost a replay on a tree
// nothing had touched, which is how TestTheCacheLivesInTheProjectAndNoCacheTouchesNone failed at load 58.
// So an unrelated directory created beside the project still hits. What the probe was for still misses: a
// package.json appearing above the project. And a directory the build listed is never recorded for its
// existence alone, even one holding nothing it read: a file added to an empty directory the include glob
// walks still misses.
func TestADirectoryOnlyProbedIsRecordedForItsExistence(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	write := func(path string, contents string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "tsconfig.json"), `{"compilerOptions":{"strict":true,"noEmit":true},"include":["**/*.ts"]}`)
	write(filepath.Join(root, "index.ts"), "export const a: number = 1;\n")
	if err := os.Mkdir(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	record := func() *program.RunCache {
		t.Helper()
		recorder := program.NewInputRecorder()
		if _, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json"), CurrentDirectory: root, Inputs: recorder}); err != nil {
			t.Fatalf("build: %v", err)
		}
		present, absent, probed := recorder.Inputs()
		cache, err := program.RecordRunCache("k", present, nil, absent, probed, nil, 0, time.Time{})
		if err != nil {
			t.Fatalf("record: %v", err)
		}
		return cache
	}

	// The parent is probed and nothing more. Since #zc57tqg it is left out of the record entirely, since the
	// project's own directory beneath it says the same thing; either way it must never be watched at full
	// signature, or the unrelated directory below would cost the replay.
	cache := record()
	rootRecorded := false
	for _, input := range cache.Inputs {
		if input.Path == parent && !input.ExistenceOnly {
			t.Fatal("the project's parent, which the build only probed, was recorded at full signature")
		}
		if input.Path == root {
			rootRecorded = true
			if input.ExistenceOnly {
				t.Error("the project's own directory, which the build lists, was recorded for its existence alone")
			}
		}
	}
	if !rootRecorded {
		t.Fatal("the project's own directory was not recorded, so nothing below is about what a listing catches")
	}

	if err := os.Mkdir(filepath.Join(parent, "unrelated"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := cache.Check("k"); err != nil {
		t.Fatalf("a directory created beside the project cost the replay: %v", err)
	}

	write(filepath.Join(parent, "package.json"), "{}\n")
	if err := cache.Check("k"); err == nil {
		t.Error("a package.json created above the project, which module resolution reads, still replayed")
	}
	if err := os.Remove(filepath.Join(parent, "package.json")); err != nil {
		t.Fatal(err)
	}
	if err := cache.Check("k"); err != nil {
		t.Fatalf("with the package.json removed again the tree did not hit: %v", err)
	}

	write(filepath.Join(root, "empty", "added.ts"), "export const added = 2;\n")
	if err := cache.Check("k"); err == nil {
		t.Error("a file added to a directory the include glob lists, holding nothing the build read, still replayed")
	}
}

// Removing a package's directory misses, though since #zc57tqg the probe of that directory is no longer
// recorded: the package.json the build read beneath it is, and it cannot stay present without the directory.
// That is the argument for leaving the probe out, so it is planted here rather than assumed: record, remove
// the package, and the check must fail; and the control, the same record over the untouched tree, must hit.
func TestRemovingAPackageWhoseProbeWasLeftOutStillMisses(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(path string, contents string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "tsconfig.json"), `{"compilerOptions":{"strict":true,"noEmit":true,"module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write(filepath.Join(root, "source", "index.ts"), "import { value } from 'pkg';\nexport const doubled: number = value * 2;\n")
	write(filepath.Join(root, "node_modules", "pkg", "package.json"), `{"name":"pkg","types":"index.d.ts"}`)
	write(filepath.Join(root, "node_modules", "pkg", "index.d.ts"), "export declare const value: number;\n")

	recorder := program.NewInputRecorder()
	if _, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json"), CurrentDirectory: root, Inputs: recorder}); err != nil {
		t.Fatalf("build: %v", err)
	}
	present, absent, probed := recorder.Inputs()
	pkg := filepath.Join(root, "node_modules", "pkg")
	if !slices.Contains(present, filepath.Join(pkg, "package.json")) {
		t.Fatal("the package's package.json was not recorded, so nothing below is about what it covers")
	}
	if slices.Contains(probed, pkg) {
		t.Errorf("%s has a recorded path beneath it and was still recorded as a probe", pkg)
	}
	cache, err := program.RecordRunCache("k", present, nil, absent, probed, nil, 0, time.Time{})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := cache.Check("k"); err != nil {
		t.Fatalf("the untouched tree missed: %v", err)
	}
	if err := os.Rename(pkg, pkg+"-moved"); err != nil {
		t.Fatal(err)
	}
	if err := cache.Check("k"); err == nil {
		t.Error("the package's directory was moved away and the record still replayed")
	}
}
