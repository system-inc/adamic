from pathlib import Path
import collections,gzip,hashlib,json,subprocess
repo=Path('/workspace/adamic');base=Path('/workspace/scratch/branch-shape');out=repo/'cloud/reports/branch-shape/evidence';out.mkdir(parents=True,exist_ok=True)
variants=['before','switch','scalar-only','scalar','landing-before','landing']
for name in variants:
 for suffix in ['.callgrind','.json','.c']:
  p=base/(name+suffix)
  if p.exists() and p.stat().st_size:
   with gzip.open(out/(p.name+'.gz'),'wb') as f:f.write(p.read_bytes())
 for suffix in ['.stdout','.stderr','-summary.log','-build.log','-emit.log']:
  p=base/(name+suffix)
  if p.exists():(out/p.name).write_bytes(p.read_bytes())
for pattern in ['*mutant*.log','*gate.log','*native-oracle.log','*vet.log','*gofmt.log','focused.log','timing*.json','timing*.log','*.asm','prototype-*','*overlay.json','corpus.sha256','*compiler.log','*final-build.log','*landing-build.log']:
 for p in base.glob(pattern):
  (out/(p.name + ".txt" if p.suffix == ".go" else p.name)).write_bytes(p.read_bytes())
for name in ['profile.py','timing.py','final-mutants.py','mutants.py','collect_evidence.py']:
 (out/name).write_bytes((base/name).read_bytes())
for name in ['batch8-go.stdout','batch8-landing.stdout']:
 with gzip.open(out/(name+'.gz'),'wb') as f:f.write((base/name).read_bytes())
for name in ['batch8-go.stderr','batch8-landing.stderr']:(out/name).write_bytes((base/name).read_bytes())
artifacts={}
for pattern in ['before','switch','scalar','scalar-only','landing','landing-before','batch8-*release','adamic-*','*.c','valgrind.deb']:
 for p in base.glob(pattern):
  if p.is_file():artifacts[p.name]={'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()}
(out/'artifact-hashes.json').write_text(json.dumps(artifacts,indent=2)+'\n')
profiles={}
for v in variants:
 p=base/(v+'.json')
 if p.exists():
  d=json.loads(p.read_text());profiles[v]=dict(zip(d['events'],d['totals']))
(out/'events.json').write_text(json.dumps(profiles,indent=2)+'\n')
p=json.loads((base/'landing-before.json').read_text());c=(base/'landing-before.c').read_text().splitlines();own=collections.defaultdict(lambda:[0]*13)
for r in p['functions']:
 if r['function'].startswith('adamic_function'):
  name=r['function'].split("'")[0];own[name]=[a+b for a,b in zip(own[name],r['costs'])]
rows=[]
for name,costs in sorted(own.items(),key=lambda r:r[1][10]+r[1][12],reverse=True)[:10]:
 hot=sorted([r for r in p['instructions'] if r['function'].split("'")[0]==name],key=lambda r:r['costs'][10]+r['costs'][12],reverse=True)[:5]
 text=['Function '+name,'Aggregated self events '+str(dict(zip(p['events'],costs)))]
 for r in hot:
  text.append(str(r))
  if r['source'].endswith('before.c'):
   i=r['line'];text.extend(f'{j+1}: {c[j]}' for j in range(max(0,i-4),min(len(c),i+3)))
 text.append(subprocess.check_output(['objdump','-d','--no-show-raw-insn','--disassemble='+name,str(base/'landing-before')],text=True))
 (out/(name+'.txt')).write_text('\n'.join(text));rows.append({'function':name,'events':dict(zip(p['events'],costs)),'hot_branches':hot})
(out/'top10.json').write_text(json.dumps(rows,indent=2)+'\n')
for pin in ['39638d9','c7991b9']:
 files=subprocess.check_output(['git','ls-tree','-r','--name-only',pin,'internal/native/runtime'],cwd=repo,text=True).splitlines()
 hashes={f:hashlib.sha256(subprocess.check_output(['git','show',pin+':'+f],cwd=repo)).hexdigest() for f in files}
 (out/(pin+'-runtime-hashes.json')).write_text(json.dumps(hashes,indent=2)+'\n')
print('Evidence collected in',out)
