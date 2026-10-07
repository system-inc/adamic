from pathlib import Path
import collections,gzip,hashlib,json,shutil,subprocess
root=Path('/workspace/adamic')
logs=Path('/workspace/wave-25-validation')
scratch=Path('/workspace/wave-25-landing')
out=root/'stage1/cohere/typeaware/validation-wave-25-landing'
out.mkdir(exist_ok=True)
results={'base':subprocess.check_output(['git','rev-parse','origin/main'],cwd=root,text=True).strip(),'tested_head':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),'populations':{},'artifacts':{}}
inputs={}
for batch in ['original','next','third','fourth','fifth','refusals','dependency']:
 directory=scratch/batch
 target=out/batch
 target.mkdir(exist_ok=True)
 for file in directory.iterdir():
  if not file.is_file() or file.is_symlink() or file.suffix not in ['.stdout','.stderr','.manifest','.json']:
   continue
  data=file.read_bytes()
  if file.suffix in ['.stdout','.stderr']:
   with gzip.GzipFile(filename=str(target/(file.name+'.gz')),mode='wb',mtime=0) as stream:stream.write(data)
  else:shutil.copyfile(file,target/file.name)
  results['artifacts'][batch+'/'+file.name]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()}
  if file.name.endswith('-go.stdout') and file.name.split('-',1)[1] in ['controls-go.stdout','controls-no-options-go.stdout','repository-go.stdout','compiler-go.stdout']:
   rows=[line.split('\t') for line in data.decode().splitlines() if line and line[0].isdigit()]
   results['populations'][batch+'/'+file.name]={'findings':len(rows),'rule_counts':dict(collections.Counter(row[2] for row in rows)),'fixes':sum(int(row[5]) for row in rows),'suggestions':sum(int(row[6]) for row in rows),'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()}
 manifest=directory/'controls.manifest'
 if manifest.exists():
  valid=directory/'controls-valid.manifest'
  accepted=set((valid if valid.exists() else manifest).read_text().splitlines())
  inputs[batch]=[{'path':name,'parse_valid':name in accepted,'sha256':hashlib.sha256(Path(name).read_bytes()).hexdigest(),'source':Path(name).read_text()} for name in manifest.read_text().splitlines()]
for name in ['landing-fetch.log','landing-rebase.log','landing-range-diff.log','landing-setup.log','landing-rules.log','landing-bridge.log','landing-node.log','landing-vet.log','landing-dependency.log','landing-benchmark.log','landing-dynamic-regexp.log','landing-constant-regexp.log','landing-constant-regexp.stdout','landing-constant-regexp.stderr']:
 with gzip.GzipFile(filename=str(out/(name+'.gz')),mode='wb',mtime=0) as stream:stream.write((logs/name).read_bytes())
for name in ['landing-gate.sh','landing-dependency-gate.sh','landing-benchmark.py','landing-package.py','landing-bench.json']:
 shutil.copyfile(logs/name,out/name)
with gzip.GzipFile(filename=str(out/'inputs.json.gz'),mode='wb',mtime=0) as stream:stream.write(json.dumps(inputs,ensure_ascii=False,indent=2).encode())
(out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results['populations'],indent=2))

shutil.copyfile(logs/'landing-dynamic-regexp.a',out/'dynamic-regexp-probe.a')
shutil.copyfile(logs/'fifth-regexp-probe.a',out/'constant-regexp-probe.a')
