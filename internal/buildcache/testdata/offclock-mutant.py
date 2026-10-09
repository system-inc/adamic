#!/usr/bin/env python3
"""Run one buildcache mutant, with an 80-second cap, and restore its source.

Use the configured Go environment, then run this script with one of:
inline-upload, inline-audit, audit-drain, normalize-non-macho.
A compiler error or timeout does not count as a kill; a test assertion must fail.
"""
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[2]
mutants = {
    "inline-upload": (
        "buildcache.go",
        "if err = spool(key, inputs.Name, product); err != nil {",
        "if err = publish(key, inputs.Name, product); err != nil {",
        "TestABuildPublishesBlobsThenItsRefAndAnotherMachineFetchesIt",
    ),
    "inline-audit": (
        "buildcache.go",
        "if err = queueAudit(key, inputs, product, testName); err != nil {",
        "if err = audit(key, inputs.Name, product, build); err != nil {",
        "TestAnAuditedFetchThatDiffersFromARebuildIsPoisoning/built",
    ),
    "audit-drain": (
        "audits.go",
        "if err := compareAudit(entry.Key, entry.Inputs.Name, got.Files, want.Files); err != nil {",
        "if err := compareAudit(entry.Key, entry.Inputs.Name, got.Files, want.Files); false && err != nil {",
        "TestAnAuditedFetchThatDiffersFromARebuildIsPoisoning/a_wrong_product",
    ),
    "normalize-non-macho": (
        "macho.go",
        "return content, nil\n\t}\n\tfile, err := macho.NewFile",
        "return make([]byte, len(content)), nil\n\t}\n\tfile, err := macho.NewFile",
        "TestAuditLeavesNonMachOBytesAlone",
    ),
}

if len(sys.argv) != 2 or sys.argv[1] not in mutants:
    sys.exit("usage: offclock-mutant.py " + "|".join(mutants))
name = sys.argv[1]
file, before, after, test = mutants[name]
source = root / "buildcache" / file
original = source.read_bytes()
text = original.decode()
if text.count(before) != 1:
    sys.exit("mutant anchor is missing or ambiguous: " + name)
try:
    source.write_text(text.replace(before, after))
    result = subprocess.run(
        ["go", "test", "./internal/buildcache", "-count=1", "-timeout=75s", "-run", "^" + test + "$"],
        cwd=root.parent,
        capture_output=True,
        text=True,
        timeout=80,
    )
    output = result.stdout + result.stderr
    print(output, end="")
    if result.returncode != 1 or "--- FAIL:" not in output:
        sys.exit("not killed by a test assertion: " + name)
    print("killed:", name)
finally:
    source.write_bytes(original)
