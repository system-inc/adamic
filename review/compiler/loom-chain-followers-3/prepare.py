import difflib,json,subprocess
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-chain-followers-3';old=root/'review/compiler/loom-chain-followers-2';source='origin/compiler/loom-chain-followers-2'
(out/'mutants').mkdir(exist_ok=True)
cases=json.loads((old/'mutants/cases.json').read_text())
for c in cases:
 target=c['file'];base=subprocess.check_output(['git','show',source+':'+target],text=True);mutant=(old/'mutants'/(c['topic']+'.go.txt')).read_text();current=(root/target).read_text();original=current
 for tag,a,b,x,y in difflib.SequenceMatcher(None,base.splitlines(True),mutant.splitlines(True),autojunk=False).get_opcodes():
  if tag=='equal':continue
  before=''.join(base.splitlines(True)[a:b]);after=''.join(mutant.splitlines(True)[x:y])
  if not before:
   anchor=''.join(base.splitlines(True)[max(0,a-1):a]);assert current.count(anchor)==1,(c['topic'],anchor);current=current.replace(anchor,anchor+after,1)
  else:
   assert current.count(before)==1,(c['topic'],before);current=current.replace(before,after,1)
 assert current!=original,c['topic']
 path=out/'mutants'/(c['topic']+'.go.txt');path.write_text(current)
 patch=out/'mutants'/(c['topic']+'.patch');patch.write_text(''.join(difflib.unified_diff(original.splitlines(True),current.splitlines(True),fromfile=target,tofile=target)))
 overlay=out/'mutants'/(c['topic']+'.json');overlay.write_text(json.dumps({'Replace':{str(root/target):str(path)}}));c['overlay']=str(overlay)
(out/'mutants/cases.json').write_text(json.dumps(cases,indent=2))
print('Prepared 14 source mutants as deltas on the reconciled tree; main changes retained.')
