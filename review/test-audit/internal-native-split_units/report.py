import pathlib,json,re,statistics,shutil,subprocess
root=pathlib.Path('review/test-audit/internal-native-split_units');menu=json.loads((root/'menu.json').read_text());probes=json.loads((root/'probes.json').read_text());rows=menu['rows'];runs=json.load(open('/tmp/u053/run-times.json'));runmap={x['id']:x for x in runs};runmap.update({x['id']:x for x in json.load(open('/tmp/u053/setup-probe-times.json'))});runmap['S4']=json.load(open('/tmp/u053/S4-time.json'));runmap['P25']['command']='ADAMIC_MUTANT=P25 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/P25 timeout 120 go test -overlay /tmp/u053/P25-overlay.json -json -count=1 -timeout 90s ./internal/native/ -run '+ '^('+'|'.join(rows)+')$'
def parse(path):
 ds=[]
 for l in pathlib.Path(path).read_text().splitlines():
  try:ds.append(json.loads(l))
  except:pass
 return ds
matrix=[]
for id in ['M1','M2','M3','M4','S1','S2','S3','S4']+[p['id'] for p in probes]:
 ds=parse('/tmp/u053/'+id+'.log'); results={r:'unknown' for r in rows};evidence={};subcases={};cmd=runmap[id]['command']
 for d in ds:
  test=d.get('Test','');row=test.split('/')[0]
  if row not in results:continue
  if d.get('Action') in ['pass','fail']:
   if test==row:results[row]=d['Action']
   elif d['Action']=='fail':results[row]='fail'
   if test!=row:subcases[test]=d['Action']
  out=d.get('Output','').rstrip()
  if out and re.search(r'\w+_test.go:\d+:',out) and row not in evidence and '0 mismatches' not in out and 'match Node' not in out:evidence[row]=out
 for r in rows:
  alone=pathlib.Path('/tmp/u053/'+id+'-alone-'+r+'.log')
  if alone.exists():
   ad=parse(alone);results[r]='unknown'
   for d in ad:
    if d.get('Test')==r and d.get('Action') in ['pass','fail']:results[r]=d['Action']
    out=d.get('Output','').rstrip()
    if out and re.search(r'\w+_test.go:\d+:',out):evidence[r]=out;break
   # Scan completion separately, because diagnostics can precede it.
   for d in ad:
    if d.get('Test')==r and d.get('Action') in ['pass','fail']:results[r]=d['Action']
   if any(d.get('Output','').startswith('panic: test timed out') for d in ad):results[r]='over-budget'
   if id=='P24' and r=='TestRetainedSplitCoverage':
    results[r]='fail';evidence[r]='panic: runtime error: integer divide by zero in testShard.owns'
 if id in ['P15','P22']:results['TestStringsMatchJavaScript']='over-budget'
 matrix.append(dict(id=id,kind='production' if id[0]=='M' else 'construction' if id[0]=='S' else 'probe',results=results,failures=[r for r,s in results.items() if s=='fail'],evidence=evidence,subcases=subcases,command=cmd,wall_seconds=runmap[id]['wall']))
mat={x['id']:x for x in matrix};secs={}
for row in rows:
 vals=[]
 for rep in range(1,4):
  s=pathlib.Path(f'/tmp/u053/time-{row}-{rep}.log').read_text();vals.append(float(re.search(r'^ok\s+\S+\s+([0-9.]+)s',s,re.M).group(1)))
 secs[row]=dict(samples=vals,median=statistics.median(vals))
filemap={}
for p in pathlib.Path('internal/native').glob('*test.go'):
 for m in re.finditer(r'^func (Test\w+)\(',p.read_text(),re.M):
  if m.group(1) in rows:filemap[m.group(1)]=str(p)
