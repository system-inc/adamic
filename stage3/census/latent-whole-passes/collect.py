#!/usr/bin/env python3
"""Copy completed measurement evidence; preserve raw bytes in deterministic gzip containers."""
import gzip,hashlib,json,pathlib,shutil,sys
scratch=pathlib.Path(sys.argv[1]);root=pathlib.Path(__file__).resolve().parent;out=root/'evidence'
def copy_run(source,dest,require_complete=True):
 if not (source/'status.json').exists():return
 dest.mkdir(parents=True,exist_ok=True)
 status=json.loads((source/'status.json').read_text())
 if require_complete and status['child_exit_code']!=0:raise ValueError(f'incomplete child: {source}')
 shutil.copy2(source/'status.json',dest/'status.json')
 for name in ['census.jsonl','passes.jsonl','units.jsonl','stderr.log']:
  p=source/name
  if p.exists():(dest/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for mode,base in [('prototype','tip-complete'),('stock','tip-complete-stock')]:
 for d in sorted((scratch/base).iterdir()):
  if d.is_dir():copy_run(d,out/'main'/mode/d.name.rsplit('-',1)[-1])
for rev,base in [('parent','parent-pin'),('87028b08','87028b08-pin')]:
 for mode in ['stock','prototype']:copy_run(scratch/base/mode,out/'pin'/rev/mode)
for mode,base,name in [('prototype','tip-runs','prototype-0000'),('stock','tip-stock-runs','stock-0000')]:
 copy_run(scratch/base/name,out/'scan-enabled-pilots'/mode,False)
inputs={}
for label,base in [('tip','tip-complete/prototype-0000'),('parent','parent-dry')]:
 header=json.loads((scratch/base/'census.jsonl').read_text().splitlines()[0])
 entries=[]
 for name in header['sources']:
  p=pathlib.Path(name);entries.append({'path':name,'bytes':p.stat().st_size,'sha256':hashlib.file_digest(p.open('rb'),'sha256').hexdigest()})
 inputs[label]={'root':header['roots'],'checker_rejected':header['checker_rejected'],'files':entries}
(out/'inputs.json').write_text(json.dumps(inputs,indent=2)+'\n')
p=out/'provenance.json';provenance=json.loads(p.read_text())
provenance['binaries']={str(p):{'sha256':hashlib.file_digest(p.open('rb'),'sha256').hexdigest(),'bytes':p.stat().st_size} for p in sorted(scratch.glob('*census*')) if p.is_file()}
p.write_text(json.dumps(provenance,indent=2)+'\n')
