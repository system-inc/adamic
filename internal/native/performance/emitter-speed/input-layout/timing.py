import json,pathlib,subprocess,os
p=pathlib.Path('scratch/emitter-speed/input-layout').resolve();helper=pathlib.Path('scratch/emitter-speed/sieve/wait4').resolve();manifest=pathlib.Path('scratch/emitter-speed/compiler.txt').resolve()
load=lambda:pathlib.Path('/proc/loadavg').read_text().strip()
report=dict(rounds=10,affinity=3,allowed_cpus=sorted(os.sched_getaffinity(0)),cpu_quota=pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(),load_before=load(),samples=[])
for round in range(10):
 for name in (['before','after'] if round%2==0 else ['after','before']):
  prefix=p/('timing-'+str(round)+'-'+name)
  command=['taskset','-c','3',str(p/name),'--manifest',str(manifest),'--count']
  subprocess.run([str(helper),str(prefix)+'.json',str(prefix)+'.stdout',str(prefix)+'.stderr','300000',*command],check=True,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
  result=json.loads(pathlib.Path(str(prefix)+'.json').read_text());result.update(variant=name,round=round,command=command)
  if result['exitCode']!=0 or result['timedOut'] or pathlib.Path(str(prefix)+'.stdout').read_bytes()!=b'0\n' or pathlib.Path(str(prefix)+'.stderr').read_bytes():raise RuntimeError('timing output differs')
  report['samples'].append(result)
report['load_after']=load();report['best']={name:{'user_seconds':min(r['userMs']/1000 for r in report['samples'] if r['variant']==name),'wall_seconds':min(r['wallMs']/1000 for r in report['samples'] if r['variant']==name)}for name in ['before','after']}
(p/'timings.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report['best']))
