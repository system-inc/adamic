#!/usr/bin/env python3
"""Key setup's warming stamp after Go has validated every dependency action.

This stamp never replaces Go's input checks. setup first uses go list -export,
which checks and restores dependency artifacts through Go's action cache. The
stamp only avoids repeating the build and optional test-binary linking afterward.
"""
import hashlib
import json
import os
import re
from pathlib import Path
import subprocess
import sys


def records(text):
    decoder = json.JSONDecoder()
    offset = 0
    while offset < len(text):
        while offset < len(text) and text[offset].isspace():
            offset += 1
        if offset == len(text):
            break
        record, offset = decoder.raw_decode(text, offset)
        yield record


def warming_key(head, sums, version, environment, packages, cache, mode):
    inputs = {
        "head": head,
        "sums": hashlib.sha256(sums).hexdigest(),
        "version": version,
        "environment": environment,
        "packages": packages,
        "cache": cache,
        "mode": mode,
    }
    return hashlib.sha256(json.dumps(inputs, sort_keys=True).encode()).hexdigest()


def manifest_bytes(paths):
    manifests = []
    for path in sorted(paths):
        file = Path(path)
        digest = hashlib.sha256(file.read_bytes()).hexdigest() if file.exists() else None
        manifests.append([str(file), digest])
    return json.dumps(manifests, sort_keys=True).encode()


def main():
    repository, package_file, mode = sys.argv[1:]
    os.chdir(repository)
    environment = json.loads(subprocess.check_output(["go", "env", "-json"]))
    # go env generates a fresh temporary prefix on every call. It remaps that prefix to
    # /tmp/go-build, so its spelling cannot affect an artifact. Keep every other flag.
    environment["GOGCCFLAGS"] = re.sub(
        r"-f(debug|file)-prefix-map=\S*/go-build\d+=/tmp/go-build",
        r"-f\1-prefix-map=<temporary>=/tmp/go-build", environment["GOGCCFLAGS"],
    )
    environment["setup-source"] = hashlib.sha256(Path("cloud/setup.sh").read_bytes()).hexdigest()
    environment["key-source"] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    packages = []
    manifests = {"go.mod", "go.sum", "cloud/markdown-width/package.json",
                 "cloud/markdown-width/package-lock.json", "cloud/markdown-width/npm-bootstrap.json",
                 "cloud/setup-markdown-width.py", "cloud/setup-stage3-api.py",
                 "stage3/api/package.json", "stage3/api/package-lock.json"}
    workspace = environment.get("GOWORK")
    if workspace and workspace != "off":
        manifests.update([workspace, workspace + ".sum"])
    for package in records(Path(package_file).read_text()):
        if package.get("Error") or package.get("DepsErrors"):
            raise RuntimeError("Go could not validate dependencies")
        module = package.get("Module", {})
        for dependency in [module, module.get("Replace", {})]:
            if dependency.get("GoMod"):
                path = Path(dependency["GoMod"])
                manifests.add(str(path))
                if path.name == "go.mod":
                    manifests.add(str(path.with_name("go.sum")))
        artifact = package.get("Export")
        if artifact:
            # Stat also notices an artifact deleted between go list and this key.
            state = os.stat(artifact)
            packages.append([package["ImportPath"], package.get("BuildID"), artifact,
                             state.st_size, state.st_mtime_ns, state.st_ctime_ns])
    cache = []
    # Link/test actions also live here. Detect removal, replacement and go clean,
    # including deletion of just one entry rather than the entire cache directory.
    for directory, _, files in os.walk(environment["GOCACHE"]):
        for name in sorted(files):
            if name.endswith(("-a", "-d")):
                path = os.path.join(directory, name)
                state = os.stat(path)
                if name.endswith("-a"):
                    # Go refreshes the action record's access timestamp even on a hit.
                    # Its action ID, output ID and size still identify the same answer.
                    cache.append([path, [word.decode("ascii") for word in Path(path).read_bytes().split()[:4]]])
                else:
                    cache.append([path, state.st_size, state.st_mtime_ns, state.st_ctime_ns])
    cache.sort()
    print(warming_key(
        subprocess.check_output(["git", "rev-parse", "HEAD"]).decode().strip(),
        manifest_bytes(manifests),
        subprocess.check_output(["go", "version"]).decode().strip(),
        environment, sorted(packages), cache, mode,
    ))


if __name__ == "__main__":
    main()
