"""Write a reviewable .a compatibility patch; production files stay untouched."""
from pathlib import Path
import difflib
import json
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[5]
evidence = Path(__file__).resolve().parent
scratch = Path(tempfile.mkdtemp(prefix="wave1-06-a-overlay-"))
paths = ["stage1/cohere/lint/registry/registry.go", "stage1/cohere/lint/lint_test.go"]
replacements = {}
patch = []
for path in paths:
    original = (repository / path).read_text()
    updated = original
    if path.endswith("registry.go"):
        updated = updated.replace('Slug            string   `json:"-"`', 'Slug            string   `json:"-"`\n\tModule          string   `json:"-"`')
        old = '\t\tmodule, err := os.ReadFile(filepath.Join(filepath.Dir(path), "rule.ts"))'
        new = '''\t\td.Module = "rule.ts"
\t\tif _, err := os.Stat(filepath.Join(filepath.Dir(path), "rule.a")); err == nil {
\t\t\tif _, err := os.Stat(filepath.Join(filepath.Dir(path), "rule.ts")); err == nil {
\t\t\t\treturn nil, fmt.Errorf("%s: ambiguous rule.a and rule.ts", path)
\t\t\t}
\t\t\td.Module = "rule.a"
\t\t}
\t\tmodule, err := os.ReadFile(filepath.Join(filepath.Dir(path), d.Module))'''
        assert updated.count(old) == 1
        updated = updated.replace(old, new)
        updated = updated.replace('mutant.File = "rule.ts"', 'mutant.File = d.Module')
        updated = updated.replace('!strings.HasSuffix(mutant.File, ".ts")', '(!strings.HasSuffix(mutant.File, ".ts") && !strings.HasSuffix(mutant.File, ".a"))')
        updated = updated.replace("../rules/%s/rule.ts';\\n\", d.Factory, i, d.Class, i, d.Slug)", "../rules/%s/%s';\\n\", d.Factory, i, d.Class, i, d.Slug, ruleModule(d))")
        updated += '''
// ruleModule retains the historical default for descriptors built by callers.
func ruleModule(d Descriptor) string {
    if d.Module == "" { return "rule.ts" }
    return d.Module
}
'''
    else:
        updated = updated.replace('strings.HasSuffix(path, ".ts")', '(strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a"))')
        updated = updated.replace('strings.HasSuffix(file, ".ts")', '(strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".a"))')
        updated = updated.replace('change.File = "rule.ts"', 'change.File = descriptor.Module')
    assert updated != original
    target = scratch / Path(path).name
    target.write_text(updated)
    subprocess.run(["gofmt", "-w", str(target)], check=True)
    updated = target.read_text()
    replacements[str(repository / path)] = str(target)
    patch.extend(difflib.unified_diff(original.splitlines(True), updated.splitlines(True), fromfile="a/" + path, tofile="b/" + path))
(evidence / "a-support.patch").write_text("".join(patch))
(scratch / "overlay.json").write_text(json.dumps({"Replace": replacements}))
print(scratch / "overlay.json")
