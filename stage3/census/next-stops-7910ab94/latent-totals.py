"""Count emitted findings, keeping incomplete ledgers distinct from complete totals."""
import collections,json
from pathlib import Path
out=Path(__file__).resolve().parent
entry=[json.loads(line) for line in (out/'entry-latent.jsonl').read_text().splitlines()]
status=json.loads((out/'entry-census-status.json').read_text())
files=[r for r in entry if 'file' in r]
findings=[f for r in entry for f in r.get('findings') or []]
counts=collections.Counter(f['kind'] for f in findings)
expected=set(entry[0].get('sources',[]));seen={r['file'] for r in files}
ordinary=[json.loads(line) for line in (out/'census.jsonl').read_text().splitlines()]
first=collections.Counter(r['kind'] for r in ordinary)
result=dict(guarded_entry=dict(finding_totals={k:counts[k] for k in ['NotYet','Refused','panic']},all_finding_kinds=dict(counts),stream_records=len(entry),completed_sources=len(seen),expected_sources=len(expected),complete=status['exit_code']==0 and expected==seen,child_status=status,checker_diagnostics=len(entry[0].get('diagnostics',[]))),ordinary_counted_observations=dict(count=len(ordinary),first_stop_totals={k:first[k] for k in ['NotYet','Refused','panic']},all_first_stop_kinds=dict(first),carries_latent_totals=False,explanation='The ordinary production stream has one kind per root and no latent findings or latent counters. These are first-stop totals only; full latent totals are unavailable.'))
(out/'latent-totals.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
