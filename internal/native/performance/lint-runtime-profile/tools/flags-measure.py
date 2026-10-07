import subprocess,time,json,os,hashlib,platform
from pathlib import Path
roots=Path('/workspace');env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1',UBSAN_OPTIONS='print_stacktrace=1')
for group,findings in [('8',161),('4',15119),('original8',161)]:
 cases=[group] if group=='original8' else ['before'+group,'after'+group]
 if group=='original8':cases=['original8']
 first=roots/('lint-flags-'+cases[0]);oracle=first/'oracle';commands={'Go':[str(oracle)]}
 for case in cases:
  d=roots/('lint-flags-'+case)
  assert (d/'oracle').read_bytes()==oracle.read_bytes(),case
  assert (d/'compiler.txt').read_bytes()==(first/'compiler.txt').read_bytes(),case
  commands[case+'_release_O2']=[str(d/'scanner')];commands[case+'_asan_ubsan_O1']=[str(d/'sanitized')]
 keys=list(commands);samples={key:[] for key in keys};orders=[];runs=[];before=Path('/proc/loadavg').read_text().strip()
 for trial in range(5):
  order=keys[trial%len(keys):]+keys[:trial%len(keys)];orders.append(order)
  for label in order:
   stem=f'flags-group-{group}-{trial}-{label}';argv=commands[label]+['--manifest',str(first/'compiler.txt'),'--count']
   with (first/(stem+'.stdout')).open('wb') as out,(first/(stem+'.stderr')).open('wb') as err:
    start=time.perf_counter();r=subprocess.run(argv,env=env,stdout=out,stderr=err);elapsed=time.perf_counter()-start
   answer=(first/(stem+'.stdout')).read_bytes();stderr=(first/(stem+'.stderr')).read_bytes()
   assert r.returncode==0,(group,label,r.returncode,stderr[:1000]);assert answer==f'{findings}\n'.encode(),(group,label,answer);assert not stderr,(group,label,stderr[:1000])
   samples[label].append(elapsed);runs.append({'round':trial,'label':label,'argv':argv,'seconds':elapsed,'stdout_sha256':hashlib.sha256(answer).hexdigest(),'stderr_bytes':len(stderr),'exit':r.returncode})
   print(group,trial,label,elapsed,flush=True)
 result={'group':group,'cases':cases,'findings':findings,'samples':samples,'best':{k:min(v) for k,v in samples.items()},'orders':orders,'load_before':before,'load_after':Path('/proc/loadavg').read_text().strip(),'environment':{k:env[k] for k in ['ASAN_OPTIONS','UBSAN_OPTIONS']},'nproc':subprocess.check_output(['nproc']).decode().strip(),'cpu_max':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'machine':platform.uname()._asdict(),'runs':runs}
 (roots/('lint-flags-group-'+group+'.json')).write_text(json.dumps(result,indent=2)+'\n');print('best',group,result['best'],'load',before,result['load_after'],flush=True)
