# Apple step 1: the dedication

Task #hkxxhvf. Prepared on Linux for verification by the lead on a Mac.

Step 1 takes `dedication/dedication.a` through the stage 0 checker and C emitter,
then builds that C and the native runtime with Xcode's clang. Both scripts must
finish by comparing application stdout byte for byte with the source run by
`node oracle/node.mjs dedication/dedication.a`. The expected output is the
97-byte dedication line, including its final newline.

A successful Mac run establishes that this program compiles and links against
the Apple SDKs, its bundles have valid plists and ad-hoc signatures, and the
macOS executable and simulator process print the same bytes as Node. Until the
lead runs the commands below, those Apple claims are unverified.

## Run on a Mac

Install full Xcode with an iOS 17 or newer simulator runtime, Go compatible with
`go.mod`, and Node 24. The Linux `cloud/setup.sh` is not a Mac setup script.
Run from the repository root with the submodules initialized:

```sh
git checkout codex/apple-step-1
git submodule update --init --recursive
xcode-select -p
xcrun --sdk macosx --show-sdk-path
xcrun --sdk iphonesimulator --show-sdk-path
go version
node --version
bash cloud/apple/build-macos.sh > /tmp/adamic-apple-macos.log 2>&1
```

The macOS script always builds arm64 and x86_64 for macOS 14 and combines them
with `xcrun lipo`. It creates this command-line application bundle:

```text
Dedication.app/
  Contents/
    Info.plist
    MacOS/Dedication
```

`LSBackgroundOnly` declares that the bundle has no UI. Run its executable
directly to capture stdout. Opening it in Finder provides no terminal.
The script executes the universal binary's host architecture. Building both
slices does not prove that the other slice runs. Verify that slice on a suitable
Mac, or under Rosetta on Apple silicon if Rosetta is installed.

For iOS use an Apple silicon Mac. Choose an available iOS 17 or newer device's
UDID from the list, then boot it if it is not already booted. Keep exactly one
simulator booted because the script targets `booted`:

```sh
xcrun simctl list devices available
# Set this to the chosen device's UDID from the list.
SIMULATOR_UDID=the-chosen-udid
xcrun simctl boot "$SIMULATOR_UDID"
xcrun simctl bootstatus "$SIMULATOR_UDID" -b
bash cloud/apple/build-ios-sim.sh > /tmp/adamic-apple-ios-sim.log 2>&1
```

The iOS script compiles for `arm64-apple-ios17.0-simulator`. The flat bundle has
`Info.plist` and `Dedication`, with identifier `org.system.adamic.dedication`,
minimum iOS version 17.0, and device families 1 and 2. It signs and verifies the
bundle, then runs:

```sh
xcrun simctl bootstatus booted -b
xcrun simctl install booted "$app"
xcrun simctl launch --console booted org.system.adamic.dedication
```

Here `$app` is the fresh bundle path printed by the script. This is a C entry
point that prints and exits, with no UIKit event loop or screen. Simulator
acceptance of that minimal app is part of the lead's verification, not an
observation made on Linux. Step 1 does not prove physical-device execution or
App Store suitability.

## Build and evidence

`common.sh` builds the driver with `go build -o "$out/adamic" ./cmd/adamic`, then
runs `"$out/adamic" c dedication/dedication.a > "$out/main.c"`. It compiles every
ordinary runtime C translation unit directly, excluding only `tsgo.c`, the
optional external Go checker bridge. No host runtime archive is reused across
Apple targets.

The clang commands use the SDK path from `xcrun --sdk ... --show-sdk-path` as
`-isysroot`, strict C11 warnings, `-O2`, `-ffp-contract=off`, and
`-fno-optimize-sibling-calls`, matching the release flags in `native.Flags`.
The macOS commands are `xcrun --sdk macosx clang -target arm64-apple-macos14`
and the corresponding `x86_64-apple-macos14` command. The iOS command is
`xcrun --sdk iphonesimulator clang -target arm64-apple-ios17.0-simulator`.
The scripts show every argument with Bash tracing.

Both scripts run `plutil -lint`, `codesign --sign - --force --deep`, and
`codesign --verify --strict --verbose=2`. An error stops the build with the file,
line and exit status. Each invocation creates a fresh directory under
`${TMPDIR:-/tmp}` and leaves it in place, including failed builds. The directory
contains `main.c`, the driver, `node.stdout`, the app bundle and captured output.
No prior stdout can satisfy a later comparison.

The macOS binary's stdout goes straight to `app.stdout`; stderr goes to
`app.stderr`. The simulator console stream goes to `console.stdout` and
`console.stderr`. If its first line is exactly the known bundle identifier,
a colon, a space and a decimal process ID, the script removes that one simctl
launch receipt. Otherwise it copies the stream unchanged. No application text,
whitespace, newline or unexpected diagnostic is normalized. The final
`cmp "$out/node.stdout" "$out/app.stdout"` must succeed. A receipt format
change fails the comparison and leaves the original console bytes for review.

## Runtime portability changes

Only `internal/native/runtime/stack.c` changes:

1. Under the existing compiler platform macro `__APPLE__`, define
   `_DARWIN_C_SOURCE` before headers and include `pthread.h` instead of
   `sys/resource.h`, exposing Darwin's stack query extensions.
