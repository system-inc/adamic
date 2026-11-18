"""Compare the original serialized top-level findings arrays, without re-encoding them."""
from pathlib import Path
import json,hashlib,sys,gzip
W=Path('/workspace');runs=W/'depthgate2-profile-runs';out=W/'depthgate2-profile-analysis';out.mkdir(exist_ok=True)
dec=json.JSONDecoder()
def arrays(p):
 result=[];count=0
 data=p.read_bytes();data=gzip.decompress(data) if p.suffix=='.gz' else data
 for raw in data.splitlines():
  line=raw.decode('utf-8');i=0
  assert line[i]=='{';i+=1
  while i<len(line):
   while line[i].isspace() or line[i]==',':i+=1
   if line[i]=='}':break
   key,end=dec.raw_decode(line,i);i=end
   while line[i].isspace():i+=1
   assert line[i]==':';i+=1
   while line[i].isspace():i+=1
   start=i;value,i=dec.raw_decode(line,i)
   if key=='findings':
    assert isinstance(value,list)
    result.append(line[start:i].encode('utf-8')+b'\n');count+=len(value)
 return b''.join(result),count
report={}
for side in ['base','head']:
 p=runs/f'{side}-utilities-r1';status=p/'status.json'
 if not status.exists():continue
 s=json.loads(status.read_text());rows=[json.loads(l) for l in (p/'latent.jsonl').read_bytes().splitlines()]
 assert s['exit_code']==0 and not s['timed_out'] and rows[-1]['status']=='relation_report'
 raw,count=arrays(p/'latent.jsonl');(out/f'{side}-findings-arrays.jsonl').write_bytes(raw)
 r={'findings':count,'serialized_array_bytes':len(raw),'sha256':hashlib.sha256(raw).hexdigest(),'comparisons':[]}
 for repeat in [1,2]:
  prior=W/'depthgate2-base/stage3/typescript/depthgate-census-2/evidence/runs'/f'{side}-utilities-r{repeat}'/'latent.jsonl.gz';old,oldcount=arrays(prior)
  r['comparisons'].append({'census2_repeat':repeat,'committed_reference':str(prior),'reference_archive_sha256':hashlib.sha256(prior.read_bytes()).hexdigest(),'findings':oldcount,'sha256':hashlib.sha256(old).hexdigest(),'byte_identical':raw==old})
  assert raw==old,f'{side} does not match census-2 repeat {repeat} byte for byte'
 report[side]=r
(out/'behavior-comparison.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps(report,indent=2))
