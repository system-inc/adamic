from pathlib import Path
import json,shutil,hashlib
root=Path('/workspace/scratch/ev14ch9-landing-proof'); before=root/'before'; after=root/'after'
if after.exists(): shutil.rmtree(after)
shutil.copytree(before,after,ignore=shutil.ignore_patterns(".git"))
rows=json.loads((root/'sites.json').read_text())
changed=[]
for file in sorted({r['site'].split(':')[0] for r in rows}):
 p=after/file; raw=p.read_bytes(); s=raw.decode(); nl='\r\n' if b'\r\n' in raw else '\n'; text=s.replace('\r\n','\n')
 names=set()
 selected=[r for r in rows if r['site'].split(':')[0]==file]
 for call in dict.fromkeys(r['call'] for r in selected):
  group=[r for r in selected if r['call']==call];row=group[0]
  assert text.count(call)==len(group),(row['site'],text.count(call),len(group))
  assert len({r['overload'] for r in group})==1
  helper='filterNarrowed' if row['overload']==1 else 'filterNarrowedOptional'
  text=text.replace(call,helper+call[len('filter'):]);names.add(helper)
 assert text.count('    filter,\n')==1,file
 text=text.replace('    filter,\n','    filter,\n'+''.join('    '+n+',\n' for n in sorted(names)),1)
 p.write_bytes(text.replace('\n',nl).encode());changed.append(file)
p=after/'src/compiler/core.ts';s=p.read_bytes().decode();nl='\r\n' if '\r\n' in s else '\n';text=s.replace('\r\n','\n'); helpers=Path('/workspace/scratch/ev14ch9-proof/isolated/helpers/core.ts').read_text().split('function isString')[0];start=text.index('export function filter<'); impl=text.index('export function filter<T>(array: readonly T[] | undefined, f: (x: T) => boolean): readonly T[] | undefined {'); end=text.index('\n}',impl)+2; group=text[start:end]; lines=group.splitlines(); ordinal=0; revised=[]
for line in lines:
 if line.startswith('export function filter') and line.endswith(';'):
  ordinal+=1
  if ordinal in (1,5): line=''
 revised.append(line)
assert ordinal==8
text=text[:start]+'\n'.join(revised)+'\n\n'+helpers.rstrip()+text[end:];p.write_bytes(text.replace('\n',nl).encode());changed.append('src/compiler/core.ts')
(root/'changed-files.json').write_text(json.dumps(changed,indent=2)+'\n')
print(json.dumps(changed))
