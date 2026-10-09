import json,statistics,re,gzip
def readlog(f):
 return f.read_text() if f.exists() else gzip.open(str(f)+".gz","rt").read()
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-lint-bv61ffb_decoded_options')
groups=json.loads((p/'groups.json').read_text());scope=json.loads((p/'scope.json').read_text())
results=[json.loads(l) for l in (p/'results.jsonl').read_text().splitlines()] if (p/'results.jsonl').exists() else []
byid={r['id']:r for r in results}
rows=[]
names=['TestDecodedOptionsAndMutant family','TestDecodedOptionsAndMutantPlantedFailure','TestDecodedOptionsAndMutant_Setup','TestClosedComparatorGaps','TestCommandDiagnosticsDropOnlyModuleDownloads','TestExecuteFailsOnStderrOtherThanModuleDownloads','TestCompilerAndStage1Agree_Setup','TestProduct_CompilerAgreement family','TestCompilerAndStage1Agree family','TestCompilerAndStage1AgreePlantedDisagreement','TestCompleteSuggestionSerialization_IsolatedShard']
roles={1:('witness','W2'),2:('setup-check','S4'),4:('setup-check','S1'),5:('witness','W1'),6:('setup-check','S2'),7:('setup-check','S5'),9:('witness','W3'),10:('setup-check','S3')}
prodids=[r['id'] for r in results if r['kind']=='production'];matrix=['TestDecodedOptionsAndMutant_'+f'{i:03d}' for i in range(6)]+['TestDecodedOptionsAndMutantUnion','TestClosedComparatorGaps']+['TestCompilerAndStage1Agree_'+f'{i:03d}' for i in (2,7,12)]
for i,g in enumerate(groups):
 seconds=[];censored=[]
 for n in range(3):
  f=p/f'timing-{i}-{n}.log'
  if not f.exists() and not Path(str(f)+".gz").exists():continue
  for l in readlog(f).splitlines():
   try:r=json.loads(l)
   except:continue
   if r.get('Action') in ('pass','fail') and not r.get('Test'):
    if r['Action']=='pass':seconds.append(r['Elapsed'])
    else:censored.append(90.0)
 samples=seconds+censored;median=statistics.median(samples) if len(samples)==3 else None
 sec=None if median is not None and median>=90 and len(censored)>=2 else median
 tested=set(g)
 # Only positive decoded agreement members prove port semantics, never its built-in-mutant preconditions.
 semantic=set(g[:3]) if i==0 else tested
 kills=[mid for mid in prodids if semantic.intersection(byid[mid]['failures'])]
 probes=[mid for mid in ('P1','P2') if mid in byid and ((mid=='P1' and i in (0,8)) or (mid=='P2' and i==3)) and semantic.intersection(byid[mid]['failures'])]
 ownprobe='P2' if i==3 else 'P1' if i in (0,8) else None
 vacuous=None
 if ownprobe in byid:
  observed=semantic.intersection(byid[ownprobe]['completed'])
  if observed:vacuous=not bool(semantic.intersection(byid[ownprobe]['failures']))
 verdict='cannot-judge';sub=[];unique=[]
 if i in roles:
  expected,mid=roles[i];r=byid.get(mid)
  if r and tested.intersection(r['failures']):verdict=expected
  elif r and tested.intersection(r['completed']):verdict='untrue'
 elif kills:verdict='pending matrix comparison'
 else:verdict='untrue' if len(prodids)==4 and any(tested.intersection(r['completed']) for r in results if r['kind']=='production') else 'cannot-judge'
 proof=next((r for r in reversed(results) if semantic.intersection(r['failures']) and (r['kind']=='production' if i not in roles else r['id']==roles[i][1])),None)
 fail_line=None
 if proof:
  for l in readlog(p/(proof['id']+'.log')).splitlines():
   try:r=json.loads(l)
   except:continue
   if r.get('Test','').split('/')[0] in semantic and 'Output' in r and ('case ' in r['Output'] and 'port ' in r['Output'] or 'native prints' in r['Output'] or 'enumerated ' in r['Output'] or 'one download: got' in r['Output'] or 'preparation failed' in r['Output'] or 'failed preparation' in r['Output'] or 'no such file' in r['Output'] or 'caught ' in r['Output'] or 'isolated shard:' in r['Output'] or 'execute passed' in r['Output'] or 'exit status' in r['Output'] or 'FAIL' in r['Output']):fail_line=r['Output'].strip();fail_line=fail_line[:350]+' [line continues in log]' if len(fail_line)>350 else fail_line;break
 oracle='Go cohere executed live; canonical diagnostic, suggestion, rejection, convergence and fixed-source comparison' if i in (0,8) else 'Node executed live plus handwritten expected outputs 2, 1, 1; checked by clean baseline' if i==3 else 'Suite-authored construction or planted-failure expectations'
 if i==0:oracle+='; M4 native catch is exit 70 (panic: missing options) before output comparison; member 003 checks only Go count, members 004/005 witness built-in mutants, Union checks registration'
 row={'test':names[i],'package':'stage1/cohere/lint','file':scope.get('locations',{}).get(g[0]),'seconds':sec,'oracle':oracle,'oracle_kind': 'external-run' if i in (0,8) else ['external-run','self'] if i==3 else 'self','kills':kills,'unique_kills':unique,'last_proven_fail':(proof['id']+' '+str(fail_line)) if proof else None,'verdict':verdict,'subsumed_by':sub,'mutants_in_matrix':len(prodids) if i in (0,3,8) else 0,'probe_kills':probes,'subsumer_seconds':None,'vacuous':vacuous,'bounded':True,'matrix_rows':matrix if i in (0,3,8) else g if len(g)<10 else ['see groups.json'],'evidence': (' '.join(proof['command'])+' > '+proof['id']+'.log; '+str(fail_line)) if proof else 'See three isolated timing logs and results.jsonl; no demonstrated failure yet','members':g if len(g)<10 else {'file':'groups.json','group':i,'count':len(g)},'timing_samples':seconds,'timing_censored_runs':len(censored),'seconds_lower_bound':90 if sec is None and len(censored)>=2 else None}
 if i==0 and 'P1' in byid:row['vacuous_subcases']=sorted(tested.intersection(byid['P1']['completed'])-set(byid['P1']['failures']))
 rows.append(row)
