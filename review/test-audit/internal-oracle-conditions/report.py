import pathlib,json,re,statistics,gzip,shutil,difflib,subprocess
root=pathlib.Path('/tmp/u059'); out=pathlib.Path('review/test-audit/internal-oracle-conditions');rows=json.loads((root/'rows.json').read_text());menu=json.loads((out/'menu.json').read_text()); witnesses={rows[1]:'W1',rows[2]:'W2',rows[3]:'W1'}
def events(file):
 result=[]
 if not file.exists():return result
 for line in file.read_text(errors='replace').splitlines():
  try:result.append(json.loads(line))
  except ValueError:pass
 return result
matrix={};evidence={};unknown={}
line_map=json.loads((out/'origin-line-map.json').read_text())
def origin_line(text):
 def replace(m):
  name,line=m.group(1),m.group(2);return name+':'+str(line_map.get(name,{}).get(line,int(line)))+':'
 return re.sub(r'([a-zA-Z0-9_]+\.go):(\d+):',replace,text)
for m in menu:
 mid=m['id'];es=events(root/(mid+'.log'))+events(root/(mid+'-counts.log'));result={};err={}
 for e in es:
  t=e.get('Test','').split('/')[0]
  if t not in rows:continue
  if e['Action'] in ['pass','fail','skip'] and e.get('Test')==t:result[t]=e['Action']
  if e.get('OutputType')=='error' and re.search(r'\.go:\d+:',e.get('Output','')):err.setdefault(t,e['Output'].strip().splitlines()[0])
 matrix[mid]=result;evidence[mid]={r:origin_line(line) for r,line in err.items()};unknown[mid]=[t for t in rows if t not in result]
(out/'matrix.json').write_text(json.dumps({'results':matrix,'unknown':unknown,'assertions':evidence},indent=2)+'\n')
seconds={}
for row in rows:
 vals=[]
 for f in sorted(root.glob('time-'+row+'-*.log')):
  match=re.search(r'^ok\s+\S+\s+([\d.]+)s',f.read_text(),re.M)
  if match:vals.append(float(match[1]))
 seconds[row]=statistics.median(vals) if len(vals)==3 else None
files={}
base=(out/'base.txt').read_text().strip()
for f in ['conditions_test.go','counters_test.go','counts_test.go','debugger_test.go','element_access_boundaries_test.go','entries_acceptance_test.go','entries_provenance_test.go','enum_initialization_reach_test.go']:
 s=subprocess.check_output(['git','show',base+':internal/oracle/'+f],text=True)
 for row in rows:
  match=re.search(r'^func '+row+r'\(',s,re.M)
  if match:files[row]='internal/oracle/'+f+':'+str(s[:match.start()].count('\n')+1)
