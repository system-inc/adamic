from pathlib import Path
import subprocess,json,os,re,tempfile
root=Path.cwd();scratch=Path(tempfile.mkdtemp(prefix='adamic-stable-splitter-measure-'));archive=scratch/'inputs'
producer=scratch/'inputs.log'
with producer.open('w') as output:
 result=subprocess.run(['go','test','./internal/native','-run','^TestStableLintUnitChanges$','-count=1','-v','-timeout','30m'],cwd=root,env={**os.environ,'ADAMIC_STABLE_LINT_PROBE':'1','ADAMIC_STABLE_C_INPUTS':str(archive),'ADAMIC_STABLE_READ_INPUTS':'0'},stdout=output,stderr=subprocess.STDOUT)
assert result.returncode==0,producer
print('evidence:',scratch,flush=True)
baseline=scratch/'units-before.go';baseline.write_bytes(subprocess.check_output(['git','show','aca41891:internal/native/units.go'],cwd=root))
overlay=scratch/'before.json';overlay.write_text(json.dumps({'Replace':{str(root/'internal/native/units.go'):str(baseline)}}))
metadata={key:subprocess.check_output(command,shell=True,text=True,cwd=root).strip() for key,command in {'commit':'git rev-parse HEAD','nproc':'nproc','cpu.max':'cat /sys/fs/cgroup/cpu.max','go':'go version','clang':'clang --version','node':'node --version'}.items()};metadata['clang']=metadata['clang'].splitlines()[0];metadata['baseline']='aca418919f345ecd41889e73832e6a93c95cd3db';metadata['jobs']=4;metadata['cache']='baseline warm, edited objects absent; fresh per-edit object cache';(scratch/'metadata.json').write_text(json.dumps(metadata,indent=2));print(json.dumps(metadata),flush=True)
for round in range(1,4):
 for version in (['before','after'] if round%2 else ['after','before']):
  log=scratch/f'{version}-{round}.log'
  env={**os.environ,'ADAMIC_STABLE_LINT_PROBE':'1','ADAMIC_STABLE_MEASURE':'1','ADAMIC_STABLE_ONE_ROUND':'1','ADAMIC_STABLE_C_INPUTS':str(archive),'ADAMIC_STABLE_READ_INPUTS':'1','ADAMIC_STABLE_BASELINE':'1' if version=='before' else '0'}
  command=['go','test']+(['-overlay='+str(overlay)] if version=='before' else [])+['./internal/native','-run','^TestStableLintUnitChanges$','-count=1','-v','-timeout','30m']
  with log.open('w') as output: result=subprocess.run(command,cwd=root,env=env,stdout=output,stderr=subprocess.STDOUT)
  print(version,round,'exit',result.returncode,flush=True)
  if result.returncode: raise SystemExit(log.read_text())
  for line in log.read_text().splitlines():
   if 'TIMING ' in line:
    print(version,round,line.strip(),flush=True)
    with (scratch/'timings.jsonl').open('a') as f: f.write(json.dumps({'version':version,'round':round,'line':line.strip(),'command':command,'metadata':metadata})+'\n')
