import pathlib,os,subprocess,time,json
p=pathlib.Path('/tmp/u152');f=pathlib.Path('/workspace/adamic/stage1/cohere/yaml/lexer.ts');old=f.read_text();records=[];env=os.environ.copy();env['ADAMIC_YAML_LIBRARY']=str(p/'library');env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/'M4-other')
try:
 f.write_text(old.replace('return unit === 44 || unit === 91','return unit === 45 || unit === 91',1))
 for row,pat in [('TestLexerGaps','TestLexerGaps'),('TestStructuralPositionRefusal','TestStructuralPositionRefusal'),('TestClosed family','TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap'),('TestSharedSliceAppendMatchesNode','TestSharedSliceAppendMatchesNode')]:
  id='M4-other-'+row.replace(' ','_');cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^('+pat+')$'];start=time.monotonic()
  with (p/(id+'.log')).open('w') as out:r=subprocess.run(cmd,env=env,cwd='/workspace/adamic',stdout=out,stderr=subprocess.STDOUT)
  rec={'id':id,'test':row,'command':' '.join(cmd),'wall':round(time.monotonic()-start,3),'exit':r.returncode};records.append(rec);(p/'m4-other-runs.json').write_text(json.dumps(records,indent=2));print(rec,flush=True)
finally:f.write_text(old)
