from pathlib import Path
import hashlib,json,difflib
W=Path('/tmp/family5-alternating');main=Path('/workspace/family5-main-lane-alone/adapted-tree');c02=Path('/workspace/family5-c02-lane/adapted-tree')
def inventory(root):
 out={}
 for d in ['built/local','lib']:
  for p in sorted((root/d).rglob('*')):
   if p.is_file():out[str(p.relative_to(root))]=p
 return out
ma,ca=inventory(main),inventory(c02);rows=[];hunks=[]
for key in sorted(ma.keys()|ca.keys()):
 a=ma.get(key);b=ca.get(key);x=a.read_bytes() if a else b'';y=b.read_bytes() if b else b''
 equal=bool(a and b and x==y);off=None
 if not equal:
  off=next((i for i,(u,v) in enumerate(zip(x,y)) if u!=v),min(len(x),len(y)))
  hunks+=['\n### '+key,f'main bytes={len(x) if a else None}; C02 bytes={len(y) if b else None}; first difference at 0-based byte {off}.']
  if key.endswith('.d.ts'):
   diff=list(difflib.unified_diff(x.decode(errors='replace').splitlines(),y.decode(errors='replace').splitlines(),fromfile='main/'+key,tofile='C02/'+key,n=2))
   hunks+=['```diff']+[l if len(l)<800 else l[:800]+' [line clipped]' for l in diff[:60]]+['```']
  else:
   start=max(0,off-90);end=off+160
   hunks+=['```text',f'@@ bytes [{start}:{end}] around first difference @@','main: '+repr(x[start:end]),'C02:  '+repr(y[start:end]),'```']
 rows.append({'file':key,'main_present':bool(a),'c02_present':bool(b),'main_bytes':len(x) if a else None,'c02_bytes':len(y) if b else None,'main_sha256':hashlib.sha256(x).hexdigest() if a else None,'c02_sha256':hashlib.sha256(y).hexdigest() if b else None,'byte_identical':equal,'first_difference_offset_0based':off})
summary={'main_tree':str(main),'c02_tree':str(c02),'c02_provenance':'Untouched original full C02 lane, before any targeted reruns or common-path source selection.','main_files':len(ma),'c02_files':len(ca),'javascript_files':sum(r['file'].endswith(('.js','.mjs','.cjs')) for r in rows),'javascript_all_identical':all(r['byte_identical'] for r in rows if r['file'].endswith(('.js','.mjs','.cjs'))),'differing_files':[r for r in rows if not r['byte_identical']]}
(W/'artifact-manifest.json').write_text(json.dumps(rows,indent=2)+'\n');(W/'artifact-summary.json').write_text(json.dumps(summary,indent=2)+'\n');(W/'artifact-hunks.md').write_text('# Built artifact differences\n'+'\n'.join(hunks)+'\n')
print('Compared',len(rows),'files; JavaScript',summary['javascript_files'],'all identical',summary['javascript_all_identical'],'different files',len(summary['differing_files']))
for r in summary['differing_files']:print(r['file'],r['main_bytes'],r['c02_bytes'],r['first_difference_offset_0based'])
