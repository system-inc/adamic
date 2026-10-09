import pathlib,json,re,subprocess,os,time,difflib
root=pathlib.Path('/workspace/adamic');ev=pathlib.Path('/tmp/u131/evidence'); cache=pathlib.Path.home()/'.cache/adamic-build'
def product(log,name,cache):
 for line in (ev/log).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  m=re.search(r'build '+re.escape(name)+r' ([0-9a-f]{12}) ',e.get('Output',''))
  if m:
   ds=[p for p in cache.glob(m[1]+'*') if p.is_dir()];assert len(ds)==1;return ds[0]
 raise Exception((name,log))
clean=product('bounded-baseline.log','markdowninline_native',cache)/'port'; mutant=product('M3.log','markdowninline_native',pathlib.Path('/tmp/u131/cache/M3'))/'port'; oracle=product('M3.log','markdowninline_Go_overlay_bridge',pathlib.Path('/tmp/u131/cache/M3'))/'go-printer'
input=ev/'M3-witness.txt';input.write_text('s===\n'); witness=[]
for id,binary,args in [('clean',clean,['--batch',str(input)]),('M3',mutant,['--batch',str(input)]),('Go',oracle,[str(input)])]:
 cmd=[str(binary),*args];start=time.monotonic()
 with (ev/('survivor-'+id+'.log')).open('w') as f:p=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 answer=(ev/('survivor-'+id+'.log')).read_text();witness.append({'side':id,'command':cmd,'exit':p.returncode,'wall':time.monotonic()-start,'stdout':answer});assert p.returncode==0
assert witness[0]['stdout']==witness[2]['stdout'] and witness[0]['stdout']!=witness[1]['stdout']
(ev/'survivor-witness.json').write_text(json.dumps(witness,indent=2));print(witness)
# Permitted suite construction change: omit the required generated C artifact.
harness='stage1/cohere/markdowninline/inline_products_shards_test.go';base=(root/harness).read_text();old='\tif err = inlineWrite(dir, "program.c", []byte(native.C(program))); err != nil {\n\t\treturn err\n\t}\n';assert base.count(old)==1;changed=base.replace(old,'',1);(ev/'S6.diff').write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+harness,tofile='b/'+harness)));target=ev/'S6.go.txt';target.write_text(changed);overlay=ev/'S6.overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/harness):str(target)}}))
names=json.loads((ev/'matrix-test-names.json').read_text()); cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-overlay='+str(overlay),'./stage1/cohere/markdowninline/','-run','^('+'|'.join(names)+')$'];env={**os.environ,'ADAMIC_MARKDOWNINLINE_LIBRARY':'/tmp/u131/prettier/node_modules/prettier','ADAMIC_BUILD_CACHE_DIR':'/tmp/u131/cache/S6'};start=time.monotonic()
with (ev/'S6.log').open('w') as f:p=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
assert p.returncode!=0
runs=json.loads((ev/'runs.json').read_text());runs.append({'id':'S6','command':cmd,'environment':{'ADAMIC_MARKDOWNINLINE_LIBRARY':env['ADAMIC_MARKDOWNINLINE_LIBRARY'],'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},'exit':p.returncode,'wall':time.monotonic()-start});(ev/'runs.json').write_text(json.dumps(runs,indent=2))
cat=json.loads((ev/'catalog.json').read_text());cat.append({'id':'S6','file':harness,'line':base[:base.index(old)].count('\n')+1,'before':old,'after':'drop required program.c write','kind':'construction'});(ev/'catalog.json').write_text(json.dumps(cat,indent=2))
with (ev/'S6-vet.log').open('w') as f:p=subprocess.run(['go','vet','-overlay='+str(overlay),'./stage1/cohere/markdowninline/'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
assert p.returncode==0