for row in rows:
 if row['verdict']!='pending matrix comparison':continue
 k=set(row['kills']);others=[x for x in rows if x is not row and x['kills']]
 unique=[m for m in k if all(m not in x['kills'] for x in others)];row['unique_kills']=sorted(unique)
 if unique:row['verdict']='slow-worthy' if row.get('seconds_lower_bound') or (row['seconds'] or 0)>60 else 'sacred'
 else:
  subs=[x for x in others if k<=set(x['kills'])]
  if subs:
   sub=min(subs,key=lambda x:x['seconds'] if x['seconds'] is not None else 100000);row['verdict']='subsumed';row['subsumed_by']=[sub['test']];row['subsumer_seconds']=sub['seconds']
  else:row['verdict']='overlapping';row['subsumed_by']=[x['test'] for x in others if k.intersection(x['kills'])]
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
menus={}
for l in (p/'menu.jsonl').read_text().splitlines():
 r=json.loads(l);menus[r['id']]=r
lines=['Unit u106: 6503 requested functions verified against fresh origin/main.','Grouped into 11 rows, with mixed agreement and witness roles preserved.','Baseline whole package cooked; enabled bounded baseline passed.','Uniqueness and subsumption are bounded, pending central replay.','Evidence includes standalone diffs, raw logs, inventories and timing samples.','','```json',json.dumps(rows,indent=2),'```','','| ID | Origin file:line | Change | Observed failing top-level members |','|---|---|---|---|']
for mid,r in byid.items():
 m=menus[mid];fails=sorted(set(x.split('/')[0] for x in r['failures']));lines.append(f"| {mid} ({r['kind']}) | {m['file']}:{m['line']} | `{m['old']}` to `{m['new']}` | {', '.join(fails) or 'none observed'} |")
