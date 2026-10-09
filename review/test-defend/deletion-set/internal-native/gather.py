import subprocess,pathlib,json,re,hashlib
ROOT=pathlib.Path('review/test-defend/deletion-set/internal-native')
CANDIDATES=['TestNumbersFormatExactlyAsJavaScriptDoes','TestParserHasNoUnusedOptionalMethodThunks','TestRecordBenchmark','TestRegExpLintPatternsNode','TestRegExpNativeStepLimit','TestRegExpSearchNode','TestRuntimeReleasePaths']
branches=[f'{kind}/internal-native-{unit}' for kind in ['test-audit','test-defend'] for unit in ['arguments_length','number','radix']]
cache={}
def read(branch,path):
 key=(branch,path)
 if key not in cache:cache[key]=subprocess.check_output(['git','show','origin/'+branch+':'+path])
 return cache[key]
def failnames(value):
 out=set()
 if isinstance(value,dict):
  for key in ['failed','failed_rows','rows_failed','failures','fail']:
   v=value.get(key)
   if isinstance(v,list):out.update(s for s in v if isinstance(s,str) and s.startswith('Test'))
   if isinstance(v,dict):out.update(s for s in v if s.startswith('Test'))
 return out
def records(value,id):
 out=set()
 if isinstance(value,dict):
  if id in value:
   v=value[id]
   if isinstance(v,list):out.update(s for s in v if isinstance(s,str) and s.startswith('Test'))
   else:out.update(failnames(v))
  if value.get('id',value.get('mutant'))==id:out.update(failnames(value))
  if value.get('test') and id in value.get('kills',[]):out.add(value['test'])
  for v in value.values():
   if isinstance(v,(list,dict)):out.update(records(v,id))
 elif isinstance(value,list):
  for v in value:out.update(records(v,id))
 return out
inventory=[];considered=[]
for branch in branches:
 prefix='review/'+branch
 paths=subprocess.check_output(['git','ls-tree','-r','--name-only','origin/'+branch,'--',prefix]).decode().splitlines()
 diffs=[s for s in paths if re.match(r'^[MD].*\.diff$',pathlib.PurePosixPath(s).name) and pathlib.PurePosixPath(s).name.startswith('M' if branch.startswith('test-audit') else 'D')]
 matrices=[s for s in paths if re.search(r'/(?:matrix|results|rows)\.json$',s)]
 for path in diffs:
  id=pathlib.PurePosixPath(path).stem;parent=pathlib.PurePosixPath(path).parent
  chosen=[]
  while str(parent).startswith(prefix):
   chosen=[s for s in matrices if pathlib.PurePosixPath(s).parent==parent]
   if chosen:break
   parent=parent.parent
  chosen=[s for s in chosen if not s.endswith("/rows.json")] or chosen
  failed=set()
  for m in chosen:failed.update(records(json.loads(read(branch,m)),id))
  candidates=sorted(set(CANDIDATES)&failed)
  considered.append(dict(branch=branch,path=path,matrices=chosen,failed=sorted(failed),selected=bool(candidates)))
  if not candidates:continue
  blob=read(branch,path);sha=hashlib.sha256(blob).hexdigest();key=f'R{len(inventory)+1:03d}'
  ROOT.joinpath('diffs').mkdir(exist_ok=True);dest=ROOT/'diffs'/f'{key}-{id}.diff';dest.write_bytes(blob)
  oldfile=None;line=0;sites=[]
  for s in blob.decode().splitlines():
   if s.startswith('--- a/'):oldfile=s[6:]
   elif s.startswith('@@'):
    match=re.search(r'@@ -(\d+)',s)
    if match:line=int(match.group(1))
   elif s.startswith('-') and not s.startswith('---'):
    if oldfile and f'{oldfile}:{line}' not in sites:sites.append(f'{oldfile}:{line}')
    line+=1
   elif s.startswith(' '):line+=1
  entry=dict(replay=key,mutant=id,file_line=sites[0] if sites else None,file_lines=sites,branch=branch,branch_sha=subprocess.check_output(['git','rev-parse','origin/'+branch]).decode().strip(),source_diff=path,diff=str(dest),sha256=sha,evidence_files=chosen,candidates_failed=candidates,other_rows_failed=sorted(failed-set(CANDIDATES)))
  inventory.append(entry)
ROOT.joinpath('mutant-list.json').write_text(json.dumps(inventory,indent=2)+'\n')
ROOT.joinpath('selection-inventory.json').write_text(json.dumps(considered,indent=2)+'\n')
ROOT.joinpath('main.txt').write_text(subprocess.check_output(['git','rev-parse','HEAD']).decode())
ROOT.joinpath('skip.json').write_text(json.dumps(CANDIDATES,indent=2)+'\n')
print('Gathered',len(inventory),'diffs from',len(considered),'eligible-name diff files')
for e in inventory:print(e['replay'],e['branch'],e['source_diff'],e['candidates_failed'],e['other_rows_failed'])
