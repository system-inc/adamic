from pathlib import Path
import subprocess, difflib, json
root = Path(__file__).resolve().parents[2]
file = 'internal/lower/predicates_proof.go'
base = subprocess.check_output(['git', 'show', 'origin/main:' + file], cwd=root, text=True)
mutants = [
 ('M1', 'if err == nil {\n\t\t\t\tv.summaries[declaration] = summary', 'if false && err == nil {\n\t\t\t\tv.summaries[declaration] = summary', 'if os.Getenv("ADAMIC_MUTANT") != "M1" && err == nil {\n\t\t\t\tv.summaries[declaration] = summary'),
 ('M2', 'p.Name().Text() == predicate.ParameterName.Text()', 'p.Name().Text() != predicate.ParameterName.Text()', '(p.Name().Text() == predicate.ParameterName.Text()) != (os.Getenv("ADAMIC_MUTANT") == "M2")'),
 ('M3', 'if result&predicateFalse != 0 && wanted {', 'if false && result&predicateFalse != 0 && wanted {', 'if os.Getenv("ADAMIC_MUTANT") != "M3" && result&predicateFalse != 0 && wanted {'),
 ('M4', 'yes = v.cells[path.cell] == constant.Text()', 'yes = v.cells[path.cell] != constant.Text()', 'yes = (v.cells[path.cell] == constant.Text()) != (os.Getenv("ADAMIC_MUTANT") == "M4")'),
 ('M5', 'if claim == nil || claim.ParameterIndex() != 0 {', 'if claim == nil {', 'if claim == nil || (os.Getenv("ADAMIC_MUTANT") != "M5" && claim.ParameterIndex() != 0) {'),
 ('M6', 'return ast.IsIdentifier(n) && v.l.checker.GetSymbolAtLocation(n) == parameter', 'return false', 'return os.Getenv("ADAMIC_MUTANT") != "M6" && ast.IsIdentifier(n) && v.l.checker.GetSymbolAtLocation(n) == parameter'),
]
assert (root / file).read_text() == base, 'verifier source differs from origin/main'
out = Path(__file__).resolve().parent
switched = base.replace('"fmt"', '"fmt"\n\t"os"', 1)
manifest = []
for mid, old, new, switch in mutants:
 assert base.count(old) == 1, (mid, base.count(old))
 line = base[:base.index(old)].count('\n') + 1
 changed = base.replace(old, new, 1)
 diff = ''.join(difflib.unified_diff(base.splitlines(True), changed.splitlines(True), fromfile='a/'+file, tofile='b/'+file))
 (out / (mid + '.diff')).write_text(diff)
 manifest.append(dict(id=mid, file=file, line=line, before=old.splitlines()[0], after=new.splitlines()[0]))
 switched = switched.replace(old, switch, 1)
(out / 'mutants.json').write_text(json.dumps(manifest, indent=2)+'\n')
(root / file).write_text(switched)
