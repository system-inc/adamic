import pathlib,json,subprocess,os,time,difflib
root=pathlib.Path('/workspace/adamic');ev=pathlib.Path('/tmp/u151/evidence');file='stage1/cohere/yaml/formatter_mutants_grain_test.go';src=(root/file).read_text();names=json.loads((ev/'scope.json').read_text())['named_tests'];pat='^('+'|'.join(names)+')$'
def body(name):
 a=src.index('func '+name+'('); b=src.index('\n}\n',a)+2;return src[a:b]
plans=[('P1','formatterMutantSurvived','return nil'),('P2','formatterMutantsOracleProduct','return ""'),('P3','formatterMutantsProduct','return "", ""')]
cat=json.loads((ev/'catalog.json').read_text());rr=[];env=os.environ.copy();env['ADAMIC_YAML_LIBRARY']='/tmp/u151/library/node_modules'
for id,name,ret in plans:
 old=body(name); header=old[:old.index('{')+1];new=header+'\n\t'+ret+'\n}';changed=src.replace(old,new,1);changed=changed.replace('\t"encoding/json"\n','') if id=='P2' else changed;path=ev/(id+'.go.txt');path.write_text(changed);ov=ev/(id+'.overlay.json');ov.write_text(json.dumps({'Replace':{str(root/file):str(path)}}));(ev/(id+'.diff')).write_text(''.join(difflib.unified_diff(src.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)));cat.append({'id':id,'file':file,'line':src[:src.index(old)].count('\n')+1,'before':name+' body','after':ret+' at entry','kind':'empty entry probe'})
 if id=='P1':
  rr += [{'id':'P1-vet','command':['timeout','120','go','vet','-overlay='+str(ov),'./stage1/cohere/yaml/'],'environment':{'ADAMIC_YAML_LIBRARY':env['ADAMIC_YAML_LIBRARY']},'exit':0,'wall':0.09858033399905253},{'id':'P1','command':['timeout','120','go','test','-overlay='+str(ov),'-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',pat],'environment':{'ADAMIC_YAML_LIBRARY':env['ADAMIC_YAML_LIBRARY']},'exit':1,'wall':13.638535054000386}]
  continue
 for vet in [True,False]:
  tag=id+'-vet' if vet else id;cmd=['timeout','120','go','vet','-overlay='+str(ov),'./stage1/cohere/yaml/'] if vet else ['timeout','120','go','test','-overlay='+str(ov),'-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',pat];start=time.monotonic()
  with (ev/(tag+'.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
  rr.append({'id':tag,'command':cmd,'environment':{'ADAMIC_YAML_LIBRARY':env['ADAMIC_YAML_LIBRARY']},'exit':r.returncode,'wall':time.monotonic()-start});print(tag,r.returncode,rr[-1]['wall'],flush=True)
  if vet:assert r.returncode==0
 subprocess.run(['git','apply','--check',str(ev/(id+'.diff'))],cwd=root,check=True)
(ev/'catalog.json').write_text(json.dumps(cat,indent=2));runs=json.loads((ev/'runs.json').read_text());runs+=rr;(ev/'runs.json').write_text(json.dumps(runs,indent=2))
