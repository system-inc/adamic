import json,re,difflib,subprocess,gzip,shutil,datetime
from pathlib import Path
p=Path('/tmp/u098');root=Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-graphql-printer-grain_mutant';out.mkdir(parents=True,exist_ok=True)
assert (p/'post-done').exists();base=(p/'base.txt').read_text().strip();scope=json.loads((p/'scope.json').read_text());rows=scope['rows'];times={x['test']:x for x in json.loads((p/'timings.json').read_text())};subs=json.loads((p/'subsumer-timing.json').read_text())
def events(n):
 result=[]
 for line in (p/(n+'.log')).read_text().splitlines():
  try:result.append(json.loads(line))
  except:pass
 return result
# Reconstruct scratch files from the unified diff, then map unchanged error lines to base.
maps={};patch=(p/'scratch-switch.diff').read_text().splitlines(True);i=0
while i<len(patch):
 if not patch[i].startswith('diff --git '):i+=1;continue
 fname=patch[i].split(' b/',1)[1].strip();old=subprocess.run(['git','show',base+':'+fname],cwd=root,capture_output=True,text=True,check=True).stdout.splitlines(True);new=[];cursor=0;i+=1
 while i<len(patch) and not patch[i].startswith('diff --git '):
  if patch[i].startswith('@@ '):
   m=re.match(r'@@ -(\d+)(?:,\d+)? \+(\d+)',patch[i]);start=int(m.group(1))-1;new+=old[cursor:start];cursor=start;i+=1
   while i<len(patch) and not patch[i].startswith(('@@ ','diff --git ')):
    line=patch[i]
    if line.startswith(' '):new.append(old[cursor]);cursor+=1
    elif line.startswith('-'):cursor+=1
    elif line.startswith('+'):new.append(line[1:])
    elif not line.startswith('\\'):break
    i+=1
  else:i+=1
 new+=old[cursor:];mapping={}
 for block in difflib.SequenceMatcher(a=old,b=new,autojunk=False).get_matching_blocks():
  for k in range(block.size):mapping[block.b+k+1]=block.a+k+1
 maps[Path(fname).name]=mapping

def mapped(s):
 return re.sub(r'([A-Za-z0-9_]+_test.go):(\d+):',lambda m:m.group(1)+':'+str(maps.get(m.group(1),{}).get(int(m.group(2)),int(m.group(2))))+':',s)
def errors(n,members):return [mapped(x.get('Output','').strip()) for x in events(n) if x.get('Test') in members and x.get('OutputType')=='error']
def command(n):
 meta=json.loads((p/(n+'.meta')).read_text());args=meta['command'];s=' '.join(args[:-1])+" '"+args[-1]+"'"
 prefix=[]
 if meta.get('selector'):prefix.append('write '+meta['selector']+' to /tmp/u098/selector')
 if meta.get('mode'):prefix.append('ADAMIC_AUDIT_MODE='+meta['mode'])
 return '; '.join(prefix+[s])

extra_groups=[dict(test='TestPrinterAsGoCohere family',members=[f'TestPrinterAsGoCohere_{i:03d}' for i in range(4)],kind='production'),dict(test='TestPrinterFileDriver',members=['TestPrinterFileDriver'],kind='production'),dict(test='TestPrinterShardPlantedDisagreement',members=['TestPrinterShardPlantedDisagreement'],kind='witness'),dict(test='TestPrinterWhitespaceGap family',members=[f'TestPrinterWhitespaceGap_{i:03d}' for i in range(5)],kind='production')]
allrows=rows+extra_groups;matrix=[]
for mid in ['M1','M2','M3','M4','P1']:
 es=events(mid);statuses={x['Test']:x['Action'] for x in es if x.get('Test') and x.get('Action') in ['pass','fail','skip']};assert set(scope['matrix_members'])<=set(statuses),(mid,'unknown row')
 assert not any('panic:' in x.get('Output','') for x in es)
 rawfailed=[n for n,s in statuses.items() if s=='fail'];failedrows=[r['test'] for r in allrows if any(statuses.get(n)=='fail' for n in r['members'])]
 qualifies=[r['test'] for r in allrows if r['kind']=='production' and r['test'] in failedrows]
 matrix.append(dict(id=mid,command=command(mid),statuses=statuses,failed_tests=rawfailed,failed_rows=failedrows,production_kill_rows=qualifies,evidence={r['test']:errors(mid,r['members']) for r in allrows if r['test'] in failedrows}))
(out/'matrix.json').write_text(json.dumps(matrix,indent=2));matrix_names=[r['test'] for r in allrows]
sourcefiles={}
for f in (root/'stage1/cohere/graphql/printer').glob('*_test.go'):
 for i,line in enumerate(f.read_text().splitlines(),1):
  m=re.match(r'func (Test\w+)\(',line)
  if m:sourcefiles[m.group(1)]=str(f.relative_to(root))+':'+str(i)
