package notices

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func committed(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(moduleRoot(t), name))
	if err != nil {
		t.Fatalf("%s is not at the module root: %v", name, err)
	}
	return string(contents)
}

// TestTheCommittedNoticesAreCurrent regenerates both files and requires the committed ones to match. It
// is what fails when a dependency moves, the compiler's NOTICE.txt changes, the Swift engine resolves a
// new package or the curated list is edited, and the fix is always the same command.
func TestTheCommittedNoticesAreCurrent(t *testing.T) {
	t.Parallel()

	files, err := Generate(moduleRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for name, generated := range files {
		if committed(t, name) != string(generated) {
			t.Errorf("%s differs from what the generator writes now. Run `go run ./internal/release/tools/notices` from the module root and commit the result", name)
		}
		// The repository stores text as LF, so a CR written here is converted on commit, and the next
		// clone fails the comparison above on bytes nobody changed. Microsoft's NOTICE.txt is CRLF.
		if bytes.ContainsRune(generated, '\r') {
			t.Errorf("%s is generated with a carriage return, which git would convert to LF on commit", name)
		}
	}
}

// TestEveryGoModRequirementIsCredited holds go.mod against the committed file directly, apart from the
// generator, so a generator that dropped a module would still be caught.
func TestEveryGoModRequirementIsCredited(t *testing.T) {
	t.Parallel()

	command := exec.Command("go", "mod", "edit", "-json")
	command.Dir = moduleRoot(t)
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	var goMod struct{ Require []struct{ Path string } }
	if err := json.Unmarshal(output, &goMod); err != nil {
		t.Fatal(err)
	}
	notices := committed(t, ThirdPartyNoticesFileName)
	credited := 0
	for _, requirement := range goMod.Require {
		if strings.HasPrefix(requirement.Path, compilerModulePrefix) || isCohere(requirement.Path) {
			continue
		}
		if !strings.Contains(notices, "\n### "+requirement.Path+"\n") {
			t.Errorf("go.mod requires %s, and %s does not credit it", requirement.Path, ThirdPartyNoticesFileName)
		}
		credited++
	}
	if credited == 0 {
		t.Fatalf("go.mod required no third-party module, so this test checked nothing")
	}
}

func TestEverySwiftPackageIsCredited(t *testing.T) {
	t.Parallel()

	pins, err := readSwiftPins(moduleRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(pins) == 0 {
		t.Fatalf("swift/Package.resolved pins nothing, so this test checked nothing")
	}
	notices := committed(t, ThirdPartyNoticesFileName)
	for _, pin := range pins {
		if !strings.Contains(notices, "\n### "+pin.Identity+"\n") || !strings.Contains(notices, "- Version: "+pin.Version+"\n") {
			t.Errorf("the Swift engine resolves %s %s, and %s does not credit that version", pin.Identity, pin.Version, ThirdPartyNoticesFileName)
		}
	}
}

func TestTheGeneratorRefusesAModuleWithNoRecordedLicense(t *testing.T) {
	t.Parallel()

	var builder strings.Builder
	err := writeModule(&builder, Module{Path: "example.com/uncredited", Version: "v1.0.0", Directory: t.TempDir()}, "carries")
	if err == nil || !strings.Contains(err.Error(), "example.com/uncredited") {
		t.Fatalf("a module with no recorded license was written rather than refused: %v", err)
	}

	// Recorded but with no license file on disk: refused too, rather than credited with nothing.
	module := Module{Path: "golang.org/x/sync", Version: "v0.0.0", Directory: t.TempDir()}
	if err := writeModule(&builder, module, "carries"); err == nil {
		t.Fatalf("a module whose license file is missing was credited")
	}
}

// TestEveryLicenseTextIsUsedAndEveryReferenceExists keeps licenses/ and the lists naming it in step: a
// text nothing names is a credit nobody reads, and a name with no text stops the generator.
func TestEveryLicenseTextIsUsedAndEveryReferenceExists(t *testing.T) {
	t.Parallel()

	named := map[string]bool{"go.txt": true}
	for _, upstream := range Upstreams {
		if (upstream.LicenseFile == "") == (upstream.Note == "") {
			t.Errorf("%s needs exactly one of a license file and a note saying why there is none", upstream.Name)
		}
		if upstream.LicenseFile != "" {
			named[upstream.LicenseFile] = true
		}
	}
	for _, license := range swiftLicenses {
		named[license.File] = true
	}
	present := map[string]bool{}
	err := fs.WalkDir(licenseTexts, "licenses", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		present[filepath.Base(path)] = true
		contents, err := licenseTexts.ReadFile(path)
		if err == nil && len(bytes.TrimSpace(contents)) == 0 {
			t.Errorf("%s is empty", path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for name := range named {
		if !present[name] {
			t.Errorf("licenses/%s is named but not there", name)
		}
	}
	for name := range present {
		if !named[name] {
			t.Errorf("licenses/%s is there but nothing names it", name)
		}
	}
}

// TestNoGPLFileIsTracked holds the copyleft check the license survey came back clean on. git's own test
// corpus, which is GPL, is read at test time from COHERE_GIT_SOURCE and must never be committed, and
// nothing else GPL may be either. A tracked file holding a GNU license's title fails here by name.
func TestNoGPLFileIsTracked(t *testing.T) {
	t.Parallel()

	// Built from pieces, so this file doesn't match its own search. Fixed strings in the two casings a
	// license title and a license header use, since a case-insensitive pattern search over every tracked
	// file took 6.9 seconds and these take a fifth of one.
	var titles []string
	for _, kind := range []string{"GENERAL", "LESSER GENERAL", "AFFERO GENERAL"} {
		title := "GNU " + kind + " PUBLIC" + " LICENSE"
		titles = append(titles, title, "GNU "+cases(kind)+" Public"+" License")
	}
	arguments := []string{"grep", "-l", "-I", "-F"}
	for _, title := range titles {
		arguments = append(arguments, "-e", title)
	}
	command := exec.Command("git", arguments...)
	command.Dir = moduleRoot(t)
	output, err := command.Output()
	var exitError *exec.ExitError
	if err != nil && !(errors.As(err, &exitError) && exitError.ExitCode() == 1) {
		t.Fatalf("git grep could not search the tracked files: %v", err)
	}
	for _, path := range strings.Fields(string(output)) {
		// Microsoft's notice mentions the LGPL in a sentence and carries no GPL code.
		if path == NoticeFileName {
			continue
		}
		t.Errorf("%s is tracked and carries a GNU license", path)
	}

	// And the search can find one: a GPL title in a file git tracks is reported.
	probe := t.TempDir()
	for _, step := range [][]string{{"init", "--quiet"}, {"add", "COPYING"}} {
		if step[0] == "add" {
			if err := os.WriteFile(filepath.Join(probe, "COPYING"), []byte("                    "+titles[0]+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if output, err := exec.Command("git", append([]string{"-C", probe}, step...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", step, err, output)
		}
	}
	probeCommand := exec.Command("git", arguments...)
	probeCommand.Dir = probe
	if found, _ := probeCommand.Output(); strings.TrimSpace(string(found)) != "COPYING" {
		t.Fatalf("the search found %q in a repository tracking a GPL file, so it can't be trusted to find one", found)
	}
}

// cases capitalizes each word of an all-caps phrase: "LESSER GENERAL" becomes "Lesser General".
func cases(phrase string) string {
	words := strings.Fields(strings.ToLower(phrase))
	for index, word := range words {
		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}
