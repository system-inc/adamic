import pathlib,json,subprocess,time,os
p=pathlib.Path('review/test-defend/stage1-cohere-yaml-scalars_match_go_shards'); menu=json.loads((p/'menu.json').read_text());runs=[]
for m in menu[1:]:
 file=pathlib.Path(m['file']);original=file.read_text();assert original.count(m['old'])==1
 subprocess.run(['git','apply','--check',str(p/'diffs'/(m['id']+'.diff'))],check=True)
 file.write_text(original.replace(m['old'],m['new'],1));env=os.environ.copy();env.update(ADAMIC_YAML_LIBRARY='/tmp/yaml-defend/library',ADAMIC_BUILD_CACHE_DIR='/tmp/yaml-defend/cache/'+m['id'])
 selections=[('target','^'+m['target']+'$'),('controls','^Test(SpeedCostProbes|ScalarsMatchGo(Union|_[0-9]+))$'),('formatter','^TestFormatterMatchesGo$'),('file-driver','^TestFileDriver(Union|_[0-9]+|_Setup)$')]
 if m['id']=='D2':selections.append(('unist-witness','^TestUnistMutants$'))
 for i in range(6):selections.append(('formatter-witness-'+str(i),'^Test(FormatterMutants_'+f'{i:03}'+ '|Product_YAMLFormatterMutant'+f'{i:03}'+'(Sources|Lowered|Native))$'))
 try:
  for label,regex in selections:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',regex];start=time.monotonic();log=m['id']+'-'+label+'.log'
   with (p/log).open('w')as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
   runs.append(dict(mutant=m['id'],selection=label,command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],exit=r.returncode,seconds=time.monotonic()-start,log=log));(p/'matrix-runs.json').write_text(json.dumps(runs,indent=2));print(runs[-1],flush=True)
 finally:file.write_text(original)
