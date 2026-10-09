import json,re,statistics
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-tsprinter-shards')
plan=json.loads((p/'mutation-plan.json').read_text());rowplan=json.loads((p/'rows-plan.json').read_text());scope=(p/'scope.txt').read_text().splitlines()
def events(id):
 out=[]
 for line in (p/f'{id}.log').read_text().splitlines():
  try:out.append(json.loads(line))
  except ValueError:pass
 return out
def members(regex):return [n for n in scope if re.search(regex,n)]
def failures(id):return [e['Test'] for e in events(id) if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']]
def caught(regex,id):return bool(set(members(regex))&set(failures(id)))
def failure(regex,id):
 failed=set(failures(id));messages=[]
 for e in events(id):
  test=e.get('Test','').split('/')[0];out=e.get('Output','')
  if test not in failed or not re.search(regex,test):continue
  if e.get('Action')=='output' and re.search(r'\.go:\d+: ',out):
   msg=re.sub(r'^\s*[^\n]*?\.go:\d+: ','',out).splitlines()[0]
   if any(w in msg for w in ['want','accepted','transport:','line 1:','assignment','selected','caught','repeated']):messages.append(msg)
 return messages[-1] if messages else 'See complete failure output in '+id+'.log'
files=['shards_test.go']*3+['statements_printer_units_test.go']+['statements_shards_test.go']*4+['statements_printer_units_test.go','tsc_printer_units_test.go','tsc_corpus_split_test.go','tsc_units_test.go','tsc_units_test.go']
proofs=['S1','W1','S2',None,'S3','S4','W2','W3','M4','M3','W4','W4','W5']
oracles=['Handwritten largest-first assignment.','Handwritten missing, extra, and unterminated-answer rejection messages.','Node runs a custom echo child; expected bytes and ordering are handwritten.','Successful preparation only; S5 returned main.ts as the statement port and this row still passed.','Handwritten stability under corpus growth and duplicate-key rejection.','Handwritten exact-once selection and invalid-selector rejection.','Handwritten planted owner, failure count, exit status, and case identity; a positive Node port run is a prerequisite.','Handwritten rejection of missing, repeated, extra IDs and wrong shard count.','Go cohere supplies expected bytes; npm/embedded Prettier checks exact outcomes with pinned differences. Node port, native, and backend compare against Go. Gap refusal labels are handwritten. Leak checks also run.','Go cohere supplies expected bytes. Node port, native, and backend check exact stdout, stderr and exit status; leak checks also run.','Go-derived positive answers plus the handwritten exactly-one-owner planted failure invariant.','Handwritten positive answers and exactly-one-owner planted failure invariant.','Handwritten rejection of missing and repeated IDs.']
rows=[]
for index,((name,regex,kind),file,proof,oracle) in enumerate(zip(rowplan,files,proofs,oracles)):
 samples=[float(re.findall(r'\t([\d.]+)s',(p/f'timing-{name.replace(" ","_")}-{i}.log').read_text())[-1]) for i in [1,2,3]]
 kills=[id for id in ['M1','M2','M3','M4'] if kind=='production' and caught(regex,id)]
 if kind=='setup-check':kills=[id for id in ['S1','S2','S3','S4','S5'] if caught(regex,id)]
 verdict='sacred' if kind=='production' else 'untrue' if proof is None else kind
 unique=[id for id in kills if id in ['M1','M4','S1','S3','S4']]
 probe=[id for id in ['P1','P2'] if kind=='production' and caught(regex,id)]
 metadata=json.loads((p/f'{proof or "S5"}.json').read_text());selector=metadata['command'][-1]
 matrix_rows=[n for n,r,k in rowplan if any(re.search(selector,m) for m in members(r))]
 text=failure(regex,proof) if proof else 'S5: setup passed with the wrong returned port path; the statement family caught it.'
 oracle_kind=['external-run','self'] if index in [2,8,10] else 'external-run' if index==9 else 'self'
 rows.append(dict(test=name,package='stage1/cohere/tsprinter',file='stage1/cohere/tsprinter/'+file,seconds=statistics.median(samples),oracle=oracle,oracle_kind=oracle_kind,kills=kills,unique_kills=unique,last_proven_fail=(proof+': '+text) if proof else None,verdict=verdict,subsumed_by=[],mutants_in_matrix=9 if kind=='production' or index==3 else 14,probe_kills=probe,subsumer_seconds=None,vacuous=False if kind=='production' else None,bounded=True,matrix_rows=matrix_rows,evidence=f'run-matrix.py [{proof or "S5"}]; command in {proof or "S5"}.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; '+text,samples=samples,members=members(regex),witness_kills=[proof] if kind=='witness' else []))
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
matrix=[]
for item in plan:
 id=item['id'];m=dict(item);m['failed_functions']=failures(id);m['failed_rows']=[n for n,r,k in rowplan if caught(r,id)];m['metadata']=json.loads((p/f'{id}.json').read_text());matrix.append(m)
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
(p/'probes.json').write_text(json.dumps([dict(id=id,failed_functions=failures(id),metadata=json.loads((p/f'{id}.json').read_text())) for id in ['P1','P2']],indent=2)+'\n')
(p/'skipped.json').write_text(json.dumps([e.get('Test') for e in events('baseline-bounded') if e.get('Action')=='skip'],indent=2)+'\n')
builds={id:json.loads((p/f'build-{id}.json').read_text()) for id in ['M1','M2','M3','M4']};assert all(v['exit']==0 for v in builds.values())
summary='Unit u141 at a7448d73cd17f16362b6cbc5c5c111080da64e43: 37 requested functions, 13 live grouped rows.\nCorrected whole baseline timed out; bounded baseline passed in 38.952 test seconds without skips.\nBounded verdicts: two sacred families, four setup-checks, six witnesses, one untrue setup row.\nFour port mutants, five construction mutations, five weakened checks, and two separate probes are recorded.\nSource restored; evidence pushed on the requested audit branch without a PR or main push.\n'
table='| ID | Origin/main location | Change | Observed failing rows |\n|---|---|---|---|\n'
for m in matrix:
 change=m['new'].replace('\n',' ') if len(m['new'])<200 else 'Bypass checker with return nil'
 table+=f"| {m['id']} | {m['file']}:{m['line']} | {change} | "+', '.join(m['failed_rows'])+' |\n'
notes='''Survivors: none among the planted mutations in the observed bounded runs. S5 passed the setup row but failed the statement family, so it is a row-level blind spot rather than a globally surviving mutation.

Brief issues and costs

The intro says 15 rows. The 38 supplied functions group into 14 rows under the family rules with witnesses separate and unions grouped with their agreement families. TestStatementsSetupSelection vanished, leaving 13 live rows. Statement wrappers, union and setup moved to statements_printer_units_test.go; TSC wrappers moved to tsc_printer_units_test.go. Each member is listed in rows.json. All mutant locations refer to starting origin/main a7448d73, not the historical 8de93800f4.

The keeper's TypeScript requirement was satisfied before the baseline: a git checkout at 050880ce59e30b356b686bd3144efe24f875ebc8. The first baseline nevertheless hit a setup error because the oracle-input hash refuses npm's .bin/prettier symlink. npm ci --no-bin-links repaired it. No mutation audit ran on that setup-red baseline. The corrected whole run timed out at 90 seconds without a preceding row failure. The bounded baseline passed all 37 surviving requested functions without skips.

Production witness failures are not verdict-bearing catches. M1 broke two witness positive preconditions. M2/M3/probes also broke some prerequisites. The matrix table records those observations; only W1 through W5 determine the six witness verdicts. M1 is unique among applicable production agreement rows, despite two witness precondition failures. Its unique_kills entry is not a claim that exactly one top-level function failed. Both sacred verdicts are bounded, and package-wide uniqueness remains unknown. M4 failed only the statement family after members were grouped.

Setup mutations are separate from port changes. S1, S2, S3 and S4 prove four setup-check verdicts. S5 changes statementPrepareCommon's returned Port from statementsMain.ts to main.ts. The setup row only calls preparation, so it still passes; the statement family rejects the wrong program. This supports untrue for that row under these construction attempts, not a claim that every preparation error is invisible or that the row should be deleted.

Probes target formatExpression and formatFile, the formatter entries used by the two drivers. They return Ok with empty text while the drivers still emit protocol rows. Both families reject their applicable probe answers. Probe IDs never support sacred, subsumed, or witness verdicts. Other rows were not probed on their own construction entries and have vacuous=null. Probe diffs are separate from mutant diffs.

M4's exact standalone flip first failed TS2367: the early return narrowed parent to '**', making a later '%' comparison disjoint. The switched version had avoided the narrowing. The final one-line diff compares parent.slice(0), an identity operation on its string argument, to prevent this typing issue without inserting a statement or changing the intended condition flip. The native build passed. Its first standalone matrix timed out while cold-building statement products and external oracle observations: fourteen statement shards failed and all eight TSC shards passed before the timeout. The final narrowed warm replay supplies completed family observations. Initial diagnostic, switched run, cooked run, and final replay logs are all retained. The repair was driven by the compiler diagnostic, not by which row caught M4.

W1 replaces the whole merger body and removes its now-unused bytes import. W2 through W5 replace checker bodies with return nil. These are the explicitly allowed witness weakenings. Every standalone Go diff applies to origin/main and passes go vet. Every final standalone port diff compiled through sanitized and release native product builds. No compile-warning kill supports a verdict. Source and helper instrumentation were restored afterward, verified against origin/main.

The function inventory is a conservative transitive declaration and helper-call inventory, not exact per-case dynamic coverage. It includes candidate methods whose precise case reachability was not resolved. Mutant sites were chosen from that source inventory before observing catches. The port switch reads a selector file at module initialization; both switched native products were built once and reused across selectors. No compiler or external oracle was mutated. No ADAMIC_NATIVE_SPLIT cache-key assumption was used. The neutral switched run passed all requested rows after refreshing external observations for changed self-corpus inputs.

The requested remote branch already contains keeper attempt aabd7a7d. Its history is preserved by a merge, not a force push. None of its timings or results support this audit. Raw logs are force-added because repository ignore rules exclude .log files. No PR or main push was made.

Timing and omissions

'''
notes+='Warm setup skipped; nproc=5. npm wall seconds: '+', '.join(f'{name} {(p/(name+".time")).read_text().strip()}' for name in ['npm-api','npm-prettier','npm-prettier-no-bin'])+'.\n'
notes+='Whole baseline command wall seconds: '+(p/'baseline.time').read_text().strip()+', corrected '+(p/'baseline-fixed.time').read_text().strip()+'. Bounded baseline '+(p/'baseline-bounded.time').read_text().strip()+'s wall, 38.952s test time.\n'
notes+='Thirty-nine timing command wall seconds: '+str(round(sum(json.loads(f.read_text())['seconds_wall'] for f in p.glob('timing-*.json')),3))+'. Switched product build '+(p/'build-switch.time').read_text().strip()+'s wall, 50.539s test time.\n'
notes+='Standalone native validation wall seconds: '+', '.join(f"{id} {v['seconds_wall']:.3f}" for id,v in builds.items())+'. First rejected M4 build: 4.038s.\n'
notes+='Neutral and final mutation/probe commands wall seconds: '+str(round(sum(json.loads((p/f'{id}.json').read_text())['seconds_wall'] for id in ['neutral','M1','M2','M3','M4','P1','P2','S1','S2','S3','S4','S5','W1','W2','W3','W4','W5']),3))+'. Cooked M4 and initial switched M4 timings are separate in their JSON files.\n'
notes+='Whole-package and repository-wide uniqueness, exact dynamic function coverage, and other construction-entry probes remain uncovered. No other package was used as an audit target. The whole package did not complete within 90 seconds. Outside-slice kills are unknown. M4 narrowing is recorded in its command selector; other construction/port matrices ran the 37 requested functions, and weakened checks ran the ten fast non-family/non-setup rows. No deletion is recommended.\n'
(p/'report.md').write_text(summary+'\n'+json.dumps(rows,indent=2)+'\n\n'+table+'\n'+notes)
print(summary)
