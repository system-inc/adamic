import subprocess,pathlib,json,re
root=pathlib.Path('/workspace/adamic');out=root/'review/test-defend/deletion-set/stage1-typescript-parser';out.mkdir(parents=True,exist_ok=True)
candidates={'TestJsxMemberNameRejection','TestWholeCompilerAgrees'};gathered=[]
for typ in ['test-audit','test-defend']:
 for unit in ['jsx_rejection','whole_mutants_split']:
  branch=typ+'/stage1-typescript-parser-'+unit;prefix='review/'+branch+'/';tag=('audit' if typ=='test-audit' else 'defend')+'-'+unit
  files=subprocess.check_output(['git','ls-tree','-r','--name-only','origin/'+branch,prefix],cwd=root,text=True).splitlines()
  matrix=json.loads(subprocess.check_output(['git','show','origin/'+branch+':'+prefix+'matrix.json'],cwd=root,text=True));fails={}
  if isinstance(matrix,dict):
   for mid,rows in matrix.items():fails[mid]={r for r,v in rows.items() if v.get('exit')==1}
  else:
   for entry in matrix:
    mid=entry.get('id',entry.get('mutant'));fails.setdefault(mid,set()).update(entry.get('fails',entry.get('failed',entry.get('rows_failed',[]))))
  (out/(tag+'-original-matrix.json')).write_text(json.dumps(matrix,indent=2))
  for file in files:
   mid=pathlib.PurePosixPath(file).stem
   if pathlib.PurePosixPath(file).parent.as_posix() not in [prefix.rstrip('/'),prefix+'diffs']:continue
   if not re.fullmatch('[MD][0-9]+',mid) or not file.endswith('.diff') or not(fails.get(mid,set())&candidates):continue
   diff=subprocess.check_output(['git','show','origin/'+branch+':'+file],cwd=root,text=True);saved=tag+'-'+mid+'.diff';(out/saved).write_text(diff)
   match=re.search(r'^\+\+\+ b/(.*)\n@@ -([0-9]+)',diff,re.M)
   check=subprocess.run(['git','apply','--check',str(out/saved)],cwd=root,capture_output=True,text=True)
   gathered.append(dict(mutant=tag+'-'+mid,original_mutant=mid,file_line=match.group(1)+':'+match.group(2) if match else 'see diff',branch=branch,diff=saved,candidates_failed=sorted(fails[mid]&candidates),other_rows_failed=sorted(fails[mid]-candidates),stale=check.returncode!=0,apply_error=check.stderr))
(out/'mutant-list.json').write_text(json.dumps(gathered,indent=2));print(json.dumps(gathered,indent=2))
