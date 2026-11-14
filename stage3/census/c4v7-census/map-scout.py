"""Map the scout's 191 sites with source hashes, including its older sys.ts."""
import difflib,hashlib,json,subprocess
from pathlib import Path
OUT=Path(__file__).resolve().parent;BASE=Path('/workspace/scratch/c4v7-census')
def blob(ref,p):return subprocess.check_output(['git','show',ref+':'+p])
ref='origin/typescript/append-constraint-scout';prefix='stage3/census/append-constraint-scout/evidence/'
coverage=json.loads(blob(ref,prefix+'coverage.json'));hashes=coverage['manifest']['source_sha256'];selected=json.loads((OUT/'selected.json').read_text())['findings']
old=json.loads(blob('894c8a2b2e7a5bd1fd6adc3738a6d92ae16de558','stage3/adapt/99-correctness-timer-closures/owners.json'))
new=json.loads(blob('5c8338c4232c37afb304ce6865458b67655c590a','stage3/adapt/99-correctness-timer-closures/owners.json'))
current=(BASE/'adapted-main/src/compiler/sys.ts').read_bytes();reconstructed=current.decode().replace('\r\n','\n');replacements=[]
for a,b in zip(new,old):
 assert a['file']==b['file'] and a['owner']==b['owner']
 if a['file']=='src/compiler/sys.ts' and a['after']!=b['after']:
  assert reconstructed.count(a['after'])==1
  reconstructed=reconstructed.replace(a['after'],b['after'])
  replacements.append(dict(owner=a['owner'],old_after_sha256=b['ownerAfterSha256'],main_after_sha256=a['ownerAfterSha256']))
raw=reconstructed.replace('\n','\r\n').encode();assert hashlib.sha256(raw).hexdigest()==hashes['sys.ts']
a=reconstructed.splitlines();b=current.decode().splitlines();mapping={}
for block in difflib.SequenceMatcher(None,a,b,autojunk=False).get_matching_blocks():
 for offset in range(block.size):mapping[block.a+offset+1]=block.b+offset+1
entries=[]
for f in selected:
 file=f['file'];rawcurrent=(BASE/'adapted-main/src/compiler'/file).read_bytes();identical=hashlib.sha256(rawcurrent).hexdigest()==hashes[file]
 line=f['line']
 if not identical:
  assert file=='sys.ts';assert line in mapping;line=mapping[line]
  assert a[f['line']-1]==b[line-1]
 entries.append(dict(file=file,scout_line=f['line'],main_line=line,column=f['column'],kind=f['kind'],reason=f['reason'],source_identical=identical))
assert sum(not e['source_identical'] for e in entries)==1
assert next(e for e in entries if not e['source_identical'])['main_line']==304
result=dict(scout_commit='894c8a2b2e7a5bd1fd6adc3738a6d92ae16de558',scout_branch_tip='04a12030beefab45f602f7541795221831ba066b',scout_sys_sha256=hashes['sys.ts'],reconstructed_scout_sys_sha256=hashlib.sha256(raw).hexdigest(),reconstruction='Replace the two changed timer-closure owner after pins in main sys.ts with their scout after pins; CRLF restored; full hash matches scout coverage.',owner_replacements=replacements,identity_sites=190,remapped_sites=1,sys_context=dict(scout='\n'.join(a[301:310]),main='\n'.join(b[298:307])),sites=entries)
(OUT/'scout-site-mapping.json').write_text(json.dumps(result,indent=2)+'\n')
print('191 scout sites mapped: 190 identity, sys.ts:307:13 -> sys.ts:304:13; full scout sys hash verified')
