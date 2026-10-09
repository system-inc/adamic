import pathlib,json,re,statistics,subprocess,datetime,difflib
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-markdownblocks-array_growth_gap';scope=json.loads((out/'scope.json').read_text());menu=json.loads((out/'menu.json').read_text());timings=json.loads((out/'timings.json').read_text())
def events(p):
 a=[]
 for l in p.read_text().splitlines():
  try:a.append(json.loads(l))
  except:pass
 return a
def failed(p):
 a=events(p);cut=next((i for i,e in enumerate(a) if 'panic: test timed out' in e.get('Output','')),len(a));return [e['Test'] for e in a[:cut] if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']]
prodlogs=[out/(x+'.log') for x in ['M1','M2','M3','M4']]+list(out.glob('M1-alone-*.log'))+list(out.glob('M3-bounded-*.log'))+list(out.glob('M4-bounded-*.log'))
kills={}
for p in prodlogs:
 id=p.name[:2]
 for n in failed(p):kills.setdefault(n,set()).add(id)
def med(pattern):
 vals=[r['binary_seconds'] for r in timings if r['row']==pattern][-3:]
 return round(statistics.median(vals),3) if len(vals)==3 and all(v is not None for v in vals) else None
def evidence(p,n):
 a=events(p);outputs=[e.get('Output','').strip() for e in a if e.get('Test','').split('/')[0]==n]
 candidates=[s for s in outputs if re.search(r'_test.go:\d+:',s) and any(w in s for w in ['first byte difference',' mismatch ',' exit ','differs:','incorrect shard','invalid union accepted','existing case moved','preparation failed','planted disagreement','list mutant survived'])]
 s=candidates[-1] if candidates else next((s for s in reversed(outputs) if '--- FAIL' in s),'no completed failing assertion')
 return s[:1000]
def proof(members,ids):
 for id in reversed(ids):
  files=([out/(id+'.log')]+[p for p in prodlogs if p.name.startswith(id+'-')]) if id.startswith('M') else [out/(id+'.log')]
  for p in reversed(files):
   if not p.exists():continue
   for n in members:
    if n in failed(p):return id,p.name,evidence(p,n)
 return None,None,None
# The three leaf mutant-only shards have a different purpose, so they form a separate witness row.
def group(name,members,timing,oracle,kind,verdict=None,weak=None,probe=None):
 ks=sorted(set().union(*(kills.get(n,set()) for n in members))) if not verdict else []
 probes=[]
 for id in ['P_C',probe]:
  if id and (out/(id+'.log')).exists() and any(n in failed(out/(id+'.log')) for n in members):probes.append(id)
 ids=[weak] if weak else ks;id,log,line=proof(members,ids)
 r=dict(test=name,package='stage1/cohere/markdownblocks',file=scope['members'][members[0]]['file'],seconds=med(timing),oracle=oracle,oracle_kind=kind,kills=ks,unique_kills=[],last_proven_fail=f'{id}: {line}' if id else None,verdict=verdict,subsumed_by=[],mutants_in_matrix=0 if verdict else 4,probe_kills=probes,subsumer_seconds=None,vacuous=(False if probe and probe in probes else False if name=='TestArrayGrowthWitness' and 'P_C' in probes else None),bounded=True,matrix_rows=[],evidence=(f'{log}: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run <recorded selector>; {line}' if log else 'No completed production-mutant failure'),members=members)
 if weak:r['check_mutants']=[weak]
 return r
rows=[]
add=lambda *a,**kw:rows.append(group(*a,**kw))
add('TestArrayGrowthWitness',['TestArrayGrowthWitness'],'TestArrayGrowthWitness','Node executes array growth; exact known native/backend exit 70, panic text and empty stdout are self-pinned. M2 catches text, not growth semantics.',['external-run','self'],probe=None)
add('TestMarkdownASTPreprocessing',['TestMarkdownASTPreprocessing'],'TestMarkdownASTPreprocessing','Actual Go cohere AST preprocessing and pinned Prettier Node; complete serialized projections. M1 caught at the clean-exit gate before byte comparison.','external-run',probe='P_ast')
add('TestWholeDocumentOraclePreflight_Setup',['TestWholeDocumentOraclePreflight_Setup'],'TestWholeDocumentOraclePreflight_Setup','Self-defined SHA256 bucket ownership checked against constructed shards; Go/Node binaries prepared, no formatter parity asserted here.','self','setup-check','S1')
add('TestMicromarkInputChunks',['TestMicromarkInputChunks'],'TestMicromarkInputChunks','Actual Go cohere and pinned Prettier Node preprocessing; full UTF16 chunk bytes, not just counts.','external-run',probe='P_chunks')
add('TestMarkdownSourceDecoding',['TestMarkdownSourceDecoding'],'TestMarkdownSourceDecoding','Actual Go cohere and pinned Prettier Node decoding, exhaustive numeric codepoints and entity transport. M1 caught at the clean-exit gate.','external-run',probe='P_decode')
add('TestTokenizerEventShardUnion',['TestTokenizerEventShardUnion'],'TestTokenizerEventShardUnion','Self-defined exact case-id union; repeated, missing and out-of-range sets must be rejected.','self','setup-check','S2')
add('TestTokenizerEventShardPlantedDisagreement',['TestTokenizerEventShardPlantedDisagreement'],'TestTokenizerEventShardPlantedDisagreement','Self-planted byte disagreement must be detected once with owning shard and case id.','self','witness','W1')
add('TestTokenizerEventShardGrowth',['TestTokenizerEventShardGrowth'],'TestTokenizerEventShardGrowth','Self-defined stable per-name occurrence keys and bucket ownership after insertion.','self','setup-check','S3')
add('TestTokenizerEvents family',['TestTokenizerEventsUnion']+[f'TestTokenizerEvents_{i:03d}' for i in range(512)],'TestTokenizerEvents(Union|_[0-9]+)','Actual Go cohere and pinned Prettier Node tokenizer event streams; full points, token fields and serialization. Union is construction. 000-002 also contain local mutant checks; only production comparisons count here.','external-run',probe='P_chunks')
add('TestFrontMatterStage',['TestFrontMatterStage'],'TestFrontMatterStage','Actual Go cohere and pinned Prettier Node front-matter stage; all serialized fields and blanked content compared bytewise.','external-run',probe='P_front')
add('TestParserRepresentationProbes',['TestParserRepresentationProbes'],'TestParserRepresentationProbes','Positive outputs compared with Node; negative Refused/NotYet text and Node truth constants are self-written. C probe does not probe Lower refusal entry.',['external-run','self'],probe=None)
add('TestMdastIdentifierScalars',['TestMdastIdentifierScalars'],'TestMdastIdentifierScalars','Actual Go NormalizeIdentifier for all 1,112,064 Unicode scalars; complete bytes. M1 caught at the clean-exit gate.','external-run',probe='P_identifier')
add('TestMarkdownLeafComposition_Setup',['TestMarkdownLeafComposition_Setup'],'TestMarkdownLeafComposition_Setup','Own Go fixture preparation, poisoned width protocol, product construction and non-nil fixture guard; setup does not assert production output parity.','self','setup-check','S4')
add('TestMarkdownLeafCompositionUnion/PlantedDisagreement family',['TestMarkdownLeafCompositionUnion','TestMarkdownLeafCompositionPlantedDisagreement'],'TestMarkdownLeafComposition(Union|PlantedDisagreement)','Both wrappers call the identical helper: exact corpus ownership plus its own planted byte comparison. This is not the production layout comparator.','self','witness','W2')
add('TestMarkdownLeafComposition production family',[f'TestMarkdownLeafComposition_{i:03d}' for i in range(4)],'TestMarkdownLeafComposition_00[0-3]','Actual Go document layout, original Prettier document printer and full original Markdown parser/layout; poisoned width inputs, per-case bytes and terminators.','external-run',probe='P_leaf_valid')
add('TestMarkdownLeafComposition mutant witness family',[f'TestMarkdownLeafComposition_{i:03d}' for i in range(4,7)],'TestMarkdownLeafComposition_00[4-6]','Built-in source Node delimiter/reference/indentation mutants must differ from Go fixture output; W3 disables precisely this comparator.',['external-run','self'],'witness','W3')
by={r['test']:r for r in rows}
for r in rows:
 if r['verdict']:continue
 if r['test']=='TestArrayGrowthWitness':r['verdict']='sacred';r['unique_kills']=['M2']
 elif r['test']=='TestMarkdownLeafComposition production family':r['verdict']='sacred';r['unique_kills']=['M4']
 else:
  candidates=[s for s in rows if s is not r and not s['verdict'] in ['setup-check','witness'] and set(r['kills'])<=set(s['kills'])]
  if candidates:
   best=min(candidates,key=lambda s:s['seconds'] if s['seconds'] is not None else 10000)
   r['verdict']='subsumed';r['subsumed_by']=[best['test']];r['subsumer_seconds']=best['seconds']
   if best['seconds'] is None:r['subsumer_seconds_lower_bound']=90
  else:r['verdict']='cannot-judge'
for r in rows:
 r['matrix_rows']=sorted(set(n for p in prodlogs for e in events(p) if e.get('Action')=='run' for n in [e.get('Test','')] if n and '/' not in n))
 if r['test']=='TestTokenizerEvents family':r['seconds_lower_bound']=90;r['timing_status']='Three full-family runs cooked at 90s; median is censored, not invented.'
 if r['test']=='TestParserRepresentationProbes':r['unprobed_subcases']='All negative Lower Refused/NotYet branches; no empty Lower-entry probe.'
 if r['test']=='TestMarkdownLeafComposition production family':r['vacuous']=None;r['probe_limitation']='P_leaf invalid excluded; P_leaf_valid compiled but whole run cooked, no completed semantic probe result. P_C rejected.'
 if r['test']=='TestTokenizerEvents family':r['probe_limitation']='P_chunks used member 003; other members were not individually probed.'
(out/'rows.json').write_text(json.dumps(rows,indent=2))
# Validate every replay diff against the actual starting commit in a temporary source tree.
check=pathlib.Path('/tmp/u126/apply-check');check.mkdir(exist_ok=True)
for m in menu:
 p=check/m['file'];p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(subprocess.check_output(['git','show',(out/'base.txt').read_text().strip()+':'+m['file']],cwd=root))
valid=[]
for p in sorted((out/'diffs').glob('*.diff')):
 r=subprocess.run(['git','apply','--check',str(p)],cwd=check,capture_output=True,text=True);valid.append(dict(diff=p.name,exit=r.returncode,output=r.stdout+r.stderr))
(out/'apply-check.json').write_text(json.dumps(valid,indent=2));assert all(v['exit']==0 for v in valid)
# Build durations are logged by the content-addressed native/Go preparation helpers.
builds=[]
for p in sorted(out.glob('*.log')):
 for e in events(p):
  s=e.get('Output','').strip();m=re.search(r'build (.+?) (?:hit|miss) ([0-9.]+)$',s)
  if m:builds.append(dict(log=p.name,product=m[1],seconds=float(m[2])))
(out/'build-times.json').write_text(json.dumps(builds,indent=2))
summary=['u126: all 534 requested names exist at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.','16 judged rows after separating production and mutant-only leaf shards.','Four production mutants caught; bounded sacred rows: array gap and leaf composition.','Eight subsumed rows, four setup checks and three witness rows; see JSON for exact counts.','Evidence includes standalone diffs, native rebuild logs, probes, three-run costs and unknown outcomes.']
counts={v:sum(r['verdict']==v for r in rows) for v in set(r['verdict'] for r in rows)};summary[3]='Verdicts: '+', '.join(f'{n} {v}' for v,n in sorted(counts.items()))+'.'
report='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n'
report+='Mutants (all line numbers are against the starting origin/main commit):\n\n| Id | File:line | Change | Observed grouped failures |\n|---|---|---|---|\n'
for m in menu:
 if m['kind']=='production':
  affected=[r['test'] for r in rows if m['id'] in r['kills']]
  report+=f"| {m['id']} | {m['file']}:{m['line']} | {m['menu']}: `{m['from_']}` to `{m['to']}` | {', '.join(affected)} |\n"
report+='\nSurvivors: none after narrowed replays. A cooked run is unknown, not a survivor. No equivalent-candidate claim.\n\n'
report+='Unclear instructions, corrections and costs:\n\n'
notes=[
'The brief says 15 rows but supplies 534 names. Numbered leaf rows 000-003 execute production parity while 004-006 only execute built-in mutation witnesses. Applying the requested witness distinction gives 16 rows. Union and PlantedDisagreement leaf wrappers are identical calls and are one witness family. Event rows retain a shared production checker, with extra local witness branches in 000-002; their production kills are counted separately from those branches.',
'The cited 8de93800f4 is not the fetched origin/main. All source locations and diffs use ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2. events_shards_test.go and identifier_test.go moved to tokenizer_events_union_test.go and identifier_independent_test.go. No requested function vanished. The current whole package contains 763 top-level Tests.',
'The clean whole-package baseline and combined slice both exceeded the test-binary 90-second budget without ordinary assertion failures. They were narrowed. All isolated production rows passed cleanly. The complete 513-member event family cooked in all three timing attempts. Go test2json attributes a timeout failure to whichever row owns the panic stream (293 in the first attempt); this is not a red assertion baseline. The timing runner initially mislabeled it RED, which was corrected after reading the panic.',
'Full native preparation plus simultaneous large-corpus tests made all four initial bounded production matrices cook. Completed results remain evidence; uncompleted results remain unknown. M1 unknown production rows were replayed alone; M3 was narrowed to chunks and event members 000,001,002,255,511; M4 was narrowed to leaf members 000-003. M2 has one observed catch among completed rows, but unknown rows and all unassigned package rows prevent an unbounded uniqueness claim. Sacred/unique findings here are bounded observations for central replay.',
'There is no production environment-variable switch: four independent rebuilds follow the permitted stage1 fallback. Every compiler or port mutant used its own ADAMIC_BUILD_CACHE_DIR. The larger number of S/W/P patches are construction/witness/empty-answer checks, not additional production mutants. All standalone diffs apply to the actual starting commit. M1 go vet passed; all production diffs produced native products in their logs. Probe-generated invalid C from P_C is the intentionally empty compiler answer, not a production mutant.',
'One first M1 overlay binary accidentally included P_C because the compiler file was being probed concurrently. Those logs are quarantined in invalid-combined-overlay and excluded from every verdict. The binary was rebuilt after restoration and all six rows rerun with a fresh M1-valid cache. The first leaf early-return probe also broke TypeScript unreachable-code narrowing; it is explicitly invalid and excluded. The repaired whole-body probe compiled but cooked before a semantic result, so leaf vacuity is null.',
'The C empty-answer probe rejects missing compiler output at compilation. It does not measure Lower refusal-entry vacuity. Negative representation probes were not empty-probed and remain unmeasured. The event inputChunks probe demonstrates rejection for member 003 only; other event members are unknown under this probe. Witness and construction rows have vacuity null.',
'Clean timing runs used copied unmodified ports and a precompiled clean binary after the initial decoder runs. The timings are real binary Elapsed values, using the last three clean-shadow records where repeated original records exist. Compiler input files were shared read-only symlinks; changing their cache fingerprints caused some fresh product builds during clean semantic runs. Setup timing is consequently cold/warm mixed and inflated (76.449,38.329,6.255s), not a steady warmed throughput claim. No elapsed shell or compilation time was substituted for binary time.',
'Several M1 catches are clean-exit gate failures (native panic), while others are byte mismatches. Oracles subsequently compare complete output on successful execution; no count-only output oracle is claimed. Array-growth M2 proves the diagnostic contract, not a repaired array-growth implementation. Leaf Union/PlantedDisagreement exercises its own local comparator, not the production layout comparator.',
'Assigned rows have no default opt-in skips and no skip events were observed. ADAMIC_MARKDOWNBLOCKS_CENSUS changes corpus breadth rather than enabling a skipped row, so the default corpus was retained. Unassigned width/dependency opt-ins were not installed or audited. Stage3/api npm ci was run before baseline; exact npm duration was not separately captured. Warm env.sh worked, so toolchain setup was skipped (0s); nproc is 5.',
'Reading the large helper files, native cold lowering, cooked matrices, isolated family timing attempts, and correcting contaminated evidence exceeded the approximate 30-minute port budget. The audit ran approximately 36 minutes, with overlapping commands. Logged build-times.json sums '+str(round(sum(x['seconds'] for x in builds),2))+' seconds of build-helper durations; these overlap and are not exclusive wall time. Whole-package baseline binary time was 90.097s and combined-slice 90.278s. timings.json, matrix.json, port-reruns.json, reruns.json and probe records retain each run and rebuild observation.',
'Not covered: repo-wide replay, all unassigned package rows, every event-family member under every mutant, exact dynamic reachability of every compiler helper, empty Lower/refusal probes, and a completed semantic empty-leaf probe. port-function-inventory.json lists the static import closure; compiler-function-inventory.txt is a conservative declaration inventory rather than measured coverage. No unsupported untrue or equivalence verdict was invented.'
]
report+='\n\n'.join(notes)+'\n\nCommands, complete failures, observed members and caches are in the adjacent JSON/log files. The invalid evidence directory is deliberately retained for provenance. Production sources were restored before commit; no PR or main push.\n'
(out/'REPORT.md').write_text(report);(out/'summary.txt').write_text('\n'.join(summary)+'\n');print(counts);print('rows',len(rows),'apply checks',len(valid),'build helper seconds',round(sum(x['seconds'] for x in builds),2))
