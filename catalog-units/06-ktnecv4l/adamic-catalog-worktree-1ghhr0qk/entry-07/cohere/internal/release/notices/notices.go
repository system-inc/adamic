// Package notices writes cohere's third-party notices: NOTICE, which carries the TypeScript compiler's
// Apache-2.0 notice, and THIRD_PARTY_NOTICES.md, which credits every other project cohere is built on
// with its license in full.
//
// Both are generated, never edited by hand, because a hand-kept list drifts the moment a dependency
// moves, and a stale notice is a license obligation quietly unmet. The Go modules come from the build
// graph of ./command/cohere on every release target, which is what actually ships, rather than from
// go.mod, which lists modules only tests use and misses ones that arrive through the vendored compiler's
// own go.mod. The projects whose code was ported into cohere's own source can't be read from any graph,
// so they are a curated list (upstreams.go) with each license text kept beside it, taken at the version
// that was ported.
//
// Regenerate with:
//
//	go run ./internal/release/tools/notices
package notices

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/system-inc/cohere/internal/release/packaging"
)

// NoticeFileName and ThirdPartyNoticesFileName are written at the module root and shipped in every
// package beside the two license texts.
const (
	NoticeFileName            = release.NoticeFileName
	ThirdPartyNoticesFileName = release.ThirdPartyNoticesFileName
)

// mainModule is cohere itself, which credits no one by being linked.
const mainModule = "github.com/system-inc/cohere"

// isCohere reports whether a module is cohere's own: the main module, or one beside it in the repository
// under its path, as static_single_assignment is, a module of its own so another repository can import it.
func isCohere(path string) bool {
	return path == mainModule || strings.HasPrefix(path, mainModule+"/")
}

// compilerModulePrefix is the vendored TypeScript compiler and the shim modules over it. The compiler is
// credited in NOTICE and its own entry, and the shims are cohere's generated code.
const compilerModulePrefix = "github.com/microsoft/TypeScript/"

// licenseTexts are the curated upstreams' license texts, one file each, named in upstreams.go.
//
//go:embed licenses
var licenseTexts embed.FS

// Module is a Go module compiled into cohere, or required by it for tests and tools only.
type Module struct {
	Path      string
	Version   string
	Directory string

	// Targets are the release targets that link it, empty when that is every one or when nothing does.
	Targets []string
}

// Files are the two generated files, by the name each is written under.
type Files map[string][]byte

// Generate renders both files for the module at moduleDirectory.
func Generate(moduleDirectory string) (Files, error) {
	compiled, err := CompiledModules(moduleDirectory)
	if err != nil {
		return nil, err
	}
	repositoryOnly, err := repositoryOnlyModules(moduleDirectory, compiled)
	if err != nil {
		return nil, err
	}
	swiftPins, err := readSwiftPins(moduleDirectory)
	if err != nil {
		return nil, err
	}

	notice, err := renderNotice(moduleDirectory)
	if err != nil {
		return nil, err
	}
	thirdParty, err := renderThirdPartyNotices(moduleDirectory, compiled, repositoryOnly, swiftPins)
	if err != nil {
		return nil, err
	}
	return Files{NoticeFileName: notice, ThirdPartyNoticesFileName: thirdParty}, nil
}

// CompiledModules are the third-party Go modules in cohere's build graph, on any release target.
//
// Every target, because the graph differs by platform: golang.org/x/sys carries different files for
// Windows, and a module only one platform links would otherwise ship there uncredited.
func CompiledModules(moduleDirectory string) ([]Module, error) {
	byPath := map[string]Module{}
	for _, target := range release.Targets {
		command := exec.Command("go", "list", "-deps", "-f",
			"{{with .Module}}{{.Path}}\t{{.Version}}\t{{.Dir}}{{end}}", "./command/cohere")
		command.Dir = moduleDirectory
		command.Env = append(os.Environ(), "GOOS="+target.GoOperatingSystem, "GOARCH="+target.GoArchitecture, "CGO_ENABLED=0")
		var standardError bytes.Buffer
		command.Stderr = &standardError
		output, err := command.Output()
		if err != nil {
			return nil, fmt.Errorf("listing cohere's build graph for %s: %w\n%s", target, err, standardError.String())
		}
		for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
			fields := strings.Split(line, "\t")
			if len(fields) != 3 || isCohere(fields[0]) || strings.HasPrefix(fields[0], compilerModulePrefix) {
				continue
			}
			module := byPath[fields[0]]
			module.Path, module.Version, module.Directory = fields[0], fields[1], fields[2]
			if !slices.Contains(module.Targets, target.String()) {
				module.Targets = append(module.Targets, target.String())
			}
			byPath[fields[0]] = module
		}
	}
	for path, module := range byPath {
		if len(module.Targets) == len(release.Targets) {
			module.Targets = nil
			byPath[path] = module
		}
	}
	return sortedModules(byPath), nil
}

