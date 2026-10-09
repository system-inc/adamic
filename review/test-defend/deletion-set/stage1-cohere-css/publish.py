import pathlib,json,gzip,shutil,subprocess,re,shlex
p=pathlib.Path('/tmp/css-delete');dest=pathlib.Path('review/test-defend/deletion-set/stage1-cohere-css');dest.mkdir(parents=True,exist_ok=True)
items=json.loads((p/'mutant-list.json').read_text());matrix=json.loads((p/'matrix.json').read_text());assert len(matrix)==len(items),(len(matrix),len(items))
skipped=['TestCSSPrinterBoundaryProofs','TestCSSPrinterThroughput','TestOptionalBooleanPrinterMatchesGo','TestSharedSliceAppendAgreesWithNode']
def events(path):
 a=[]
 for line in path.read_text().splitlines():
  try:a.append(json.loads(line))
  except:pass
 return a
baseline=events(p/'baseline.log');assert (p/'baseline.exit').read_text().strip()=='0'
main=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
report={'package':'stage1/cohere/css','main':main,'skipped':skipped,'mutants':[],'keep':[],'deletable':[]}
for row in matrix:
 if not row['stale']:
  assert row['still_caught_by'] or any(any('Test'not in e and e['Action']in ['pass','fail'] for e in events(p/(r['id']+'.log'))) for r in row['runs'] if not r['panicking_tests']),row['mutant']
  for r in row['runs']:
   candidate_skip=skipped.copy()
   prior=[]
   for earlier in row['runs']:
    if earlier is r:break
    prior.extend(earlier['panicking_tests'])
   candidate_skip+=list(dict.fromkeys(prior))
   args=['go','test','-json','-count=1','-timeout','30m','./stage1/cohere/css/','-skip','^('+ '|'.join(candidate_skip)+')($|_|/)']
   r['argv']=args;r['command']='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/css-delete/cache/'+r['id']+' '+shlex.join(args)+' > /tmp/css-delete/'+r['id']+'.log 2>&1'
   (p/(r['id']+'.run.json')).write_text(json.dumps(r,indent=2)+'\n')
 out={k:row[k] for k in ['mutant','file_line','branch','candidates_failed','stale']};out['still_caught_by']=row.get('still_caught_by',[])
 if row.get('witness_failures'):out['witness_failures']=row['witness_failures']
 if row.get('panicking_tests'):out['panicking_tests']=row['panicking_tests']
 report['mutants'].append(out)
for test in skipped:
 relevant=[r for r in matrix if test in r['candidates_failed']]
 lost=[r for r in relevant if not r.get('still_caught_by')]
 if lost:report['keep'].append({'test':test,'because':'; '.join(r['mutant']+(' is stale, so deletion is not established' if r['stale'] else ' loses its last non-witness catcher without the set') for r in lost)})
 else:report['deletable'].append(test)
 if not relevant:report.setdefault('notes',{})[test]='No gathered mutant failed this candidate; it guards nothing shown by this evidence.'
report['notes']=report.get('notes',{})|{'scope':'Includes all 12 eligible M/D diffs plus three conservatively included B-prefixed production diffs from the same defender evidence.','early_stop':'Runs with a clean non-witness catch stop there; later outcomes are unknown. Witness-only failures do not count.'}
wall=sum(r['wall_seconds'] for row in matrix for r in row.get('runs',[]));bw=float((p/'baseline.wall').read_text());times=[r['wall_seconds'] for row in matrix for r in row.get('runs',[])]
report['timing']={'baseline_wall_seconds':round(bw,3),'replay_wall_seconds':round(wall,3),'replay_min_seconds':min(times),'replay_max_seconds':max(times)}
(p/'report.json').write_text(json.dumps(report,indent=2)+'\n');(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
(p/'baseline-summary.json').write_text(json.dumps({'main':main,'wall_seconds':bw,'passed':[e['Test'] for e in baseline if e['Action']=='pass' and 'Test'in e and '/'not in e['Test']],'skipped':[e['Test'] for e in baseline if e['Action']=='skip' and 'Test'in e and '/'not in e['Test']],'command':'ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/css-delete/cache/baseline '+shlex.join(['go','test','-json','-count=1','-timeout','30m','./stage1/cohere/css/','-skip','^('+ '|'.join(skipped)+')($|_|/)'])+' > /tmp/css-delete/baseline.log 2>&1'},indent=2)+'\n')
(p/'publication-notes.txt').write_text('Clean detached main '+main+'. Submodules were updated recursively to their recorded pins. Warm env.sh used; stage3/api npm ci ran before baseline. Initial /tmp free 5.9 GB, /workspace 17 GB. Removed the completed prior unit /tmp/def-nodefs; /tmp free rose to 6 GB. The /tmp filesystem has only 8.8 GB capacity, so 15 GB free is impossible. No disk failure occurred. /usr/bin/time was unavailable before Go started; baseline timing used Python instead. The initial 12-entry M/D list was saved before baseline; B1-B3 were added conservatively after finding their naming mismatch, before their own replays. No mutant was invented or rebased. Each run had a fresh ADAMIC_BUILD_CACHE_DIR and ADAMIC_GATE_UNCACHED=1; completed scratch build caches were removed to control disk use. Witness classification follows actual bodies, including mutant-only printer shards 032-127, sanitizer/range checks, planted disagreements and mutant product builders. Every production diff was reversed after its replay.\n')
(p/'pins.txt').write_text(subprocess.check_output(['git','submodule','status','--recursive'],text=True))
(p/'initial-mutant-list.json').write_text(json.dumps(items[:12],indent=2)+'\n')
for file in p.rglob('*'):
 if not file.is_file() or 'cache' in file.relative_to(p).parts:continue
 rel=file.relative_to(p)
 if file.suffix=='.log':
  target=dest/(str(rel)+'.gz');target.parent.mkdir(parents=True,exist_ok=True)
  with gzip.open(target,'wb') as f:f.write(file.read_bytes())
 else:
  target=dest/rel;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(file,target)
print(json.dumps(report,indent=2))
