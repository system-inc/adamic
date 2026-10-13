"""Independent coverage, source-provenance and observation audit with record mutants."""
import collections,hashlib,json,pathlib,sys
root=pathlib.Path(__file__).resolve().parent
ledger=json.loads((root/'evidence/ledger-additions.json').read_text())
expected=[r for r in ledger if r['kind']=='as_cast' and r['disposition']=='open']
data=json.loads((root/'contracts.json').read_text());rows=data['rows']
if '--drop-row' in sys.argv:rows.pop()
key=lambda r:(r['file'],r['node_start'],r['node_end'])
assert len(rows)==31 and collections.Counter(map(key,rows))==collections.Counter(map(key,expected)),'exact 31-row ledger coverage'
outside=json.loads((root/'evidence/outside-stock-casts.json').read_text())['adapted_additions']
assert collections.Counter((r['file'],r['line'],r['column'],r['text']) for r in rows)==collections.Counter((r['file'],r['line'],r['column'],r['text']) for r in outside if r['kind']=='as_cast'),'independent cast-classifier population'
assert collections.Counter(r['decision'] for r in rows)==data['counts']
measurement=json.loads((root/'evidence/source-measurement.json').read_text())
assert collections.Counter(map(key,measurement['attribution']))==collections.Counter(map(key,rows))
if '--wrong-owner' in sys.argv:rows[0]['adaptation']='stage3/adapt/45-regex-captures/'
source={key(r):r for r in measurement['attribution']}
for row in rows:
 assert row['adaptation']==source[key(row)]['introduction']['adaptation'],'immediate source introduction owner'
 assert source[key(row)]['introduction']['after']>source[key(row)]['introduction']['before']
 assert row['why_added'] and row['truthful_type'] and row['proof'] and row['proposal']
 assert row['proven_adamic_upcast'] is False
assert measurement['baseline_semantic_diagnostics']==[]
assert sum(p['sites'] for p in measurement['proposals'])==21
assert all(p['semantic_diagnostics']==[] and p['byte_identical_javascript'] for p in measurement['proposals'])
fixtures=json.loads((root/'fixtures/status.json').read_text());assert len(fixtures)==3
for row in fixtures:
 p=root/'fixtures'/row['file'];header,body=p.read_bytes().split(b'\n',1)
 assert header.decode()==row['header'] and hashlib.sha256(body).hexdigest()==row['body_sha256']
 assert row['node']['exit']==0 and row['node']['stderr']==''
 assert row['mutant']['observation']['exit']==0 and row['mutant']['observation']['stderr']=='' and row['mutant']['observation']['stdout']!=row['node']['stdout']
 assert row['stage0']['outcome']=='Refused'
acheck=json.loads((root/'evidence/a-check.json').read_text())
assert acheck['baseline']['exit']==acheck['final']['exit']==0
assert len(acheck['mutants'])==6 and all(r['exit']==1 for r in acheck['mutants'])
print('PASS: exact 31-row coverage, measured owners, 21 byte-identical upstream proposals, three Node goldens and mutants, truthful headers and six a-check mutants')
