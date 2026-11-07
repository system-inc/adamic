package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// commonJSSource reads __filename and __dirname, which bind only in a .ts file of a build that declares CommonJS
// (Adamic's ruling, Nov 6, #redxes2). Node's globals are typed only for a program that imports a node: module.
const commonJSSource = "import 'node:process';\nconst file: string = __filename;\nconst directory: string = __dirname;\nconsole.log(file);\nconsole.log(directory);\n"

// commonJSProgram writes commonJSSource as a .ts program of the test's own.
func commonJSProgram(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(commonJSSource), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

// --module-format takes commonjs or esm, once, after c's or js's file or among build's flags. Anything else is a
// usage error, before anything is loaded. Mutants: build's flag accepting a value moduleFormats doesn't hold, or a
// second --module-format, and its row exits 0 or 1.
func TestModuleFormatParsing(t *testing.T) {
	t.Parallel()
	for _, arguments := range [][]string{
		{"c", "missing.ts", "--module-format", "umd"},
		{"js", "missing.ts", "--module-format"},
		{"js", "missing.ts", "--module-format", "commonjs", "--module-format", "esm"},
		{"c", "--module-format", "commonjs", "missing.ts"},
		{"build", "missing.ts", "-o", "out", "--module-format"},
		{"build", "missing.ts", "-o", "out", "--module-format", "CommonJS"},
		{"build", "missing.ts", "-o", "out", "--module-format", "esm", "--module-format", "esm"},
	} {
		if code, output, diagnostic := runGuardCommand(t, arguments); code != 2 || output != "" || diagnostic != usage+"\n" {
			t.Errorf("%v: got exit %d, stdout %q, stderr %q; want exit 2 and usage text", arguments, code, output, diagnostic)
		}
	}
}

// The format a command is given reaches lowering: under commonjs the program builds and its C reads the image's path
// and directory; under esm, or naming nothing, it's NotYet and exits 1, as it was before the flag. Mutant:
// buildWithNativeExports dropping the format on its way to compileLibrary, and the commonjs row exits 1.
func TestModuleFormatReachesBuild(t *testing.T) {
	t.Parallel()
	path := commonJSProgram(t)
	for _, one := range []struct {
		arguments []string
		want      int
	}{
		{[]string{"--module-format", "commonjs"}, 0},
		{[]string{"--module-format", "esm"}, 1},
		{nil, 1},
	} {
		built := false
		code := buildWithNative(path, filepath.Join(t.TempDir(), "out"), one.arguments, func(source, output string, options native.Options) error {
			built = true
			for _, function := range []string{"adamic_node_filename()", "adamic_node_dirname()"} {
				if !strings.Contains(source, function) {
					t.Errorf("%v: the C doesn't call %s", one.arguments, function)
				}
			}
			return nil
		})
		if code != one.want || built != (one.want == 0) {
			t.Errorf("%v: got exit %d, built %t; want exit %d", one.arguments, code, built, one.want)
		}
	}
}

// c and js take the format after their file. Under commonjs, c's program calls the two runtime functions and js's
// reads the module's own path; under esm it's NotYet. Mutant: renderArguments answering an ES module whatever it's
// given, and the commonjs rows exit 1.
func TestModuleFormatReachesCAndJavaScript(t *testing.T) {
	t.Parallel()
	path := commonJSProgram(t)
	for _, one := range []struct {
		arguments []string
		want      int
		holds     string
	}{
		{[]string{"c", path, "--module-format", "commonjs"}, 0, "adamic_node_dirname()"},
		{[]string{"js", path, "--module-format", "commonjs"}, 0, "adamicNodeURL.fileURLToPath(import.meta.url)"},
		{[]string{"js", path, "--module-format", "commonjs", "--explain-checks"}, 0, "adamicNodePath.dirname("},
		{[]string{"c", path, "--module-format", "esm"}, 1, ""},
		{[]string{"js", path}, 1, ""},
	} {
		code, output, diagnostic := runGuardCommand(t, one.arguments)
		if code != one.want || !strings.Contains(output, one.holds) {
			t.Errorf("%v: got exit %d, stderr %q; want exit %d and stdout holding %q", one.arguments, code, diagnostic, one.want, one.holds)
		}
	}
}
