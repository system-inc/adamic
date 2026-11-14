// adamic-apple-bindings generates bindings using the selected Apple SDK's clang.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	if headerRoot == "" {
		output, err := generate.FromSDK(context.Background(), sdk, platform, names)
		if err != nil {
			return err
		}
		return output.Write(directory)
	}
	// Fixture headers, outside any SDK.
	if platform == "" {
		platform = "macos"
	}
	configuration, err := generate.FrameworkConfiguration(headerRoot, platform, names)
	if err != nil {
		return err
	}
	if umbrella == "" {
		file, err := os.CreateTemp("", "adamic-apple-*.h")
		if err != nil {
			return err
		}
		defer os.Remove(file.Name())
		for _, name := range names {
			if _, err := fmt.Fprintf(file, "#import <%s/%s.h>\n", name, name); err != nil {
				file.Close()
				return err
			}
		}
		if err := file.Close(); err != nil {
			return err
		}
		umbrella = file.Name()
	} else {
		configuration.Umbrella = filepath.Base(umbrella)
	}
	clangArguments := append(generate.ClangArguments(), "-fobjc-runtime=macosx-10.13", "-F", headerRoot, umbrella)
	output, err := generate.RunClang(context.Background(), clang, clangArguments, configuration)
	if err != nil {
		return err
	}
	return output.Write(directory)
}
