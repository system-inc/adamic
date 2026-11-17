import os,json,pathlib,subprocess,sys,time,datetime
s=pathlib.Path('/workspace/scratch/overload-certificates-census');label=sys.argv[1];directory=s/'results'/label/'full-directory';directory.mkdir(exist_ok=True)
output=directory/'core.jsonl';env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str(s/'adapted/src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TAIL='1',CENSUS_MODULE_FORMAT='CommonJSBundle',GOMAXPROCS='1',LATENT_TIMES=str(directory/'core.times.jsonl'))
args=[str(s/(label+'-tools')/'latent-census'),str(s/'adapted/src/compiler'),str(output)];start=time.monotonic()
with (directory/'core.log').open('wb') as log:
 p=subprocess.Popen(args,env=env,stdout=log,stderr=subprocess.STDOUT,cwd=s/label);_,status,usage=os.wait4(p.pid,0);p.returncode=os.waitstatus_to_exitcode(status)
row={'label':label,'mode':'full-directory','file':'src/compiler/core.ts','root':'src/compiler','args':args,'state':'complete' if p.returncode==0 else 'failed','exit':p.returncode,'seconds':time.monotonic()-start,'maxRssKiB':usage.ru_maxrss,'output':str(output.relative_to(s))}
if output.exists():
 records=[json.loads(line) for line in output.read_text().splitlines() if line.strip()];row['streamComplete']=bool(records and records[-1].get('status')=='relation_report');row['findings']=sum(len(r.get('findings') or []) for r in records)
(directory/'receipt.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row));sys.exit(p.returncode)
