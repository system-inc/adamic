from pathlib import Path
import collections,gzip,hashlib,json,shutil
root=Path('/workspace/adamic');scratch=Path('/workspace/wave-25-fifth-final');logs=Path('/workspace/wave-25-validation');out=root/'stage1/cohere/typeaware/validation-wave-25-fifth';out.mkdir(exist_ok=True)
for p in scratch.iterdir():
 if p.is_file() and (p.suffix in ('.stdout','.stderr','.manifest') or p.name=='controls-config.json'):
  if p.suffix=='.stdout' and p.stat().st_size>0:
   with gzip.GzipFile(filename=str(out/(p.name+'.gz')),mode='wb',mtime=0) as target:target.write(p.read_bytes())
  else:shutil.copyfile(p,out/p.name)
for name in ['fifth-fetch-all.log','fifth-selection.log','fifth-claim-push.log','fifth-gate-final.log','fifth-bridge-gate.log','fifth-annotation-mutant.log','fifth-refusals-final.log','fifth-regexp-refusal.log','fifth-vet-final.log','fifth-benchmark.log']:
 p=logs/name
 with gzip.GzipFile(filename=str(out/(name+'.gz')),mode='wb',mtime=0) as target:target.write(p.read_bytes())
for origin,name in [(Path('/workspace/wave-25-fifth-selection.json'),'selection.json'),(logs/'fifth-bench.json','benchmark.json')]:shutil.copyfile(origin,out/name)
shutil.copyfile(logs/'fifth-benchmark.py',out/'benchmark.py');shutil.copyfile(logs/'fifth-select.py',out/'select.py')
valid=set((scratch/'controls-valid.manifest').read_text().splitlines());inputs=[{'path':p,'sha256':hashlib.sha256(Path(p).read_bytes()).hexdigest(),'parse_valid':p in valid,'source':Path(p).read_text()} for p in (scratch/'controls.manifest').read_text().splitlines()]
(out/'inputs.json').write_text(json.dumps(inputs,ensure_ascii=False,indent=2)+'\n')
findings=[line.split('\t') for line in (scratch/'008-controls-go.stdout').read_text().splitlines() if not line.startswith(('file\t','findings '))]
summary={'controls':{'total':len(inputs),'valid':len(valid),'findings':len(findings),'rule_counts':dict(collections.Counter(r[2] for r in findings)),'message_counts':dict(collections.Counter(r[3] for r in findings)),'fix_count':sum(int(r[5]) for r in findings),'findings_with_fixes':sum(int(r[5])>0 for r in findings),'suggestions':sum(int(r[6]) for r in findings)},'files':{}}
for pop,prefix in [('controls','008-controls'),('repository','020-repository'),('compiler','026-compiler')]:
 data=(scratch/(prefix+'-go.stdout')).read_bytes();summary['files'][pop]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()}
(out/'results.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))

refusals=Path('/workspace/wave-25-fifth-refusals-final')
for p in refusals.iterdir():
 if p.suffix in ('.stderr','.stdout'):
  with gzip.GzipFile(filename=str(out/('refusal-'+p.name+'.gz')),mode='wb',mtime=0) as target:target.write(p.read_bytes())
shutil.copyfile(logs/'fifth-regexp-probe.a',out/'regexp-probe.a')
