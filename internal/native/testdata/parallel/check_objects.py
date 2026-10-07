"""Compare every parent release object with this checkout, using identical paths."""
import hashlib
import json
import pathlib
import subprocess
import sys
import tempfile

parent = sys.argv[1] if len(sys.argv) > 1 else "0f67284"
root = pathlib.Path("internal/native/runtime")
names = subprocess.check_output(
    ["git", "ls-tree", "--name-only", parent, "internal/native/runtime/"]
).decode().splitlines()
old = {
    pathlib.Path(name).name: subprocess.check_output(["git", "show", f"{parent}:{name}"])
    for name in names
    if name.endswith((".c", ".h"))
}
current = {path.name: path.read_bytes() for path in root.iterdir() if path.suffix in (".c", ".h")}
if "--mutant" in sys.argv[2:]:
    seam = b"return grain == 0 ? 1 : grain > 256 ? 256 : grain;"
    assert current["parallel.c"].count(seam) == 1
    current["parallel.c"] = current["parallel.c"].replace(seam, b"return 256;")
flags = ["-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-Wno-unused-variable",
         "-Wno-unused-but-set-variable", "-Wno-unused-function", "-Wno-unused-parameter",
         "-Wno-self-assign", "-ffp-contract=off", "-pthread", "-fno-optimize-sibling-calls", "-O2"]
rows = []
with tempfile.TemporaryDirectory(prefix="adamic-object-proof-") as directory:
    directory = pathlib.Path(directory)
    for label, files in [("before", old), ("after", current)]:
        for name, data in files.items():
            (directory / name).write_bytes(data)
        hashes = {}
        for name in sorted(old):
            if not name.endswith(".c"):
                continue
            target = directory / (name + ".o")
            subprocess.run(["clang", *flags, "-c", name, "-o", str(target)], cwd=directory, check=True)
            hashes[name] = hashlib.sha256(target.read_bytes()).hexdigest()
            symbols = subprocess.check_output(["nm", str(target)]).decode()
            assert "adamic_tsan_" not in symbols, (label, name, symbols)
        if label == "before":
            baseline = hashes
        else:
            for name, digest in hashes.items():
                rows.append({"file": name, "before": baseline[name], "after": digest,
                             "identical": digest == baseline[name]})
    # ASan has line tables, so object identity is not claimed for debug builds.
    # Compile every ASan object and verify it defines or calls no hook symbol.
    for name in sorted(old):
        if not name.endswith(".c"):
            continue
        target = directory / (name + ".asan.o")
        subprocess.run(["clang", *flags[:-1], "-O1", "-g", "-fsanitize=address,undefined",
                        "-fno-sanitize-recover=all", "-c", name, "-o", str(target)], cwd=directory, check=True)
        assert "adamic_tsan_" not in subprocess.check_output(["nm", str(target)]).decode(), name
    guarded = subprocess.run(["clang", *flags, "-DADAMIC_TSAN_TEST", "-c", "parallel.c",
                              "-o", str(directory / "invalid.o")], cwd=directory, capture_output=True)
    assert guarded.returncode != 0 and b"requires ThreadSanitizer" in guarded.stderr
print(json.dumps({"parent": parent, "objects": rows, "release_hook_symbols": 0,
                  "asan_hook_symbols": 0, "non_tsan_define_refused": True}, indent=2))
assert all(row["identical"] for row in rows), "release objects changed"
