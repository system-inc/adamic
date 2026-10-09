import pathlib,json,subprocess,difflib,time
root=pathlib.Path('/workspace/adamic'); ev=root/'review/test-audit/stage1-cohere-lint-shards'; path=root/'stage1/cohere/lint/shards/shards.go'; original=(ev/'original.go.txt').read_text(); timings=[]
try:
 for mid in ['clean','C5','C6']:
  source=original
  if mid!='clean':
   m=next(x for x in json.loads((ev/'menu.json').read_text())if x['id']==mid);source=original.replace(m['old'],m['new'])
  path.write_text(source);start=time.monotonic()
  with (ev/('witness-'+mid+'.log')).open('w')as log:r=subprocess.run(['go','run','/tmp/u125/witness.go'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  timings.append(dict(id=mid,command='go run /tmp/u125/witness.go',exit=r.returncode,seconds=time.monotonic()-start));assert r.returncode==0
finally:path.write_text(original)
for mid in ['C5','C6']:(ev/('witness-'+mid+'.diff')).write_text(''.join(difflib.unified_diff((ev/'witness-clean.log').read_text().splitlines(True),(ev/('witness-'+mid+'.log')).read_text().splitlines(True),fromfile='clean',tofile=mid)))
(ev/'witness-commands.json').write_text(json.dumps(timings,indent=2))
