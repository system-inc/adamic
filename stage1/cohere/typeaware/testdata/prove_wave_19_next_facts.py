"""Prove the isolated raw fact contracts fail without touching shared files."""
import json
import os
from pathlib import Path
import subprocess
import sys

repository = Path(__file__).resolve().parents[4]
artifacts = Path(sys.argv[1]).resolve()
artifacts.mkdir(parents=True, exist_ok=True)
changes = [
    ("ancestry-alias", "declaration_ancestry.go", "symbol = c.GetAliasedSymbol(symbol)", "symbol = c.GetSymbolAtLocation(node)", "TestWave19AncestryAliasAndBindingNames", "wrong alias ancestry"),
    ("ancestry-name-guard", "declaration_ancestry.go", "identifier != nil && identifier.Kind == ast.KindIdentifier", "identifier != nil", "TestWave19AncestryAliasAndBindingNames", "Unhandled case in Node.Text: *ast.BindingPattern"),
    ("ancestry-global", "declaration_ancestry.go", "out.yes(ancestor.Kind == ast.KindModuleDeclaration && ast.IsGlobalScopeAugmentation(ancestor))", "out.yes(false)", "TestWave19DeclarationAncestry", "wrong global-augmentation metadata"),
    ("callee-return", "resolved_callee.go", "out.number(flags)", "_ = flags; out.number(0)", "TestWave19ResolvedCalleeAndProgramModules", "wrong resolved callee"),
    ("module-target", "program_modules.go", "resolved = target.FileName()", "_ = target; resolved = \"\"", "TestWave19ResolvedCalleeAndProgramModules", "wrong resolved module edge"),
]
for name, filename, before, after, test, message in changes:
    original = repository / "bridge/tsgo/checker" / filename
    source = original.read_text()
    if source.count(before) != 1:
        raise RuntimeError(f"nonunique mutant {name}")
    replacement = artifacts / filename
    replacement.write_text(source.replace(before, after))
    overlay = artifacts / (name + ".json")
    overlay.write_text(json.dumps({"Replace": {str(original): str(replacement)}}))
    with (artifacts / (name + ".log")).open("wb") as output:
        result = subprocess.run(["go", "test", "-overlay", str(overlay), "./bridge/tsgo/checker", "-run", "^" + test + "$", "-v", "-count=1"], cwd=repository, stdout=output, stderr=subprocess.STDOUT)
    observed = (artifacts / (name + ".log")).read_text()
    if result.returncode == 0 or message not in observed or "build failed" in observed:
        raise RuntimeError(f"mutant not killed by its contract: {name}\n{observed}")
    print(f"{name}: checker contract caught the deliberately wrong fact; compile succeeded")
