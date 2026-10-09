package guard

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// guardOverlay is the overlay the guards here read the rule packages through: the one GOFLAGS names, as
// file path to contents, or nil when GOFLAGS names none.
//
// The guards load the rule packages with go/packages, which runs `go list` and then parses every file
// itself, from disk. `go test -overlay x.json` reaches neither: the flag builds the test binary and
// stops there, and nothing in the binary can see it (measured: no build setting records it, and the
// environment is untouched). So a guard run that way checked the working tree while the command line
// said it checked the overlay, and two wrong "dispatch green" claims went through on it. GOFLAGS is
// inherited by the test binary and by the `go list` go/packages starts, so `GOFLAGS=-overlay=x.json`
// is the spelling every layer reads; this hands its contents to go/packages' own parser.
//
// The spelling that cannot work is refused rather than ignored, when it can be seen: on Unix the `go`
// command that started this binary is asked for its arguments, and an -overlay there that GOFLAGS does
// not carry fails the guard by name.
func guardOverlay(t *testing.T) map[string][]byte {
	t.Helper()
	return guardOverlayFrom(t, os.Getenv("GOFLAGS"))
}

// guardOverlayFrom is guardOverlay reading goFlags in place of the environment's GOFLAGS, so the tests
// that prove a named overlay is what the guards read can name one without t.Setenv, and run in parallel
// (#nxgt2ca).
func guardOverlayFrom(t *testing.T, goFlags string) map[string][]byte {
	t.Helper()

	refuseUnreachableOverlayFlag(t)

	overlayFile := ""
	for _, flag := range strings.Fields(goFlags) {
		if value, found := strings.CutPrefix(flag, "-overlay="); found {
			overlayFile = value
		} else if value, found := strings.CutPrefix(flag, "--overlay="); found {
			overlayFile = value
		}
	}
	if overlayFile == "" {
		return nil
	}
	overlay, err := readOverlay(overlayFile)
	if err != nil {
		t.Fatalf("GOFLAGS names the overlay %s, and the guards cannot read it: %v", overlayFile, err)
	}
	return overlay
}

// readOverlay reads a `go build -overlay` file into file path to contents.
//
// Paths must be absolute. The go command resolves a relative one against its own working directory, and
// this binary runs in its package's directory and loads from the module root, so a relative path would
// name a different file at each layer. A deleted file ("" as its replacement) is refused too: go/packages
// can replace a file's contents but has no way to say a file is absent, so the guard would read it.
func readOverlay(path string) (map[string][]byte, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("the overlay path %s is relative; name it absolutely", path)
	}
	encoded, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file struct {
		Replace map[string]string
	}
	if err := json.Unmarshal(encoded, &file); err != nil {
		return nil, fmt.Errorf("reading %s as an overlay: %w", path, err)
	}
	overlay := map[string][]byte{}
	for target, replacement := range file.Replace {
		if !filepath.IsAbs(target) || (replacement != "" && !filepath.IsAbs(replacement)) {
			return nil, fmt.Errorf("the overlay entry %s -> %s is relative; name both absolutely", target, replacement)
		}
		if replacement == "" {
			return nil, fmt.Errorf("the overlay deletes %s, which go/packages cannot express, so the guard would still read it", target)
		}
		contents, err := os.ReadFile(replacement)
		if err != nil {
			return nil, fmt.Errorf("reading the overlay's replacement for %s: %w", target, err)
		}
		overlay[target] = contents
	}
	return overlay, nil
}

// refuseUnreachableOverlayFlag fails the guard when the `go test` that started this binary was given
// -overlay on its command line rather than through GOFLAGS. Best effort: where the parent's arguments
// cannot be read, nothing is refused.
func refuseUnreachableOverlayFlag(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	output, err := exec.Command("ps", "-o", "args=", "-p", fmt.Sprint(os.Getppid())).Output()
	if err != nil {
		return
	}
	if overlayFlagOutsideGoFlags(string(output), os.Getenv("GOFLAGS")) {
		t.Fatalf("this guard was run by `%s`, whose -overlay builds the test binary and never reaches the guard, "+
			"so it would check the working tree; pass the overlay as GOFLAGS=-overlay=<absolute path> instead",
			strings.TrimSpace(string(output)))
	}
}

// overlayFlagOutsideGoFlags reports a go command line that carries -overlay when GOFLAGS does not.
func overlayFlagOutsideGoFlags(arguments string, goFlags string) bool {
	if strings.Contains(goFlags, "overlay") {
		return false
	}
	for _, argument := range strings.Fields(arguments) {
		if argument == "-overlay" || argument == "--overlay" || strings.HasPrefix(argument, "-overlay=") || strings.HasPrefix(argument, "--overlay=") {
			return true
		}
	}
	return false
}

