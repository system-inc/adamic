#!/usr/bin/env python3
"""Retain actual logs, patches and identities; no scratch binary becomes an oracle."""
import difflib,gzip,hashlib,json,re,shutil
from pathlib import Path
P=Path(__file__).resolve().parent;E=P/'evidence';W=Path('/workspace/cache/step31-binder-walk-final');original=Path('/workspace/cache/step31-binder-driver-slice-v2');tree=W/'tree'
def gz(name,data):
 with gzip.GzipFile(filename=str(E/name),mode='wb',mtime=0) as f:f.write(data)
logs={'setup':'/tmp/step31-binder-setup.log','fetch':'/tmp/step31-binder-fetch.log','compiler-build':'/tmp/step31-binder-area-build.log','slice':'/tmp/step31-binder-driver-slice.log','slice-verify':'/tmp/step31-binder-driver-slice-verify.log','split0':'/tmp/step31-binder-split0.log','split1':'/tmp/step31-binder-split1.log','walk':'/tmp/step31-binder-walk-final.log','fixtures':'/tmp/step31-binder-fixtures-final.log','node-comparison':'/tmp/step31-binder-port-final.log','comparator-mutants':'/tmp/step31-binder-comparator-mutants.log','flags-mutant':'/tmp/step31-binder-flags-mutant.log'}
for name,file in logs.items():gz(name+'.log.gz',Path(file).read_bytes())
for i in range(1,16):
 for ext in ['stdout','stderr']:gz(f'stop-{i:02}.{ext}.gz',(W/f'stop-{i:02}.{ext}').read_bytes())
 shutil.copyfile(W/f'patch-{i:02}.json',E/f'patch-{i:02}.json')
 shutil.copyfile(W/f'stop-{i:02}.source.gz',E/f'stop-{i:02}.source.gz')
 gz(f'patch-{i:02}.diff.gz',(W/f'patch-{i:02}.diff').read_bytes())
legacy=Path('/workspace/adamic/stage3/drivers/scanner/evidence/combined-records-library/stops.json');scanner=json.loads(legacy.read_text());shutil.copyfile(legacy,E/'scanner-stops.json')
rows=json.loads((W/'stops.json').read_text())
for r in rows:
 i=r['order'];r['owner']='library' if i==2 else 'adaptation' if i in [10,11,12,14,15] else 'compiler'
 r['ownerReason']='ErrorConstructor captureStackTrace declaration; host stack formatting is separate runtime work' if i==2 else 'placeholder-induced initialization/narrowing obligation' if r['owner']=='adaptation' else 'iterator completion attribution' if i==1 else 'enum/namespace module initialization proof'
 r['scannerExactMessageOrders']=[s['order'] for s in scanner if s['message']==r['message']]
 r['scannerTaskId']='not present in repository; compare the pinned available fifteen-stop artifact'
 r['qualification']='depends on accumulated throwing replacements; not a native acceptance result'
 if i in [10,11,12,14,15]:r['qualification']+='; site is inside or caused by a placeholder'
 r['probe']=f'fixtures/{i:02}.a';r['patch']=f'evidence/patch-{i:02}.json'
 r['sourceHashAtFinalWalk']=hashlib.sha256((tree/r['file']).read_bytes()).hexdigest()
(P/'stops.json').write_text(json.dumps(rows,indent=2)+'\n')
changes=[]
for file in sorted(tree.rglob('*')):
 if file.is_file() and (original/file.relative_to(tree)).exists() and file.suffix in ['.ts','.a']:
  before=(original/file.relative_to(tree)).read_text();after=file.read_text()
  if before!=after:changes.extend(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='original/'+str(file.relative_to(tree)),tofile='discovery/'+str(file.relative_to(tree))))
gz('discovery.patch.gz',''.join(changes).encode())
gz('slice.json.gz',(original/'slice.json').read_bytes())
hashes={str(file.relative_to(original)):hashlib.sha256(file.read_bytes()).hexdigest() for file in sorted(original.rglob('*')) if file.is_file() and file.suffix in ['.ts','.a']};(P/'input-hashes.json').write_text(json.dumps(hashes,indent=2)+'\n')
manifest={}
for name,file in [('request','/workspace/cache/step31-binder-acceptance.json'),('stock','/tmp/step31-binder-golden.jsonl'),('sourceNode','/workspace/cache/step31-binder-port-final/actual.stdout'),('compiler','/workspace/cache/step31-binder-area-adamic')]:
 data=Path(file).read_bytes();manifest[name]={'path':file,'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()}
assert manifest['stock']['sha256']==manifest['sourceNode']['sha256']
manifest['projects']=301;manifest['nativeExecution']=False;manifest['splitModes']={'0':{'exit':1},'1':{'exit':1,'jobs':5}};manifest['scannerArtifactSHA256']=hashlib.sha256(legacy.read_bytes()).hexdigest()
for label,file in [('node-control-report','/workspace/cache/step31-binder-port-final/report.json'),('flags-mutant-report','/workspace/cache/step31-binder-flags-mutant/report.json')]:shutil.copyfile(file,E/(label+'.json'))
gz('flags-mutant.jsonl.gz',Path('/workspace/cache/step31-binder-flags-mutant/actual.stdout').read_bytes())
shutil.copyfile('/workspace/cache/step31-binder-flags-mutant/actual.exit',E/'flags-mutant.exit')
shutil.copyfile('/workspace/cache/step31-binder-flags-mutant/actual.stderr',E/'flags-mutant.stderr')
gz('stock-golden.jsonl.gz',Path('/tmp/step31-binder-golden.jsonl').read_bytes());gz('request.json.gz',Path('/workspace/cache/step31-binder-acceptance.json').read_bytes())
(P/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
lines=['| Order | Progressive scratch location | Message | Owner | Scanner exact match | Witness |','|---:|---|---|---|---|---|']
for r in rows:lines.append(f"| {r['order']} | {r['file']}:{r['line']}:{r['column']} | {r['message'].replace('|','&#124;')} | {r['owner']} | "+(','.join(map(str,r['scannerExactMessageOrders'])) or 'none')+f" | [{r['order']:02}.a]({r['probe']}) |")
(P/'STOPS.md').write_text('\n'.join(lines)+'\n')
print('15 stops packaged; 301 source Node projects equal stock; native unavailable')
