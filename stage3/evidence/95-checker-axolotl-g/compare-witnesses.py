import json
from pathlib import Path
root = Path(__file__).parent
final = root / "scoped-final"
def load(name): return json.loads((final / name).read_text())
def clean(x):
    if isinstance(x, dict): return {k:clean(v) for k,v in x.items() if k not in ("stack", "scopedExitStacks")}
    if isinstance(x, list): return [clean(v) for v in x]
    return x
before = load("95g-services-before.json")
after = load("95g-services-scoped-built.json")
assert clean(after) == clean(load("95g-services-scoped-audit.json"))
resolved = unchanged = 0
moved = []
for b,a in zip(before, after):
    assert (b["fixture"], b["operation"], b["diagnostics"]) == (a["fixture"], a["operation"], a["diagnostics"])
    if "error" in b:
        if "error" in a: moved.append((a["fixture"], a["operation"], b["error"], a["error"]))
        else: resolved += 1
    else:
        assert clean(a) == clean(b)
        unchanged += 1
assert (len(before), len(after), resolved, unchanged, len(moved)) == (57,57,27,29,1)
assert moved[0][:2] == ("original-super", "implementation")
assert "parent" in moved[0][2] and "valueDeclaration" in moved[0][3]
assert json.loads((root / "super-before.json").read_text()) == load("95g-super-scoped-built.json")
assert json.loads((root / "decl-free-symbol-control-before.json").read_text()) == load("95g-decl-free-scoped-built.json")
assert load("95g-success-before.json") == load("95g-success-scoped-built.json")
unchanged = changed = 0
for b,a in zip(load("95g-controls-before.json"),load("95g-controls-scoped-built.json")):
    assert b["diagnostics"] == a["diagnostics"] and b["diagnosticOutput"] == a["diagnosticOutput"]
    assert len(b["queries"]) == len(a["queries"])
    for u,v in zip(b["queries"],a["queries"]):
        if clean(u) == clean(v): unchanged += 1
        else:
            assert b["fixture"] == "consumer.ts" and "parent" in u["error"] and "error" not in v
            changed += 1
assert (unchanged,changed) == (232,2)
original = load("95g-original-scoped-built.json")
assert len(original) == 2
assert all(x["result"] == "undefined" and x["diagnostics"] == [] and x["diagnosticOutput"] == "" for x in original)
print(json.dumps({"services":57,"resolved":resolved,"unchangedSuccessful":29,"moved":moved,"newCrashes":0,"controlsUnchanged":232,"controlsResolved":2,"originalFixed":2},indent=2))
