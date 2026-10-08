"""Prove the independent corpus auditor rejects forged host/body certifications."""
import copy,json,pathlib,sys
from importlib.machinery import SourceFileLoader
module=SourceFileLoader('shape_audit',str(pathlib.Path(__file__).with_name('audit.py'))).load_module()
result=json.loads(pathlib.Path(sys.argv[1]).read_text());mapped=json.loads(pathlib.Path(sys.argv[2]).read_text());root=pathlib.Path(sys.argv[3])
module.audit(result,mapped,root)
for reason in ('host metadata','diagnosed body'):
 mutant=copy.deepcopy(result);row=next(r for r in mutant['sites'] if (bool(r['host_values']) if reason=='host metadata' else r['detail']==['containing function body has checker diagnostics; body skipped']))
 old=row['outcome'];row['outcome']=module.OUTCOMES[0]
 if reason=='diagnosed body':row['allocation_sites']=[result['allocation_schemas'][0]['site']]
 mutant['counts'][row['kind']][old]-=1;mutant['counts'][row['kind']][row['outcome']]+=1
 old_reason=row['reason']
 mutant['unknown_reasons'][row['kind']][old_reason]-=1
 if mutant['unknown_reasons'][row['kind']][old_reason]==0:del mutant['unknown_reasons'][row['kind']][old_reason]
 try:module.audit(mutant,mapped,root)
 except AssertionError as error:print(reason,'certification mutant caught:',error)
 else:raise AssertionError(reason+' forged certification escaped')
