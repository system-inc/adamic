from pathlib import Path
import subprocess,json,difflib
out=Path(__file__).resolve().parent;source='05eb9043dc8b01b32402ce33b91f97aadf524305';base='dbdf977f5b8b0979d414ebe6f20e7aadd312e309'
def git(*args):return subprocess.check_output(['git',*args],text=True)
a=git('diff',source+'^',source,'--','internal/native/runtime');b=git('diff',base,'--','internal/native/runtime');(out/'source-runtime-patch.json').write_text(json.dumps(dict(patch=a),indent=2)+'\n');(out/'rebased-runtime-patch.json').write_text(json.dumps(dict(patch=b),indent=2)+'\n')
def lines(patch):
 files={};name=None
 for line in patch.splitlines():
  if line.startswith('diff --git '):name=line.split(' b/',1)[1];files[name]=[]
  elif line[:1] in ['+','-'] and not line.startswith(('+++','---')):files[name].append(line)
 return files
x,y=lines(a),lines(b);records=[];md=['Every difference in the source and re-cut runtime edit lines follows. Minus/plus inside each fenced diff describes the edit-sequence comparison; the second minus/plus is the actual runtime deletion/addition. Full patches preserve context.','']
for name in sorted(x.keys()|y.keys()):
 if x.get(name)==y.get(name):continue
 delta='\n'.join(difflib.unified_diff(x.get(name,[]),y.get(name,[]),fromfile='source own edits',tofile='re-cut own edits',lineterm=''))
 records.append(dict(file=name,source=x.get(name,[]),rebased=y.get(name,[]),comparison=delta));md += ['### '+name,'','```diff',delta,'```','']
(out/'runtime-line-differences.md').write_text('\n'.join(md));(out/'runtime-comparison.json').write_text(json.dumps(dict(source=source,base=base,identical=not records,differences=records),indent=2)+'\n')
print([r['file'] for r in records])
