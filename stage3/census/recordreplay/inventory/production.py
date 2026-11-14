from pathlib import Path
import sys,importlib.util,json,hashlib,shutil,ast,subprocess
D=Path('/workspace/scratch/recordreplay');R=D/'head';O=D/'inventory-audit';tree=D/'candidate2';sys.dont_write_bytecode=True;sys.path.insert(0,str(R/'stage3'));sys.path.insert(0,str(R/'stage3/lane'))
def module(name,file):
 spec=importlib.util.spec_from_file_location(name,file);m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m);return m
apply=module('reader_apply',R/'stage3/apply.py');artifact=module('reader_artifact',R/'stage3/lane/artifact.py');view=module('reader_view',R/'stage3/lane/view.py');result={}
def manifest(root):return {str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in root.rglob('*') if p.is_file()}
# Exercise the exact source inventory with no node_modules/build absolute links.
payload=O/'payload-input';payload.mkdir(exist_ok=True)
if not (payload/'src').exists():shutil.copytree(tree/'src',payload/'src',copy_function=__import__('os').link)
without=O/'payload-without-added';without.mkdir(exist_ok=True)
if not (without/'src').exists():
 shutil.copytree(tree/'src',without/'src',copy_function=__import__('os').link);(without/'src/recordreplay.d.ts').unlink()
source=manifest(tree/'src');assert len(source)==748
new_digest=artifact.payload_digest(payload);old_digest=artifact.payload_digest(without);assert new_digest!=old_digest
result['lane/artifact.payload_digest']={'status':'pass','amended_src_files':len(source),'contains_added':True,'digest_with_added':new_digest,'digest_without_added_only':old_digest,'added_file_changes_identity':True,'scope':'source-only payload fixture containing byte-identical full amended src; unrelated build/tests omitted'}
old_key=subprocess.check_output(['python3',str(D/'base/stage3/apply.py'),'--key'],text=True).strip();new_key=apply.product_key();assert old_key!=new_key
result['apply.product_key']={'status':'pass','unamended_inputs_key':old_key,'amended_working_inputs_key':new_key,'different':True,'note':'observed before writing new proof evidence; evidence itself also participates in future keys'}
key='inventory-probe-'+new_key;store=O/'product-store';apply.save_product(str(store),key,payload);restored=O/'product-restored';apply.restore_product(str(store),key,restored);assert manifest(restored/'src')==source
result['apply.save_product/restore_product']={'status':'pass','src_files':len(source),'added_file_roundtrip_sha256':hashlib.sha256((restored/'src/recordreplay.d.ts').read_bytes()).hexdigest(),'manifest':json.loads((store/(key+'.manifest')).read_text()),'scope':'private source-only payload fixture; not published as a production build product'}
# The actual lane view links src as one immutable input, so the added file is visible.
output=O/'lane-view';view.make_view(tree,output);assert manifest(output/'src')==source
result['lane/view.make_view']={'status':'pass','src_files':len(source),'src_is_symlink':(output/'src').is_symlink(),'added_visible':(output/'src/recordreplay.d.ts').is_file()}
fixed=[]
for entry in json.loads((O/'catalog.json').read_text())['fixed_python_count_guards']:
 row={**entry,'results':{}}
 for label,count in [('control',747),('amended',748)]:
  if 'sum(f.is_file()' in entry['source']:ns={'scoped':D/('control' if label=='control' else 'candidate2')}
  elif "i['after']" in entry['source']:ns={'i':{'after':dict.fromkeys(range(count)), 'previous':dict.fromkeys(range(count)),'built_compiler_declarations_exist':False,'publication_compiler_diff':False}}
  else:ns={'actual':dict.fromkeys(range(count)),'expected':dict.fromkeys(range(count))}
  try:exec(compile(entry['source'],entry['file'],'exec'),ns);row['results'][label]={'status':'pass','actual':count}
  except AssertionError:row['results'][label]={'status':'reject','actual':count,'expected':746}
 fixed.append(row)
result['fixed_python_guards_before_repair']=fixed
# Execute the historical frozen-manifest inventory function, without its server jobs.
script=(R/'stage3/census/speculative/server-run.sh').read_text();start=script.index('def source_checks(');end=script.index('\n\ndef ',start+1);code=script[start:end];ns={'digest':lambda b:hashlib.sha256(b).hexdigest()};exec(code,ns)
for label,identity_root in [('fresh',tree),('stale_names',D/'control')]:
 names=[p for p in (identity_root/'src').rglob('*') if p.is_file() and p.suffix in ('.ts','.a')]
 # Keep byte checks current, isolating the inventory membership guard.
 identity={'files':[{'file':str(p.relative_to(identity_root)),'bytes':(tree/p.relative_to(identity_root)).stat().st_size,'sha256':hashlib.sha256((tree/p.relative_to(identity_root)).read_bytes()).hexdigest()} for p in names]}
 try:ns['source_checks'](tree,identity);result['speculative.server.source_checks/'+label]={'status':'pass','files':len(names)}
 except AssertionError as e:result['speculative.server.source_checks/'+label]={'status':'reject','reason':str(e),'files':len(names)}
(O/'production.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v.get('status') if isinstance(v,dict) else len(v) for k,v in result.items()}))