lines+=['','Survivors:']
for mid in prodids:
 if not byid[mid]['failures']:lines.append(mid+': equivalent candidate in this bounded experiment; no differing-output witness produced. Not classified as unguarded.')
lines+=['','Brief issues and limits:','The brief says 13 rows but enumerates 6503 functions. Family grouping produces 11 rows. All requested names exist. The stated historical commit differs from current origin/main; base.txt pins the actual start.','The decoded family combines positive port comparisons, a Go-only count check, built-in-mutant witnesses and a registration union. Production failures in the built-in-mutant members are preserved as raw evidence but excluded from semantic kills.','The compiler family includes Go-only shards and empty buckets. A completed Go-only or empty shard cannot establish port correctness. Full family timing cooks at 90 seconds; only selected source-Node shards enter the production matrix. Other backend and package kills are unknown.','Product declaration tests use one shared recipe checker and are construction checks, rather than semantic agreement rows. The stderr tests also test the suite harness itself; their edits are separately marked construction or witness experiments.','The function inventories are static conservative supersets; exact dynamic reachability was not measured. Four native rebuild mutations were chosen instead of 20 variants. No claim covers unexecuted functions.','Warm tools did not include the pinned TypeScript corpus. It was installed at /tmp/u106/typescript from v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8, to enable compiler agreement rows. npm ci ran before baseline.','Cold product construction can exceed the 90-second budget even when warm checks pass. Censored timing samples are not successful medians. Native rebuild durations are preserved in each build log; process wall times are in results.jsonl.','Initial compiler failures for M1/M2/M3/P1 were corpus cleanliness refusals, not semantic kills. The corpusfiles loader compares selected tracked TypeScript inputs to HEAD. Clean temporary variant commits allowed real comparisons without weakening that guard; replays replaced the refused outcomes. Central replay must likewise commit applied port diffs or use a clean variant checkout. Scratch commit ids and initial logs are retained in results.jsonl.',
'The initial P1 unconditional early return made later TypeScript discriminated-union refinements unreachable and failed compilation. That attempt is excluded. The corrected probe uses if (row.length >= 0) return 0 at entry. String length is always nonnegative, but the guard preserves type checking of the remaining source and its built-in-mutant anchor. Its actual native rebuild and agreement failures are logged.',
'P2 returns empty C source. The compiler-gap row rejects it during clang compilation, before semantic output comparison. This proves an empty artifact is rejected by the row pipeline, not that its semantic assertions run on an empty native executable.',
'Source Node compiler shard 012 passed M3, while 002 and 007 caught it. Those are observations, not evidence that 012 is permanently vacuous. P1 made all three selected source-Node shards fail. Shards 003, 004, 005 and the decoded Union passed P1: 003 is Go-only, 004/005 accept any inequality rather than identifying why the built-in mutant differs, and Union checks enumeration.',
'One native builder family construction fault was run on the Go-oracle member, using the common preparation-success checker. The other two members were timed but were not separately faulted; no semantic verdict is assigned to builders.',
'The 13-row count appears to omit family grouping of the three product declarations. Numbered wrappers are verified programmatically rather than pretending that a filename alone establishes scope. The complete member lists and source locations are in groups.json and scope.json.',
'No repo-wide uniqueness, full-package per-mutant run, exhaustive native compiler corpus matrix or exact coverage profile was completed. Empty-answer probes are excluded from all production kills.','Setup: warm Go 1.27.1, nproc 5, no cloud setup. npm and corpus installation logs are saved. Timing driver wall costs are in commands.jsonl. Audit elapsed time is recorded in completion.txt.']
(p/'REPORT.md').write_text('\n'.join(lines)+'\n')