2. On Darwin, obtain the current thread's top and size using `pthread_self`,
   `pthread_get_stackaddr_np` and `pthread_get_stacksize_np`. Place the limit
   above the actual bottom with an eighth of the stack reserved, capped at
   256 KiB. This avoids using the desktop process stack limit for a smaller iOS
   stack. Linux retains its existing `getrlimit` path and argument reservation.

There is no thread creation or new concurrency support. The global stack limit
still belongs to the startup thread. Step 2 must revisit it before running
Adamic code on other threads. The dedication needs only stdout and the startup
thread. The other inspected OS interfaces are POSIX file, directory, signal
and output APIs provided by Darwin; no additional guards were needed by
inspection. Compiling them against the actual iOS SDK remains unverified.
The simulator is not a proof of unrestricted file access in an iOS sandbox.

## Linux verification and limits

Run the plumbing check with the configured Linux toolchain:

```sh
source /workspace/adamic-tools/env.sh
python3 cloud/apple/check-linux.py > /tmp/apple-linux-check.log 2>&1
```

It syntax-checks all scripts, runs ShellCheck when installed, and confirms both
production scripts refuse Linux. Mock Apple tools run the real driver and
Linux clang to exercise bundling, plist contents, console capture and
comparisons. The mocks do not produce Mach-O binaries, universal slices or
signatures. They do not validate an SDK or launch a simulator.

Mutants change one byte of the generated dedication C, add unexpected console
text, or remove the final newline. Each must compile and run, then fail only
at `cmp`. Injected compile, sign, verify, boot, install and launch errors must
stop without a success message. A Darwin pthread shim also exercises 512 KiB,
8 MiB, zero and invalid stack bounds. Subtracting rather than adding the safety
margin is valid C but must fail the executed bounds check. The shim does not
validate Apple's header declarations or actual stack mapping.

The following commands and behavior cannot be verified in this Linux box:
all real `xcrun` SDK queries, all three target clang invocations, `xcrun lipo`,
Apple `plutil`, both `codesign` operations, execution of the macOS Mach-O bundle,
`simctl boot`, `bootstatus`, `install`, and `launch --console`, the real console
receipt format, and the final comparisons using Apple-built executables.
The Go driver and Node command ran on Linux, not on macOS. The lead should
retain the two Mac logs and the artifact directories as the evidence.

## What step 2 needs

Keep the bundle paths, identifiers, per-target SDK selection, signing order and
Node comparison. Add a small Objective-C entry point that calls the compiled
Adamic program through an explicit C function boundary. The generated C owns
`main` today; rename it at compilation or extract a callable entry point before
adding `NSApplicationMain` or `UIApplicationMain`. Do not introduce two mains.
Compile Objective-C sources separately with the same target and SDK, link the
appropriate Foundation/AppKit or UIKit frameworks, and sign only after all
executables and resources are complete.

Step 2 supplies the event loop, delegate, window or view, lifetime rules at the
Objective-C boundary, and a way to capture the dedication without launch
metadata. It also needs stack limits appropriate to any thread that can enter
Adamic. This layout is the starting point for those changes; step 1 supplies
no Objective-C runtime bridge.

## Worker gate on Linux

The October 7, 2026 worker gate used main `ef3d907ecdc4c771b016f7d9c52372def057a340`.
`bash cloud/setup.sh` succeeded: Go ready at 0s, clang ready at 1s, Node ready at
1s, submodules ready at 1s, build cache warm at 140s, done at 140s.
`nproc` printed 5; cgroup CPU quota was 4 cores. Go was 1.27.1, clang 20.1.8,
and Node 24.19.0. The environment file was `/workspace/adamic-tools/env.sh`.

These checks completed successfully, with test output written directly to logs:

```sh
source /workspace/adamic-tools/env.sh
gofmt -l cmd internal > /tmp/apple-gofmt.log
go vet ./... > /tmp/apple-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native > /tmp/apple-native-test.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/(dedication|.*stack_)|TestTheOracleCatchesOneByte|TestLongArgumentsLeaveTheStackItsLimit|TestSmallStacksStillPanic' > /tmp/apple-oracle-verbose.log 2>&1
```

Formatting and vet logs were empty. The native package passed in 116.491s.
The verbose filtered oracle passed in 1.600s, with all cache hits zero. Its
fixture selection ran the dedication only; the stack coverage came from the
long-argument, empty-environment and small-stack tests (1024, 512 and 256 KiB).
`TestTheOracleCatchesOneByte` appended `!` to the lowered dedication string and
caught it through stdout disagreement with Node. The complete repository test
gate was not run. ShellCheck was unavailable; `bash -n` passed for all scripts.

The final corrected plumbing check also passed:

```sh
python3 cloud/apple/check-linux.py > /tmp/apple-linux-verified.log 2>&1
```

Both mock bundle paths reported `97 bytes` matching Node. The generated-C
byte mutant failed `cmp` on both paths; extra simulator text and the missing
final newline failed `cmp`; the Darwin margin mutant compiled without warnings
and failed the executed bound check. All nine injected errors stopped the
scripts: compile, sign and verify on both paths, and simulator bootstatus,
install and launch. Successful simulator capture was checked with and without
a launch receipt. Evidence remains under
`/tmp/adamic-gate/adamic-apple-check-wslbbr8l`. These are Linux observations
with mocked Apple commands, not Apple verification.
