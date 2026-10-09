import pathlib,json,subprocess,os,time
p=pathlib.Path('/tmp/defend-mdast')
for m in json.loads(p.joinpath('plan.json').read_text()):
 if m['id'] not in ['D1','D4']:continue
 mid=m['id'];f=pathlib.Path(m['file']);s=f.read_text();assert s.count(m['before'])==1;f.write_text(s.replace(m['before'],m['after']))
 run='^(TestMdastIdentifierWitnesses|TestNativeMdastConstruction|TestMdastMalformedEvents.*|TestProduct_MarkdownMalformedEvents.*|TestMdastIdentifierScalars|TestProduct_MarkdownIdentifier.*)$'
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',run]
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/(mid+'-expanded'));start=time.monotonic()
 try:
  with p.joinpath(mid+'-expanded.log').open('w') as log:rc=subprocess.call(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 finally:f.write_text(s)
 p.joinpath(mid+'-expanded-run.json').write_text(json.dumps(dict(command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],exit=rc,wall=time.monotonic()-start),indent=2));print(mid,rc,round(time.monotonic()-start,3),flush=True)
