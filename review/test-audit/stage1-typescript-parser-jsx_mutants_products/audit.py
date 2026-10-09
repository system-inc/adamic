import pathlib,json,re,subprocess,os,time,difflib
root=pathlib.Path('/workspace/adamic');ev=pathlib.Path('/tmp/u155/evidence');file='stage1/typescript/parser/jsx_mutants_products_test.go';src=(root/file).read_text();names=re.findall(r'^func (Test\w+)\(',src,re.M);live=[s for s in (ev/'list.log').read_text().splitlines() if s.startswith('Test')];assert len(names)==31 and not(set(names)-set(live));groups={'TestJsxMutants family':[n for n in names if n.startswith('TestJsxMutants')],'TestProduct_JsxMutants family':[n for n in names if n.startswith('TestProduct_JsxMutants')]};(ev/'members.json').write_text(json.dumps(groups,indent=2));pat='^('+'|'.join(names)+')$';(ev/'matrix-test-names.json').write_text(json.dumps(names,indent=2));(ev/'scope.json').write_text(json.dumps({'commit':'ef819b8e03c26ee3c5ac7bc1d8f067b77d553640','nproc':5,'package_tests':len(live),'named_tests':names,'missing':[],'functions':re.findall(r'^func (\w+)\(',src,re.M)},indent=2))
env=os.environ.copy();env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u155/typescript';runs=[]
def run(id,pattern=pat,overlay=None,cache=None,vet=False):
 cmd=['timeout','120','go','vet','./stage1/typescript/parser/'] if vet else ['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run',pattern]
 if overlay:cmd.insert(4,'-overlay='+str(overlay))
 e=env.copy()
 if cache:e['ADAMIC_BUILD_CACHE_DIR']=cache
 start=time.monotonic()
 with (ev/(id+'.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=e,stdout=out,stderr=subprocess.STDOUT)
 runs.append({'id':id,'command':cmd,'environment':{k:v for k,v in e.items() if k in ['ADAMIC_TYPESCRIPT_SOURCE','ADAMIC_BUILD_CACHE_DIR']},'wall':time.monotonic()-start,'exit':r.returncode});(ev/'runs.json').write_text(json.dumps(runs,indent=2));print(id,r.returncode,runs[-1]['wall'],flush=True);return r.returncode
# All edits fixed before reading their outcomes.
def body(name):
 a=src.index('func '+name+'(');b=src.index('\n}\n',a)+2;return src[a:b]
plans=[('W1','if bytes.Equal(side.output, want) {','if false && bytes.Equal(side.output, want) {','weaken comparison',[]),('S1','os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(ir)), 0644)','os.WriteFile(filepath.Join(dir, "missing-program.c"), []byte(native.C(ir)), 0644)','construction artifact option change',[]),('S2','\t\t\tseen[id] = true\n','','construction drop statement',[])]
for id,name,ret,imports in [('P1','jsxMutantsRun','return',['bytes']),('P2','jsxMutantsUnion','return',[]),('P3','jsxMutantsOracle','return ""',[]),('P4','jsxMutantsLower','return ""',['github.com/system-inc/adamic/internal/load','github.com/system-inc/adamic/internal/lower']),('P5','jsxMutantsNative','return ""',['crypto/sha256'])]:
 old=body(name);new=old[:old.index('{')+1]+'\n\t'+ret+'\n}';plans.append((id,old,new,'empty entry probe',imports))
cat=[]
for id,old,new,kind,imports in plans:
 assert src.count(old)==1,(id,src.count(old));changed=src.replace(old,new,1)
 for imp in imports:changed=changed.replace('\t"'+imp+'"\n','')
 target=ev/(id+'.go.txt');target.write_text(changed);(ev/(id+'.overlay.json')).write_text(json.dumps({'Replace':{str(root/file):str(target)}}));(ev/(id+'.diff')).write_text(''.join(difflib.unified_diff(src.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)));cat.append({'id':id,'file':file,'line':src[:src.index(old)].count('\n')+1,'before':old,'after':new,'kind':kind,'removed_unused_imports':imports})
(ev/'catalog.json').write_text(json.dumps(cat,indent=2))
for rep in [1,2,3]:
 for i,(g,mm) in enumerate(groups.items()):assert run(f'timing-{i}-{rep}',pattern='^('+'|'.join(mm)+')$')==0
for id,old,new,kind,imports in plans:
 overlay=ev/(id+'.overlay.json');assert run(id+'-vet',overlay=overlay,vet=True)==0
 run(id,overlay=overlay,cache='/tmp/u155/cache/S1' if id=='S1' else None)
 subprocess.run(['git','apply','--check',str(ev/(id+'.diff'))],cwd=root,check=True)
