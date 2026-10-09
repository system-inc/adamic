import pathlib,json,re,statistics,difflib,subprocess,shutil
root=pathlib.Path('review/test-audit/internal-oracle-oct6_mutant');scope=json.loads((root/'scope.json').read_text());menu=json.loads((root/'menu.json').read_text());rows=scope['rows'];witness=set(scope['witness_rows']);runs=json.load(open('/tmp/u066/runs.json'));alone=json.load(open('/tmp/u066/P1-alone-runs.json'))
def data(p):
 ds=[]
 for l in pathlib.Path(p).read_text().splitlines():
  try:ds.append(json.loads(l))
  except:pass
 return ds
# Map every diagnostic line back to the starting commit.
originals=json.load(open('/tmp/u066/originals.json'));scratch=pathlib.Path('/tmp/u066/mapping');scratch.mkdir(exist_ok=True)
for f,s in originals.items():p=scratch/f;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(s)
subprocess.run(['git','apply','--unsafe-paths',str((root/'switch.diff').resolve())],cwd=scratch,check=True)
line_maps={}
for f,s in originals.items():
 a=s.splitlines();b=(scratch/f).read_text().splitlines();mapping={}
 for tag,i,j,k,l in difflib.SequenceMatcher(a=a,b=b,autojunk=False).get_opcodes():
  for x in range(k,l):mapping[x+1]=(i+x-k+1 if tag=='equal' else i+1)
 line_maps[pathlib.Path(f).name]=mapping
def origin_line(s):
 return re.sub(r'(\w+\.(?:go|c)):(\d+):',lambda m:m.group(1)+':'+str(line_maps.get(m.group(1),{}).get(int(m.group(2)),int(m.group(2))))+':',s)
matrix=[]
for m in menu:
 id=m['id'];records=[x for x in runs if x['id']==id];results={r:'unknown' for r in rows};ev={};raw={};subcases={}
 for part in records:
  for d in data('/tmp/u066/'+id+'-'+part['part']+'.log'):
   test=d.get('Test','');row=test.split('/')[0]
   if row not in results:continue
   if d.get('Action') in ['pass','fail']:
    if test==row:results[row]=d['Action']
    elif d['Action']=='fail':results[row]='fail'
    if test!=row:subcases[test]=d['Action']
   out=d.get('Output','').strip()
   if out and re.search(r'\w+_test.go:\d+:',out) and row not in ev and 'caught:' not in out and ('caught by' not in out or 'not caught' in out) and 'padding caught' not in out and 'retain caught' not in out:raw[row]=out;ev[row]=origin_line(out)
 if id=='P1':
  for record in alone:
   row=record['test'];ds=data('/tmp/u066/P1-alone-'+row+'.log');results[row]='unknown'
   for d in ds:
    if d.get('Test','').split('/')[0]==row and d.get('Action')=='fail':results[row]='fail'
    out=d.get('Output','').strip()
    if out and re.search(r'\w+_test.go:\d+:',out) and row not in ev:raw[row]=out;ev[row]=origin_line(out)
   if results[row]=='unknown' and any(d.get('Output','').startswith('panic:') for d in ds):results[row]='fail';ev[row]='panic: nil pointer dereference after Lower returned nil, nil'
 matrix.append(dict(id=id,kind='production' if id[0]=='M' else 'witness-check' if id[0]=='W' else 'probe',results=results,subcases=subcases,failures=[r for r in rows if results[r]=='fail'],production_kills=[r for r in rows if results[r]=='fail' and r not in witness] if id[0]=='M' else [],evidence=ev,raw_evidence=raw,commands=records))
mat={x['id']:x for x in matrix};timings={}
for row in rows:
 a=[]
 for rep in range(1,4):
  p=pathlib.Path('/tmp/u066/time-'+row+'-'+str(rep)+'.log')
  if not p.exists():continue
  m=re.search(r'^ok\s+\S+\s+([0-9.]+)s',p.read_text(),re.M)
  if m:a.append(float(m.group(1)))
 timings[row]=dict(samples=a,median=statistics.median(a) if len(a)==3 else None)
