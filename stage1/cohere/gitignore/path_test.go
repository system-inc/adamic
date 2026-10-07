package gitignore

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// sweepPaths are the paths path.ts is held to Go's path package on: every string of up to seven units
// over '/', '.' and 'a', which is every arrangement of separators, dots and dot-dots there is at that
// length; every string of up to four over '/', '.', '\u00e9', '\U0001f600' and '.a', characters a UTF-16 unit and a
// byte count differently; and longer ones by hand.
func sweepPaths() []string {
	var paths []string
	seen := map[string]bool{}
	var extend func(prefix string, alphabet []string, left int)
	extend = func(prefix string, alphabet []string, left int) {
		if !seen[prefix] {
			seen[prefix] = true
			paths = append(paths, prefix)
		}
		if left == 0 {
			return
		}
		for _, piece := range alphabet {
			extend(prefix+piece, alphabet, left-1)
		}
	}
	extend("", []string{"/", ".", "a"}, 7)
	extend("", []string{"/", ".", "..", "a"}, 6)
	extend("", []string{"/", ".", "\u00e9", "\U0001f600", ".a"}, 4)
	return append(paths,
		"/home/user/adamic/stage1/cohere/gitignore", "a/./b/../c", "./a/b/../../c/d", "/a/./b/c/../../d/", "/home/user/adamic/stage1/../stage1/./cohere//gitignore/",
		"../../a/b/../../..", "a/b/c/../../../../d", "/../../..", "//a//b//", "a/.../b", "a/..b/..", "...", "..a/../b",
		strings.Repeat("a/", 50)+strings.Repeat("../", 49), "a\nb/../c", "\\/a",
	)
}

// The sweep: path.ts's clean, base and dir answer as Go's path.Clean, path.Base and path.Dir do on every
// sweep path, natively, on Node and through the JavaScript backend; and each mutant of path.ts below is
// caught on Node and natively.
func TestPathAnswersAsGosPathPackage(t *testing.T) {
	t.Parallel()
	paths := sweepPaths()
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n")
	var input, want strings.Builder
	for _, each := range paths {
		input.WriteString(escape.Replace(each))
		input.WriteByte('\n')
		want.WriteString(path.Clean(each) + "\t" + path.Base(each) + "\t" + path.Dir(each) + "\n")
	}
	pathsFile := filepath.Join(t.TempDir(), "paths.txt")
	if err := os.WriteFile(pathsFile, []byte(input.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	sweep := func(t *testing.T, directory string) (run, run, run) {
		t.Helper()
		program := lowered(t, filepath.Join(directory, "pathsweep.ts"))
		nativeRun, sanitized := natively(t, program, pathsFile)
		if leaked := leaks(t, program, sanitized, pathsFile); leaked != "" {
			t.Errorf("leaks:\n%s", leaked)
		}
		return onNode(t, filepath.Join(directory, "pathsweep.ts"), pathsFile), nativeRun, onJavaScriptBackend(t, program, pathsFile)
	}

	t.Run("as Go's path package", func(t *testing.T) {
		t.Parallel()
		directory := sweepDirectory(t, "", "")
		nodeRun, nativeRun, backendRun := sweep(t, directory)
		for _, side := range []struct {
			name string
			run  run
		}{{"Node", nodeRun}, {"native", nativeRun}, {"the JavaScript backend", backendRun}} {
			if side.run.exitCode != 0 || len(side.run.stderr) > 0 {
				t.Fatalf("%s: exit %d, stderr %q", side.name, side.run.exitCode, side.run.stderr)
			}
			if difference := firstDifference(string(side.run.stdout), want.String()); difference != "" {
				t.Errorf("%s and Go's path package differ: %s", side.name, difference)
			}
		}
		t.Logf("%d paths: clean, base and dir the same from Go's path package, path.ts natively, on Node and through the JavaScript backend", len(paths))
	})

	for _, mutant := range []struct{ name, from, to string }{
		// Each takes one test out of isClean, so a path it no longer sees as unclean comes back as it is.
		{"a run of slashes taken for clean", "if(path === '' || path.includes('//') || (path.endsWith('/') && path !== '/')) {", "if(path === '' || (path.endsWith('/') && path !== '/')) {"},
		{"a trailing slash taken for clean", "if(path === '' || path.includes('//') || (path.endsWith('/') && path !== '/')) {", "if(path === '' || path.includes('//')) {"},
		{"a trailing . taken for clean", "if(path.startsWith('./') || path.includes('/./') || path.endsWith('/.')) {", "if(path.startsWith('./') || path.includes('/./')) {"},
		{"a rooted path's .. taken for clean", "return !path.includes('/../') && !path.endsWith('/..');", "return !path.includes('/../');"},
		{"a .. after a name taken for clean", "!rest.includes('/../') &&\n        !rest.endsWith('/..') &&", "!rest.endsWith('/..') &&"},
		// dir cleaning from the last slash's position on, where Go cleans up to and with it.
		{"dir keeping the last element's first unit", "return clean(path.slice(0, lastSlash));", "return clean(path.slice(0, lastSlash + 2));"},
	} {
		t.Run("catches "+mutant.name, func(t *testing.T) {
			t.Parallel()
			directory := sweepDirectory(t, mutant.from, mutant.to)
			nodeRun, nativeRun, _ := sweep(t, directory)
			for _, side := range []struct {
				name string
				run  run
			}{{"on Node", nodeRun}, {"natively", nativeRun}} {
				if side.run.exitCode != 0 {
					t.Errorf("%s the mutant exits %d (stderr %q); it must be caught by its answers", side.name, side.run.exitCode, side.run.stderr)
					continue
				}
				difference := firstDifference(string(side.run.stdout), want.String())
				if difference == "" {
					t.Errorf("%s the mutant agrees with Go's path package: the sweep cannot see it", side.name)
					continue
				}
				t.Logf("%s, caught: %s", side.name, difference)
			}
		})
	}
}

// sweepDirectory copies path.ts and its driver into a directory of their own, with from replaced by to
// in path.ts when from is set.
func sweepDirectory(t *testing.T, from string, to string) string {
	t.Helper()
	directory := t.TempDir()
	for _, name := range []string{"path.ts", "pathsweep.ts"} {
		contents, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		source := string(contents)
		if from != "" && name == "path.ts" {
			if strings.Count(source, from) != 1 {
				t.Fatalf("the mutant must change exactly one place in path.ts: %q", from)
			}
			source = strings.Replace(source, from, to, 1)
		}
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}
