import json,pathlib,subprocess,os,time
p=pathlib.Path('review/test-defend/stage1-cohere-yaml-scalars_match_go_shards');runs=[]
for m in json.loads((p/'runtime-menu.json').read_text()):
 f=pathlib.Path(m['file']);s=f.read_text();assert s[m['offset']:].startswith(m['old']);subprocess.run(['git','apply','--check',str(p/'diffs'/(m['id']+'.diff'))],check=True);f.write_text(s[:m['offset']]+s[m['offset']:].replace(m['old'],m['new'],1))
 e=os.environ.copy();e.update(ADAMIC_YAML_LIBRARY='/tmp/yaml-defend/library',ADAMIC_BUILD_CACHE_DIR='/tmp/yaml-defend/cache/'+m['id']);regex='^Test(SpeedCostProbes|ScalarsMatchGo(Union|_[0-9]+))$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',regex];st=time.monotonic()
 try:
  with (p/(m['id']+'-runtime.log')).open('w')as out:r=subprocess.run(cmd,env=e,stdout=out,stderr=subprocess.STDOUT)
  runs.append(dict(mutant=m['id'],command=cmd,cache=e['ADAMIC_BUILD_CACHE_DIR'],exit=r.returncode,seconds=time.monotonic()-st));(p/'runtime-runs.json').write_text(json.dumps(runs,indent=2));print(runs[-1],flush=True)
 finally:f.write_text(s)
