import pathlib,json,difflib,subprocess,time
root=pathlib.Path('/workspace/adamic'); ev=root/'review/test-audit/stage1-cohere-lint-registry'; path=root/'stage1/cohere/lint/registry/registry.go'; original=(ev/'original.go.txt').read_text(); records=[]
try:
 for m in json.loads((ev/'menu.json').read_text()):
  changed=original.replace(m['old'],m['new'])
  if m['id']=='P2': changed=changed.replace('\t"go/format"\n','')
  diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/stage1/cohere/lint/registry/registry.go',tofile='b/stage1/cohere/lint/registry/registry.go')); (ev/(m['id']+'.diff')).write_text(diff); path.write_text(original)
  apply=subprocess.run(['git','apply','--check',str(ev/(m['id']+'.diff'))],cwd=root,capture_output=True); assert apply.returncode==0
  path.write_text(changed); start=time.monotonic()
  with (ev/(m['id']+'-vet.log')).open('w') as log:r=subprocess.run(['go','vet','./stage1/cohere/lint/registry/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  records.append(dict(id=m['id'],apply_exit=apply.returncode,vet_exit=r.returncode,seconds=time.monotonic()-start)); assert r.returncode==0,m['id']
finally:path.write_text(original)
(ev/'validation.json').write_text(json.dumps(records,indent=2))
