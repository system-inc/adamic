import pathlib,json,difflib,subprocess,os,time
p=pathlib.Path('review/test-audit/stage1-cohere-css-top_level_shards');f='stage1/cohere/css/parser_shards_test.go';source=pathlib.Path(f).read_text();env=os.environ.copy();env.update(ADAMIC_CSS_FIXTURES='/tmp/u080-css-fixtures',ADAMIC_CSS_LIBRARY='/tmp/u080-css-library');target=str(pathlib.Path(f).resolve())
for id in ['S01','P02']:
 if id=='S01':new=source.replace('for variant := -1; variant < len(mutants); variant++ {','for variant := -1; variant < len(mutants)-1; variant++ {',1)
 else:new=source.replace('func cssParserShards(keys []string) []cssParserShard {','func cssParserShards(keys []string) []cssParserShard {\n\tif os.Getenv("ADAMIC_AUDIT_PROBE") == "P02" { return nil }',1);env['ADAMIC_AUDIT_PROBE']='P02'
 path=pathlib.Path('/tmp/u080-'+id+'.go');path.write_text(new);overlay=pathlib.Path('/tmp/u080-'+id+'-overlay.json');overlay.write_text(json.dumps({'Replace':{target:str(path)}}));(p/(id+'.diff')).write_text(''.join(difflib.unified_diff(source.splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 cmd=['timeout','120','go','test','-overlay='+str(overlay),'-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^TestThePortParsesAsGoCohereDoes_Setup$'];start=time.monotonic()
 with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 (p/(id+'-run.json')).write_text(json.dumps({'command':'ADAMIC_CSS_FIXTURES=/tmp/u080-css-fixtures ADAMIC_CSS_LIBRARY=/tmp/u080-css-library ADAMIC_AUDIT_PROBE='+env.get('ADAMIC_AUDIT_PROBE','')+' '+' '.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start},indent=2))
 with (p/(id+'-vet.log')).open('w') as log:r=subprocess.run(['timeout','90','go','vet','-overlay='+str(overlay),'./stage1/cohere/css/'],env=env,stdout=log,stderr=subprocess.STDOUT)
 print(id,'vet',r.returncode,flush=True)
