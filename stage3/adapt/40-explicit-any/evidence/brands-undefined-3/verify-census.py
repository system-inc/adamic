import sys,json,gzip,shutil,importlib.util,hashlib,collections
from pathlib import Path
sys.dont_write_bytecode=True
repo=Path('/workspace/adamic');work=Path('/workspace/scratch/brands3');out=repo/'stage3/adapt/40-explicit-any/evidence/brands-undefined-3'
spec=importlib.util.spec_from_file_location('meter_report',repo/'stage3/meter/report.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
summary={}
for label,tree in [('control',Path('/workspace/scratch/brands2/control-tree')),('retyped',work/'retyped-tree')]:
 source=work/'census'/(label+'.jsonl');records=[json.loads(s) for s in source.read_text().splitlines()];expected={str(p.resolve()) for p in (tree/'src/compiler').rglob('*.ts')};r=m.lowering_summary(records,expected,'measured on a checker-rejected program');assert records[0]['latent_mode']=='first-error'
 summary[label]={'files':r['source_files'],'totals':r['totals'],'checker_diagnostics':len(records[0]['diagnostics']),'void_container':sum(v for k,v in r['per_reason'].items() if "a read of a void container's elements" in k),'instantiated_at_void':sum(v for k,v in r['per_reason'].items() if 'instantiated at void' in k),'Path_value':sum(v for k,v in r['per_reason'].items() if k.endswith(': a value of type Path')),'all_primitive_brand_not_void':sum(v for k,v in r['per_reason'].items() if 'a primitive brand member ' in k and 'whose type is not void' in k)}
 original_five={'__compilerOptionsKey','__incrementalBuildInfoFileIdBrand','__incrementalBuildInfoFileIdListIdBrand','__modeAwareCacheKey','__pathBrand'}
 summary[label]['original_five_primitive_brand_sites']=sum(v for k,v in r['per_reason'].items() if any('a primitive brand member '+name+' whose type is not void' in k for name in original_five))
 with source.open('rb') as f,(out/(label+'-latent.jsonl.gz')).open('wb') as g:
  with gzip.GzipFile(filename='',mode='wb',fileobj=g,mtime=0) as z:shutil.copyfileobj(f,z)
 findings=set()
 for record in records[1:]:
  for f in record.get('findings') or []:
   if f['kind'] not in ['NotYet','Refused']:findings.add(tuple(f[k] for k in ['kind','where','reason','text']))
 (out/(label+'-other-findings.json')).write_text(json.dumps([dict(zip(['kind','where','reason','text'],f)) for f in sorted(findings)],indent=2)+'\n')
(out/'meter-verification.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
