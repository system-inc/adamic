import subprocess,pathlib,json,re
R=pathlib.Path('review/test-defend/deletion-set/internal-fuzz');C=['TestGeneratedProgramsCheckAndLower','TestOneSeedOneProgram','TestRegexProgramsPassTheChecker'];items=[];allfiles=[]
for branch in ['test-audit/internal-fuzz','test-defend/internal-fuzz']:
 prefix='review/'+branch;paths=subprocess.check_output(['git','ls-tree','-r','--name-only','origin/'+branch,'--',prefix]).decode().splitlines()
 for path in paths:
  name=pathlib.PurePosixPath(path).name
  standard=re.match(r'M.*\.diff$',name) if branch.startswith('test-audit') else re.match(r'D.*\.diff$',name)
  supplemental=branch.startswith('test-defend') and '/seed-checker-defense/' in path and re.match(r'[GRS]\d+\.diff$',name)
  if not (standard or supplemental):continue
  mid=pathlib.PurePosixPath(path).stem;parent=pathlib.PurePosixPath(path).parent;matrix=str(parent/'matrix.json')
  data=json.loads(subprocess.check_output(['git','show','origin/'+branch+':'+matrix]));failed=[]
  if isinstance(data,dict):failed=[t for t,v in data.get(mid,{}).items() if v=='fail']
  else:
   for row in data:
    if row.get('mutant')==mid:failed=row.get('rows_failed',[])
  candidates=sorted(set(C)&set(failed));allfiles.append(dict(branch=branch,file=path,failed=failed,selected=bool(candidates)))
  if not candidates:continue
  blob=subprocess.check_output(['git','show','origin/'+branch+':'+path]);rid='R%03d'%(len(items)+1);local=R/'diffs'/(rid+'-'+mid+'.diff');local.write_bytes(blob)
  file=None;line=0;site=None
  for s in blob.decode().splitlines():
   if s.startswith('--- a/'):file=s[6:]
   elif s.startswith('@@'):line=int(re.search(r'@@ -(\d+)',s)[1])
   elif s.startswith('-') and not s.startswith('---'):
    if site is None:site=f'{file}:{line}'
    line+=1
   elif s.startswith(' '):line+=1
  items.append(dict(replay=rid,mutant=mid,file_line=site,branch=branch,source_diff=path,diff=str(local),candidates_failed=candidates,other_rows_failed=sorted(set(failed)-set(C)),supplemental_filename=bool(supplemental)))
(R/'mutant-list.json').write_text(json.dumps(items,indent=2)+'\n');(R/'selection-inventory.json').write_text(json.dumps(allfiles,indent=2)+'\n');(R/'skip.json').write_text(json.dumps(C,indent=2)+'\n');(R/'main.txt').write_text(subprocess.check_output(['git','rev-parse','HEAD']).decode())
for e in items:print(json.dumps(e))
