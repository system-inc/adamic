package generate

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/apple/naming"
)

// The generator's own sources, so Version changes whenever what it would write could.
//
//go:embed *.go
var sources embed.FS

// Version names this generator: a hash of its sources and the naming layer's. A cache of what it
// wrote is good only for the Version that wrote it.
func Version() string {
	hash := sha256.New()
	for _, files := range []fs.FS{sources, naming.Sources} {
		names, _ := fs.Glob(files, "*.go")
		sort.Strings(names)
		for _, name := range names {
			content, _ := fs.ReadFile(files, name)
			fmt.Fprintf(hash, "%s %d\n", name, len(content))
			hash.Write(content)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))[:16]
}

// SDKBuild is the build of Apple's SDK selected by sdk (macosx, iphoneos, ...), as Xcode names it.
func SDKBuild(sdk string) (string, error) {
	data, err := exec.Command("xcrun", "--sdk", sdk, "--show-sdk-build-version").Output()
	if err != nil {
		return "", fmt.Errorf("Apple's SDK %s: %w", sdk, err)
	}
	build := strings.TrimSpace(string(data))
	if !regexp.MustCompile(`^[A-Za-z0-9.]+$`).MatchString(build) {
		return "", fmt.Errorf("Apple's SDK %s has an unreadable build %q", sdk, build)
	}
	return build, nil
}

// FromSDK generates bindings for frameworks from Apple's SDK selected by sdk, through the SDK's own
// clang, for platform (macos, ios, ...; empty takes the SDK's).
func FromSDK(ctx context.Context, sdk, platform string, frameworks []string) (Output, error) {
	if platform == "" {
		platform = map[string]string{"macosx": "macos", "iphoneos": "ios", "iphonesimulator": "ios", "appletvos": "tvos", "appletvsimulator": "tvos", "watchos": "watchos", "watchsimulator": "watchos", "xros": "visionos", "xrsimulator": "visionos"}[sdk]
		if platform == "" {
			return Output{}, fmt.Errorf("unknown SDK %s: name its platform", sdk)
		}
	}
	data, err := exec.Command("xcrun", "--sdk", sdk, "--show-sdk-path").Output()
	if err != nil {
		return Output{}, fmt.Errorf("Apple's SDK %s: %w", sdk, err)
	}
	sdkPath := strings.TrimSpace(string(data))
	versionData, err := exec.Command("xcrun", "--sdk", sdk, "--show-sdk-version").Output()
	if err != nil {
		return Output{}, fmt.Errorf("Apple's SDK %s version: %w", sdk, err)
	}
	version := strings.TrimSpace(string(versionData))
	if !regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)*$`).MatchString(version) {
		return Output{}, fmt.Errorf("invalid SDK version %q", version)
	}
	operatingSystem := map[string]string{"macos": "macos", "ios": "ios", "tvos": "tvos", "watchos": "watchos", "visionos": "xros"}[platform]
	if operatingSystem == "" {
		return Output{}, fmt.Errorf("unsupported target platform %q", platform)
	}
	target := "arm64-apple-" + operatingSystem + version
	if strings.Contains(sdk, "simulator") {
		target += "-simulator"
	}
	clangData, err := exec.Command("xcrun", "--sdk", sdk, "--find", "clang").Output()
	if err != nil {
		return Output{}, err
	}
	headerRoot := filepath.Join(sdkPath, "System", "Library", "Frameworks")
	configuration, err := FrameworkConfiguration(headerRoot, platform, frameworks)
	if err != nil {
		return Output{}, err
	}
	umbrella, err := os.CreateTemp("", "adamic-apple-*.h")
	if err != nil {
		return Output{}, err
	}
	defer os.Remove(umbrella.Name())
	for _, name := range frameworks {
		if _, err := fmt.Fprintf(umbrella, "#import <%s/%s.h>\n", name, name); err != nil {
			umbrella.Close()
			return Output{}, err
		}
	}
	if err := umbrella.Close(); err != nil {
		return Output{}, err
	}
	arguments := append(ClangArguments(), "-isysroot", sdkPath, "-target", target, "-F", headerRoot, umbrella.Name())
	return RunClang(ctx, strings.TrimSpace(string(clangData)), arguments, configuration)
}

// ClangArguments are the arguments every run of clang for the generator starts with.
func ClangArguments() []string {
	return []string{"-x", "objective-c", "-fsyntax-only", "-fblocks", "-Werror", "-Wno-nullability-completeness", "-Xclang", "-ast-dump=json"}
}

// FrameworkConfiguration names each framework's public headers below headerRoot, refusing a name
// twice or a framework with no headers.
func FrameworkConfiguration(headerRoot, platform string, frameworks []string) (Configuration, error) {
	configuration := Configuration{Platform: platform}
	seen := map[string]bool{}
	for _, name := range frameworks {
		if seen[name] {
			return Configuration{}, fmt.Errorf("duplicate framework %s", name)
		}
		seen[name] = true
		if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`).MatchString(name) {
			return Configuration{}, fmt.Errorf("invalid framework name %q", name)
		}
		headers := filepath.Join(headerRoot, name+".framework", "Headers")
		info, err := os.Stat(headers)
		if err != nil {
			return Configuration{}, err
		}
		if !info.IsDir() {
			return Configuration{}, fmt.Errorf("not a header directory: %s", headers)
		}
		configuration.Frameworks = append(configuration.Frameworks, Framework{Name: name, Headers: headers})
	}
	return configuration, nil
}

// CheckSDK compiles the check file written into directory against Apple's SDK selected by sdk,
// in the bridge's own mode, so a binding whose tag disagrees with its header never reaches a
// program.
func CheckSDK(ctx context.Context, sdk, directory string) error {
	command := exec.CommandContext(ctx, "xcrun", "--sdk", sdk, "clang", "-x", "objective-c", "-fsyntax-only", "-fno-objc-arc", "-Werror", "-Wno-deprecated-declarations", "-fmodules", "bindings-check.m")
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 10 {
			lines = lines[:10]
		}
		return fmt.Errorf("the generated bindings disagree with Apple's headers: %w\n%s", err, strings.Join(lines, "\n"))
	}
	return nil
}
