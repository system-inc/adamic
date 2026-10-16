#!/usr/bin/env python3
"""Run isolated index mutations, each against its bounded named witness."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
path = root / "internal/lower/cycles.go"
source = path.read_text()
mutants = [
    ("missing-shape-edge", "f.shapeIndex[property.Name] = append(f.shapeIndex[property.Name], shape)", "f.shapeIndex[property.Name] = nil", "TestCycleShapeIndexKeepsHiddenEdge"),
    ("optional-filter", "property.Flags&ast.SymbolFlagsOptional == 0 &&", "", "TestCycleShapeIndexOptionalProperties"),
    ("stale-reach-cache", "f.reachEdges == nil || f.reachShapes != len(f.shapes)", "f.reachEdges == nil", "TestCycleShapeIndexInvalidatesReachability"),
    ("no-layout-sharing", "actual = prior", "actual = proven", "TestCycleShapeIndexRepeatedLayouts"),
    ("erased-slot-type", "proven:   f.l.checker.GetTypeOfSymbol(property)", "proven: nil", "TestCycleLiteralLayoutsKeepTargetsDistinct"),
    ("missing-unknown-write", "f.unknownWrite = index", "f.unknownWrite = -1", "TestCycleWriteIndexKeepsUnknown"),
    ("last-write", "first >= 0 && index >= first", "false", "TestCycleWriteIndexKeepsFirstWrite"),
]
accessor_start = source.index("\t\t// A paired setter")
accessor_end = source.index("\t\ttrie :=", accessor_start)
mutants.append(("erased-accessor-layout", source[accessor_start:accessor_end], "", "TestCycleLiteralLayoutsKeepAccessorsDistinct"))
selected = os.environ.get("ADAMIC_CYCLE_MUTANTS", "").split(",")
with tempfile.TemporaryDirectory(prefix="cycle-index-mutants-") as temporary:
    for name, before, after, test in mutants:
        if selected != [""] and name not in selected:
            continue
        if before not in source:
            raise RuntimeError("missing mutation anchor: " + name)
        mutation = Path(temporary) / (name + ".go.txt")
        changed = source.replace(before, after, 1)
        if name == "erased-accessor-layout":
            changed = changed.replace("flags:    property.Flags", "flags:    0").replace("held:     (property.Flags&ast.SymbolFlagsMethod == 0 || literalMethod(property)) && !accessorSymbol(property)", "held:     true").replace("readonly: f.l.checker.IsReadonlySymbol(property)", "readonly: false")
        mutation.write_text(changed)
        overlay = Path(temporary) / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(path): str(mutation)}}))
        log = Path("/tmp/cycle-index-mutant-" + name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./internal/lower", "-run", "^" + test + "$", "-count=1", "-timeout=55s", "-v"], cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=58)
        text = log.read_text()
        if result.returncode == 0 or "--- FAIL: " + test not in text or "[build failed]" in text:
            raise RuntimeError(name + " was not caught by " + test + ": " + text)
        print(name + " caught by " + test, flush=True)
