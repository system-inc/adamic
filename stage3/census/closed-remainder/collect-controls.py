#!/usr/bin/env python3
"""Freeze successful ordinary scans and direct controls; label unknown optional exits explicitly."""
import gzip,hashlib,json,pathlib,shutil,sys
scratch=pathlib.Path(sys.argv[1]);out=pathlib.Path(sys.argv[2]);out.mkdir(parents=True,exist_ok=True)
def freeze(src,dst):
 s=json.loads((src/'status.json').read_text());assert s['child_exit_code']==0 and not s['deadline_expired'],src
 dst.mkdir(parents=True,exist_ok=True);shutil.copyfile(src/'status.json',dst/'status.json')
 for name in ['census.jsonl','units.jsonl','passes.jsonl','stderr.log']:
  p=src/name
  if p.exists():(dst/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for side in ['base','head']:
 for start in ['0','1500','3000']:freeze(scratch/'scans'/side/start,out/'ordinary-scans'/side/start)
 freeze(scratch/'combined'/side,out/'controls'/'combined'/side)
 d=out/'ordinary-scans'/'overlays'/side;d.mkdir(parents=True,exist_ok=True)
 for name in ['hook.go','passes.go']:shutil.copyfile(scratch/(side+'-scan')/name,d/(name+'.txt'))
 p=scratch/(side+'-scan-build.log');(out/(side+'-scan-build.log.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for name in ['ordered-summary.json','ordered-units.json.gz','ordered-changed-first-stops.json','ordered-regressions.json','ordered-new-panics.json']:shutil.copyfile(scratch/'ordered-analysis'/name,out/name)
for part in ['0','1500']:
 src=scratch/'scans/known'/part
 if (src/'status.json').exists():freeze(src,out/'controls'/'historical-scans'/part)
# Optional follow-ups whose supervisor status was lost are evidence, not accepted runs.
unknown=out/'controls/unobserved-followups.json'
if unknown.exists():
 for r in json.loads(unknown.read_text()):
  src=scratch/'scans'/r['side']/r['range'];dst=out/'controls'/'unobserved'/r['side']/r['range'];dst.mkdir(parents=True,exist_ok=True)
  for name in ['census.jsonl','stderr.log']:
   p=src/name
   if p.exists():(dst/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
p=out/'provenance.json';r=json.loads(p.read_text())
for name in ['base-scan','head-scan','known','known-scan']:
 f=scratch/(name+'-census');r['binaries'][name]={'sha256':hashlib.sha256(f.read_bytes()).hexdigest(),'bytes':f.stat().st_size}
r['whole_entry_scan_priority_counts']={'lowers':3,'kind_through_Node':0,'elsewhere':22};r['primary_first_stop_order']=['lowering','deferred (pair definition)'];r['primary_counts']={'lowers':3,'kind_through_Node':0,'elsewhere':34};r['preflight_trace']={'base':{'events':347,'shared_result':0},'head':{'events':347,'shared_result':0}}
p.write_text(json.dumps(r,indent=2)+'\n')
for name,source in [('known-scan-hook.go.txt',scratch/'known-scan-hook.go'),('known-scan-observer.go.txt',scratch/'known-original/observer.go')]:
 dst=out/'controls'/'historical-scans'/name;dst.write_bytes(source.read_bytes())
print('ordinary scan, direct combined, binary provenance, and excluded-control evidence frozen')
