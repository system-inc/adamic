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
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
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
	// Target is empty for native, or wasm32-wasi for a WASI module.
	Target string

	// Request selects a WASI reactor containing the emitted request ABI.
	Request bool

	// Split compiles generated functions in separate translation units. Off by default, and only
	// for native targets.
	Split bool

	// Jobs bounds parallel clang processes in a split build; zero means one.
	Jobs int

	// Sanitize compiles with the address and undefined-behavior sanitizers, as the tests do.
	Sanitize bool

	// Count makes the binary count its allocations, frees, retains and releases, and write them to
	// stderr as it exits (runtime/count.h). Only a counted build does; the counts table is made of them.
	Count bool

	// cpu, for tests, compiles for a particular processor (-march), so a test on an x86 machine can
	// see what fused multiply-adds would do, as on arm64.
	cpu string

	// slabs, for tests, keeps the size-class allocator on in a sanitized build (heap.c), where every
	// value otherwise comes from malloc, to show a use after a free is still caught with it on.
	slabs bool
}

// Flags are what clang compiles a program and the runtime with. The fuzzer (internal/fuzz) compiles
// with the same ones, so what it finds is what Build would.
func Flags(options Options) []string {
	// A program may declare a variable, a function or a parameter it never uses, or assign a variable
	// to itself, as JavaScript allows; that's the linter's business (cohere's no-unused-vars and
	// no-self-assign), not a reason the C can't compile.
	flags := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-Wcast-function-type-strict", "-pedantic", "-Wno-unused-variable", "-Wno-unused-but-set-variable", "-Wno-unused-function", "-Wno-unused-parameter", "-Wno-self-assign"}
	// JavaScript rounds every operation on its own. clang otherwise fuses a * b + c into one
	// multiply-add wherever the processor has one (every arm64, so every Apple silicon Mac), and
	// 0.1 * 10 - 1 is then 5.551115123125783e-17 instead of 0. V8 builds itself the same way.
	flags = append(flags, "-ffp-contract=off")
	// Every function checks its frame against the stack's limit (stack.c), and a call in tail position
	// that clang turns into a jump never makes a frame: a self tail call becomes a loop, and recursion
	// with no end runs forever where Node's runs out of stack. Every call keeps its frame, as V8's do.
	flags = append(flags, "-fno-optimize-sibling-calls")
	if options.Target == "wasm32-wasi" {
		flags = append(flags, "--target=wasm32-wasi", "--sysroot="+os.Getenv("WASI_SYSROOT"), "-DADAMIC_TARGET_WASI=1", "-mno-atomics")
	}
	if options.Count {
		flags = append(flags, "-DADAMIC_COUNT")
	}
	if options.slabs {
		flags = append(flags, "-DADAMIC_SLABS")
	}
	if options.cpu != "" {
		flags = append(flags, "-march="+options.cpu)
	}
	if options.Sanitize {
		return append(flags, "-O1", "-g", "-fsanitize=address,undefined", "-fno-sanitize-recover=all")
	}
	return append(flags, "-O2")
}

// Build compiles C source and the runtime into a native binary at output.
func Build(source string, output string, options Options) error {
	if err := ValidateOptions(options); err != nil {
		return err
	}
	if options.Target == "" && (options.Split || os.Getenv("ADAMIC_NATIVE_SPLIT") == "1") {
		return buildUnits(source, output, options)
	}
	return buildOne(source, output, options, Flags(options))
}

// BuildFlags is Build with the compiler flags the caller gives in place of Flags(options), in one translation unit
// (markdownblocks' mutant canary, at -O0): the binary names no build directory, as Build's doesn't.
func BuildFlags(source string, output string, options Options, flags []string) error {
	if err := ValidateOptions(options); err != nil {
		return err
	}
	return buildOne(source, output, options, flags)
}

