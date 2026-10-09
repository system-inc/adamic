import pathlib,json,re,statistics,subprocess,gzip,shutil,datetime
root=pathlib.Path('/tmp/u082');out=pathlib.Path('review/test-audit/stage1-cohere-cssnumbers');groups=json.loads((out/'scope.json').read_text());base=(out/'base.txt').read_text().strip()
def events(name):
 result=[]
 for l in (root/(name+'.log')).read_text(errors='replace').splitlines():
  try:result.append(json.loads(l))
  except ValueError:pass
 return result
def map_lines(diff):
 maps={};file=None;old=new=1
 for line in diff.splitlines():
  if line.startswith('+++ b/'):file=pathlib.Path(line[6:]).name;maps.setdefault(file,{})
  elif line.startswith('@@'):
   m=re.match(r'@@ -(\d+)(?:,\d+)? \+(\d+)',line);a,b=map(int,m.groups())
   if file:
    while new<b:maps[file][new]=old;new+=1;old+=1
   old,new=a,b
  elif file and not line.startswith(('diff ','index ','---')):
   if line.startswith(' '):maps[file][new]=old;old+=1;new+=1
   elif line.startswith('-'):old+=1
   elif line.startswith('+'):new+=1
 for file,d in maps.items():
  if d:
   last=max(d);delta=d[last]-last
   for n in range(last+1,2000):d[n]=n+delta
 return maps
first_map=map_lines((out/'scratch-switch.diff').read_text());extra_map=map_lines((out/'construction-empty-switch.diff').read_text())
def normalize(s,extra=False):
 mp=extra_map if extra else first_map
 return re.sub(r'([\w]+\.go):(\d+):',lambda m:m[1]+':'+str(mp.get(m[1],{}).get(int(m[2]),int(m[2])))+':',s)
def status(name,members):
 es=events(name);ss={e['Test']:e['Action'] for e in es if e.get('Test') in members and e['Action'] in ['pass','fail','skip']}
 assert len(ss)==len(members),(name,len(ss),len(members))
 return 'fail' if 'fail' in ss.values() else 'skip' if set(ss.values())=={'skip'} else 'pass'
def failure(name,members):
 for e in events(name):
  if e.get('Test','').split('/')[0] in members and e.get('OutputType')=='error':return normalize(e['Output'].strip().splitlines()[0],name in ['P2','P3','P4','P5','P6','P7'])
 return None
