import json
from pathlib import Path
raw=Path(__file__).parent

def findings(side):
 records=[json.loads(line) for line in (raw/(side+'-census.jsonl')).read_text().splitlines()]
 header=records[0]
 files=[row for row in records if 'file' in row]
 assert len(files)==1 and files[0]['file'].endswith('/src/compiler/tracing.ts')
 assert header['latent_mode']=='full' and not header['diagnostics']
 assert records[-1]['status']=='relation_report'
 def clean(row):
  return {key:(value.split('/src/compiler/')[-1] if isinstance(value,str) and '/src/compiler/' in value else value) for key,value in row.items() if key in ('unit','phase','kind','where','reason')}
 return [clean(row) for row in files[0]['findings'] if row.get('unit','').endswith('tracing.ts:59:5') and row['kind'] in ('Refused','NotYet')]
a,b,m=(findings(side) for side in ['before','after','mutant'])
added=[row for row in b if row not in a];removed=[row for row in a if row not in b]
report=dict(before=a,after=b,mutant=m,added=added,removed=removed,restoringMutantMatchesBefore=m==a)
(raw/'census-comparison.json').write_text(json.dumps(report,indent=2)+'\n')
assert m==a,'restoring mutant does not match before'
assert any(row['kind']=='Refused' and row['where']=='tracing.ts:103:67' and row['reason']=='a spread after the first field' for row in a)
assert not any('spread after' in row['reason'] for row in b)
assert not added,'new startTracing findings: '+str(added)
print(json.dumps(report,indent=2))