func buildOne(source string, output string, options Options, flags []string) error {
	directory, err := os.MkdirTemp("", "adamic-build-")
	if err != nil {
		return fmt.Errorf("native: %w", err)
	}
	defer os.RemoveAll(directory)

	library, err := RuntimeLibraryForSource("", source, options)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644); err != nil {
		return fmt.Errorf("native: %w", err)
	}
	// Neither the build directory nor this machine's runtime cache may reach the binary, so its bytes are a
	// function of its inputs wherever it's built (see cachedRuntime). clang is never given an absolute path to read: it
	// compiles main.c from inside the build directory, and finds the runtime's headers and archive through runtime, a link
	// there to the cache. The prefix maps cover debug info and __FILE__, but ASan's metadata for globals keeps each source
	// file as clang was given it (Workshop, Oct 11: <tmp>/adamic-build-<random>/main.c and <cache>/.../adamic.h in .rodata).
	if err := os.Symlink(filepath.Dir(library), filepath.Join(directory, "runtime")); err != nil {
		return fmt.Errorf("native: %w", err)
	}
	runtimeArchive := filepath.Join("runtime", filepath.Base(library))
	compileFlags := append(append(append(slices.Clone(flags), prefixMaps(directory, "/adamic-build")...), prefixMaps(filepath.Dir(library), "/adamic-runtime")...), "-I", "runtime")
	// On macOS ld's debug map records every object it links that carries debug info (a sanitized build's do) by its path,
	// and a compile and link in one call links main.c's object from clang's own temporary file, <TMPDIR>/main-<random>.o,
	// random per build (Tests, from Compiler, Oct 12: a whole sanitized build's binary named it). There main.c is
	// compiled to main.o in the build directory first and linked by that name, -oso_prefix taking the directory off, so
	// the debug map names it main.o alone, as a split build's names its units. An ELF linker records no input's path.
	program := "main.c"
	if goruntime.GOOS == "darwin" && options.Target != "wasm32-wasi" {
		compile := toolchain.CommandIn(directory, compilerName(options), append(slices.Clone(compileFlags), "-c", "main.c", "-o", "main.o")...)
		if combined, err := compile.CombinedOutput(); err != nil {
			return fmt.Errorf("native: clang failed: %w\n%s", err, combined)
		}
		// The debug map records each object's modification time beside its name, so main.o's is the epoch's, as
		// ZERO_AR_DATE has ar give an archive member's: a fresh main.o's would differ per build.
		if err := os.Chtimes(filepath.Join(directory, "main.o"), time.Unix(0, 0), time.Unix(0, 0)); err != nil {
			return fmt.Errorf("native: %w", err)
		}
		program = "main.o"
	}
	compilerLink, err := regexpCompilerLinkFlags(source, options, directory)
	if err != nil {
		return err
	}
	arguments := append(compileFlags, "-o", output)
	if !options.Request {
		arguments = append(arguments, program)
	}
	if options.Target == "wasm32-wasi" {
		arguments = append(arguments, "-Xlinker", "--whole-archive", runtimeArchive, "-Xlinker", "--no-whole-archive")
	} else {
		arguments = append(arguments, RuntimeLinkFlags(runtimeArchive)...)
	}
	// The runtime calls libm (trunc, floor, sqrt). On macOS that's part of libSystem and comes free; on
	// Linux it's its own library, and only the sanitizers' runtime happened to pull it in.
	if options.Request {
		// Runtime constructors precede module initialization at the same default priority.
		arguments = append(arguments, program)
	}
	arguments = append(arguments, compilerLink...)
	if program == "main.o" {
		arguments = append(arguments, objectPrefixFlags()...)
	}
	arguments = append(arguments, "-lm")
	if options.Target == "wasm32-wasi" {
		arguments = append(arguments, WASILinkFlags(options)...)
	}
	command := toolchain.CommandIn(directory, compilerName(options), arguments...)
	if combined, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("native: clang failed: %w\n%s", err, combined)
	}
	return nil
}