matrix={mid:{g['test']:status(mid,g['members']) for g in groups} for mid in ['M1','M2','M3','M4']};(out/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
metadata={}
for f in root.glob('*.meta'):
 meta=json.loads(f.read_text());meta['binary_seconds']=None
 for e in events(f.stem):
  if e.get('Output'):
   m=re.search(r'(?:ok\s+|FAIL\s+)github.com/system-inc/adamic/stage1/cohere/cssnumbers\s+([\d.]+)s',e['Output'])
   if m:meta['binary_seconds']=float(m[1])
 metadata[f.stem]=meta
seconds={}
for i,g in enumerate(groups):
 vs=[]
 for n in [1,2,3]:
  m=re.search(r'^ok\s+\S+\s+([\d.]+)s',(root/f'time-{i}-{n}.log').read_text(),re.M)
  if m:vs.append(float(m[1]))
 seconds[g['test']]=statistics.median(vs) if len(vs)==3 else None
pids=['P4','P5','P6','P7','P1','P3',None,'P2'];proofs=['S1','S1','S2','S3','M3','S4','W1','S4'];verdicts=['untrue','setup-check','untrue','untrue','sacred','setup-check','witness','setup-check']
oracles=[('Build completion only; no executable existence or behavior assertion.','self'),('Go cohere executes corpus; only clean completion is checked, no expected output.','external-run'),('Own compiler preparation accepts port and built-in variants; generated outputs are not checked.','self'),('Own native preparation reports success; executable existence and output are not checked.','self'),('Go cohere actual CSS printer and Prettier 3.9.6, exact output plus clean exit/stderr, native sanitizers/leaks and Node/backend execution.','external-run'),('Independent self-authored census IDs, missing/repeated ranges and shard-selection partition checks.','self'),('Self-authored planted case IDs; child must fail exactly the owner shard with its case ID.','self'),('Self-authored exact 357-unit census and preparation success.','self')]
rows=[]
for i,g in enumerate(groups):
 members=g['members'];name=g['test'];files=[]
 for f in ['port_test.go','shards_test.go','setup_products_test.go','top_level_shards_test.go']:
  path='stage1/cohere/cssnumbers/'+f;s=subprocess.check_output(['git','show',base+':'+path],text=True)
  for member in members:
   m=re.search(r'^func '+member+r'\(',s,re.M)
   if m:files.append(path+':'+str(s[:m.start()].count('\n')+1))
 kills=[mid for mid in matrix if matrix[mid][name]=='fail'];unique=[mid for mid in kills if sum(v=='fail' for v in matrix[mid].values())==1]
 pid=pids[i];ps=status(pid,members) if pid else None;proof=proofs[i];fl=failure(proof,members)
 cmd=metadata[proof]['command'];extra='ADAMIC_AUDIT_MODE='+proof+' ' if proof.startswith(('S','W')) else 'selector='+proof+' '
 row=dict(test=name,package='stage1/cohere/cssnumbers',file=files[0],seconds=seconds[name],oracle=oracles[i][0],oracle_kind=oracles[i][1],kills=kills,unique_kills=unique,last_proven_fail=proof+': '+fl if fl else None,verdict=verdicts[i],subsumed_by=[],mutants_in_matrix=4,probe_kills=[pid] if ps=='fail' else [],subsumer_seconds=None,vacuous=ps=='pass' if ps else None,bounded=False,matrix_rows=[g['test'] for g in groups],evidence=extra+' '.join(cmd)+' => '+(fl if fl else 'PASS although construction produced no promised artifact; '+proof+'-artifacts.json is []'))
 row['members']=members;row['member_files']=files
 if pid:row['entry_probe']=dict(id=pid,result=ps,command=metadata[pid]['command'])
 if i==4:
  passed=[e['Test'] for e in events('P1') if e.get('Test') in members and e['Action']=='pass']
  row['vacuous_subcases']=[m for m in passed if m not in ['TestCSSNumbers_350','TestCSSNumbers_351','TestCSSNumbers_352','TestCSSNumbers_353','TestCSSNumbers_356']]
  row['non_port_members']=['TestCSSNumbers_353','TestCSSNumbers_356'];row['witness_members']=['TestCSSNumbers_350','TestCSSNumbers_351','TestCSSNumbers_352'];row['witness_weakening']='W2, each witness member failed when its comparison treated every result as agreement'
 rows.append(row)
(out/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
for f in root.glob('*.log'):
 with f.open('rb') as src,gzip.open(out/(f.name+'.gz'),'wb') as dst:shutil.copyfileobj(src,dst)
for pattern in ['*.meta','*.py','*.exit','*-exit','*-artifacts.json','*.stdout','*.stderr','*-input.txt','start']:
 for f in root.glob(pattern):shutil.copy2(f,out/f.name)
(out/'timings.json').write_text(json.dumps(seconds,indent=2)+'\n')
builds=[]
for name in ['baseline','clean-switch','M1','M2','M3','M4','P1']:
 for e in events(name):
  for line in e.get('Output','').splitlines():
   if 'build cssnumbers_' in line:builds.append(dict(run=name,line=line.strip()))
(out/'build-records.json').write_text(json.dumps(builds,indent=2)+'\n')
(out/'costs.json').write_text(json.dumps(dict(nproc=5,setup_seconds=0,setup='warm env.sh works; stage3/api npm ci reported 407ms; pinned Prettier installed and npm ci completed',baseline_binary_seconds=54.529,native_standalone_builds={mid:metadata[mid+'-build'] for mid in matrix},invocations=metadata,start_utc=(root/'start').read_text().strip(),finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat()),indent=2)+'\n')
(out/'survivors.json').write_text(json.dumps([dict(id='M4',classification='unguarded behavior in this package',input='d\u00801.000px\n',before=(root/'M4-before.stdout').read_text(),after=(root/'M4-after.stdout').read_text(),native_after=(root/'M4-native.stdout').read_text(),go_oracle=(root/'M4-go.stdout').read_text(),exit_codes={side:int((root/('M4-'+side+'.exit')).read_text()) for side in ['before','after','native','go']})],indent=2)+'\n')
print(json.dumps([{k:r[k] for k in ['test','seconds','kills','verdict','probe_kills','vacuous']} for r in rows],indent=2))
