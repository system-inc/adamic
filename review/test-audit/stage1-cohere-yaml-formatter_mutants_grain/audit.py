import pathlib,json,re,subprocess,os,time,difflib
repo=pathlib.Path('/workspace/adamic');ev=pathlib.Path('/tmp/u151/evidence');file='stage1/cohere/yaml/formatter_mutants_grain_test.go';src=(repo/file).read_text();names=re.findall(r'^func (Test\w+)\(',src,re.M);live=[x for x in (ev/'list.log').read_text().splitlines() if x.startswith('Test')];assert not(set(names)-set(live))
(ev/'scope.json').write_text(json.dumps({'commit':'a7448d73cd17f16362b6cbc5c5c111080da64e43','nproc':5,'named_tests':names,'missing':[],'code_under_test_functions':re.findall(r'^func (\w+)\(',src,re.M)},indent=2))
groups={'TestProduct_YAMLFormatterMutantsOracle':['TestProduct_YAMLFormatterMutantsOracle'],'TestProduct_YAMLFormatterMutant family':[n for n in names if re.fullmatch(r'TestProduct_YAMLFormatterMutant\d+(Sources|Lowered|Native)',n)],'TestFormatterMutants family':[n for n in names if re.fullmatch(r'TestFormatterMutants_\d+',n)],'TestFormatterMutantsPlantedFailure':['TestFormatterMutantsPlantedFailure']};(ev/'members.json').write_text(json.dumps(groups,indent=2));pattern='^('+'|'.join(names)+')$';(ev/'matrix-test-names.json').write_text(json.dumps(names,indent=2))
baseenv=os.environ.copy();baseenv['ADAMIC_YAML_LIBRARY']='/tmp/u151/library/node_modules';runs=json.loads((ev/'runs.json').read_text()) if os.environ.get('AUDIT_SKIP_TIMING')=='1' else []
def run(id,pat=pattern,overlay=None,cache=None,vet=False):
 cmd=['timeout','120','go','vet','./stage1/cohere/yaml/'] if vet else ['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',pat]
 if overlay:cmd.insert(4,'-overlay='+str(overlay))
 env=baseenv.copy()
 if cache:env['ADAMIC_BUILD_CACHE_DIR']=cache
 start=time.monotonic()
 with (ev/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT)
 runs.append({'id':id,'command':cmd,'environment':{k:v for k,v in env.items() if k in ['ADAMIC_YAML_LIBRARY','ADAMIC_BUILD_CACHE_DIR']},'wall':time.monotonic()-start,'exit':r.returncode});(ev/'runs.json').write_text(json.dumps(runs,indent=2));print(id,r.returncode,runs[-1]['wall'],flush=True);return r.returncode
# Before checking any failure: fixed list from construction/check functions.
old='func formatterMutantSurvived(actual, expected []byte) error {\n\tif bytes.Equal(actual, expected) {\n\t\treturn fmt.Errorf("missed mutant")\n\t}\n\treturn nil\n}'
assert old in src
# Keep bytes import used by command output buffers; drop complete check statement.
plans=[('W1',old,'func formatterMutantSurvived(actual, expected []byte) error {\n\treturn nil\n}','weakened comparison'),('S1','os.WriteFile(filepath.Join(directory, "cases.txt"), data, 0644)','os.WriteFile(filepath.Join(directory, "missing-cases.txt"), data, 0644)','construction artifact option change'),('S2','os.WriteFile(filepath.Join(directory, file), source, 0644)','os.WriteFile(filepath.Join(directory, file+".missing"), source, 0644)','construction artifact option change')]
cat=[]
for id,old,new,kind in plans:
 assert src.count(old)==1
 changed=src.replace(old,new,1);target=ev/(id+'.go.txt');target.write_text(changed);overlay=ev/(id+'.overlay.json');overlay.write_text(json.dumps({'Replace':{str(repo/file):str(target)}}));(ev/(id+'.diff')).write_text(''.join(difflib.unified_diff(src.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)));cat.append({'id':id,'file':file,'line':src[:src.index(old)].count('\n')+1,'before':old,'after':new,'kind':kind})
(ev/'catalog.json').write_text(json.dumps(cat,indent=2))
if os.environ.get('AUDIT_SKIP_TIMING') != '1':
 for i in [1,2,3]:
  for g,nn in groups.items():
   tag=list(groups).index(g);assert run(f'timing-{tag}-{i}',pat='^('+'|'.join(nn)+')$')==0
for id,old,new,kind in plans:
 overlay=ev/(id+'.overlay.json');assert run(id+'-vet',overlay=overlay,vet=True)==0
 run(id,overlay=overlay,cache='/tmp/u151/cache/'+id if id!='W1' else None)
for c in cat:subprocess.run(['git','apply','--check',str(ev/(c['id']+'.diff'))],cwd=repo,check=True)
