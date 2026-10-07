"""Verify declarations against pinned Go listener keys and named syntax kinds."""
import json
import re
from pathlib import Path

REPOSITORY = Path(__file__).resolve().parents[4]
ROOT = Path(__file__).resolve().parent
KINDS = {
    name: name.removeprefix("Kind")
    for name, value in re.findall(
        r"x\[(Kind\w+)-(\d+)\]",
        (REPOSITORY / "cohere/TypeScript/tsc/internal/ast/kind_stringer_generated.go").read_text(),
    )
}


def check(path, declaration):
    name = declaration["name"]
    if name.startswith("@next/next/"):
        family, stem = "next", name.removeprefix("@next/next/")
    elif name.startswith("react-hooks/"):
        family, stem = "react", name.split("/", 1)[1]
    elif "/" in name:
        family, stem = name.split("/", 1)
    else:
        family, stem = "core", name
    source = REPOSITORY / "cohere/internal/lint/rules" / family / (stem.replace("-", "_") + ".go")
    expected = [KINDS[kind] for kind in re.findall(r"^\t{3}(Kind\w+):", source.read_text().replace("ast.Kind", "Kind"), re.MULTILINE)]
    assert expected, f"No production listeners found for {name}"
    actual = declaration["kinds"]
    assert all(type(kind) is str for kind in actual), f"Unnamed kinds: {path}"
    assert actual == expected, f"Listener mismatch for {name}: {actual} != {expected}"


paths = sorted(ROOT.glob("*/rule.json"))
assert len(paths) == 15, f"Expected 15 declarations, got {len(paths)}"
for path in paths:
    check(path, json.loads(path.read_text()))
print("PASS: 15 manifests match production listener keys and pinned ast.Kind names")
mutant = json.loads(paths[0].read_text())
mutant["kinds"][0] = KINDS["KindUnknown"]
try:
    check(paths[0], mutant)
except AssertionError as error:
    print("CAUGHT named-listener mutant:", error)
else:
    raise AssertionError("Named-listener mutant survived")
