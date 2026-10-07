from pathlib import Path
import subprocess,time,os,json,hashlib,resource
root=Path('/workspace/lint-cost-baseline');env=dict(os.environ)
commands={'Go':[str(root/'oracle')],'today':[str(root/'scanner')],'kind_table':['/workspace/lint-cost-prototype/scanner'],'all_listeners_table':['/workspace/lint-cost-all-listeners/scanner']}
samples={k:[] for k in commands};runs=[];before=Path('/proc/loadavg').read_text().strip()
for trial in range(5):
 keys=list(commands);order=keys[trial%4:]+keys[:trial%4]
 for name in order:
  argv=commands[name]+['--manifest',str(root/'compiler.txt'),'--count'];stem=root/f'bench-{trial}-{name}'
  with Path(str(stem)+'.stdout').open('wb') as out,Path(str(stem)+'.stderr').open('wb') as err:
   prior=resource.getrusage(resource.RUSAGE_CHILDREN);start=time.perf_counter();r=subprocess.run(argv,stdout=out,stderr=err);elapsed=time.perf_counter()-start;after=resource.getrusage(resource.RUSAGE_CHILDREN)
  assert r.returncode==0 and Path(str(stem)+'.stdout').read_bytes()==b'161\n' and not Path(str(stem)+'.stderr').read_bytes(),(trial,name,r.returncode)
  samples[name].append(elapsed);runs.append({'round':trial,'name':name,'argv':argv,'wall':elapsed,'user':after.ru_utime-prior.ru_utime,'system':after.ru_stime-prior.ru_stime});print(trial,name,elapsed,flush=True)
result={'best':{k:min(v) for k,v in samples.items()},'samples':samples,'runs':runs,'load_before':before,'load_after':Path('/proc/loadavg').read_text().strip(),'nproc':subprocess.check_output(['nproc']).decode().strip(),'cpu_max':Path('/sys/fs/cgroup/cpu.max').read_text().strip()};Path('/workspace/lint-cost-timing.json').write_text(json.dumps(result,indent=2)+'\n');print(result['best'],flush=True)