// repositoryOnlyModules are the modules go.mod requires that no release target compiles in: the ones
// only tests and tools use. They ship in no package, and are credited because the repository holds code
// that imports them.
func repositoryOnlyModules(moduleDirectory string, compiled []Module) ([]Module, error) {
	command := exec.Command("go", "mod", "edit", "-json")
	command.Dir = moduleDirectory
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}
	var goMod struct {
		Require []struct{ Path, Version string }
	}
	if err := json.Unmarshal(output, &goMod); err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}

	byPath := map[string]Module{}
	for _, requirement := range goMod.Require {
		isCompiled := slices.ContainsFunc(compiled, func(module Module) bool { return module.Path == requirement.Path })
		if isCompiled || isCohere(requirement.Path) || strings.HasPrefix(requirement.Path, compilerModulePrefix) {
			continue
		}
		download := exec.Command("go", "mod", "download", "-json", requirement.Path+"@"+requirement.Version)
		download.Dir = moduleDirectory
		output, err := download.Output()
		if err != nil {
			return nil, fmt.Errorf("finding %s@%s to read its license: %w", requirement.Path, requirement.Version, err)
		}
		var downloaded struct{ Dir string }
		if err := json.Unmarshal(output, &downloaded); err != nil {
			return nil, err
		}
		byPath[requirement.Path] = Module{Path: requirement.Path, Version: requirement.Version, Directory: downloaded.Dir}
	}
	return sortedModules(byPath), nil
}

func sortedModules(byPath map[string]Module) []Module {
	modules := make([]Module, 0, len(byPath))
	for _, module := range byPath {
		modules = append(modules, module)
	}
	slices.SortFunc(modules, func(left Module, right Module) int { return strings.Compare(left.Path, right.Path) })
	return modules
}

// moduleLicenseFileNames are the names a Go module's license is found under, in the order they are tried.
var moduleLicenseFileNames = []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "LICENCE", "COPYING"}

// readModuleLicense reads a module's license from its directory, refusing one that has none: a module
// with no license file can't be credited, and that is a question for a person, not something to skip.
func readModuleLicense(module Module) (string, error) {
	for _, name := range moduleLicenseFileNames {
		contents, err := os.ReadFile(filepath.Join(module.Directory, name))
		if err == nil {
			return string(contents), nil
		}
	}
	return "", fmt.Errorf("%s@%s has no license file in %s (looked for %s)", module.Path, module.Version, module.Directory, strings.Join(moduleLicenseFileNames, ", "))
}

// swiftPin is one package the Swift engine resolves, from swift/Package.resolved.
type swiftPin struct {
	Identity string
	Location string
	Version  string
}

func readSwiftPins(moduleDirectory string) ([]swiftPin, error) {
	contents, err := os.ReadFile(filepath.Join(moduleDirectory, "swift", "Package.resolved"))
	if err != nil {
		return nil, fmt.Errorf("reading the Swift engine's resolved packages: %w", err)
	}
	var resolved struct {
		Pins []struct {
			Identity string
			Location string
			State    struct{ Version, Revision string }
		}
	}
	if err := json.Unmarshal(contents, &resolved); err != nil {
		return nil, fmt.Errorf("reading swift/Package.resolved: %w", err)
	}
	pins := make([]swiftPin, 0, len(resolved.Pins))
	for _, pin := range resolved.Pins {
		version := pin.State.Version
		if version == "" {
			version = pin.State.Revision
		}
		pins = append(pins, swiftPin{Identity: pin.Identity, Location: strings.TrimSuffix(pin.Location, ".git"), Version: version})
	}
	slices.SortFunc(pins, func(left swiftPin, right swiftPin) int { return strings.Compare(left.Identity, right.Identity) })
	return pins, nil
}
