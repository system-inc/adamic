"""Verify an already built road (b) product against current main plus the accepted adapter."""
from pathlib import Path
import hashlib,json,subprocess,tarfile,io
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/getfunctionname');fresh=s/'tree';old=Path('/workspace/scratch/debug-messages/after')
skip={'.git','node_modules','built','test.config','mytest.config','.failed-tests','patch-set.md'}
def manifest(root):
 result={}
 for p in sorted(root.rglob('*')):
  rel=p.relative_to(root)
  if rel.parts[0] in skip or rel.parts[:3]==('tests','baselines','local') or p.is_symlink() or not p.is_file():continue
  result[str(rel)]=hashlib.sha256(p.read_bytes()).hexdigest()
 return result
new=manifest(fresh);previous=manifest(old);different=[k for k in new.keys()|previous.keys() if new.get(k)!=previous.get(k)];assert not different,different
archive=subprocess.check_output(['git','show','385ccad0bf6cfd800443566d947cc638a7e776aa:stage3/adapt/99-correctness-core-debug-messages/after-javascript.tar.gz'],cwd=repo)
(s/'raw/reused-javascript.tar.gz').write_bytes(archive)
with tarfile.open(fileobj=io.BytesIO(archive),mode='r:gz') as tf:
 compiled=json.loads(tf.extractfile('manifest.json').read())
 for name,digest in compiled.items():assert hashlib.sha256((old/'built/local'/name).read_bytes()).hexdigest()==digest,name
maps=[]
for name in ['_tsc.js','typescript.js','run.js']:
 mp=old/'built/local'/(name+'.map');data=json.loads(mp.read_text());matches=[]
 for index,source in enumerate(data['sources']):
  content=data.get('sourcesContent',[None]*len(data['sources']))[index]
  resolved=(mp.parent/source).resolve()
  try:rel=resolved.relative_to(old)
  except ValueError:continue
  if str(rel) in ['src/compiler/debug.ts','src/compiler/core.ts','src/deprecatedCompat/deprecate.ts']:
   if content is not None:assert content==(fresh/rel).read_bytes().decode(),(name,str(rel))
   assert str(rel) in new;matches.append(str(rel))
 maps.append({'file':name+'.map','sha256':hashlib.sha256(mp.read_bytes()).hexdigest(),'namedSourceInputsVerified':matches,'sourceContentsEmbedded':'sourcesContent' in data})
result={'main_sha':'3214dd465e1155d7bf4d41653e95b29a3f1060a6','accepted_road_b_sha':'2a95448b8fadbcd30549a148bd28da5f8b75f907','accepted_adapter_blob':'6953c4e927326bb13673ea717e7a19664bd7575d','accepted_pins_blob':'f68fe1717dc92fd1a392d9f7683733928c582bd9','source_apply_product_key':'d2af18f8b5611dd97bc936f672250b934cd97593a02dfc7f1fcfe1932066dbcf','new_tree':str(fresh),'reused_product':str(old),'input_files_compared':len(new),'input_differences':different,'excluded_non_inputs':sorted(skip)+['tests/baselines/local'],'compiled_files_verified_against_prior_commit':compiled,'prior_build_evidence_commit':'385ccad0bf6cfd800443566d947cc638a7e776aa','source_maps':maps,'compiler_or_harness_build_run':False,'adaptation_constructed':False,'manifest':new}
(s/'raw/reuse-validation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items() if k not in ['manifest','compiled_files_verified_against_prior_commit']}))
