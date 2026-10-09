import pathlib,json,subprocess,os,time
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-lint-harness';plan=json.loads((out/'plan.json').read_text());records=[]
for m in plan:
 if m['id'] not in ['M1','M2','M3']:continue
 path=root/m['file'];s=path.read_text();path.write_text(s.replace(m['old'],m['new'],1))
 try:
  env=os.environ.copy();env['ADAMIC_NATIVE_SPLIT']='1';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u109/cache/'+m['id']
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run','^TestJsxLintTreesSetupIsolation$']
  start=time.monotonic()
  with (out/(m['id']+'-isolation.log')).open('w') as f:p=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  records.append(dict(id=m['id'],exit=p.returncode,wall=time.monotonic()-start,command=cmd));(out/'isolation-results.json').write_text(json.dumps(records,indent=2)+'\n');print(m['id'],p.returncode,records[-1]['wall'],flush=True)
 finally:path.write_text(s)
