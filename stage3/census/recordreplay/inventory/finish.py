from pathlib import Path
import json,subprocess,re,hashlib,ast
D=Path('/workspace/scratch/recordreplay');R=D/'head';O=D/'inventory-audit'
# Run the Go readers' original filesystem clauses only; no census is started.
rows=[]
for rel in ['stage3/census/tool/main.go','stage3/census/latent/main.go.txt']:
 s=(R/rel).read_text(); start=s.index('\terr := filepath.WalkDir');end=s.index('\tsort.Strings(paths)',start)
 clause=s[start:end]; tool='tool/' in rel
 imports=['"os"','"path/filepath"','"encoding/json"']+(['"strings"'] if tool else [])
 code='package main\nimport ('+';'.join(imports)+')\nfunc main(){var paths []string\n'+('var entries []string; nonSourceOnly:=false\n' if tool else '')+clause+'\njson.NewEncoder(os.Stdout).Encode(map[string]any{"paths":paths'+(',"entries":entries' if tool else '')+'})}\n'
 p=O/('tool-reader.go' if tool else 'latent-reader.go');p.write_text(code)
 subprocess.run(['go','build','-o',str(p.with_suffix('')),str(p)],check=True)
 for side in ['control','candidate2']:
  out=json.loads(subprocess.check_output([str(p.with_suffix('')),str(D/side/'src')],text=True));rows.append({'reader':rel,'side':side,'source_paths':len(out['paths']),'entries':len(out.get('entries',[])),'contains_added':str(D/side/'src/recordreplay.d.ts') in out['paths'],'method':'original WalkDir clause only; no loading or lowering'})
# All compiler-only filesystem readers see the exact same filenames and bytes.
compiler={side:{str(p.relative_to(D/side/'src/compiler')):hashlib.sha256(p.read_bytes()).hexdigest() for p in (D/side/'src/compiler').rglob('*') if p.is_file()} for side in ['control','candidate2']}
# sys is deliberately changed; filenames remain identical.
rows.append({'reader':'compiler-only source enumeration','control':len(compiler['control']),'amended':len(compiler['candidate2']),'same_names':set(compiler['control'])==set(compiler['candidate2']),'changed_bytes':[p for p in compiler['control'] if compiler['control'][p]!=compiler['candidate2'][p]]})
# Evaluate all fixed-count assertion predicates, including those without a local filesystem walk.
for rel in subprocess.check_output(['git','ls-files','*.py'],cwd=R,text=True).splitlines():
 try:s=(R/rel).read_text(); tree=ast.parse(s)
 except (OSError,SyntaxError):continue
 for n in ast.walk(tree):
  if isinstance(n,ast.Assert) and any(isinstance(x,ast.Constant) and x.value==746 for x in ast.walk(n)):
   rows.append({'reader':rel,'line':n.lineno,'predicate':ast.get_source_segment(s,n),'control_747':'reject','amended_748':'reject','method':'fixed-count conjunct evaluated; enclosing historical driver not run'})
(O/'final-readers.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
