from pathlib import Path
import json,re,collections
out=Path('/tmp/void-split-evidence');rows=[json.loads(x) for x in (out/'admission.jsonl').read_text().splitlines()]
def normalize(s):
 s=s.replace('/tmp/adamic-void-split-base/','<repository>/').replace('/workspace/adamic/','<repository>/')
 s=re.sub(r'0x[0-9a-fA-F]+','<address>',s)
 s=re.sub(r'goroutine \d+','goroutine <id>',s)
 return s
panics=[];other=[]
for r in rows:
 if r['base']['exit']==2 or r['head']['exit']==2:
  panics.append({**r,'normalized_equal':normalize(r['base']['stderr'])==normalize(r['head']['stderr'])})
 if r['base']['exit']!=r['head']['exit'] and 'adamic/no-void-widening' not in r['head']['stderr']:other.append(r)
result={'panics':panics,'other_exit_movement':other,'current_exit_pairs':dict(collections.Counter(str((r['base']['exit'],r['head']['exit'])) for r in rows if r['current'])),'original_exit_pairs':dict(collections.Counter(str((r['base']['exit'],r['head']['exit'])) for r in rows if r['original']))}
(out/'admission-other-movement.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({'panics':len(panics),'normalized_equal':sum(r['normalized_equal'] for r in panics),'other_exit_movement':len(other)}))
