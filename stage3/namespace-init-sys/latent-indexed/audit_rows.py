"""Independent recount: conserve the ledger and require raw evidence for each claim."""
import collections,csv,gzip,json,sys,copy
from pathlib import Path
repo=Path(__file__).resolve().parents[3];out=Path(sys.argv[1]);source=Path(sys.argv[2])
ledger={r['id']:r for r in csv.DictReader((repo/'stage3/stricter-indexed-all/evidence/ledger-rows.csv').open()) if r['option']=='noUncheckedIndexedAccess'}
rows=json.loads((out/'rows.json').read_text());summary=json.loads((out/'summary.json').read_text())
raw=[json.loads(x) for x in gzip.open(out/'raw.jsonl.gz','rt')]
def normal(s):return s.replace(str(source)+'/', '')
findings={(f['kind'],normal(f['where']),normal(f['text'])) for r in raw[1:] for f in r['findings']}
findings|={(x['failure']['kind'],normal(x['failure']['where']),normal(x['failure']['text'])) for r in raw[1:] for x in r['function_attempts'] or [] if x.get('failure')}
checks={(normal(c['unit']),normal(c['where'])) for r in raw[1:] for c in r['checks'] or [] if c['kind']=='indexed-presence'}
file_records={normal(r['file']):r for r in raw[1:]}
sites={(normal(d['site']['file']),d['site']['line'],d['site']['column']):d['site']['position'] for d in raw[0]['option_dispositions'] if 'noUncheckedIndexedAccess' in d['site']['options']}
function_kinds={'KindFunctionDeclaration','KindFunctionExpression','KindArrowFunction','KindMethodDeclaration','KindConstructor','KindGetAccessor','KindSetAccessor'}
def validate(rs):
 assert len(rs)==99 and {r['id'] for r in rs}==set(ledger)
 for r in rs:
  l=ledger[r['id']];assert (r['file'],r['line'],r['column'])==(l['file'],int(l['line']),int(l['column']))
  pos=sites[(r['file'],r['line'],r['column'])]
  owners=[d for d in file_records[r['file']]['declarations'] if d['kind'] in function_kinds and d['start']<=pos<d['end']]
  nearest=sorted(owners,key=lambda d:d['end']-d['start'])[0] if owners else None
  assert (r['owner']['start'],r['owner']['end'])==(nearest['start'],nearest['end']) if nearest else r['owner'] is None

  if r['state']=='check emitted':assert r['checks'] and all((x['unit'],x['where']) in checks for x in r['checks'])
  elif r['state']=='function blocked by a refusal':
   b=r['blocker'];assert r['owner'] and b and b['kind'] in ('Refused','NotYet')
  else:assert r['state']=='not reached' and r['why']
  if r['blocker']:
   b=r['blocker'];assert (b['kind'],b['where'],b['text']) in findings
validate(rows)
assert dict(collections.Counter(r['state'] for r in rows))==summary['counts']
for table in summary['blocking_refusals']:
 matches=[r for r in rows if r['blocker'] and (r['blocker']['kind'],r['blocker']['where'],r['blocker']['text'])==(table['kind'],table['where'],table['message'])]
 assert len(matches)==table['rows'] and {r['id'] for r in matches}==set(table['ids'])
for name,mutate in [('lost-row',lambda rs:rs.pop()),('fabricated-emission',lambda rs:rs[0].update(state='check emitted',checks=[{'unit':'invented','where':'invented'}])),('wrong-function-span',lambda rs:next(r for r in rs if r['owner'])['owner'].update(start=-1)),('wrong-blocker-location',lambda rs: next(r for r in rs if r['blocker'])['blocker'].update(where='invented:1:1'))]:
 changed=copy.deepcopy(rows);mutate(changed)
 try:validate(changed)
 except AssertionError:print(name+' mutant caught')
 else:raise AssertionError(name+' survived')
print('99 exact ledger identities; raw IR/refusal witnesses; counts and blocker table independently recounted: pass')
