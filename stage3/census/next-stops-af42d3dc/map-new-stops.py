import json
from pathlib import Path
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent
cores=json.loads((repo/'stage3/census/core-sites-af42d3dc/rows.json').read_text())['sites'];em=json.loads((repo/'stage3/census/core-sites-af42d3dc/emitters.json').read_text());groups=json.loads((out/'grouped-first-stops.json').read_text());result={}
for g in groups:
 if not g['new_group'] or g['where']=='-':continue
 match=next((f for f in cores if 'core.ts:'+f['where'].rsplit('/core.ts:',1)[1]==g['where'] and f['kind']==g['kind'] and f['reason']==g['reason']),None)
 if match:result[g['where']]=em[match['id']]
 else:raise RuntimeError('Unmapped new ordinary group: '+str(g))
 for e in result[g['where']]:
  if e['function']=='(*lowering).methodValueRefusal':e['note']='The deciding function starts at line 260. Its guard at line 261 selects a property access read outside a direct call, without an allowed method-read proof. The method-symbol check at line 269 selects this refusal, emitted at line 282. The enclosing source comments that Intl.Collator.prototype.compare is bound; this report records the current compiler decision without changing it.'
  file=repo/e['file'];(out/(file.name+'.txt')).write_bytes(file.read_bytes())
(out/'new-stop-emitters.json').write_text(json.dumps(result,indent=2)+'\n');print('mapped new groups',len(result))
