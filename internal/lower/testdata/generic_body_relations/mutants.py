#!/usr/bin/env python3
"""Run isolated Go-overlay mutants without changing the working tree."""
import json
import pathlib
import subprocess
import sys

root = pathlib.Path(__file__).resolve().parents[4]
output = pathlib.Path(sys.argv[1]).resolve()
output.mkdir(parents=True, exist_ok=True)
mutants = [
    ("drop-body", "internal/lower/refusals.go",
     "\t\tif err := l.genericBodyRefusal(node); err != nil {\n\t\t\tfound = err\n\t\t\treturn true\n\t\t}\n",
     "", "^TestGenericBodyRelationsRefuse$/^(initializer|field_initializer|index_signature)$",
     "want generic body refusal, got internal compiler error: generic body initializer witness"),
    ("drop-witness", "internal/lower/generic_body_relations.go",
     "func (l *lowering) genericBodyWitness(declaration *ast.Node) error {\n",
     "func (l *lowering) genericBodyWitness(declaration *ast.Node) error {\n\treturn nil\n",
     "^TestGenericBodyRelationsWitness$", "want resolved initializer witness 2 into 1, got <nil>"),
    ("drop-mapper", "internal/lower/generic.go",
     "if mapper := genericSignatureMapper(resolved); mapper != nil {",
     "if mapper := genericSignatureMapper(nil); mapper != nil {",
     "^TestGenericBodyRelationsIndexedReturnMapper$", "--- FAIL: TestGenericBodyRelationsIndexedReturnMapper"),
]
for name, relative, before, after, test, evidence in mutants:
    original = root / relative
    source = original.read_text()
    assert source.count(before) == 1, (name, "mutation site drift")
    replacement = output / (name + ".go.txt")
    replacement.write_text(source.replace(before, after))
    overlay = output / (name + ".json")
    overlay.write_text(json.dumps({"Replace": {str(original): str(replacement)}}))
    log = output / (name + ".log")
    with log.open("w") as capture:
        result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./internal/lower",
                                 "-run", test, "-count=1", "-v"], cwd=root,
                                stdout=capture, stderr=subprocess.STDOUT)
    text = log.read_text()
    assert result.returncode != 0 and evidence in text, (name, result.returncode, text)
    assert "[build failed]" not in text, (name, "compile failure is not a killed mutant")
    if name == "drop-body":
        assert "internal compiler error: generic body field initializer witness" in text, (name, text)
        for fixture in ("initializer", "field_initializer", "index_signature"):
            assert "FAIL: TestGenericBodyRelationsRefuse/" + fixture in text, (name, fixture, text)
    print(name + ": caught; " + str(log))
