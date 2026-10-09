import pathlib,json,subprocess,time,difflib
root=pathlib.Path('/workspace/adamic'); ev=root/'review/test-audit/stage1-cohere-lint-registry'; p=root/'stage1/cohere/lint/registry/registry.go'; switched=p.read_text(); original=(ev/'original.go.txt').read_text(); (ev/'switch.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),switched.splitlines(True),fromfile='a/'+str(p.relative_to(root)),tofile='b/'+str(p.relative_to(root))))); results=[]
try:
 for m in json.loads((ev/'menu.json').read_text()):
  p.write_text(original)
  apply=subprocess.run(['git','apply','--check',str(ev/(m['id']+'.diff'))],cwd=root,capture_output=True,text=True)
  assert apply.returncode==0,apply.stderr
  p.write_text(original.replace(m['old'],m['new']))
  start=time.monotonic()
  with (ev/(m['id']+'-vet.log')).open('w') as log: r=subprocess.run(['go','vet','./stage1/cohere/lint/registry/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  results.append(dict(id=m['id'],apply_exit=apply.returncode,vet_exit=r.returncode,seconds=time.monotonic()-start))
finally:p.write_text(original)
(ev/'validation.json').write_text(json.dumps(results,indent=2)); print(results)
