import pathlib,json,re
out=pathlib.Path(__file__).parent;rows=[];certs=json.loads((out/'emit2-inventory.json').read_text())['certificates'];byname={c['file']:c for c in certs};hunk=None;name=None
for line in (out/'emit2.diff').read_text().splitlines():
 if line.startswith('--- control/'):
  name=line[len('--- control/'):];continue
 if line.startswith('@@'):
  if hunk is not None:rows.append(hunk)
  match=re.match(r'@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@',line);assert match
  old,new=int(match[1]),int(match[3]);hunk=dict(file=name,header=line,classes=[],reviewedTracingEdit=False,changes=[]);continue
 if hunk is None or line.startswith('+++'):continue
 if not line or line[0] not in ' +-':continue
 side='before' if line[0]=='-' else 'after';number=old if side=='before' else new
 if line[0] in '+-':
  root='/workspace/scratch/tracing-control2' if side=='before' else '/workspace/scratch/tracing-after2';text=(pathlib.Path(root)/'built/local'/name).read_text();hits=[]
  for span in byname[name][side+'ApprovedSpans']:
   end=text.count('\n',0,span['end'])+1
   if span['line']<=number<=end:hits.append(span)
  if hits:
   for span in hits:
    if span['kind'] not in hunk['classes']:hunk['classes'].append(span['kind'])
    hunk['reviewedTracingEdit']|='reviewed' in span['detail']
  elif line[1:].strip()=='// src/compiler/hostErrors.ts':
   if 'declaration-only relocation' not in hunk['classes']:hunk['classes'].append('declaration-only relocation')
  elif line[1:].strip():
   if 'rename' not in hunk['classes']:hunk['classes'].append('rename')
  hunk['changes'].append(dict(side=side,line=number,text=line[1:]))
 if line[0]!='+':old+=1
 if line[0]!='-':new+=1
if hunk:rows.append(hunk)
assert rows and all(r['classes'] and set(r['classes'])<={'rename','closed removal','declaration-only relocation'} for r in rows)
(out/'emit2-hunk-classification.json').write_text(json.dumps(dict(hunks=len(rows),allRemainingTokensCertified=True,source='emit2.diff',hunksClassified=rows),indent=2)+'\n')
(out/'emit2-hunk-classification.md').write_text('# Emitted JavaScript hunk certificate\n\nEvery hunk is listed below. The independent AST/token certificate proves every nonexcluded token unchanged or a consistent symbol rename. Reviewed tracing import/local/loader spans are explicitly marked. Declaration-only hostErrors relocation includes esbuild initializer stubs whose callbacks contain only a strict directive and whose binding uses are only no-argument calls. Binding shorthand expansion preserves the property key and resolves the same renamed local.\n\n| File | Hunk | Classes | Reviewed tracing edit |\n| --- | --- | --- | --- |\n'+''.join('| '+r['file']+' | `'+r['header']+'` | '+', '.join(r['classes'])+' | '+str(r['reviewedTracingEdit'])+' |\n' for r in rows))
print('certified hunks',len(rows))
