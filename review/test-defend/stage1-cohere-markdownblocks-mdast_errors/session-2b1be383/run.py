import pathlib,json,subprocess,os,time
p=pathlib.Path('/tmp/defend-mdast');plans=json.loads(p.joinpath('plan.json').read_text())
base='TestMdastIdentifierWitnesses|TestNativeMdastConstruction|TestMdastMalformedEvents.*|TestProduct_MarkdownMalformedEvents.*'
for m in plans:
 mid=m['id'];f=pathlib.Path(m['file']);s=f.read_text();assert s.count(m['before'])==1
 f.write_text(s.replace(m['before'],m['after']))
 run='^('+base+('|TestMdastIdentifierScalars|TestProduct_MarkdownIdentifier.*' if mid in ['D2','D3'] else '')+')$'
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',run]
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/mid)
 start=time.monotonic()
 try:
  with p.joinpath(mid+'.log').open('w') as log:rc=subprocess.call(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 finally:f.write_text(s)
 p.joinpath(mid+'-run.json').write_text(json.dumps(dict(command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],exit=rc,wall=time.monotonic()-start),indent=2))
 print(mid,rc,round(time.monotonic()-start,3),flush=True)
 if rc==124:print('Cooked outer run; narrow before using unknown observations.',flush=True);break
