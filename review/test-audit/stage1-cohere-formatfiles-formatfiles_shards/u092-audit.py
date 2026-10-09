import pathlib,json,subprocess,os,time,difflib,re
p=pathlib.Path('review/test-audit/stage1-cohere-formatfiles-formatfiles_shards');(p/'diffs').mkdir(exist_ok=True);(p/'probes').mkdir(exist_ok=True);scratch=pathlib.Path('/tmp/u092');scratch.mkdir(exist_ok=True);selector=scratch/'mutant';selector.write_text('M00')
files=['stage1/cohere/formatfiles/'+f for f in ['enumerate.ts','disk.ts','golang.ts','main.ts']];original={f:pathlib.Path(f).read_text() for f in files};plan=[]
def add(id,file,old,new,kind):
 file='stage1/cohere/formatfiles/'+file;s=original[file];assert s.count(old)==1,(id,s.count(old));plan.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind))
add('M01','enumerate.ts','enumeration.walked++;','enumeration.walked += 2;','off-by-one count')
add('M02','disk.ts',"return status.kind === 'Ok' && status.type === 'directory';","return status.kind === 'Ok' && status.type === 'file';",'change constant')
add('M03','golang.ts',"character === 'İ' ? 'i' : character.toLowerCase()","character === 'İ' ? 'j' : character.toLowerCase()",'change constant')
add('M04','main.ts',"const hexDigits = '0123456789abcdef';","const hexDigits = '0123456789ABCDEF';",'change constant')
(p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
# Declared before inspecting outcomes: inventory source declarations and direct imports.
inv=[]
for file in files+['stage1/cohere/gitignore/'+x+'.ts' for x in ['path','glob','gitignore']]:
 s=pathlib.Path(file).read_text()
 for i,line in enumerate(s.splitlines(),1):
  if re.match(r'(?:export )?function |\s*(?:constructor|[a-zA-Z_][a-zA-Z0-9_]*)\([^;]*\)\s*:\s*[^;]*\{',line):inv.append(dict(file=file,line=i,declaration=line.strip()))
(p/'source-function-inventory.json').write_text(json.dumps(inv,indent=2));(p/'code-under-test.txt').write_text('Port entry: main.ts module execution. Formatfiles functions in main.ts (escapeOf, needsEscape, quote, unescapeOf, unescape, field, sortedKeys); golang.ts (goToLower, compareCodePoints, ext, join, rel); disk.ts (lstat, statExists, statIsDirectory, followedEntry, DiskTree constructor/load); enumerate.ts (Enumeration constructor, covers, walkTree, walk, walkDirectories, walkDir, parentOf, hasOwnRepository, repositoryBoundaryBetween, nestedRepositoryContaining, scopeOf, nestedRepositoriesBelow, enumerate; visit/count/handles callbacks). Imported gitignore path/glob/matcher function declarations are in source-function-inventory.json. This is a conservative transitive inventory; V8 coverage records observed entry counts separately. Witness code under test: firstDifference, suite-owned comparison; only it may be weakened. Oracle: Go cohere formatfiles with same on-disk corpus; witness self-written caught-owner expectation.\n')
def diff(q,after,folder):
 before=original[q['file']];(p/folder/(q['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+q['file'],tofile='b/'+q['file'])))
def run(id,regex='.',extra=None):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/formatfiles/','-run',regex];start=time.monotonic();env=dict(os.environ,**(extra or {}))
 with (p/'logs'/(id+'.log')).open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 es=[]
 for l in (p/'logs'/(id+'.log')).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return dict(id=id,command=' '.join(cmd),environment=(extra or {}),exit=r.returncode,wall_seconds=time.monotonic()-start,events=es,kills=sorted(set(e['Test'].split('/')[0] for e in es if e['Action']=='fail' and e.get('Test'))),cooked=any('test timed out' in e.get('Output','') for e in es))
results=[]
try:
 for q in plan:
  after=original[q['file']].replace(q['old'],q['new']);diff(q,after,'diffs');pathlib.Path(q['file']).write_text(after)
  v=run('validate-'+q['id'],'^TestProduct_FormatfilesNative$',dict(ADAMIC_BUILD_CACHE_DIR='/tmp/u092/cache/validate-'+q['id']));assert v['exit']==0,'standalone build failed';q['validation']=dict(command=v['command'],environment=v['environment'],wall_seconds=v['wall_seconds'],exit=v['exit']);pathlib.Path(q['file']).write_text(original[q['file']]);print('validated',q['id'],v['wall_seconds'],flush=True)
 (p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
 main='stage1/cohere/formatfiles/main.ts';entry=original[main].index('const casesPath = programArguments()[0]');probe=dict(id='P01',file=main,line=original[main][:entry].count('\n')+1,old=original[main][entry:],new='',kind='empty module entry: drop entire executable driver')
 diff(probe,original[main][:entry],'probes');pathlib.Path(main).write_text(original[main][:entry]);v=run('validate-P01','^TestProduct_FormatfilesNative$',dict(ADAMIC_BUILD_CACHE_DIR='/tmp/u092/cache/validate-P01'));assert v['exit']==0;probe['validation']=dict(command=v['command'],environment=v['environment'],wall_seconds=v['wall_seconds']);(p/'probe-plan.json').write_text(json.dumps(probe,indent=2));pathlib.Path(main).write_text(original[main])
 # Selector scaffolding only, absent from standalone diffs; preserves existing built-in mutation sites.
 switched=dict(original);g='stage1/cohere/formatfiles/golang.ts';helper="\nimport { readTextFile } from 'adamic';\nconst auditRead = readTextFile('/tmp/u092/mutant');\nconst auditId = auditRead.kind === 'Ok' ? auditRead.text.trim() : '';\nexport function auditSelected(id: string): boolean { return auditId === id; }\n"
 switched[g]=switched[g].replace("import { clean } from '../gitignore/path.ts';","import { clean } from '../gitignore/path.ts';"+helper)
 switched['stage1/cohere/formatfiles/enumerate.ts']=switched['stage1/cohere/formatfiles/enumerate.ts'].replace('import { ext, goToLower, join, rel }','import { auditSelected, ext, goToLower, join, rel }')
 switched['stage1/cohere/formatfiles/disk.ts']=switched['stage1/cohere/formatfiles/disk.ts'].replace("import { fileStatus, readDirectory, readTextFile } from 'adamic';","import { fileStatus, readDirectory, readTextFile } from 'adamic';\nimport { auditSelected } from './golang.ts';")
 switched[main]=switched[main].replace('import { compareCodePoints, ext, goToLower }','import { auditSelected, compareCodePoints, ext, goToLower }')
 replacements={'M01':"enumeration.walked += auditSelected('M01') ? 2 : 1;",'M02':"return status.kind === 'Ok' && status.type === (auditSelected('M02') ? 'file' : 'directory');",'M03':"character === 'İ' ? (auditSelected('M03') ? 'j' : 'i') : character.toLowerCase()",'M04':"const hexDigits = auditSelected('M04') ? '0123456789ABCDEF' : '0123456789abcdef';"}
 for q in plan:switched[q['file']]=switched[q['file']].replace(q['old'],replacements[q['id']])
 entry=switched[main].index('const casesPath = programArguments()[0]');switched[main]=switched[main][:entry]+"if (!auditSelected('P01')) {\n"+switched[main][entry:]+"\n}\n"
 for f,s in switched.items():pathlib.Path(f).write_text(s)
 cache=dict(ADAMIC_BUILD_CACHE_DIR='/tmp/u092/cache/switch')
 baseline=run('switch-baseline','.',cache);(p/'switch-baseline.json').write_text(json.dumps(baseline,indent=2));assert baseline['exit']==0,'red switched baseline'
 for q in plan:
  selector.write_text(q['id']);r=run(q['id'],'.',cache);results.append(r);(p/'matrix.json').write_text(json.dumps(results,indent=2));print(q['id'],r['exit'],r['kills'],flush=True)
  if r['cooked']:raise Exception('matrix cooked: narrow this mutant')
 selector.write_text('P01');r=run('P01','.',cache);assert not r['cooked'];(p/'probe-result.json').write_text(json.dumps(r,indent=2));print('P01',r['kills'],flush=True)
finally:
 for f,s in original.items():pathlib.Path(f).write_text(s)
 selector.write_text('M00');(p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
