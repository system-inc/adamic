from pathlib import Path
import json,difflib,re,subprocess
out=Path('review/test-audit/internal-lower-enums'); (out/'diffs').mkdir(exist_ok=True); (out/'switch').mkdir(exist_ok=True)
menu=json.loads((out/'menu.json').read_text())
# Exact negation is within the condition-flip menu.
menu[3]['new']='!(math.IsNaN(value) || math.IsInf(value, 0))';menu[3]['menu']='flip non-finite rejection condition'
(out/'menu.json').write_text(json.dumps(menu,indent=2)+'\n')
files={m['file'] for m in menu}|{'internal/lower/lower.go'}
base={f:subprocess.check_output(['git','show','origin/main:'+f],text=True) for f in files}; switched=base.copy()
def body_span(s,entry):
 a=s.index(entry); b=s.index('{',a); depth=1;i=b+1;mode=None
 while depth:
  ch=s[i]; nxt=s[i:i+2]
  if mode=='line':
   if ch=='\n':mode=None
  elif mode=='block':
   if nxt=='*/':mode=None;i+=1
  elif mode in ['"',"'",'`']:
   if ch==mode:mode=None
   elif ch=='\\' and mode!='`':i+=1
  elif nxt=='//':mode='line';i+=1
  elif nxt=='/*':mode='block';i+=1
  elif ch in ['"',"'",'`']:mode=ch
  elif ch=='{':depth+=1
  elif ch=='}':depth-=1
  i+=1
 return b,i

def empty_body(s,entry,ret):
 b,e=body_span(s,entry);return s[:b+1]+'\n\t'+ret+'\n'+s[e-1:]

def clean_imports(s):
 # Only remove an import if none of its package selectors survive.
 for pkg in re.findall(r'^\s*"([^"]+)"$',s,re.M):
  name=pkg.split('/')[-1]
  if not re.search(r'\b'+re.escape(name)+r'\.',s):s=re.sub(r'^\s*"'+re.escape(pkg)+r'"\n','\n',s,flags=re.M)
 return s

def diff(mid,f,mut):
 mut=clean_imports(mut);(out/'diffs'/f'{mid}.diff').write_text(''.join(difflib.unified_diff(base[f].splitlines(True),mut.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
for m in menu:
 mid,f,old,new=m['id'],m['file'],m['old'],m['new']; s=base[f];assert s.count(old)==1,(mid,s.count(old))
 if m['entry_return']:
  diff(mid,f,empty_body(s,old,new));switched[f]=switched[f].replace(old,old+'\n\tif auditMutant("'+mid+'") { '+new+' }',1)
 elif mid=='M08':
  diff(mid,f,s.replace('\t'+old+'\n','',1)); switched[f]=switched[f].replace(old,'if !auditMutant("M08") { '+old+' }',1)
 elif mid=='M09':
  diff(mid,f,s.replace(old,new,1));switched[f]=switched[f].replace('l.genericDepth >= maximumGenericDepth','l.genericDepth >= auditChoice("M09", maximumGenericDepth, 0)',1)
 elif mid=='M13':
  diff(mid,f,s.replace(old,new,1));switched[f]=switched[f].replace(old,'range auditChoice("M13", inputFunctions, inputFunctions[:len(inputFunctions)-1]) {',1)
 elif old.startswith('return '):
  diff(mid,f,s.replace(old,new,1));switched[f]=switched[f].replace(old,'return auditChoice("'+mid+'", '+old[7:]+', '+new[7:]+')',1)
 elif old.startswith('if '):
  diff(mid,f,s.replace(old,new,1));switched[f]=switched[f].replace(old,'if auditChoice("'+mid+'", '+old[3:-2]+', '+new[3:-2]+') {',1)
 elif mid=='M18':
  diff(mid,f,s.replace(old,new,1));switched[f]=switched[f].replace(old,'cyclic = auditChoice("M18", true, false)',1)
 else:
  diff(mid,f,s.replace(old,new,1));switched[f]=switched[f].replace(old,'auditChoice("'+mid+'", '+old+', '+new+')',1)
probes=[('PLower','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return nil, nil'),('POrder','internal/lower/modules.go','func esmModuleOrder(typeChecker *checker.Checker, entry *ast.SourceFile) ([]*ast.SourceFile, bool) {','return nil, false'),('PProve','internal/lower/load_time_reads.go','func (l *lowering) proveModuleReads(files []*ast.SourceFile) {','return')]
for mid,f,entry,ret in probes:
 diff(mid,f,empty_body(base[f],entry,ret));switched[f]=switched[f].replace(entry,entry+'\n\tif auditMutant("'+mid+'") { '+ret+' }',1)
for f,s in switched.items():Path(f).write_text(s)
Path('internal/lower/audit_switch.go').write_text('''package lower

import "os"

// Scratch-only audit switch scaffolding, excluded from standalone diffs.
func auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }
func auditChoice[T any](id string, normal, mutant T) T {
 if auditMutant(id) { return mutant }; return normal
}
''')
(out/'base-files.json').write_text(json.dumps(sorted(files),indent=2)+'\n')
