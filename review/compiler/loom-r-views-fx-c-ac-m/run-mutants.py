import json,subprocess,time,difflib,os,re
from pathlib import Path
out=Path(__file__).resolve().parent;root=Path.cwd();results=[]
# Read through the positive run's immutable cache; publish new mutant actions only on the root volume.
source=Path('/tmp/adamic-gate/loom-chain-go-cache');cache=Path('/workspace/scratch/views-fx-c-mutant-go-cache');cache.mkdir(parents=True,exist_ok=True)
linked=0
for p in source.rglob('*'):
 relative=p.relative_to(source);target=cache/relative
 if p.is_dir():target.mkdir(exist_ok=True)
 elif p.is_file() and not target.exists() and not target.is_symlink():target.symlink_to(p);linked+=1
(out/'mutant-cache.json').write_text(json.dumps(dict(source=str(source),new_writes=str(cache),linked=linked),indent=2))
environment=dict(os.environ,ADAMIC_GOCACHE_DIR=str(cache))
for row in json.loads((out/'mutant-plan.json').read_text()):
 originals={};patches=[]
 try:
  for relative,before,after in row['changes']:
   p=root/relative;original=p.read_bytes();originals[p]=original;text=original.decode();assert text.count(before)==1,(row['name'],relative,text.count(before));changed=text.replace(before,after,1);p.write_text(changed);patches.extend(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/'+relative,tofile='b/'+relative))
  (out/(row['name']+'.patch')).write_text(''.join(patches))
  cmd=['go','test',row['package'],'-run',row['test'],'-count=1','-timeout','90s','-parallel','4','-p','1','-json'];start=time.monotonic();log=out/('mutant-'+row['name']+'.log')
  with log.open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=180,env=environment)
  text=log.read_text();events=[]
  for line in text.splitlines():
   try:e=json.loads(line)
   except ValueError:continue
   if isinstance(e,dict) and 'Test' in e and e.get('Action') in ('pass','fail'):events.append(e)
  bad=[e for e in events if e.get('Elapsed',0)>=60 and not any(x['Test'].startswith(e['Test']+'/') for x in events)]
  failed=[e for e in events if e['Action']=='fail'];verified=(r.returncode==0 and bool(events)) if row['green'] else r.returncode==1 and bool(failed) and '[build failed]' not in text and 'panic: test timed out' not in text and not re.search(r'\.c:[0-9]+:[0-9]+: (?:fatal )?error:',text)
  record=dict(name=row['name'],command=cmd,exit=r.returncode,expected_green=row['green'],verified=verified,seconds=round(time.monotonic()-start,3),catchers=failed);results.append(record);print(row['name'],r.returncode,record['seconds'],verified,flush=True)
  assert verified and not bad,(row['name'],bad)
 finally:
  for p,original in originals.items():p.write_bytes(original)
  (out/'mutant-results.json').write_text(json.dumps(results,indent=2))
