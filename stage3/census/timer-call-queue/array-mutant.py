from pathlib import Path
import subprocess,json,os,hashlib
p=Path(__file__).parent;repo=Path('/tmp/adamic-timer-queue-work');tree=Path('/tmp/adamic-timer-queue-baseline');file=tree/'src/compiler/sys.ts'
oldpins=json.loads(subprocess.check_output(['git','-C',str(repo),'show','61603dc6761a2f1c8312a636ecb74be4a256ed86:stage3/adapt/99-correctness-timer-closures/owners.json'],text=True))
newpins=json.loads((repo/'stage3/adapt/99-correctness-timer-closures/owners.json').read_text())
old=next(x['after']for x in oldpins if x['owner']=='createDynamicPriorityPollingWatchFile');new=next(x['after']for x in newpins if x['owner']=='createDynamicPriorityPollingWatchFile')
saved=file.read_bytes();text=saved.decode();assert text.count(new)==1;mutant=text.replace(new,old);(p/'sys.array-mutant.ts').write_text(mutant)
try:
 file.write_text(mutant)
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(file))
 with (p/'census.array-mutant.log').open('w')as log:r=subprocess.run(['/workspace/timer-call-queue-scratch/census',str(tree/'src/compiler'),str(p/'census.array-mutant.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT)
 assert r.returncode==0,r.returncode
 rows=[json.loads(x)for x in (p/'census.array-mutant.jsonl').read_text().splitlines()];assert rows[-1]['status']=='relation_report'
 found=[f for row in rows for f in row.get('findings',[])if f.get('kind')=='Refused'and f.get('where','').endswith('sys.ts:272:23')and f.get('reason')=="a cast the runtime can't check"]
 assert found,'restored array cast was not refused'
 result={'census_exit':r.returncode,'refused_again':found,'source_restored':True,'restored_source_sha256':hashlib.sha256(saved).hexdigest()}
finally:file.write_bytes(saved)
assert file.read_bytes()==saved
(p/'array-mutant-result.json').write_text(json.dumps(result,indent=2)+'\n');print('restoring array mutant: Refused sys.ts:272:23; composition restored')
