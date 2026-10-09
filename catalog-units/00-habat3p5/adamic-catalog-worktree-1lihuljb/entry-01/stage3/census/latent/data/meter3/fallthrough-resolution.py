from pathlib import Path
import re,subprocess
p=Path('/tmp/latent3-tree');assert subprocess.check_output(['git','branch','--show-current'],cwd=p,text=True).strip().startswith('scratch/latent-')
files=subprocess.check_output(['git','diff','--name-only','--diff-filter=U'],cwd=p,text=True).splitlines()
pat=re.compile(r'^<<<<<<<[^\n]*\n(.*?)^=======\n(.*?)^>>>>>>>[^\n]*\n',re.M|re.S)
for name in files:
 f=p/name;s=f.read_text()
 def resolve(m):
  a,b=m.groups()
  if name=='docs/0.1.md':return a.split('- `switch` cases')[0]+b[b.index('- `switch` cases'):] if '- Conditions' in a else a
  if name=='tsconfig.json':return '\t\t"erasableSyntaxOnly": false,\n'
  if name=='internal/load/load.go':return '\t\tErasableSyntaxOnly:         core.TSFalse,\n'
  if name=='internal/load/load_test.go':return '' if 'noImplicitReturns' in a else a+'}\n\n'+b
  if name=='internal/lower/refusals.go':return a # main's accessor support
  if name=='internal/ir/ir.go':return '\tBreak struct { Label string; Depth int }\n\tContinue struct { Label string }\n'
  if name=='internal/lower/statements.go':return a # retain full taste label/continue support; fallthrough paths outside this hunk
  if name=='internal/flow/build.go':return a.replace('depth := len(b.jumps) - 1','depth := len(b.jumps) - 1 - statement.Depth')
  if name=='internal/javascript/javascript.go':return a.replace('e.line("break;")', 'e.line("break %s;", e.breakables[len(e.breakables)-1-statement.Depth])')
  if name=='internal/native/emit_statements.go':return a[:a.index('\t\t// A try')] + b
  if name=='internal/oracle/counts.md':
   keys={line.split('|')[1].strip() for line in a.splitlines() if line.startswith('|')}
   return a+''.join(line+'\n' for line in b.splitlines() if not line.startswith('|') or line.split('|')[1].strip() not in keys)
  raise RuntimeError(name)
 s=pat.sub(resolve,s);assert '<<<<<<<' not in s;f.write_text(s)
 subprocess.run(['git','add',name],cwd=p,check=True)
 print(name)
subprocess.run(['gofmt','-w',*[str(p/f) for f in files if f.endswith('.go')]],check=True)
subprocess.run(['git','add','--update'],cwd=p,check=True)
