"""Apply report.py's attempted-source, readiness and Node-throw criteria."""
import json,sys
from pathlib import Path
out=Path(__file__).resolve().parent
sys.path.insert(0,str(out.parents[1]/'meter'))
from report import READINESS_REASONS,stop_file
rows=[r for n in range(1,5) for r in json.loads((out/f'group-{n}-rows.json').read_text())]
probes=json.loads((out/'node-readiness/results.json').read_text())
status=json.loads((out/'entry-census-status.json').read_text())
latent=[json.loads(line) for line in (out/'entry-latent.jsonl').read_text().splitlines()]
expected=set(latent[0]['sources']);seen={r['file'] for r in latent[1:] if 'file' in r}
complete=status['exit_code']==0 and expected==seen
attempted={r['file'] for r in latent[1:] if any(u['status'] in ('attempted','panic','speculative') for u in r.get('units',[]))}
readiness={stop_file(f['where']) for r in latent[1:] for f in r.get('findings') or [] if f['reason'] in READINESS_REASONS}
checks=[]
for r in rows:
 if r.get('reason') not in READINESS_REASONS:continue
 root=r['roots'][0];probe=probes[root]
 eligible=stop_file(r['where']) in attempted-readiness
 label='artifact classification incomplete' if not complete else 'root-order artifact' if eligible and probe['throws'] else 'counts against root file'
 checks.append(dict(root=root.split('/src/compiler/')[-1],where=r['where'].split('/src/compiler/')[-1],reason=r['reason'],classification=label,node=probe,entry_stop_source_attempted=stop_file(r['where']) in attempted,entry_stop_source_readiness=stop_file(r['where']) in readiness))
result=dict(status='complete' if complete else 'incomplete',entry_status=status,entry_stream_records=len(latent),expected_sources=len(expected),completed_sources=len(seen),stopping_reason=None if complete else 'The bounded entry census did not complete its per-file ledger. report.py requires attempted-source and finding evidence before classifying artifacts.',checks=checks)
(out/'root-order.json').write_text(json.dumps(result,indent=2)+'\n')
print(result['status'],[(r['root'],r['classification']) for r in checks])