// writeOverlay writes an overlay file replacing target with contents, and returns the overlay file's path.
func writeOverlay(t *testing.T, target string, contents string) string {
	t.Helper()
	directory := t.TempDir()
	replacement := filepath.Join(directory, filepath.Base(target))
	if err := os.WriteFile(replacement, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(map[string]map[string]string{"Replace": {target: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlayFile := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayFile, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	return overlayFile
}

// The creating-getter guard reads an overlay GOFLAGS names: a GetLocals call that exists only in the
// overlay is caught, and the working tree, which has none, is not what was read.
func TestTheCreatingGetterGuardReadsTheGoFlagsOverlay(t *testing.T) {
	t.Parallel()

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "internal/lint/rules/typescript/no_shadow.go")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	planted := strings.Replace(string(original), "range node.Locals()", "range ast.GetLocals(node)", 1)
	if planted == string(original) {
		t.Fatal("no-shadow no longer ranges over node.Locals(), so the plant missed; update the probe")
	}
	calls := creatingGetterCalls(t, guardOverlayFrom(t, "-overlay="+writeOverlay(t, path, planted)))
	if len(calls) != 1 || !strings.Contains(calls[0], "no_shadow.go") || !strings.HasSuffix(calls[0], "GetLocals") {
		t.Fatalf("a GetLocals call that exists only in the GOFLAGS overlay was not caught: %v", calls)
	}
}

// The TypeReach guard reads an overlay GOFLAGS names: a Shapes claim added only in the overlay, to a rule
// the scan already finds reading imported bodies, is caught.
func TestTheTypeReachGuardReadsTheGoFlagsOverlay(t *testing.T) {
	t.Parallel()

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	baseline := scanTypeReach(t, nil)
	claimed := map[string]bool{}
	for _, name := range baseline.claims {
		claimed[name] = true
	}
	// The first such rule by name, so every run plants into the same file.
	var candidates []string
	for name := range baseline.reason {
		if !claimed[name] {
			candidates = append(candidates, name)
		}
	}
	if len(candidates) == 0 {
		t.Fatal("no rule reads imported bodies without claiming Shapes, so there is nothing to plant a claim in")
	}
	slices.Sort(candidates)
	target := candidates[0]
	file := filepath.Join(root, "internal/lint/rules", baseline.reason[target])
	original, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	planted, err := plantShapesClaim(original)
	if err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	scan := scanTypeReach(t, guardOverlayFrom(t, "-overlay="+writeOverlay(t, file, string(planted))))
	caught := false
	for _, name := range scan.claims {
		caught = caught || (name == target && scan.mayReadImportedBodies[name])
	}
	if !caught {
		t.Fatalf("a Shapes claim on %s that exists only in the GOFLAGS overlay was not caught", target)
	}
}

// plantShapesClaim returns a rule file whose rule.Rule literal claims TypeReach Shapes: the literal's own
// TypeReach replaced, or one added before its first element. Found by parsing, not by matching text, so
// a column-aligned `Name:` or any other spelling of the literal is planted the same way, and spelled with
// the literal's own package qualifier.
func plantShapesClaim(source []byte) ([]byte, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "", source, 0)
	if err != nil {
		return nil, err
	}
	var literal *ast.CompositeLit
	var qualifier string
	ast.Inspect(file, func(node ast.Node) bool {
		if literal != nil {
			return false
		}
		candidate, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		selector, ok := candidate.Type.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Rule" {
			return true
		}
		if identifier, ok := selector.X.(*ast.Ident); ok && len(candidate.Elts) > 0 {
			literal, qualifier = candidate, identifier.Name
		}
		return false
	})
	if literal == nil {
		return nil, fmt.Errorf("no rule.Rule literal with elements")
	}
	// Offsets in a file parsed alone start at 1.
	offset := func(position token.Pos) int { return int(position) - 1 }
	claim := qualifier + ".TypeReachShapes"
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if key, isIdentifier := pair.Key.(*ast.Ident); ok && isIdentifier && key.Name == "TypeReach" {
			return slices.Concat(source[:offset(pair.Value.Pos())], []byte(claim), source[offset(pair.Value.End()):]), nil
		}
	}
	first := offset(literal.Elts[0].Pos())
	return slices.Concat(source[:first], []byte("TypeReach: "+claim+",\n"), source[first:]), nil
}

// The plant works on every rule there is, read back the way the scan reads a claim, so the probe above
// cannot miss on whichever rule it happens to pick.
func TestAShapesClaimPlantsIntoEveryRuleFile(t *testing.T) {
	t.Parallel()

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "internal/lint/rules/*/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	planted := 0
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		original, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, original, 0)
		if err != nil {
			t.Fatal(err)
		}
		name, _ := ruleLiteralTypeReach(parsed)
		if name == "" {
			continue
		}
		result, err := plantShapesClaim(original)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		reparsed, err := parser.ParseFile(token.NewFileSet(), path, result, 0)
		if err != nil {
			t.Errorf("%s: the planted file does not parse: %v", path, err)
			continue
		}
		if plantedName, claims := ruleLiteralTypeReach(reparsed); plantedName != name || !claims {
			t.Errorf("%s: planted, the rule reads as %q claiming Shapes %v, want %q claiming it", path, plantedName, claims, name)
		}
		planted++
	}
	if planted < 100 {
		t.Fatalf("planted into %d rule files, too few to be every rule", planted)
	}
}

// The command line `go test -overlay` is refused, since it cannot reach the guard, and GOFLAGS is not.
func TestAnOverlayFlagGoFlagsDoesNotCarryIsRefused(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		arguments string
		goFlags   string
		refused   bool
	}{
		{"go test -overlay /tmp/x.json ./internal/guard", "", true},
		{"go test -overlay=/tmp/x.json ./internal/guard", "", true},
		{"go test ./internal/guard", "-overlay=/tmp/x.json", false},
		{"go test -overlay=/tmp/x.json ./internal/guard", "-overlay=/tmp/x.json", false},
		{"go test ./internal/guard", "", false},
	} {
		if got := overlayFlagOutsideGoFlags(testCase.arguments, testCase.goFlags); got != testCase.refused {
			t.Errorf("%q with GOFLAGS %q: refused %v, want %v", testCase.arguments, testCase.goFlags, got, testCase.refused)
		}
	}
}
