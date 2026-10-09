package release

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The expression is spelled out here rather than read from LicenseExpression, so a change to the constant
// fails rather than passing on whatever it became. A lowercase or reads to npm and every license scanner as
// one unknown license instead of a choice of two, and message tooling lowercased it once already.
const expectedLicenseExpression = "MIT OR Apache-2.0"

func TestEveryPackageDeclaresTheDualLicense(t *testing.T) {
	t.Parallel()

	if LicenseExpression != expectedLicenseExpression {
		t.Fatalf("the license expression is %q, wanted exactly %q", LicenseExpression, expectedLicenseExpression)
	}

	manifests := map[string]map[string]any{
		DispatcherPackageName: decodeManifest(t, mustDispatcherManifest(t, "1.2.3")),
	}
	for _, target := range Targets {
		manifests[target.PackageName()] = decodeManifest(t, mustPlatformManifest(t, target, "1.2.3"))
	}
	for name, manifest := range manifests {
		if manifest["license"] != expectedLicenseExpression {
			t.Errorf("%s declares license %v, wanted exactly %q", name, manifest["license"], expectedLicenseExpression)
		}
		// Listed in files, or npm publish drops the texts the license field promises.
		files, _ := manifest["files"].([]any)
		for _, license := range LegalFileNames {
			if !slices.Contains(files, any(license)) {
				t.Errorf("%s's files are %v, so npm publish would leave %s out", name, files, license)
			}
		}
	}

	// The VS Code extension's manifest is a file in the repository rather than generated, so it is read.
	contents, err := os.ReadFile(filepath.Join(moduleRoot(t), "editors", "vscode", "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var extension struct {
		License string `json:"license"`
	}
	if err := json.Unmarshal(contents, &extension); err != nil {
		t.Fatal(err)
	}
	if extension.License != expectedLicenseExpression {
		t.Errorf("the VS Code extension declares license %q, wanted exactly %q", extension.License, expectedLicenseExpression)
	}
}

func TestPackagesShipBothLicenseTextsAsCommitted(t *testing.T) {
	t.Parallel()

	module := moduleRoot(t)
	output := t.TempDir()
	dispatcher, err := buildDispatcherPackage(Options{ModuleDirectory: module, OutputDirectory: output, Version: "1.0.0"}, placeholderChecksums)
	if err != nil {
		t.Fatal(err)
	}
	platform := filepath.Join(output, "cohere-linux-x64")
	if err := os.MkdirAll(platform, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := stageLicenses(module, platform); err != nil {
		t.Fatal(err)
	}

	for _, directory := range []string{dispatcher.Directory, platform} {
		for _, name := range LegalFileNames {
			committed, err := os.ReadFile(filepath.Join(module, name))
			if err != nil {
				t.Fatal(err)
			}
			shipped, err := os.ReadFile(filepath.Join(directory, name))
			if err != nil {
				t.Fatalf("%s does not ship %s: %v", directory, name, err)
			}
			if !bytes.Equal(shipped, committed) {
				t.Fatalf("%s ships a %s that differs from the committed one", directory, name)
			}
		}
	}

	empty := t.TempDir()
	for _, name := range LegalFileNames {
		writeFile(t, filepath.Join(empty, name), "\n")
	}
	if err := stageLicenses(empty, t.TempDir()); err == nil {
		t.Fatalf("an empty license was staged, so a package would ship terms with nothing in them")
	}
}

func TestTheLicenseTextsAreTheStandardOnes(t *testing.T) {
	t.Parallel()

	module := moduleRoot(t)

	// The Apache text verbatim from apache.org/licenses/LICENSE-2.0.txt, whose SHA-256 this is. A hand edit
	// to a license is a change of terms, so any byte that moves fails here.
	apache, err := os.ReadFile(filepath.Join(module, "LICENSE-APACHE"))
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256Hex(apache); got != "cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30" {
		t.Errorf("LICENSE-APACHE hashes to %s, not apache.org's Apache License 2.0 text", got)
	}

	// MIT's text is the same everywhere but its copyright line, so the rest is compared against the
	// tsgolint license this repository already carries, which is the standard text.
	mit, err := os.ReadFile(filepath.Join(module, "LICENSE-MIT"))
	if err != nil {
		t.Fatal(err)
	}
	reference, err := os.ReadFile(filepath.Join(module, "internal", "lint", "checking", "LICENSE"))
	if err != nil {
		t.Fatal(err)
	}
	if mitTerms(mit) != mitTerms(reference) || mitTerms(mit) == "" {
		t.Errorf("LICENSE-MIT's terms differ from the standard MIT text:\n%s", mit)
	}
	if !strings.HasPrefix(string(mit), "MIT License\n\nCopyright (c) 2026 ") {
		t.Errorf("LICENSE-MIT does not open with its title and a 2026 copyright line:\n%s", mit)
	}
}

func TestTheCopyrightHolderIsSystemInc(t *testing.T) {
	t.Parallel()

	// Spelled out, so a change to the constant fails rather than passing on whatever it became.
	const wanted = "System, Inc."
	if CopyrightHolder != wanted {
		t.Fatalf("the copyright holder is %q, wanted exactly %q", CopyrightHolder, wanted)
	}

	// The line itself, so a placeholder or a stray edit in the license fails here by what it says.
	mit, err := os.ReadFile(filepath.Join(moduleRoot(t), "LICENSE-MIT"))
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Split(string(mit), "\n"); len(lines) < 3 || lines[2] != "Copyright (c) 2026 "+wanted {
		t.Fatalf("LICENSE-MIT's copyright line is not %q:\n%s", "Copyright (c) 2026 "+wanted, mit)
	}

	manifests := map[string]map[string]any{
		DispatcherPackageName: decodeManifest(t, mustDispatcherManifest(t, "1.2.3")),
	}
	for _, target := range Targets {
		manifests[target.PackageName()] = decodeManifest(t, mustPlatformManifest(t, target, "1.2.3"))
	}
	for name, manifest := range manifests {
		if manifest["author"] != wanted {
			t.Errorf("%s names its author as %v, wanted %q", name, manifest["author"], wanted)
		}
	}

	contents, err := os.ReadFile(filepath.Join(moduleRoot(t), "editors", "vscode", "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var extension struct {
		Author string `json:"author"`
	}
	if err := json.Unmarshal(contents, &extension); err != nil {
		t.Fatal(err)
	}
	if extension.Author != wanted {
		t.Errorf("the VS Code extension names its author as %q, wanted %q", extension.Author, wanted)
	}
}

// mitTerms is an MIT license from its first paragraph on, past the title and the copyright line.
func mitTerms(license []byte) string {
	_, terms, found := strings.Cut(string(license), "Permission is hereby granted")
	if !found {
		return ""
	}
	return terms
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	module, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return module
}
