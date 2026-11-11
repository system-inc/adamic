import json
from pathlib import Path
raw=Path(__file__).parent

def findings(side):
 records=[json.loads(line) for line in (raw/('current-'+side+'-census.jsonl')).read_text().splitlines()]
 assert records[0]['latent_mode']=='full' and records[0]['diagnostics']==[]
 assert records[-1]['status']=='relation_report'
 files=[row for row in records if 'file' in row];assert len(files)==1
 def clean(row):
  return {key:(value.split('/src/compiler/')[-1] if isinstance(value,str) else value) for key,value in row.items() if key in ('unit','phase','kind','where','reason')}
 return [clean(row) for row in files[0]['findings']]
a,b,m=(findings(side) for side in ['before','after','mutant'])
assert m==a,'restoring mutant differs from before'
added=[row for row in b if row not in a];removed=[row for row in a if row not in b]
reason='JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)'
expected=dict(unit='tracing.ts:59:5',phase='lowering',kind='NotYet',where='tracing.ts:104:46',reason=reason)
assert added==[expected],('unexpected new finding anywhere in tracing.ts',added)
assert removed==[dict(unit='tracing.ts:59:5',phase='lowering',kind='Refused',where='tracing.ts:103:67',reason='a spread after the first field')]
existing=[r for r in a if r['where']=='tracing.ts:194:67' and r['reason']==reason];assert len(existing)==1
assert existing[0] in b
scoped=lambda rows:[row for row in rows if row['unit']=='tracing.ts:59:5' and row['kind'] in ('NotYet','Refused')]
report=dict(passed=True,scope='all findings anywhere in tracing.ts',beforeFindings=len(a),afterFindings=len(b),before=scoped(a),after=scoped(b),mutant=scoped(m),added=added,removed=removed,restoringMutantMatchesBefore=True,existingSameStop=existing[0],allowedNewlyVisibleStop=expected,otherNewFindings=0)
(raw/'current-census-comparison.json').write_text(json.dumps(report,indent=2)+'\n')
print('whole-file comparison: only the authorized next JSON.stringify stop becomes visible; restoring mutant matches')
