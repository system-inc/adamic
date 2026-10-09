from pathlib import Path
import os,subprocess,time,json,re,shutil
root=Path(__file__).resolve().parents[3];catalog=root/'verify/catalog';env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',PYTHONDONTWRITEBYTECODE='1');main=subprocess.check_output(['git','rev-parse','origin/main'],cwd=root,text=True).strip()
metadata={'commit_sha':main,'nproc':subprocess.check_output(['nproc'],text=True).strip(),'cpu_max':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'go_version':subprocess.check_output(['go','version'],text=True).strip(),'clang_version':subprocess.check_output(['clang','--version'],text=True).splitlines()[0],'node_version':subprocess.check_output(['node','--version'],text=True).strip(),'go_flags':subprocess.check_output(['go','env','GOFLAGS'],text=True).strip(),'oracle_cache':'ADAMIC_GATE_UNCACHED=1','before_flags':'go test -count=1; first entry control also -x for build profiling','after_flags':'go test -trimpath -count=1; jobs default nproc; one worktree per active entry'}
(catalog/'environment-parallel.json').write_text(json.dumps(metadata,indent=2)+'\n')
measurements=[]
def run(kind,command,iteration):
 log=catalog/'logs'/f'parallel-{kind}-{iteration}.log';load_before=Path('/proc/loadavg').read_text().strip();started=time.monotonic()
 with log.open('w') as f:p=subprocess.run(command,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 elapsed=time.monotonic()-started;text=log.read_text();assert p.returncode==0,text[-6000:]
 assert re.findall(r'^(\d\d) [a-z].*: (?:applies-and-fails-as-recorded|skipped)',text,re.MULTILINE)==[f'{n:02}' for n in range(1,17)]
 logs=Path(next(line[5:] for line in text.splitlines() if line.startswith('logs ')))
 record={'instrument':' '.join(command),'exit_code':p.returncode,'elapsed_seconds':elapsed,'load_before':load_before,'load_after':Path('/proc/loadavg').read_text().strip(),'log':'logs/'+log.name,'raw_logs':str(logs)}
 if kind=='before':
  phase=json.loads((logs/'01-control.timing.json').read_text());control=(logs/'01-control.log').read_text();duration=re.search(r'^ok\s+\S+\s+([0-9.]+)s',control,re.MULTILINE)
  record['entry01_control']=phase;record['entry01_control']['go_reported_test_seconds']=float(duration[1]) if duration else None
  record['entry01_compile_commands']=len(re.findall(r'/pkg/tool/[^ /]+/compile ',control))
 else:
  manifest=json.loads((logs/'results.json').read_text());assert manifest['jobs']==int(metadata['nproc'])
  active=[e for e in manifest['entries'] if 'worktree' in e];assert len(active)==11 and len({e['worktree'] for e in active})==11
  assert all(e['ok'] for e in manifest['entries']);record['manifest']=manifest
 # Keep timing inventories and actual controls/mutants for each loop without reprinting build traces.
 destination=catalog/'logs'/f'parallel-{kind}-{iteration}-evidence';destination.mkdir(exist_ok=True)
 for file in logs.rglob('*'):
  if file.is_file():
   output=destination/file.relative_to(logs);output.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(file,output)
 return record
for iteration in range(1,4):
 print(f'Loop {iteration}: before sequential',flush=True)
 before=run('before',['python3',str(catalog/'history'/'check-sequential.py'),main],iteration)
 print(f'  before {before["elapsed_seconds"]:.3f}s; entry01 control {before["entry01_control"]["seconds"]:.3f}s',flush=True)
 print(f'Loop {iteration}: after parallel',flush=True)
 after=run('after',['bash','verify/catalog/check.sh',main],iteration)
 print(f'  after {after["elapsed_seconds"]:.3f}s',flush=True)
 measurements.append({'loop':iteration,'before':before,'after':after});(catalog/'parallel-measurements.json').write_text(json.dumps(measurements,indent=2)+'\n')
 if max(before['elapsed_seconds'],after['elapsed_seconds'])>300:
  print('A run exceeded five minutes; using the requested single-run exception',flush=True);break
print('Interleaved checker measurements complete',flush=True)
