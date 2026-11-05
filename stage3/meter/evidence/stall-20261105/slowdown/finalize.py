from pathlib import Path
import json,csv,gzip,shutil,subprocess,hashlib,os
r=Path('/tmp/meter-stall-20261105');s=r/'slowdown';d=r/'report/stage3/meter/evidence/stall-20261105/slowdown';d.mkdir(parents=True,exist_ok=True)
summary=json.loads((s/'comparison.json').read_text());rows=json.loads((s/'top30-named.json').read_text())
lines=['| Declaration and unit identity | Good s | Bad s | Increase s | Ratio | Good snapshot s | Bad snapshot s |','|---|---:|---:|---:|---:|---:|---:|']
for x in rows:lines.append(f"| {x['declaration']} (`{x['unit']}`) | {x['good']:.3f} | {x['bad']:.3f} | {x['increase']:.3f} | {x['ratio']:.2f}x | {x['good_snapshot']:.3f} | {x['bad_snapshot']:.3f} |")
top='\n'.join(lines)+'\n';(s/'top30.md').write_text(top)
plain_names=['prepare.py','prepare-unit.py','compare.py','names.cjs','run-pair.py','test-unit.py','repeat.py','plain.sh','status.py','unit-main.go.txt','comparison.json','predicate.json','top30.md','top30-named.json','modules.go.txt','load_time_reads.go.txt','entries_records.go.txt','interface_cast.go.txt','bisect.log','test-timings.jsonl','first-probe.log','parent-probe.log','repeat.log']
for name in plain_names:shutil.copyfile(s/name,d/name)
def zipped(src,dest):
 with src.open('rb') as f,gzip.open(dest,'wb') as g:shutil.copyfileobj(f,g)
for name in ['first-bad-lowering.patch','bisect-driver.log']:zipped(s/name,d/(name+'.gz'))
zipped(s/'common-units.csv',d/'common-units.csv.gz')
for label in ['good','bad']:
 target=d/label;target.mkdir(exist_ok=True)
 for name in ['start.json','result.json','cpu.pprof','pprof-flat.txt','pprof-cum.txt']:shutil.copyfile(s/label/name,target/name)
 for name in ['times.jsonl','latent.jsonl','census.log']:zipped(s/label/name,target/(name+'.gz'))
 target=d/'plain'/label;target.mkdir(parents=True,exist_ok=True)
 for name in ['records.jsonl','run.log','exit.txt','start.txt','end.txt','build.log']:shutil.copyfile(s/'plain'/label/name,target/name)
for sha in ['b2b32fb8914d28e5ae911ec1dd8a475933f2b10c','4a8171994b6a63a42962ddaa462aa28d8e9caa3d']:
 for label in ['confirm-1','confirm-2']:
  source=s/'unit-tests'/sha/label;target=d/'unit-tests'/sha/label;target.mkdir(parents=True,exist_ok=True)
  for name in ['result.json','times.jsonl','run.log','cpu.pprof','pprof-flat.txt','pprof-cum.txt']:shutil.copyfile(source/name,target/name)
  zipped(source/'record.jsonl',target/'record.jsonl.gz')
  if (source/'build.log').exists():shutil.copyfile(source/'build.log',target/'build.log')
for name in ['good-build.log','bad-build.log','plain-driver.log','bisect-reset.log']:shutil.copyfile(s/name,d/name)
# Artifact hashes establish the exact binaries used without committing executables.
with (d/'binaries.sha256').open('w') as f:
 for p in [s/'good-census',s/'bad-census',*[s/'unit-tests'/sha/'confirm-1/unit' for sha in ['b2b32fb8914d28e5ae911ec1dd8a475933f2b10c','4a8171994b6a63a42962ddaa462aa28d8e9caa3d']],s/'plain/good/census',s/'plain/bad/census']:
  f.write(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p)+'\n')
with (d/'environment.txt').open('w') as f:
 subprocess.run(['/workspace/adamic-tools/go/bin/go','version'],stdout=f,text=True,check=True)
 subprocess.run(['uname','-a'],stdout=f,text=True,check=True)
 f.write('os.cpu_count='+str(os.cpu_count())+'\n')
 for p in ['/sys/fs/cgroup/cpu.max','/sys/fs/cgroup/memory.max']:
  if Path(p).exists():f.write(p+'='+Path(p).read_text())
# Confirmation matrix includes measured body time, not load or registration.
with (d/'per-commit-timings.csv').open('w') as f:
 w=csv.writer(f,lineterminator='\n');w.writerow(['commit','run','body_wall_seconds','snapshot_wall_seconds','body_cpu_seconds','threshold_seconds','bad'])
 for line in (s/'test-timings.jsonl').read_text().splitlines():
  x=json.loads(line);t=x['unit_time'];w.writerow([x['commit'],x['label'],t['wall_seconds'],t['snapshot_wall_seconds'],t['cpu_seconds'],x['threshold_seconds'],x['bisect_exit']])
