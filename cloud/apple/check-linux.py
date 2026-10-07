#!/usr/bin/env python3
"""Exercise bundle plumbing on Linux. Apple commands are mocks, not SDK validation."""
import os
import pathlib
import plistlib
import shlex
import shutil
import subprocess
import tempfile

repository = pathlib.Path(__file__).resolve().parents[2]
work = pathlib.Path(tempfile.mkdtemp(prefix="adamic-apple-check-"))
print(f"Evidence: {work}", flush=True)


def run(arguments, *, environment=None, expected=0, log=None):
    print("+ " + shlex.join(str(value) for value in arguments), flush=True)
    if log is None:
        result = subprocess.run(arguments, cwd=repository, env=environment, check=False)
    else:
        with log.open("wb") as output:
            result = subprocess.run(arguments, cwd=repository, env=environment,
                                    stdout=output, stderr=subprocess.STDOUT, check=False)
    assert result.returncode == expected, (arguments, result.returncode, log)


for script in repository.glob("cloud/apple/*.sh"):
    run(["bash", "-n", script])
if shutil.which("shellcheck"):
    run(["shellcheck", "-x", *sorted(repository.glob("cloud/apple/*.sh"))])
else:
    print("shellcheck unavailable; bash -n only", flush=True)

run(["node", "cloud/apple/check-simulator.mjs"])
helper = (repository / "cloud/apple/simulator.mjs").read_text()
for name, before, after in (
        ("oldest-runtime", "return candidates.at(-1)", "return candidates.at(0)"),
        ("first-line-only", "if (!removed &&", "if (start === 0 && !removed &&")):
    mutant_helper = work / f"{name}.mjs"
    mutant_helper.write_text(helper.replace(before, after))
    run(["node", "cloud/apple/check-simulator.mjs", mutant_helper], expected=1,
        log=work / f"{name}.log")
    print(f"Caught {name} regression by simulator input probes", flush=True)

# No platform override is exposed by the production scripts.
for platform in ("macos", "ios-sim"):
    log = work / f"linux-refusal-{platform}.log"
    run(["bash", f"cloud/apple/build-{platform}.sh"], expected=1, log=log)
    assert "Xcode on macOS is required" in log.read_text()