oracles={rows[0]:('Node source execution, exact stdout/stderr/exit against native and JavaScript; ledger coverage uses repository-authored manifests.',['external-run','self']),rows[1]:('Node source execution; witness requires both deliberately changed products to exit 0 and differ only in stdout.',['external-run','self']),rows[2]:('Repository-authored historical ledger and coverage manifests; witness requires missing-site diagnostic substring.','self'),rows[3]:('Node source execution; witness requires both deliberately changed products to exit 0 and differ only in stdout.',['external-run','self']),rows[4]:('Node source execution for stdout; hand-calculated integer eligibility checked in emitted C.',['external-run','self']),rows[5]:('Repository snapshot counts.md, measured by the counted native runtime; self-authored allocation/reference counts.','self'),rows[6]:('Self-authored generated-C equivalence and absence-of-traps checks, plus exactly one emitted JavaScript debugger statement.','self'),rows[7]:('Node executes each source and must print witness; repository-authored exact NotYet/Refused labels.',['external-run','self']),rows[8]:('Missing pinned acceptance corpus; intended Node comparison plus self-authored reflection and checked-failure contracts.',['external-run','self']),rows[9]:('Node exact output for successful programs; repository-authored exact exit 70 and panic messages for rejected enumeration, plus IR provenance checks.',['external-run','self']),rows[10]:('Repository-authored exact checked-readiness panic and exit 70; unchecked variant must match Node.',['external-run','self']),rows[11]:('Node exact differential output for successful programs; Node TypeError substring for pending reads and exact repository-authored lowering refusal.',['external-run','self']),rows[12]:('Node must print before and report TypeError; native/JavaScript must match repository-authored exact ReferenceError panic and exit 70.',['external-run','self'])}
kills={r:[m['id'] for m in menu if matrix[m['id']].get(r)=='fail'] if r not in witnesses else [] for r in rows}
report=[]
for r in rows:
 unique=[mid for mid in kills[r] if [x for x in rows if x not in witnesses and matrix[mid].get(x)=='fail']==[r]]
 probes=[];probe_results={}
 for pid in ['P1','P2','P3']:
  if r in witnesses or r==rows[8] or (pid=='P1' and r==rows[10]) or (pid=='P3' and r==rows[4]) or (pid in ['P2','P3'] and r==rows[7]):continue
  f=root/(pid+'-'+r+'.log') if pid=='P1' or (pid=='P2' and r==rows[5]) else root/(pid+'.log')
  es=events(f);status=None
  for e in es:
   if e.get('Test')==r and e['Action'] in ['pass','fail','skip']:status=e['Action']
  if pid=='P1' and 'panic:' in f.read_text(errors='replace') if f.exists() else False:status='fail'
  if status is not None:probe_results[pid]=status
  if status=='fail':probes.append(pid)
 subs=[];verdict='cannot-judge';last=None;ev=''
 if r in witnesses:
  mid=witnesses[r];es=events(root/(mid+'.log'));failed=any(e.get('Test')==r and e['Action']=='fail' for e in es)
  verdict='witness' if failed else 'untrue'
  line=next((e['Output'].strip().splitlines()[0] for e in es if e.get('Test','').split('/')[0]==r and e.get('OutputType')=='error'),'')
  line=origin_line(line);last=mid+': '+line if failed else None;ev='ADAMIC_MUTANT='+mid+' '+json.loads((root/(mid+'.meta')).read_text())['command']+' => '+line
 elif r==rows[8]:ev='clean discovery and all runs: SKIP, missing stage3/fixtures/entries/expectations.json'
 elif unique:verdict='slow-worthy' if seconds[r] and seconds[r]>60 else 'sacred'
 elif kills[r]:
  candidates=[x for x in rows if x!=r and x not in witnesses and set(kills[r])<=set(kills[x])]
  if candidates:subs=[min(candidates,key=lambda x:seconds[x] if seconds[x]!=None else float('inf'))];verdict='subsumed'
  else:subs=sorted({x for mid in kills[r] for x in rows if x!=r and x not in witnesses and mid in kills[x]});verdict='overlapping'
 else:verdict='untrue'
 if kills[r]:
  mid=kills[r][-1];line=evidence[mid].get(r,'completed FAIL without a captured assertion line');last=mid+': '+line
  suffix='-counts' if r==rows[5] else ''
  ev='ADAMIC_MUTANT='+mid+' ADAMIC_BUILD_CACHE_DIR=/tmp/u059/cache/'+mid+' '+json.loads((root/(mid+suffix+'.meta')).read_text())['command']+' => '+line
 vacuous=None if not probe_results else any(v=='pass' for v in probe_results.values())
 obj=dict(test=r,package='internal/oracle',file=files[r],seconds=seconds[r],oracle=oracles[r][0],oracle_kind=oracles[r][1],kills=kills[r],unique_kills=unique,last_proven_fail=last,verdict=verdict,subsumed_by=subs,mutants_in_matrix=8,probe_kills=probes,subsumer_seconds=seconds[subs[0]] if verdict=='subsumed' else None,vacuous=vacuous,bounded=True,matrix_rows=rows,evidence=ev)
 if probe_results:obj['entry_probe_results']=probe_results
 if verdict=='subsumed':obj['subsumption_basis_mutants']=len(kills[r])
 report.append(obj)
(out/'rows.json').write_text(json.dumps(report,indent=2)+'\n')
for f in root.glob('*.log'):
 with f.open('rb') as src,gzip.open(out/(f.name+'.gz'),'wb') as dst:shutil.copyfileobj(src,dst)
for pattern in ['*.meta','*-exit','*.exit','*-start','*.py']:
 for f in root.glob(pattern):shutil.copy2(f,out/f.name)
(out/'timings.json').write_text(json.dumps(seconds,indent=2)+'\n')
print(json.dumps([{k:r[k] for k in ['test','seconds','kills','unique_kills','verdict','probe_kills','vacuous']} for r in report],indent=2))
