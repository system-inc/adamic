import json
from pathlib import Path
raw=Path(__file__).parent

def read(name):
 rows=[json.loads(line) for line in (raw/name).read_text().splitlines()]
 assert rows[0]['latent_mode']=='full' and rows[0]['diagnostics']==[] and rows[-1]['status']=='relation_report'
 files=[row for row in rows if 'file' in row];assert len(files)==1
 return [{k:(v.split('/src/compiler/')[-1] if isinstance(v,str) else v) for k,v in f.items() if k in ('unit','phase','kind','where','reason')} for f in files[0]['findings']]
a=read('current-before-census.jsonl');b=read('catalog-after-census.jsonl');length=read('catalog-length-mutant-census.jsonl');both=read('catalog-all-mutant-census.jsonl')
reason='JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)'
expected=dict(unit='tracing.ts:59:5',phase='lowering',kind='NotYet',where='tracing.ts:104:46',reason=reason)
added=[r for r in b if r not in a];removed=[r for r in a if r not in b];assert added==[expected]
assert {(r['where'],r['kind']) for r in removed}=={('tracing.ts:72:9','NotYet'),('tracing.ts:72:9','Boundary'),('tracing.ts:103:67','Refused')}
assert all(r['unit']=='tracing.ts:59:5' for r in removed)
assert both==a
assert length==read('current-after-census.jsonl')
assert any(r['where']=='tracing.ts:72:9' and r['kind']=='NotYet' and r['reason']=='assigning a field of a value' for r in length)
assert not any(r['where']=='tracing.ts:72:9' for r in b)
existing=[r for r in a if r['where']=='tracing.ts:194:67' and r['reason']==reason];assert len(existing)==1 and existing[0] in b
scope=lambda rows:[r for r in rows if r['unit']=='tracing.ts:59:5' and r['kind'] in ('NotYet','Refused')]
result=dict(passed=True,scope='all findings anywhere in tracing.ts',before=scope(a),after=scope(b),lengthOnlyMutant=scope(length),allRestoringMutant=scope(both),added=added,removed=removed,allowedNewlyVisibleStop=expected,existingSameStop=existing[0],otherNewFindings=0,lengthOnlyMutantRestores72=True,allRestoringMutantMatchesBefore=True)
(raw/'catalog-census-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print('72:9 and 103:67 removed; both restoring controls pass; only authorized 104:46 is newly visible')
