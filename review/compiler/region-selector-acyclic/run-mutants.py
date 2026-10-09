from pathlib import Path
import difflib
import os
import subprocess

root = Path.cwd()
evidence = root / "review/compiler/region-selector-acyclic"

def run(name, path, change, package, test, diagnostic):
    source = root / path
    original = source.read_text()
    altered = change(original)
    assert altered != original, name
    (evidence / (name + ".patch")).write_text("".join(difflib.unified_diff(
        original.splitlines(True), altered.splitlines(True),
        fromfile="a/" + path, tofile="b/" + path)))
    try:
        source.write_text(altered)
        with (evidence / (name + ".log")).open("w") as log:
            result = subprocess.run(["timeout", "--kill-after=2s", "90", "go", "test",
                package, "-run", "^" + test + "$", "-count=1", "-timeout=90s", "-v"],
                stdout=log, stderr=subprocess.STDOUT)
        output = (evidence / (name + ".log")).read_text()
        assert result.returncode not in (0, 124, 137) and diagnostic in output, (name, output)
        print(name + ": caught by " + test + ": " + diagnostic, flush=True)
    finally:
        source.write_text(original)

run("reinclude-acyclic-containers", "internal/lower/program_region_selection.go",
    lambda s: s.replace(" && f.programOwnedCycle(value)", "", 1).replace("\t\tif programScalarValue(value) {\n\t\t\treturn false\n\t\t}\n", "", 1),
    "./internal/lower", "TestProgramRegionCensusMembership", "membership changed; regenerate")
run("reinclude-payload-control", "internal/lower/program_region_selection.go",
    lambda s: s.replace(" && f.programOwnedCycle(value)", "", 1).replace("\t\tif programScalarValue(value) {\n\t\t\treturn false\n\t\t}\n", "", 1),
    "./internal/lower", "TestProgramRegionAcyclicContainerPayloads", "member=true, want false")
run("drop-cyclic-member", "internal/lower/program_region.go",
    lambda s: s.replace("import (", 'import (\n "strings"', 1).replace(
        "case ir.ObjectLiteral:\n\t\tvalue.ProgramRegion = member",
        'case ir.ObjectLiteral:\n\t\tvalue.ProgramRegion = member\n\t\tif strings.Contains(sourceExpression(node),"root${1}"){value.ProgramRegion=false}', 1),
    "./internal/oracle", "TestProgramRegionOwnership", "regions 1, want 2")

run("phantom-brand-owns", "internal/lower/program_region_selection.go",
    lambda s: s.replace("\t\tif programScalarValue(value) {\n\t\t\treturn false\n\t\t}\n", "", 1),
    "./internal/lower", "TestProgramRegionAcyclicContainerPayloads", "container pending member=true, want false")
run("ignore-method-captures", "internal/lower/program_region_selection.go",
    lambda s: s.replace("accessorSymbol(property) || f.isFunction(f.l.checker.GetTypeOfSymbol(property))", "accessorSymbol(property)", 1),
    "./internal/lower", "TestProgramRegionAcyclicContainerPayloads", "container methods member=false, want true")