mock = work / "mock"
mock.mkdir()
# A single mock implements transport and plist validation. clang and Adamic stay real.
(mock / "apple-tool").write_text('''#!/usr/bin/env python3
import json, os, pathlib, plistlib, shutil, subprocess, sys
name = pathlib.Path(sys.argv[0]).name
args = sys.argv[1:]
operation = name + " " + " ".join(args)
with open(os.environ["APPLE_CALLS"], "a") as calls:
    calls.write(operation + "\\n")
if os.environ.get("APPLE_FAIL") and os.environ["APPLE_FAIL"] in operation:
    sys.exit(19)
if name == "uname":
    print("Darwin" if args == ["-s"] else "arm64")
elif name == "plutil":
    with open(args[-1], "rb") as source:
        plistlib.load(source)
elif name == "codesign":
    pass
elif name == "xcrun":
    if args[:1] == ["--sdk"]:
        args = args[2:]
    if args == ["--show-sdk-path"]:
        print("/mock-sdk")
    elif args[0] == "clang":
        filtered = []
        index = 1
        while index < len(args):
            if args[index] in ("-target", "-isysroot"):
                index += 2
            else:
                filtered.append(args[index])
                index += 1
        if os.environ.get("APPLE_MUTANT") == "program-byte":
            main = next(pathlib.Path(a) for a in filtered if a.endswith("/main.c"))
            main.write_text(main.read_text().replace("Kenneth", "Xenneth"))
        sys.exit(subprocess.call([os.environ["APPLE_CLANG"], *filtered]))
    elif args[0] == "lipo":
        shutil.copy(args[2], args[-1])
    elif args[:3] == ["simctl", "list", "devices"]:
        state = "Shutdown" if os.environ.get("APPLE_BOOTED") == "no" else "Booted"
        print(json.dumps({"devices": {
            "com.apple.CoreSimulator.SimRuntime.iOS-26-9": [{"name": "iPhone 17", "udid": "older", "state": state, "isAvailable": True}],
            "com.apple.CoreSimulator.SimRuntime.iOS-26-10": [{"name": "iPhone 17 Pro", "udid": "newest", "state": "Shutdown", "isAvailable": True}],
            "com.apple.CoreSimulator.SimRuntime.iOS-27-0": [{"name": "iPad Pro", "udid": "ipad", "state": "Shutdown", "isAvailable": True}]
        }}))
    elif args[:2] == ["simctl", "boot"]:
        assert args[-1] == "newest", args
    elif args[:2] == ["simctl", "bootstatus"]:
        pass
    elif args[:2] == ["simctl", "install"]:
        pathlib.Path(os.environ["APPLE_INSTALLED"]).write_text(args[-1])
    elif args[:2] == ["simctl", "launch"]:
        app = pathlib.Path(pathlib.Path(os.environ["APPLE_INSTALLED"]).read_text())
        result = subprocess.run([app / "Dedication"], capture_output=True, check=True)
        if os.environ.get("APPLE_RECEIPT", "yes") == "yes":
            print("org.system.adamic.dedication: 1234", flush=True)
        data = result.stdout
        if os.environ.get("APPLE_MUTANT") == "extra-console":
            data += b"unexpected console text\\n"
        elif os.environ.get("APPLE_MUTANT") == "missing-newline":
            data = data.rstrip(b"\\n")
        if os.environ.get("APPLE_RECEIPT") == "after":
            data += b"org.system.adamic.dedication: 1234\\n"
        sys.stdout.buffer.write(data)
        sys.stderr.buffer.write(result.stderr)
    else:
        raise AssertionError(args)
else:
    raise AssertionError(name)
''')
(mock / "apple-tool").chmod(0o755)
for name in ("uname", "plutil", "codesign", "xcrun"):
    (mock / name).symlink_to("apple-tool")
base_environment = dict(os.environ, PATH=str(mock) + os.pathsep + os.environ["PATH"],
                        TMPDIR=str(work), APPLE_CLANG=shutil.which("clang"),
                        APPLE_INSTALLED=str(work / "installed"))


def bundle_check(platform, label, *, fail="", mutant="", receipt="yes", booted="yes"):
    calls = work / f"{platform}-{label}.calls"
    log = work / f"{platform}-{label}.log"
    environment = dict(base_environment, APPLE_CALLS=str(calls), APPLE_FAIL=fail,
                       APPLE_MUTANT=mutant, APPLE_RECEIPT=receipt, APPLE_BOOTED=booted)
    expected = 19 if fail else 1 if mutant else 0
    run(["bash", f"cloud/apple/build-{platform}.sh"], environment=environment,
        expected=expected, log=log)
    text = log.read_text()
    if expected == 0:
        assert "stdout matches Node byte for byte (" in text
    else:
        assert "stdout matches Node byte for byte (" not in text
        assert "apple: failed at" in text
    if mutant:
        assert "differ" in text or "EOF on" in text, text
        print(f"Caught {platform} {mutant} by cmp", flush=True)
    return calls.read_text()


for platform in ("macos", "ios-sim"):
    calls = bundle_check(platform, "good")
    assert "codesign --sign - --force --deep" in calls
    assert "codesign --verify --strict" in calls
    assert "-ffp-contract=off" in calls and "-fno-optimize-sibling-calls" in calls
    if platform == "macos":
        assert "arm64-apple-macos14" in calls and "x86_64-apple-macos14" in calls
        assert "xcrun lipo -create" in calls
    else:
        assert "arm64-apple-ios17.0-simulator" in calls
        assert "simctl install older" in calls and "simctl launch --console older" in calls
        assert "simctl boot " not in calls
    bundle_check(platform, "program-byte", mutant="program-byte")
    for failed in ("clang", "codesign --sign", "codesign --verify"):
        bundle_check(platform, failed.replace(" ", "-"), fail=failed)