settings=[('W1',None,'witness','self','Unchanged Go cohere answers; requires built-in native and Node mutants to finish cleanly and differ. W1 removes difference detection. Two sides of the same port are products.',None),('S1','P7','setup-check','self','Self-authored mutant/case census; whole and shards share printerMutantCases. Empty enumeration passes.',None),('S5','P5','untrue','self','Preparation completion and nonempty path only; no executable existence or behavior assertion.','S5-artifacts.json'),('S6','P6','setup-check','external-run','Go cohere generates defaults cases and expected answers; this row checks preparation/readability, not port agreement. Empty entry passes.',None),('S2','P2','untrue','self','Source construction completion only; promised main.ts presence is not asserted.','S2-artifacts.json'),('S3','P3','untrue','self','Lowering construction completion only; promised port.c and backend presence are not asserted.','S3-artifacts.json'),('S4','P4','untrue','self','Sanitized compilation construction completion only; promised executable presence is not asserted.','S4-artifacts.json'),('W2',None,'witness','self','Self-authored planted survivor IDs; unchanged answer must be rejected only by its owning shard.',None),('M4','P1','subsumed','external-run','Go cohere expected bytes, plus actual Go cohere and Prettier 3.9.6 execution. Native and source Node outputs must match. Logs throughput; has no speed threshold.',None)]
report=[]
for r,(proof,probe,verdict,kind,oracle,artifact) in zip(rows,settings):
 errs=errors(proof,r['members']);line=errs[0] if errs else None
 evidence=command(proof)+' => '+(line or 'PASS')
 if artifact:
  absent=json.loads((p/artifact).read_text());assert absent==[];evidence+=' despite absent promised artifacts; '+artifact+'=[]'
 kills=['M1','M2','M3','M4'] if r['kind']=='production' else []
 x=dict(test=r['test'],package='stage1/cohere/graphql/printer',file=sourcefiles[r['members'][0]],seconds=times[r['test']]['median'],oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=[],last_proven_fail=(proof+': '+line if line else None),verdict=verdict,subsumed_by=[subs['test']] if kills else [],mutants_in_matrix=4,probe_kills=['P1'] if kills else [],subsumer_seconds=subs['median'] if kills else None,vacuous=(False if kills else True if probe else None),bounded=True,matrix_rows=matrix_names,evidence=evidence,members=r['members'],member_files={n:sourcefiles[n] for n in r['members']})
 if probe:x['entry_probe']=dict(id=probe,command=command(probe),result='fail' if errors(probe,r['members']) else 'pass')
 report.append(x)
(out/'rows.json').write_text(json.dumps(report,indent=2));(out/'survivors.json').write_text('[]\n')
experiments=[]
for r,(proof,probe,verdict,kind,oracle,artifact) in zip(rows,settings):
 if proof!='M4':experiments.append(dict(id=proof,type='witness weakening' if proof.startswith('W') else 'construction mutation',rows=r['members'],failing_lines=errors(proof,r['members']),command=command(proof),diff=proof+'.diff'))
 if probe:experiments.append(dict(id=probe,type='empty-answer probe',rows=r['members'],failing_lines=errors(probe,r['members']),command=command(probe),diff=probe+'.diff'))
(out/'experiments.json').write_text(json.dumps(experiments,indent=2))
for f in p.iterdir():
 if f.is_file() and f.suffix=='.log':
  with f.open('rb') as src,gzip.open(out/(f.name+'.gz'),'wb') as dst:shutil.copyfileobj(src,dst)
 elif f.is_file() and f.suffix in ['.json','.meta','.diff','.py','.txt']:shutil.copy2(f,out/f.name)
# The independent probe validation records take precedence over unsuccessful initial scratch diff attempts.
for mid in ['M1','M2','M3','M4','P1','W1','W2','S1','S2','S3','S4','S5','S6','P2','P3','P4','P5','P6','P7']:
 check=subprocess.run(['git','apply','--check',str(p/(mid+'.diff'))],cwd=root,capture_output=True,text=True);assert check.returncode==0,(mid,check.stderr)
(out/'replay-validation.json').write_text(json.dumps(dict(base=base,all_diffs_apply=True,production_builds=[json.loads((p/(n+'-build.meta')).read_text()) for n in ['M1','M2','M3','M4']],probe_build=json.loads((p/'P1-build.meta').read_text()),harness_vet=json.loads((p/'probe-validation.json').read_text())),indent=2))
summary=[dict(test=x['test'],verdict=x['verdict'],seconds=x['seconds'],last_proven_fail=x['last_proven_fail']) for x in report];print(json.dumps(summary,indent=2))