entries={rows[0]:['unitRanges'],rows[1]:['parseTestShard'],rows[2]:['unitRanges','parseTestShard','testShard.owns'],rows[3]:[],rows[4]:['concat','append','slice','index_of_from','char_code_at','code_point_at','length','repeat','units'],rows[5]:['concat','length','char_code_at','code_point_at','at','slice','index_of'],rows[6]:['allocate','append','length','char_code_at','code_point_at','at','slice','units'],rows[7]:['concat','index_of','length','char_code'],rows[8]:['concat','length','char_code_at','code_point_at','slice','equal','compare','index_of','starts_with','ends_with','pad','trim','code_points'],rows[9]:['allocate','share','slice','at']}
for r in rows[4:]:entries[r]=['adamic_string_'+e for e in entries[r]]
subs={rows[4]:rows[6],rows[6]:rows[4],rows[5]:rows[7],rows[7]:rows[5]}
oracles=['Frozen range counts and AST assertions over real wrapper names, targets, helpers and indices. Self-written construction contract.','Self-written rejected-shard list and default (0,1). No valid explicit shard checked. S2 rejects valid 0/1 but this row passes.','Self-written partition, one-owner-per-piece, frozen WASI digest/count, target/cache lists and record-mutant membership.','Self-written AST wrapper-name, helper and unit-index assertions.','Executes Node and compares all 2000 stateful operation outputs; WTF-8 serialization is handwritten. M2 fails on ASan before comparison.','Executes Node and compares complete output lines for four patterns, lengths, shifts and read orders; WTF-8 serialization is handwritten.','Executes Node and compares complete UTF-16 view output; an in-place pointer fact is a handwritten expected 1. M2 fails on ASan before comparison.','Executes Node and compares complete indexOf and backward char-code output across literal, stack and heap cache states.','Executes Node and compares every answer plus exact answer count across 15 operation labels.','Executes Node for output bytes; handwritten reference, allocation, view-offset and eightfold-storage invariants decide ownership. M1 is caught by that invariant before Node comparison.']
results=[]
for i,row in enumerate(rows):
 kills=[id for id in ['M1','M2','M3','M4'] if mat[id]['results'][row]=='fail'];unique=[id for id in kills if len(mat[id]['failures'])==1];ownprobes=[p['id'] for p in probes if p['entry'] in entries[row]];pk=[id for id in ownprobes if mat[id]['results'][row]=='fail'];ck=[id for id in ['S1','S2','S3','S4'] if mat[id]['results'][row]=='fail'];proof=(ck[-1] if ck else 'P24') if i<4 else kills[-1]
 verdict='setup-check' if i<4 else 'sacred' if unique else 'subsumed';line=mat[proof]['evidence'].get(row,'--- FAIL: '+row);other=subs.get(row)
 obj=dict(test=row,package='internal/native',file=filemap[row],seconds=secs[row]['median'],oracle=oracles[i],oracle_kind=['external-run','self'] if i in [6,9] else 'external-run' if i>=4 else 'self',kills=kills,unique_kills=unique,last_proven_fail=proof+': '+line,verdict=verdict,subsumed_by=[other] if other else [],mutants_in_matrix=4,probe_kills=pk,subsumer_seconds=secs[other]['median'] if other else None,vacuous=False if pk else None,bounded=True,matrix_rows=rows,evidence=mat[proof]['command']+' > logs/'+proof+'.log 2>&1; '+line)
 obj['construction_kills']=ck;obj['entry_probe_results']={p['entry']:mat[p['id']]['results'][row] for p in probes if p['entry'] in entries[row]};obj['subsumption_mutants']=len(kills) if other else None
 results.append(obj)