bundle_check("ios-sim", "without-receipt", receipt="no")
for mutant in ("extra-console", "missing-newline"):
    bundle_check("ios-sim", mutant, mutant=mutant)
calls = bundle_check("ios-sim", "boot-newest-receipt-after", receipt="after", booted="no")
assert "simctl boot newest" in calls and "simctl bootstatus newest -b" in calls
assert "simctl install newest" in calls and "simctl launch --console newest" in calls
assert "booting simulator iPhone 17 Pro (iOS 26.10, newest)" in (work / "ios-sim-boot-newest-receipt-after.log").read_text()
bundle_check("ios-sim", "boot-failure", fail="simctl boot ", booted="no")
for failed in ("simctl list", "simctl bootstatus", "simctl install", "simctl launch"):
    bundle_check("ios-sim", failed.replace(" ", "-"), fail=failed)

for app in work.glob("adamic-*/Dedication.app"):
    macos = (app / "Contents").exists()
    plist = app / "Contents/Info.plist" if macos else app / "Info.plist"
    # A deliberately failed compile leaves a partial bundle without its plist.
    if not plist.exists():
        continue
    with plist.open("rb") as source:
        values = plistlib.load(source)
    assert values["CFBundleIdentifier"] == "org.system.adamic.dedication"
    assert values["CFBundleExecutable"] == "Dedication"
    if macos:
        assert values["LSBackgroundOnly"] is True
    else:
        assert values["MinimumOSVersion"] == "17.0" and values["UIDeviceFamily"] == [1, 2]

# Use a shim to exercise Darwin's branch without claiming Apple headers compiled.
shim = work / "shim"
shim.mkdir()
(shim / "pthread.h").write_text('''#include <stddef.h>
typedef unsigned long pthread_t;
pthread_t pthread_self(void);
void *pthread_get_stackaddr_np(pthread_t thread);
size_t pthread_get_stacksize_np(pthread_t thread);
''')
probe = work / "stack-probe.c"
probe.write_text('''#include "stack.c"
static uintptr_t test_top = 0x10000000;
static size_t test_size = 512 * 1024;
pthread_t pthread_self(void) { return 1; }
void *pthread_get_stackaddr_np(pthread_t thread) { (void)thread; return (void *)test_top; }
size_t pthread_get_stacksize_np(pthread_t thread) { (void)thread; return test_size; }
_Noreturn void adamic_panic(const char *message, size_t length) {
    (void)message; (void)length; __builtin_trap();
}
int main(void) {
    find_stack_limit();
    if (adamic_stack_limit != 0x0ff90000) return 1;
    test_size = 8 * 1024 * 1024;
    find_stack_limit();
    if (adamic_stack_limit != 0x0f840000) return 2;
    test_size = 0;
    find_stack_limit();
    if (adamic_stack_limit != 0) return 3;
    test_top = 1;
    test_size = 512 * 1024;
    find_stack_limit();
    if (adamic_stack_limit != 0) return 4;
    return 0;
}
''')
runtime = repository / "internal/native/runtime"
flags = ["clang", "-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-D__APPLE__",
         "-I", str(shim), "-I", str(runtime)]
run([*flags, probe, "-o", work / "stack"])
run([work / "stack"])
# A wrong limit is still valid C. Only the executed bound check catches this mutant.
(shim / "stack.c").write_text((runtime / "stack.c").read_text().replace(
    "top - size + margin", "top - size - margin"))
run([*flags, probe, "-o", work / "stack-mutant"])
run([work / "stack-mutant"], expected=1)
print("Caught Darwin stack margin mutant by executed bound check", flush=True)
print("PASS: Linux plumbing and mutants; Xcode and simulator remain unverified", flush=True)
