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
	goruntime "runtime"
	"strings"
)

//go:embed runtime/*.c runtime/*.h
var runtime embed.FS

// cString is a C expression for a NUL-terminated string holding exactly value's bytes: a string
// literal, or, for one longer than C11 promises a literal can be (longestLiteral; a field name of
// 5,000 bytes, say), a compound literal of its bytes, which has static storage at file scope and
// lives through the call it's an argument to inside a function. A message array's initializer is
// cArray instead.
//
// Every byte outside printable ASCII becomes a three-digit octal escape, which can't run into the
// character after it the way \x can, and quote, backslash and question mark are escaped too, the last
// so no ?? can ever read as a trigraph.
func cString(value string) string {
	if len(value) > longestLiteral {
		return "((const char[]){" + cBytes(value) + ", 0})"
	}
	return cLiteral(value)
}

// cArray is what initializes a char array to value's bytes and a terminator, as
// static const char message[] = ...; does: the literal, or the bytes when it would be too long.
func cArray(value string) string {
	if len(value) > longestLiteral {
		return "{" + cBytes(value) + ", 0}"
	}
	return cLiteral(value)
}

// cLiteral is the string literal itself, of any length.
func cLiteral(value string) string {
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
	// Release selects the native build people ship. Tests leave it false.
	// Sanitized, counted and WASI builds retain their existing compilation policy.
	Release bool

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

	// ThreadSanitize is a separate race-check build, never combined with ASan.
	ThreadSanitize bool

	// Count makes the binary count its allocations, frees, retains and releases, and write them to
	// stderr as it exits (runtime/count.h). Only a counted build does; the counts table is made of them.
	Count bool

	// cpu, for tests, compiles for a particular processor (-march), so a test on an x86 machine can
	// see what fused multiply-adds would do, as on arm64.
	cpu string

	// Slabs keeps the size-class allocator on in a sanitized build (heap.c), where every value
	// otherwise comes from malloc: the oracle runs every fixture this way too, so the classes
	// themselves run under the sanitizers, and a use after a free is still caught with them on.
	Slabs bool

	// Malloc takes every value from malloc in a build without sanitizers, as a sanitized build does.
	// macOS's leaks tool needs it: a chunk of the size classes stays reachable from the runtime's own
	// table, so a value leaked into one is never reported.
	Malloc bool
}

// Flags are what clang compiles a program and the runtime with. The fuzzer (internal/fuzz) compiles
// with the same ones, so what it finds is what Build would. Build also passes the complete
// list to the link, where ThinLTO generates code. See docs/native-builds.md.
func Flags(options Options) []string {
	// A program may declare a variable, a function or a parameter it never uses, or assign a variable
	// to itself, as JavaScript allows; that's the linter's business (cohere's no-unused-vars and
	// no-self-assign), not a reason the C can't compile.
	flags := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-Wno-unused-variable", "-Wno-unused-but-set-variable", "-Wno-unused-function", "-Wno-unused-parameter", "-Wno-self-assign"}
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
	} else {
		flags = append(flags, "-pthread")
	}
	if options.Count {
		flags = append(flags, "-DADAMIC_COUNT")
	}
	if options.Slabs {
		flags = append(flags, "-DADAMIC_SLABS")
	}
	if options.Malloc {
		flags = append(flags, "-DADAMIC_MALLOC")
	}
	if options.cpu != "" {
		flags = append(flags, "-march="+options.cpu)
	}
	if options.Sanitize && options.ThreadSanitize {
		panic("native: ASan and TSan cannot be combined")
	}
	if options.ThreadSanitize {
		return append(flags, "-O1", "-g", "-fsanitize=thread", "-DADAMIC_TSAN_TEST")
	}
	if options.Sanitize {
		return append(flags, "-O1", "-g", "-fsanitize=address,undefined", "-fno-sanitize-recover=all")
	}
	if options.Target == "wasm32-wasi" {
		return append(flags, "-Oz")
	}
	flags = append(flags, "-O2")
	if shippedRelease(options) {
		flags = append(flags, "-flto=thin")
	}
	return flags
}

func shippedRelease(options Options) bool {
	return options.Release && !options.Sanitize && !options.Count && options.Target == ""
}

// LinkFlags retains every compilation flag, especially -ffp-contract=off and
// -fno-optimize-sibling-calls, when ThinLTO performs code generation at the link.
// Linker selection belongs here, never in the runtime's warning-strict clang -c.
func LinkFlags(options Options) []string {
	flags := Flags(options)
	if shippedRelease(options) && goruntime.GOOS != "darwin" {
		flags = append(flags, "-fuse-ld=lld")
	}
	return flags
}

// Build compiles C source and the runtime into a native binary at output.
func Build(source string, output string, options Options) error {
	if err := ValidateOptions(options); err != nil {
		return err
	}
	if options.Target == "" && (options.Split || os.Getenv("ADAMIC_NATIVE_SPLIT") == "1") {
		return buildUnits(source, output, options)
	}
	directory, err := os.MkdirTemp("", "adamic-build-")
	if err != nil {
		return fmt.Errorf("native: %w", err)
	}
	defer os.RemoveAll(directory)

	library, err := RuntimeLibrary("", options)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644); err != nil {
		return fmt.Errorf("native: %w", err)
	}
	arguments := append(LinkFlags(options), "-I", filepath.Dir(library), "-o", output)
	if !options.Request {
		arguments = append(arguments, filepath.Join(directory, "main.c"))
	}
	if options.Target == "wasm32-wasi" {
		arguments = append(arguments, "-Xlinker", "--whole-archive", library, "-Xlinker", "--no-whole-archive")
	} else {
		arguments = append(arguments, RuntimeLinkFlags(library)...)
	}
	// The runtime calls libm (trunc, floor, sqrt). On macOS that's part of libSystem and comes free; on
	// Linux it's its own library, and only the sanitizers' runtime happened to pull it in.
	if options.Request {
		// Runtime constructors precede module initialization at the same default priority.
		arguments = append(arguments, filepath.Join(directory, "main.c"))
	}
	arguments = append(arguments, "-lm")
	if options.Target == "wasm32-wasi" {
		arguments = append(arguments, WASILinkFlags(options)...)
	}
	command := exec.Command(compilerName(options), arguments...)
	if combined, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("native: clang failed: %w\n%s", err, combined)
	}
	return nil
}
