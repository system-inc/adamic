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
	return append(
		paths,
		"/home/user/adamic/stage1/cohere/gitignore",
		"a/./b/../c",
		"./a/b/../../c/d",
		"/a/./b/c/../../d/",
		"/home/user/adamic/stage1/../stage1/./cohere//gitignore/",
		"../../a/b/../../..",
		"a/b/c/../../../../d",
		"/../../..",
		"//a//b//",
		"a/.../b",
		"a/..b/..",
		"...",
		"..a/../b",
		strings.Repeat("a/", 50)+strings.Repeat("../", 49),
		"a\nb/../c",
		"\\/a",
	)
}

// The sweep: path.ts's clean, base and dir answer as Go's path.Clean, path.Base and path.Dir do on every
// sweep path, natively, on Node and through the JavaScript backend; and each mutant of path.ts below is
// caught on Node and natively.
func TestPathAnswersAsGosPathPackage(test *testing.T) {
	test.Parallel()
	paths := sweepPaths()
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n")
	var input, want strings.Builder
	for _, each := range paths {
		input.WriteString(escape.Replace(each))
		input.WriteByte('\n')
		want.WriteString(path.Clean(each) + "\t" + path.Base(each) + "\t" + path.Dir(each) + "\n")
	}
	pathsFile := filepath.Join(test.TempDir(), "paths.txt")
	if writeError := os.WriteFile(pathsFile, []byte(input.String()), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	sweep := func(test *testing.T, mutant pathMutant) (run, run, run) {
		test.Helper()
		directory := sweepDirectory(test, mutant.from, mutant.to)
		program := lowered(test, filepath.Join(directory, "pathsweep.ts"))
		sanitized := filepath.Join(pathSweepSanitized(test, mutant), "port")
		nativeRun := sanitizedRun(test, sanitized, pathsFile)
		if leaked := leaks(test, program, sanitized, pathsFile); leaked != "" {
			test.Errorf("leaks:\n%s", leaked)
		}
		return onNode(
			test,
			filepath.Join(directory, "pathsweep.ts"),
			pathsFile,
		), nativeRun, onJavaScriptBackend(test, program, pathsFile)
	}

	test.Run("as Go's path package", func(test *testing.T) {
		test.Parallel()
		nodeRun, nativeRun, backendRun := sweep(test, pathMutant{})
		for _, side := range []struct {
			name string
			run  run
		}{{"Node", nodeRun}, {"native", nativeRun}, {"the JavaScript backend", backendRun}} {
			if side.run.exitCode != 0 || len(side.run.stderr) > 0 {
				test.Fatalf("%s: exit %d, stderr %q", side.name, side.run.exitCode, side.run.stderr)
			}
			if difference := firstDifference(string(side.run.stdout), want.String()); difference != "" {
				test.Errorf("%s and Go's path package differ: %s", side.name, difference)
			}
		}
		test.Logf(
			"%d paths: clean, base and dir the same from Go's path package, path.ts natively, on Node and through the JavaScript backend",
			len(paths),
		)
	})

	for _, mutant := range pathMutants {
		test.Run("catches "+mutant.name, func(test *testing.T) {
			test.Parallel()
			nodeRun, nativeRun, _ := sweep(test, mutant)
			for _, side := range []struct {
				name string
				run  run
			}{{"on Node", nodeRun}, {"natively", nativeRun}} {
				if side.run.exitCode != 0 {
					test.Errorf(
						"%s the mutant exits %d (stderr %q); it must be caught by its answers",
						side.name,
						side.run.exitCode,
						side.run.stderr,
					)
					continue
				}
				difference := firstDifference(string(side.run.stdout), want.String())
				if difference == "" {
					test.Errorf("%s the mutant agrees with Go's path package: the sweep cannot see it", side.name)
					continue
				}
				test.Logf("%s, caught: %s", side.name, difference)
			}
		})
	}
}

// A pathMutant takes one place out of path.ts, the from text replaced by to; the zero one is path.ts itself.
type pathMutant struct{ name, from, to string }

// pathMutants are the mutants of path.ts TestPathAnswersAsGosPathPackage must catch.
var pathMutants = []pathMutant{
	// Each takes one test out of isClean, so a path it no longer sees as unclean comes back as it is.
	{
		"a run of slashes taken for clean",
		"if(path === '' || path.includes('//') || (path.endsWith('/') && path !== '/')) {",
		"if(path === '' || (path.endsWith('/') && path !== '/')) {",
	},
	{
		"a trailing slash taken for clean",
		"if(path === '' || path.includes('//') || (path.endsWith('/') && path !== '/')) {",
		"if(path === '' || path.includes('//')) {",
	},
	{
		"a trailing . taken for clean",
		"if(path.startsWith('./') || path.includes('/./') || path.endsWith('/.')) {",
		"if(path.startsWith('./') || path.includes('/./')) {",
	},
	{
		"a rooted path's .. taken for clean",
		"return !path.includes('/../') && !path.endsWith('/..');",
		"return !path.includes('/../');",
	},
	{
		"a .. after a name taken for clean",
		"        !rest.includes('/../') &&\n        !rest.endsWith('/..') &&\n",
		"        !rest.endsWith('/..') &&\n",
	},
	// dir cleaning from the last slash's position on, where Go cleans up to and with it.
	{
		"dir keeping the last element's first unit",
		"return clean(path.slice(0, lastSlash));",
		"return clean(path.slice(0, lastSlash + 2));",
	},
}

// sweepDirectory copies path.ts and its driver into a directory of their own, with from replaced by to
// in path.ts when from is set.
func sweepDirectory(test *testing.T, from string, to string) string {
	test.Helper()
	directory := test.TempDir()
	for _, name := range []string{"path.ts", "pathsweep.ts"} {
		contents, readError := os.ReadFile(name)
		if readError != nil {
			test.Fatal(readError)
		}
		source := string(contents)
		if from != "" && name == "path.ts" {
			if strings.Count(source, from) != 1 {
				test.Fatalf("the mutant must change exactly one place in path.ts: %q", from)
			}
			source = strings.Replace(source, from, to, 1)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return directory
}
