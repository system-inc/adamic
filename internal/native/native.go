// Package native compiles Adamic's IR to C, and C to a native binary with clang.
//
// C is the first backend because clang already runs everywhere Adamic means to: Xcode's clang for
// macOS and iOS, the NDK's for Android, and decades of optimization come with it. The C is written
// for -Wall -Wextra -Werror, and the tests compile it under the address and undefined-behavior
// sanitizers.
package native

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed runtime/*.c runtime/*.h
var runtime embed.FS

// cString is a C string literal holding exactly value's bytes.
//
// Every byte outside printable ASCII becomes a three-digit octal escape, which can't run into the
// character after it the way \x can, and quote, backslash and question mark are escaped too, the last
// so no ?? can ever read as a trigraph.
func cString(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for index := 0; index < len(value); index++ {
		character := value[index]
		switch {
		case character == '"' || character == '\\' || character == '?':
			builder.WriteByte('\\')
			builder.WriteByte(character)
		case character >= 0x20 && character < 0x7f:
			builder.WriteByte(character)
		default:
			fmt.Fprintf(&builder, "\\%03o", character)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

// Options says how to compile.
type Options struct {
	// Sanitize compiles with the address and undefined-behavior sanitizers, as the tests do.
	Sanitize bool
}

// Build compiles C source and the runtime into a native binary at output.
func Build(source string, output string, options Options) error {
	directory, err := os.MkdirTemp("", "adamic-build-")
	if err != nil {
		return fmt.Errorf("native: %w", err)
	}
	defer os.RemoveAll(directory)

	// The runtime is every file in runtime/, written beside the program and compiled with it.
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644); err != nil {
		return fmt.Errorf("native: %w", err)
	}
	units := []string{filepath.Join(directory, "main.c")}
	entries, err := runtime.ReadDir("runtime")
	if err != nil {
		return fmt.Errorf("native: %w", err)
	}
	for _, entry := range entries {
		contents, err := runtime.ReadFile("runtime/" + entry.Name())
		if err != nil {
			return fmt.Errorf("native: %w", err)
		}
		path := filepath.Join(directory, entry.Name())
		if err := os.WriteFile(path, contents, 0o644); err != nil {
			return fmt.Errorf("native: %w", err)
		}
		if strings.HasSuffix(entry.Name(), ".c") {
			units = append(units, path)
		}
	}

	// A program may declare a variable or a function it never uses, as JavaScript allows; that's the
	// linter's business (cohere's no-unused-vars), not a reason the C can't compile.
	arguments := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-Wno-unused-variable", "-Wno-unused-but-set-variable", "-Wno-unused-function"}
	if options.Sanitize {
		arguments = append(arguments, "-O1", "-g", "-fsanitize=address,undefined", "-fno-sanitize-recover=all")
	} else {
		arguments = append(arguments, "-O2")
	}
	arguments = append(arguments, "-o", output)
	arguments = append(arguments, units...)
	// The runtime calls libm (trunc, floor, sqrt). On macOS that's part of libSystem and comes free; on
	// Linux it's its own library, and only the sanitizers' runtime happened to pull it in.
	arguments = append(arguments, "-lm")
	command := exec.Command("clang", arguments...)
	if combined, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("native: clang failed: %w\n%s", err, combined)
	}
	return nil
}
