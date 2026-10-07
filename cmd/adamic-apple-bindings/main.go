// adamic-apple-bindings generates bindings using the selected Apple SDK's clang.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/system-inc/adamic/internal/apple/generate"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "adamic-apple-bindings:", err)
		os.Exit(1)
	}
}
func run(arguments []string) error {
	directory := ""
	sdk := "macosx"
	platform := ""
	clang := "clang"
	umbrella := ""
	headerRoot := ""
	names := []string{}
	// Accept flags after framework operands, as documented by the command's API.
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]
		switch arg {
		case "-o", "-sdk", "-platform", "-clang", "-umbrella", "-headers":
			if i+1 == len(arguments) {
				return fmt.Errorf("%s needs a value", arg)
			}
			i++
			value := arguments[i]
			switch arg {
			case "-o":
				directory = value
			case "-sdk":
				sdk = value
			case "-platform":
				platform = value
			case "-clang":
				clang = value
			case "-umbrella":
				umbrella = value
			case "-headers":
				headerRoot = value
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag %s", arg)
			}
			names = append(names, arg)
		}
	}
	if directory == "" || len(names) == 0 {
		return fmt.Errorf("usage: adamic-apple-bindings <framework>... -o <directory> [-sdk macosx] [-platform macos] [-umbrella file -headers framework-directory]")
	}
	if platform == "" {
		switch sdk {
		case "macosx":
			platform = "macos"
		case "iphoneos", "iphonesimulator":
			platform = "ios"
		case "appletvos", "appletvsimulator":
			platform = "tvos"
		case "watchos", "watchsimulator":
			platform = "watchos"
		case "xros", "xrsimulator":
			platform = "visionos"
		default:
			return fmt.Errorf("unknown SDK %s: specify -platform", sdk)
		}
	}
	configuration := generate.Configuration{Platform: platform}
	clangArguments := []string{"-x", "objective-c", "-fsyntax-only", "-fblocks", "-Werror", "-Wno-nullability-completeness", "-Xclang", "-ast-dump=json"}
	if headerRoot == "" {
		data, err := exec.Command("xcrun", "--sdk", sdk, "--show-sdk-path").Output()
		if err != nil {
			return fmt.Errorf("Apple SDK lookup: %w; fixtures can use -headers and -umbrella", err)
		}
		sdkPath := strings.TrimSpace(string(data))
		headerRoot = filepath.Join(sdkPath, "System", "Library", "Frameworks")
		versionData, versionError := exec.Command("xcrun", "--sdk", sdk, "--show-sdk-version").Output()
		if versionError != nil {
			return fmt.Errorf("Apple SDK version lookup: %w", versionError)
		}
		version := strings.TrimSpace(string(versionData))
		if !regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)*$`).MatchString(version) {
			return fmt.Errorf("invalid SDK version %q", version)
		}
		operatingSystem := map[string]string{"macos": "macos", "ios": "ios", "tvos": "tvos", "watchos": "watchos", "visionos": "xros"}[platform]
		if operatingSystem == "" {
			return fmt.Errorf("unsupported target platform %q", platform)
		}
		target := "arm64-apple-" + operatingSystem + version
		if strings.Contains(sdk, "simulator") {
			target += "-simulator"
		}
		clangArguments = append(clangArguments, "-isysroot", sdkPath, "-target", target)
		data, err = exec.Command("xcrun", "--sdk", sdk, "--find", "clang").Output()
		if err != nil {
			return err
		}
		if clang == "clang" {
			clang = strings.TrimSpace(string(data))
		}
	} else {
		clangArguments = append(clangArguments, "-fobjc-runtime=macosx-10.13")
	}
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			return fmt.Errorf("duplicate framework %s", name)
		}
		seen[name] = true
		if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`).MatchString(name) {
			return fmt.Errorf("invalid framework name %q", name)
		}
		headers := filepath.Join(headerRoot, name+".framework", "Headers")
		info, err := os.Stat(headers)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("not a header directory: %s", headers)
		}
		configuration.Frameworks = append(configuration.Frameworks, generate.Framework{Name: name, Headers: headers})
	}
	temporary := ""
	if umbrella == "" {
		file, err := os.CreateTemp("", "adamic-apple-*.h")
		if err != nil {
			return err
		}
		temporary = file.Name()
		defer os.Remove(temporary)
		for _, name := range names {
			if _, err := fmt.Fprintf(file, "#import <%s/%s.h>\n", name, name); err != nil {
				file.Close()
				return err
			}
		}
		if err := file.Close(); err != nil {
			return err
		}
		umbrella = temporary
	} else {
		configuration.Umbrella = filepath.Base(umbrella)
	}
	clangArguments = append(clangArguments, "-F", headerRoot, umbrella)
	output, err := generate.RunClang(context.Background(), clang, clangArguments, configuration)
	if err != nil {
		return err
	}
	return output.Write(directory)
}