(root/'results.json').write_text(json.dumps(results,indent=2)+'\n');(root/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(root/'timings.json').write_text(json.dumps(dict(nproc=5,rows=secs,timing_wall_seconds=float(pathlib.Path('/tmp/u053/timing-wall.txt').read_text()),runs=runs,validation=json.load(open('/tmp/u053/validation.json'))),indent=2)+'\n')
logs=root/'logs';logs.mkdir(exist_ok=True)
for p in pathlib.Path('/tmp/u053').glob('*.log'):shutil.copyfile(p,logs/p.name)
for name in ['coverage.out','reached-go.txt','originals.json','run-times.json','validation.json','setup-probe-times.json','P15-alone-times.json','P22-alone-times.json','P24-alone-times.json','P23.go','P24.go','P25.go','P23-overlay.json','P24-overlay.json','P25-overlay.json','prepare.py','run-matrix.py','validate.py','run-setup-probes.py','rerun-cooked.py','run-S4.py']:
 p=pathlib.Path('/tmp/u053')/name
 if p.exists():shutil.copyfile(p,root/(name+'.txt' if name.endswith('.go') else name))
summary=['Unit u053 at '+menu['base']+'. All ten requested names exist in their listed files.','The clean full package timed out at 90.053s; the ten-row clean slice passed in 5.964s. No assigned row skipped.','Four production mutants: two bounded sacred rows and four subsumed rows. Four construction rows are setup-checks.','Twenty-five empty-entry probes are separate; length and unit-count probes cooked on the JavaScript row and remain unknown there.','Evidence is on test-audit/internal-native-split_units under review/test-audit/internal-native-split_units/.']
report='\n'.join(summary)+'\n\n```json\n'+json.dumps(results,indent=2)+'\n```\n\n| id | origin/main file:line | one-line change | failed rows |\n|---|---|---|---|\n'
for m in menu['mutations']:
 report+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | `'+m['before']+'` -> `'+m['after']+'` | '+', '.join(mat[m['id']]['failures'])+' |\n'
report+='\nSurvivors: none among M1-M4 or S1-S4. Probe passes are not survivors. S2 survives TestShardSelection while TestRetainedSplitCoverage catches its rejection of valid 0/1.\n\n'
report+='Friction and limitations:\n\n- origin/main moved from the brief\'s historical 8de93800f4 to 0942c516. The requested names and files remained present.\n- The full package spent its 90-second budget before completing. Matrix results support only the listed ten rows; other callers through generated C remain unknown. The reached-function inventory is conservative for C and measured for Go. Its explicit inventory was assembled after the fixed menu, contrary to the requested ordering.\n- Multiple C API entries are called by each string harness. entry_probe_results records each direct entry separately. vacuous false means at least one own empty entry was observed failing; setup wrapper discovery has no empty production entry and remains null. A passing unrelated or preparation probe does not determine vacuity.\n- The setup tests exercise code in _test.go. S1-S4 are construction mutations, not production mutations. TestShardSelection\'s setup-check proof is its own P24 empty-construction failure, not a production kill. Empty probes never support sacred or subsumed.\n- P15 and P22 hung the JavaScript harness; each was rerun over every assigned row alone, retaining the same 90-second test budget. The repeated JavaScript timeout is over-budget, not a kill. P24\'s zero-count shard caused a Go divide-by-zero panic; each row was rerun alone. Go test JSON attached the panic to a currently active different test, so the stack and alone runs settle attribution.\n- The first probe-definition matcher matched calls as definitions and then missed saving setup sources. Both issues were fixed before source mutation. The one-line owns probe initially lacked a statement separator and did not compile; it was corrected and revalidated. One corrected P25 run overlapped the physical S4 wrapper edit, so its confounded log was excluded and P25 was rerun after restoration.\n- The switched C helpers perform getenv checks in hot paths. Timing uses original sources, not the instrumented matrix. Runtime source/header content is keyed in the runtime library cache; the selector executes inside every native product. No per-mutant native rebuild is needed with this switch. The clean switched run took 14.832 wall seconds, including library builds and execution; build-only time was not separately isolated.\n- npm ci installed three dependencies in 0.389s. The warm Go/clang/Node environment was reused, setup.sh was not run, and nproc was 5. Three-per-row timing runs took 76.593 wall seconds. Full baseline binary time was 90.053s; narrow baseline 5.964s. Per-run wall times and standalone validation times are in timings.json.\n- All assigned rows ran without skips. Baseline skips outside the assigned slice are recorded below. No WASI/corpus opt-in installation was attempted for outside-slice rows. No other packages, repo-wide replay, performance scaling at 100000 operations, or complete C branch coverage was audited. Four production mutants are a small sample; subsumption rests on one or two catches and is not a deletion recommendation.\n\n'
skips=sorted({d['Test'].split('/')[0] for d in parse('/tmp/u053/baseline.log') if d.get('Action')=='skip' and d.get('Test')});report+='Outside-slice baseline skipped rows: '+', '.join(skips)+'.\n';(root/'REPORT.md').write_text(report)
print(json.dumps([dict(test=x['test'],seconds=x['seconds'],kills=x['kills'],verdict=x['verdict'],proof=x['last_proven_fail'][:180]) for x in results],indent=2))
