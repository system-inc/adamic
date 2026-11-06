"""Join independently located real casts to before/final refusal-scan findings."""
import json,sys
from pathlib import Path
work=Path(sys.argv[1]);out=Path(__file__).resolve().parent
sites=json.loads((work/'site-locations-final.json').read_text())
def record(name):return json.loads((work/name).read_text().splitlines()[1])
debug={'before':record('latent-before.jsonl'),'after':record('latent-final.jsonl')}
emitter={'before':record('emitter-scan-before.jsonl'),'after':record('emitter-scan-final.jsonl')}
for site in sites:
 for label in ['before','after']:
  observations=(debug if site['file'].endswith('debug.ts') else emitter)[label]
  matches=[f for f in observations['findings'] if f['where']==site[label]['where'] and f['phase']=='refusal_scan' and f['kind'] in ('Refused','NotYet')]
  first=next(f for f in matches if f['kind']==('Refused' if label=='before' else 'NotYet'))
  if label=='before':assert 'which can write Record<string, string | number>' in first['reason']
  else:
   assert first['reason']=='an ESM namespace object used as a value; use qualified live bindings'
   assert not any(f['kind']=='Refused' for f in matches)
  site[label].update({'stop':first['kind'],'reason':first['reason'],'phase':'refusal_scan','compiler_source':'internal/lower/invariance.go:673' if label=='before' else 'internal/lower/module_namespace.go:27','finding':first,'lowering_findings_at_site':[f for f in observations['findings'] if f['where']==site[label]['where'] and f['phase']=='lowering' and f['kind'] in ('Refused','NotYet')]})
assert len(sites)==19
(out/'site-stops.json').write_text(json.dumps(sites,indent=2)+'\n')
lines=['# The 19 enum-view sites','','All 19 real cast locations are Refused before and NotYet after in the existing refusal scan. No final cast has a Refused finding. Before diagnostics originate in invariance.go:673; final namespace diagnostics originate in module_namespace.go:27. These are measurement observations, not successful native lowering.','','| Rule | Site / owner | Before | Final |','| --- | --- | --- | --- |']
for s in sites:
 where=s['after']['where'].split('/src/compiler/')[-1]
 lines.append(f"| {s['id']} | `{where}`, {s['after']['owner']}, {s['key']} | Refused: writable record | NotYet: namespace value |")
lines+=['','The 18 debug.ts sites use the full latent census restricted to debug.ts. The nineteenth, emitter.ts:5790:141 in makeName, uses the identical refusal scan without the unrelated emitter unit walk. The scan-only variant matches all 57 before-debug refusal-scan findings exactly. Its no-output guards remain enabled.','','The independent emitter single-unit lowering probe stops earlier, at emitter.ts:5790:84 reading autoGenerate, on both trees. The cast-specific scan result does not claim that this earlier boundary has been removed.','','Full raw records, exact reasons, phase-specific observations and actual namespace locations are retained in site-stops.json and the bounded piece directories.']
(out/'SITE-STOPS.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({'sites':19,'before':{'Refused':19},'final':{'NotYet':19},'final_refused_casts':0},indent=2))
