"""Attach owner and phase only for measured reasons verified against source."""
import json,os
from pathlib import Path
base=Path(__file__).resolve().parent;out=base/os.environ['CENSUS_MODULE_FORMAT'];repo=base.parents[2];tree=Path('/workspace/scratch/next-stops-516c3ded-adapted')
groups=json.loads((out/'grouped-first-stops.json').read_text());owners=json.loads((base/'owners.json').read_text());evidence=[]
for g in groups:
 reason=g['reason'];where=g['where'];source=[]
 if g['kind']=='accepted':phase='accepted';emitter='lower.LowerWithOptions (accepted; no stop)'
 elif 'before runtime initialization' in reason or 'before its runtime initialization' in reason:
  phase='lowering (initialization preflight)';emitter='internal/lower/namespaces.go: (*lowering).namespaceInitialization';source=['namespaces.go','lower.go','modules.go']
 elif 'absent from structural source' in reason:
  phase='refusal_scan';emitter='internal/lower/optional_widening.go: (*lowering).optionalWideningRefusal (via refuseOptionalWidening)';source=['optional_widening.go','lower.go','refusals.go','refusal_rules.go']
 elif 'whose readonly field' in reason:
  phase='refusal_scan';emitter='internal/lower/invariance.go: (*lowering).wideningRefusal (via refuseWidening)';source=['invariance.go','library_string.go','lower.go','refusals.go','refusal_rules.go']
 elif 'a method read as a value' in reason:
  phase='refusal_scan';emitter='internal/lower/refusals.go: (*lowering).methodValueStop (via methodValueRefusal)';source=['refusals.go','lower.go']
 elif reason == "a cast the runtime can't check":
  phase='refusal_scan';emitter='internal/lower/cast_proof.go: (*lowering).castProof (via castRefusal)';source=['cast_proof.go','refusals.go','lower.go']
 elif 'assigning an element of a value' in reason:
  phase='lowering';emitter='internal/lower/object.go: (*lowering).setIndex';source=['object.go']
 else:raise AssertionError(('unresolved emitter',g))
 g['phase']=phase;g['owner']=owners.get(reason,emitter);g['emitter']=emitter
 item={'where':where,'reason':reason,'phase':phase,'owner':g['owner'],'emitter':emitter,'source_files':source}
 if where!='-':
  name,line,col=where.rsplit(':',2);lines=(tree/'src/compiler'/name).read_text().splitlines();item['adapted_excerpt']='\n'.join(f'{i+1}: {lines[i]}' for i in range(max(0,int(line)-4),min(len(lines),int(line)+3)))
 for name in source:(base/(name+'.txt')).write_bytes((repo/'internal/lower'/name).read_bytes())
 evidence.append(item)
(out/'grouped-first-stops.json').write_text(json.dumps(groups,indent=2)+'\n')
(out/'source-evidence.json').write_text(json.dumps(evidence,indent=2)+'\n')
