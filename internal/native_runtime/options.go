package native_runtime

import "os"

// Options says how to compile.
type Options struct {
	// Debug emits source locations and asks clang for debug information.
	Debug bool
	// DebugDirectory is the entry program's absolute directory, used only by its own compile.
	DebugDirectory string
	// Release enables ThinLTO only for shipped native, uncounted, unsanitized builds.
	Release bool
	// ProgramRegion enables the one-shot CLI Program region.
	ProgramRegion bool

	// Target is the triple clang compiles for: a native one of Triple's table, or wasm32-wasi for a WASI module. Empty
	// is the build's own native triple (ResolveTarget), which every flag and key names in its place.
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

	// ThreadSanitize enables the separate race-check build.
	ThreadSanitize bool

	// Count makes the binary count its allocations, frees, retains and releases, and write them to
	// stderr as it exits (runtime/count.h). Only a counted build does; the counts table is made of them.
	Count bool

	// CPU, Contract and Slabs are exported only so internal/native's tests and this package's can set them; every build a
	// product, a program or the command asks for leaves them unset. A setter outside a test is a decision about what the
	// archives are built with, so it's said on its task, never made in passing (Runtime's should on #tb722h2).

	// CPU compiles for a particular processor (-march). Tests set it, so a test on an x86 machine can see what fused
	// multiply-adds would do, as on arm64.
	CPU string

	// Contract lets clang fuse multiply-adds (-ffp-contract=fast, which disregards ieee754.c's FP_CONTRACT pragma).
	// Tests set it, for the twin of a program built as Node's arm64 release is, which GCC compiled with contraction on
	// (internal/native's fusedTwin, #4z1nx1s).
	Contract bool

	// Slabs keeps the size-class allocator on in a sanitized build (-DADAMIC_SLABS, heap.c), where every value
	// otherwise comes from malloc. Tests set it, to show a use after a free is still caught with it on.
	Slabs bool

	// FirstLinkTraps lets internal/native's emitter write the C of a program lowered in first-link trap mode
	// (lower's Options.FirstLinkTraps, the same name on purpose, so its one-door test guards this one too).
	// Without it, a program whose Traps isn't empty is refused: it's measurement, never a product. No flag or
	// key reads it, since the C it allows is a program's like any other.
	FirstLinkTraps bool
}

// Flags are what clang compiles a program and the runtime with. The fuzzer (internal/fuzz) compiles
// with the same ones, so what it finds is what Build would. They always name the target, a native build's triple
// resolved from the build's platform when Target is empty, so every key that holds them holds the target too.
func Flags(options Options) []string {
	options, err := ResolveTarget(options)
	if err != nil {
		// Every build validates its options before clang runs (ValidateOptions), refusing a platform the table holds no
		// row for by its name, so this is only a caller asking for a key's flags on such a platform: there is no native
		// build there to name.
		panic(err)
	}
	// A program may declare a variable, a function or a parameter it never uses, or assign a variable
	// to itself, as JavaScript allows; that's the linter's business (cohere's no-unused-vars and
	// no-self-assign), not a reason the C can't compile.
	flags := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-Wcast-function-type-strict", "-pedantic", "-Wno-unused-variable", "-Wno-unused-but-set-variable", "-Wno-unused-function", "-Wno-unused-parameter", "-Wno-self-assign"}
	// JavaScript rounds every operation on its own. clang otherwise fuses a * b + c into one
	// multiply-add wherever the processor has one (every arm64, so every Apple silicon Mac), and
	// 0.1 * 10 - 1 is then 5.551115123125783e-17 instead of 0. V8's own build doesn't fuse either;
	// Node's arm64 release does, since GCC contracts by default, and its last bits aren't the
	// language's answer (#4z1nx1s): what Node gives on x86-64 is.
	if options.Debug {
		flags = append(flags, "-g")
	}
	if options.Contract {
		flags = append(flags, "-ffp-contract=fast")
	} else {
		flags = append(flags, "-ffp-contract=off")
	}
	// Every function checks its frame against the stack's limit (stack.c), and a call in tail position
	// that clang turns into a jump never makes a frame: a self tail call becomes a loop, and recursion
	// with no end runs forever where Node's runs out of stack. Every call keeps its frame, as V8's do.
	flags = append(flags, "-fno-optimize-sibling-calls")
	if options.Target == "wasm32-wasi" {
		flags = append(flags, "--target=wasm32-wasi", "--sysroot="+os.Getenv("WASI_SYSROOT"), "-DADAMIC_TARGET_WASI=1", "-mno-atomics")
	}
	if options.Target != "wasm32-wasi" {
		flags = append(flags, "--target="+options.Target, "-pthread")
	}
	if options.ProgramRegion {
		flags = append(flags, "-DADAMIC_PROGRAM_REGION")
	}
	if options.Count {
		flags = append(flags, "-DADAMIC_COUNT")
	}
	if options.Slabs {
		flags = append(flags, "-DADAMIC_SLABS")
	}
	if options.CPU != "" {
		flags = append(flags, "-march="+options.CPU)
	}
	if options.Sanitize && options.ThreadSanitize {
		panic("native: ASan and TSan cannot be combined")
	}
	if options.ThreadSanitize {
		return append(flags, "-O1", "-g", "-fsanitize=thread", "-DADAMIC_TSAN_TEST")
	}
	if options.Sanitize {
		// UBSan keeps each check's source location as a string in the object, a header's by the path clang found it at,
		// and no prefix map reaches those strings: a sysroot under the tools directory put its root in every object that
		// checks code from its headers (Hot's arm64, #8wgk65j). A report names its file alone; the line and column stay.
		return append(flags, "-O1", "-g", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-fsanitize-undefined-strip-path-components=-1")
	}
	if options.Target == "wasm32-wasi" {
		return append(flags, "-Oz")
	}
	flags = append(flags, "-O2")
	if ShippedRelease(options) {
		flags = append(flags, "-flto=thin")
	}
	return flags
}

// ShippedRelease says whether options build a shipped native release, uncounted and unsanitized, the one build that
// compiles and links with ThinLTO (-flto=thin in Flags, and internal/native's linkFlags).
func ShippedRelease(options Options) bool {
	return options.Release && !options.Sanitize && !options.ThreadSanitize && !options.Count && options.Target != "wasm32-wasi"
}
