"""Pair refusal-scan findings by the requested identity key; retain complete deltas."""
import collections,gzip,json,re
from pathlib import Path
OUT=Path(__file__).resolve().parent

def normalize(s):
 return re.sub(r'/workspace/scratch/c4v7-census/adapted-(?:head|main)/src/compiler/','',s)
def key(f):
 file,line,col=f['where'].rsplit(':',2)
 return f['kind'],file,int(line),int(col),f['reason']
def load(label):
 found={};counts=[]
 for i in range(1,80):
  dest=OUT/'part2'/label
  status=json.loads((dest/f'{i:02}-status.json').read_text());assert status['validated'] and status['exit_code']==0 and status['zero_lowering_units']
  rows=[json.loads(x) for x in gzip.decompress((dest/f'{i:02}.jsonl.gz').read_bytes()).splitlines()]
  assert not rows[-1]['findings'],rows[-1]['findings']
  assert not rows[-1]['relation_overflow']['after_load'] and not rows[-1]['relation_overflow']['end']
  record=next(x for x in rows if x.get('file'))
  for f in record['findings']:
   raw_text=f['text'];raw_where=f['where']
   f={k:normalize(v) if isinstance(v,str) else v for k,v in f.items()}
   f['raw_text']=raw_text;f['raw_where']=raw_where
   k=key(f)
   if k in found: assert found[k]['text']==f['text']
   found[k]=f
  counts.append({'file':status['file'],'findings':len(record['findings'])})
 return found,counts
before,bc=load('main');after,ac=load('head')
gone=sorted(before.keys()-after.keys());new=sorted(after.keys()-before.keys());changed=sorted(k for k in before.keys()&after.keys() if before[k]['text']!=after[k]['text'])
byreason=collections.defaultdict(lambda:dict(before=0,after=0,gone=0,new=0,changed=0))
for data,label in [(before,'before'),(after,'after')]:
 for k in data:byreason[k[-1]][label]+=1
for keys,label in [(gone,'gone'),(new,'new'),(changed,'changed')]:
 for k in keys:byreason[k[-1]][label]+=1
selected=json.loads((OUT/'selected.json').read_text());assert selected['count']==len(selected['findings'])==191
selected_result=[]
scout_mapping=json.loads((OUT/'scout-site-mapping.json').read_text())['sites']
for f,m in zip(selected['findings'],scout_mapping):
 assert f['file']==m['file'] and f['line']==m['scout_line'] and f['column']==m['column']
 k=(f['kind'],f['file'],m['main_line'],f['column'],f['reason'])
 selected_result.append({'original_scout_line':f['line'],'key':k,'present_before':k in before,'present_after':k in after,'gone':k not in after,'removed_by_head':k in before and k not in after})
result=dict(pairing=['kind','file','line','column','reason'],changed_definition='same key, changed normalized text; reason or kind changes count as gone plus new',before=len(before),after=len(after),gone=len(gone),new=len(new),changed=len(changed),by_reason=dict(sorted(byreason.items())),selected_191=dict(present_before=sum(x['present_before'] for x in selected_result),gone=sum(x['gone'] for x in selected_result),removed_by_head=sum(x['removed_by_head'] for x in selected_result),present_after=sum(x['present_after'] for x in selected_result)),per_file={'main':bc,'head':ac})
for name,obj in [('main-findings.json',list(before.values())),('head-findings.json',list(after.values())),('pairing.json',result),('new-findings.json',[after[k] for k in new]),('gone-findings.json',[before[k] for k in gone]),('changed-findings.json',[dict(before=before[k],after=after[k]) for k in changed]),('selected-191-pairing.json',selected_result)]:
 (OUT/name).write_text(json.dumps(obj,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k not in ('by_reason','per_file')},indent=2))
