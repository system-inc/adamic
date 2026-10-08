"""Generate the pinned Go file inventory; lexical reference sets are conservative.
Go has package-wide declarations, not imports between files. References include method names
and collisions; planning groups resolve cycles rather than claiming a false file DAG.
"""
import pathlib,re
root=pathlib.Path(__file__).resolve().parents[4]
package=root/'cohere/internal/lint/ecmascript/high_level_intermediate_representation'
files=sorted(package.glob('*.go'))
def symbols(text):
 return sorted(set(re.findall(r'^type\s+(\w+)|^func\s+(?:\([^\n]*?\)\s+)?(\w+)',text,re.M)),key=str)
def names(text):
 result={a or b for a,b in symbols(text)}
 for block in re.findall(r'(?m)^const\s*\((.*?)^\)',text,re.S):
  result.update(re.findall(r'(?m)^\s*([A-Za-z_]\w*)\s*(?:[A-Za-z_][\w.]*\s*)?(?:=|$)',scrub(block)))
 result.update(re.findall(r'(?m)^const\s+(\w+)',text))
 return sorted(result)
def scrub(text):return re.sub(r'//[^\n]*|/\*.*?\*/|"(?:\\.|[^"\\])*"|`[^`]*`',' ',text,flags=re.S)
data={f.name:f.read_text() for f in files}
owners={}
for name,text in data.items():
 if name.endswith('_test.go'):continue
 for symbol in names(text):owners.setdefault(symbol,set()).add(name)
rows=[]
for name,text in data.items():
 definitions=names(text);tokens=set(re.findall(r'\b\w+\b',scrub(text)))
 deps=sorted({owner for token in tokens for owner in owners.get(token,[]) if owner!=name})
 imports=[]
 for block in re.findall(r'(?m)^import\s+\((.*?)\)|^import\s+("[^"\n]+")',text,re.S):
  imports+=re.findall(r'"([^"]+)"',''.join(block))
 tests=sorted(test for test,source in data.items() if test.endswith('_test.go') and set(definitions)&set(re.findall(r'\b\w+\b',scrub(source)))) if not name.endswith('_test.go') else []
 rows.append((name,len(text.splitlines()),definitions,deps,imports,tests))
out=['# File inventory','',f'Pinned cohere `{__import__("subprocess").check_output(["git","-C",str(root/"cohere"),"rev-parse","HEAD"],text=True).strip()}`. Go has no intra-package import statements. “Package references” below are conservative declaration-name references after comments and literals are stripped; method-name collisions can add edges. This is a review aid, not a promised acyclic file graph. Test associations are symbol references; same-stem tests additionally test the named pass. All test files are inventoried too.','']
for name,lines,defs,deps,imports,tests in rows:
 purpose = next((line.removeprefix('//').strip() for line in data[name].splitlines() if line.startswith('//') and not line.startswith('//go:')), 'Tests of the declarations listed below.' if name.endswith('_test.go') else 'HIR declarations and operations.')
 out += [f'## {name} — {lines} lines','',f'Purpose: {purpose}',f'Defines: {", ".join(defs) or "constants / initialization"}.',f'Package references: {", ".join(deps) or "none"}.',f'Imports: {", ".join(imports) or "none"}.']
 if tests:out += [f'Test references: {", ".join(tests)}.']
 out+=['']
(root/'stage1/cohere/high_level_intermediate_representation/INVENTORY.md').write_text('\n'.join(out))
print('production',sum(not row[0].endswith('_test.go') for row in rows),sum(row[1] for row in rows if not row[0].endswith('_test.go')),'tests',sum(row[1] for row in rows if row[0].endswith('_test.go')))
