import gzip
import hashlib
import json
from pathlib import Path
lane=Path(__file__).resolve().parent
root=lane.parents[3]
source=root/'stage3/interface-downcasts/lazy/census/read-demand-pairs.json.gz'
rows=[r for r in json.loads(gzip.decompress(source.read_bytes())) if 'object plus primitive union' in r['families']]
rows.sort(key=lambda r:(-r['read_count'],r['type_id'],r['field']))
assert len(rows)==42 and sum(r['read_count'] for r in rows)==181
shapes={
 'true | Node | undefined':'node-indicator',
 'string | DiagnosticMessageChain':'diagnostic',
 'string | number | PseudoBigInt':'literal',
 'string | NodeArray<JSDocComment> | undefined':'comment',
}
certification=json.loads((lane.parent/'original/certification.json').read_text())
certified={(p['type_id'],p['field']) for p in certification['pairs']}
for fixture,expected in certification['fixtures'].items():
 assert hashlib.sha256((lane.parent/'original'/fixture).read_bytes()).hexdigest()==expected,fixture
for certificate,expected in certification.get('shared_certificates',{}).items():
 assert hashlib.sha256((lane.parent/'original'/certificate).read_bytes()).hexdigest()==expected,certificate
certified_rows=[r for r in rows if (r['type_id'],r['field']) in certified]
modeled=[r for r in rows if r['declared_type'] in shapes]
for index,row in enumerate(rows,1):
 row['rank']=index
 row['status']='certified standalone original field contract; execution reachability unmeasured' if (row['type_id'],row['field']) in certified else 'pending original-pair witness'
 if row['declared_type'] in shapes: row['reduced_shape_control']=shapes[row['declared_type']]
result={'basis':'lazy static candidates; exact production reachability unmeasured','integration_sha':'432d4913d31daaa49d8ca4eb46f30f21f90da5ee','source_sha256':hashlib.sha256(source.read_bytes()).hexdigest(),'pairs':42,'candidate_reads':181,'completed_pairs':len(certified_rows),'completed_reads':sum(r['read_count'] for r in certified_rows),'completion_basis':certification['basis'],'remaining_candidate_pairs':len(rows)-len(certified_rows),'remaining_candidate_reads':sum(r['read_count'] for r in rows if (r['type_id'],r['field']) not in certified),'reduced_shape_candidate_pairs':len(modeled),'reduced_shape_candidate_reads':sum(r['read_count'] for r in modeled),'candidate_pairs_without_shape_control':42-len(modeled),'candidate_reads_without_shape_control':181-sum(r['read_count'] for r in modeled),'rows':rows}
(lane/'lazy-pair-progress.json').write_text(json.dumps(result,indent=2)+'\n')
print('42 candidate pairs / 181 candidate reads; exact reachability unmeasured')
for row in rows: print(row['rank'],row['read_count'],row['type'],row['field'],row['declared_type'])
