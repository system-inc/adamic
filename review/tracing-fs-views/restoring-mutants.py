import pathlib,shutil,subprocess,os,json,time
out=pathlib.Path(__file__).resolve().parent
for number,old,new,line in [
(62,'fs.writeSync(typesFd, JSON.stringify(descriptor)!)','(fs as { writeSync(fd: number, data: string | undefined): number }).writeSync(typesFd, JSON.stringify(descriptor))',340),
(63,'fs.writeFileSync(legendPath, JSON.stringify(legend)!)','(fs as { writeFileSync(path: string, data: string | undefined): void }).writeFileSync(legendPath, JSON.stringify(legend))',359)]:
 tree=pathlib.Path('/workspace/tracing-mutant'+str(number))
 shutil.copytree('/workspace/tracing-measured-after/src',tree/'src')
 file=tree/'src/compiler/tracing.ts';before=file.read_bytes();assert before.count(old.encode())==1;file.write_bytes(before.replace(old.encode(),new.encode()))
 env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str(file),LATENT_ASSERT_NO_OUTPUT='1')
 name='mutant'+str(number)
 with (out/(name+'.log')).open('w') as log:
  r=subprocess.run(['/workspace/tracing-census-build/census',str(tree/'src/compiler'),str(out/(name+'.jsonl'))],stdout=log,stderr=subprocess.STDOUT,env=env)
 (out/(name+'.exit.json')).write_text(json.dumps({'exit':r.returncode})+'\n');assert r.returncode==0
 records=[json.loads(s) for s in (out/(name+'.jsonl')).read_text().splitlines()]
 assert not records[0]['diagnostics']
 findings=[f for rec in records for f in rec.get('findings',[]) if f.get('kind')=='Refused' and f':{line}:' in f.get('where','') and 'tsc relates a method' in f['reason']]
 assert findings,(number,'missing restored refusal')
 print(name,'restored',[(f['phase'],f['where'],f['reason']) for f in findings],flush=True)