bounded=[]
for rep in range(1,4):
 s=pathlib.Path('/tmp/u066/time-Native-bounded-'+str(rep)+'.log').read_text();bounded.append(float(re.search(r'^ok\s+\S+\s+([0-9.]+)s',s,re.M).group(1)))
timings['TestNativeAgreesWithNode']['bounded_samples']=bounded;timings['TestNativeAgreesWithNode']['bounded_median']=statistics.median(bounded);timings['TestNativeAgreesWithNode']['whole_row']='over budget at 90.039s; further whole-row timings stopped'
filemap={}
for p in pathlib.Path('internal/oracle').glob('*test.go'):
 for m in re.finditer(r'^func (Test\w+)\(',p.read_text(),re.M):
  if m.group(1) in rows:filemap[m.group(1)]=str(p)
oracles=['Node stdout disagreement detects the two built-in inheritance mutants; witness weakened disagreement.','Node comparison stays clean; LeakSanitizer report text must contain LeakSanitizer or leaked. Witness weakened leakSanitizer.','Node stdout disagreement must report stdout differs for built-in present-zero padding; clean exit/stderr and leaks are preconditions.','Executes Node for the source and compares native and JavaScript results. Handwritten Node 11 and leak-free expectations also apply.','Node stdout disagreement must report stdout differs for built-in reader zero padding; clean exit/stderr and leaks are preconditions.','Node source versus native release, sanitized, and JavaScript output. Full row also contains self-written refusal and inserted-check expectations; matrix covers twelve lowering fixtures only.','Node stdout comparison must notice the built-in extra byte; witness weakened disagreement.','Node execution on a shared stdout/stderr file decides byte order and exit status. Handwritten observed-output containment checks prevent a missing case.','Node execution on closed stdout decides exit and stderr; a handwritten Node exit-70 assertion also applies. M1 produces exit 71 with identical stderr.','Node execution with streams sharing one pipe decides exact output order; handwritten first/second/third lines must also appear.','Executed Node dialogue decides prompt-before-read and output; handwritten ready/got yes guard applies.','Executed Node protocol decides signal disposition, stdout and empty stderr; CPU-based readiness prevents premature signaling.','Executed Node outputs are checked against handwritten strings. Lower must produce lower.Refused containing a handwritten overload label. The ruling oracle is self; Node does not prove the static refusal.','Node stdout disagreement must catch each built-in missing/late parameter store; exit and sanitizer cleanliness are preconditions.','ASan stderr must contain AddressSanitizer: heap-use-after-free. Exit status alone is not checked. Disabling ASan makes the built-in mutant exit 0 with empty stderr and the witness fail.']
entries={rows[3]:['P1'],rows[5]:['P1','P2'],rows[7]:['P2','P3'],rows[8]:['P2','P4'],rows[9]:['P2','P3'],rows[10]:['P2','P3'],rows[11]:['P2','P4'],rows[12]:['P1']}
proofs={r:('W2' if r==rows[1] else 'W3' if r==rows[14] else 'W1') for r in witness}
group=[rows[i] for i in [3,5,9,10,12]]
subs={row:[other for other in group if other!=row] for row in group}
results=[]
for i,row in enumerate(rows):
 kills=[] if row in witness else [id for id in ['M1','M2','M3','M4'] if mat[id]['results'][row]=='fail'];unique=[id for id in kills if len(mat[id]['production_kills'])==1];pk=[] if row in witness else [id for id in entries.get(row,[]) if mat[id]['results'][row]=='fail'];verdict='witness' if row in witness else 'sacred' if unique else 'subsumed';proof=proofs.get(row,unique[0] if unique else 'M4');line=mat[proof]['evidence'].get(row,'--- FAIL: '+row);line=line[:600];part='native' if row==rows[5] else 'rows';command=next(x['command'] for x in runs if x['id']==proof and x['part']==part);others=subs.get(row,[]);other=min((x for x in others if timings[x]['median'] is not None),key=lambda x:timings[x]['median'],default=None)
 obj=dict(test=row,package='internal/oracle',file=filemap[row],seconds=timings[row]['median'],oracle=oracles[i],oracle_kind=['external-run','self'] if i in [1,3,5,7,8,9,10,11,12,14] else 'external-run',kills=kills,unique_kills=unique,last_proven_fail=proof+': '+line,verdict=verdict,subsumed_by=others,mutants_in_matrix=4,probe_kills=pk,subsumer_seconds=timings[other]['median'] if other else None,vacuous=False if pk else None,bounded=True,matrix_rows=rows,evidence=command+' > logs/'+proof+'-'+part+'.log 2>&1; '+line,witness_kills=[proof] if row in witness else [],subsumption_mutants=len(kills) if other else None)
 if row==rows[5]:obj['bounded_seconds']=timings[row]['bounded_median'];obj['bounded_subcases']=scope['native_actual_subcases']
 obj['entry_probe_results']={id:mat[id]['results'][row] for id in entries.get(row,[])};obj['witness_check']='disagreement' if proof=='W1' else 'leakSanitizer' if proof=='W2' else 'ASan enabled' if proof=='W3' else None
 results.append(obj)
