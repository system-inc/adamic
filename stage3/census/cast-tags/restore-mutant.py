from pathlib import Path
import shutil,json,time,subprocess,os,concurrent.futures,hashlib
r=Path('/workspace/scratch/cast-tags');e=r/'evidence'
while not (e/'census-status.json').exists() or not (e/'c-dts-status.json').exists():time.sleep(5)
assert json.loads((e/'c-dts-status.json').read_text())['exit']==0
m=r/'mutant-tree';m.mkdir();shutil.copytree(r/'c-tree/src',m/'src');pins=json.loads((e/'c-probe-owners.json').read_text())
for pin in pins:
 p=m/'src/compiler'/pin['file'];raw=p.read_bytes();crlf=b'\r\n' in raw;text=raw.decode().replace('\r\n','\n');assert text.count(pin['after'])==1;text=text.replace(pin['after'],pin['before']);p.write_bytes(text.replace('\n','\r\n').encode() if crlf else text.encode())
files=list((r/'main-tree/src').rglob('*'));count=0
for a in files:
 if a.is_file():
  b=m/'src'/a.relative_to(r/'main-tree/src');assert a.read_bytes()==b.read_bytes(),str(a);count+=1
(e/'restoring-mutant-source.json').write_text(json.dumps({'restored_interfaces':len(pins),'byte_identical_source_files':count})+'\n')
def run(file):
 tag='mutant-'+Path(file).stem;env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(m/'src/compiler'/file),LATENT_PROGRESS_FILE=str(e/(tag+'-progress.json')))
 with (e/(tag+'.log')).open('w') as log:p=subprocess.run([str(r/'latent'),str(m/'src/compiler'),str(e/(tag+'.jsonl'))],env=env,stdout=log,stderr=subprocess.STDOUT)
 result={'case':tag,'exit':p.returncode};(e/(tag+'-status.json')).write_text(json.dumps(result)+'\n');print(result,flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:list(pool.map(run,['transformers/generators.ts']))

with (e/'mutant-isolated-casts.json').open('w') as out,(e/'mutant-isolated-casts.log').open('w') as log:
 p=subprocess.run([str(r/'cast-probe'),str(m/'src/compiler')],env=dict(os.environ,CAST_TAG_TRACE='1'),stdout=out,stderr=log)
(e/'mutant-isolated-casts-status.json').write_text(json.dumps({'exit':p.returncode})+'\n')
assert p.returncode==0