(root/'results.json').write_text(json.dumps(results,indent=2)+'\n');(root/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(root/'timings.json').write_text(json.dumps(dict(nproc=5,rows=timings,timing_wall=float(pathlib.Path('/tmp/u066/timing-wall.txt').read_text()),runs=runs,panic_reruns=alone,validation=json.load(open('/tmp/u066/validation.json'))),indent=2)+'\n')
(root/'logs').mkdir(exist_ok=True)
for p in pathlib.Path('/tmp/u066').glob('*.log'):shutil.copyfile(p,root/'logs'/p.name)
for name in ['originals.json','bounded-baseline.cov','native-baseline.cov','validation.json','runs.json','P1-alone-runs.json','init.py','timing.py','prepare.py','validate.py','run.py','rerun-P1.py','report.py']:
 shutil.copyfile('/tmp/u066/'+name,root/name)
summary=['u066 starts at '+scope['base']+'; all fifteen requested names remain in their listed files.','Full package baseline timed out at 90.088s without a completed test failure; narrowed clean runs passed.','Seven witnesses proved; three bounded sacred rows; five subsumed rows based on one generic Lower-entry mutant.','Four production mutants, three weakened witness checks, and four distinct-entry probes have standalone validated diffs.','Production and harness changes restored; evidence is on test-audit/internal-oracle-oct6_mutant under review/test-audit/internal-oracle-oct6_mutant/.']
report='\n'.join(summary)+'\n\n```json\n'+json.dumps(results,indent=2)+'\n```\n\n| ID | origin/main file:line | change | failed ordinary/witness rows |\n|---|---|---|---|\n'
for m in menu:
 if m['id'][0]=='P':continue
 failures=mat[m['id']]['production_kills'] if m['id'][0]=='M' else [r for r in mat[m['id']]['failures'] if r in witness]
 report+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | `'+m['before'].replace('\n',' ')+'` to `'+m['after'].replace('\n',' ')+'` | '+', '.join(failures)+' |\n'
report+='\nSurvivors: none among M1-M4. All three weakened witness checks were caught. Probe passes are neither survivors nor production kills.\n\nFriction, ambiguity, and limits:\n\n'
report+='- The requested historical commit 8de93800f4 differs from current origin/main 6c60da09. No requested row moved or vanished.\n- The package and TestNativeAgreesWithNode alone both exceed the prescribed 90-second budget. Its whole-row median is null; three bounded timings cover twelve actual fixtures. The selector initially named fourteen fixture patterns, but status_of_stdout and killed_after_output are absent from this row\'s fixture list. Actual matched subcases are explicit in scope.json and results.json.\n- Go coverage listed 579 reached lower/native/oracle production functions before the fixed menu. Test-file helper functions were read and listed using rg. Exact runtime C branch coverage was not obtained; the C support inventory is conservative.\n- The seven planted-failure tests are witnesses. Their M4 failures are broken preparation and cannot prove the comparison or sanitizer check. Those failures remain in raw matrix results but never count toward their kills or uniqueness. W1, W2, W3 supply their verdicts.\n- W1 and W2 return the empty comparison/report value, under the explicit witness-check exception. They are not production mutants. W3 disables the check by changing the sanitizer option, leaving the planted retain mutation and expected report predicate intact. With ASan disabled the observed mutant exits 0 with empty stderr, demonstrating why an exit-only check would miss it.\n- Five subsumption findings rest on only M4, which breaks the general one-file entry requirement. This does not measure omitted-slot, overload-proof, or individual fixture sensitivity. The five rows mutually subsume each other within this single-mutant matrix; each names the others. The fastest measured subsumer is recorded. These judgments and the other single-mutant findings are hints, not deletion recommendations.\n- Runtime mutations use a native-process environment selector; the runtime source/header content remains identical for every selection, so cached libraries execute the selected behavior rather than stale compiled branches. ADAMIC_GATE_UNCACHED=1 forces actual Node/native/protocol observations for every matrix run; ADAMIC_BUILD_CACHE_DIR is separate per selection.\n- The nil Lower probe caused a Go panic. Every assigned row was rerun alone, retaining the twelve-fixture restriction for NativeAgrees. Probe results on witnesses are precondition effects and vacuity is null there. Ordinary output rows are judged by their own runtime entries, not by Lower preparation.\n- Three-per-row timing uses normal warm result caches, as the brief\'s commands specify. The first observation is often slower than subsequent cached observations. The matrix bypasses those caches. The full NativeAgrees row was stopped after its first 90.039-second timeout, not repeated twice more.\n- M1 is caught on exit status, but the test also compares exact stderr; the mutant preserved stderr. M2 and M3 change actual byte order/content. The overload ruling oracle is a self-written refusal label even though Node runs each program: Node runtime behavior cannot establish the static refusal policy.\n- All requested rows ran without skips. Outside-slice skip names from the unfinished full baseline are listed below. No outside-slice SDK or corpus opt-in was installed. Other package rows, unselected NativeAgrees fixtures, repo-wide uniqueness, and exhaustive C branch behavior remain unknown.\n'
skips=sorted({d['Test'].split('/')[0] for d in data('/tmp/u066/baseline.log') if d.get('Action')=='skip' and d.get('Test')});report+='\nOutside-slice observed skips: '+(', '.join(skips) if skips else 'none before the timeout')+'.\n\n'
validation=json.load(open('/tmp/u066/validation.json'));report+='Warm setup reused; setup.sh not run; nproc=5. npm ci: 0.537s. Clean timing runs: '+pathlib.Path('/tmp/u066/timing-wall.txt').read_text()+' wall seconds, including the 90.039s whole-row timeout. Standalone validation: '+str(sum(x['seconds'] for x in validation))+' wall seconds. Matrix: '+str(sum(x['seconds'] for x in runs))+' wall seconds. Per-run and rerun times are in timings.json. The clean switched control was 28.502 wall seconds, including compilation, runtime-library rebuilding, and 20.759 binary seconds; build-only native time was not separately isolated. The native library builds once per flag set because the selector bytes are shared. All final standalone validations passed and all diffs apply to the starting commit. Restored rows passed; restored bounded NativeAgrees passed three times.\n';(root/'REPORT.md').write_text(report)
print(json.dumps([dict(test=x['test'],seconds=x['seconds'],verdict=x['verdict'],kills=x['kills'],proof=x['last_proven_fail'][:180]) for x in results],indent=2))
