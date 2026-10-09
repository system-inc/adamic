import json,pathlib,re,statistics,subprocess
R=pathlib.Path('/workspace/adamic');E=pathlib.Path('/tmp/u159/evidence');P='stage1/typescript/scanner';base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip()
def events(id):return [json.loads(l) for l in (E/(id+'.log')).read_text().splitlines() if l.startswith('{')]
def median(n):
 a=[next(e['Elapsed'] for e in events('timing-'+n+'-'+str(i)) if e['Action']=='pass' and 'Test' not in e) for i in (1,2,3)];return statistics.median(a)
top=[l for l in (E/'list.log').read_text().splitlines() if l.startswith('Test')]
A='TestScannerAgreesWithTypescriptGo family';B='TestProduct_Scanner family'
def group(t):return A if t.startswith('TestScannerAgreesWithTypescriptGo_') else B if t.startswith('TestProduct_Scanner') else t
groups={g:[t for t in top if group(t)==g] for g in dict.fromkeys(map(group,top))}
matrix={id:{g:('fail' if any(e['Action']=='fail' and e.get('Test') in mm for e in events(id)) else 'pass' if all(any(e['Action']=='pass' and e.get('Test')==t for e in events(id)) for t in mm) else 'unknown') for g,mm in groups.items()} for id in ['M1','M2','M3','M4']}
assert all(x!='unknown' for m in matrix.values() for x in m.values());assert len(top)==29 and len(groups)==8
(E/'matrix.json').write_text(json.dumps(dict(starting_commit=base,rows=groups,primary=matrix,invalid_runs=['M1-invalid','M2-invalid','M3-invalid','P1-invalid','P1-partial-invalid'],bounded=False),indent=2))
files={}
for p in (R/P).glob('*_test.go'):
 for m in re.finditer(r'^func (Test\w+)\(',p.read_text(),re.M):files[m[1]]=str(p.relative_to(R))+':'+str(p.read_text()[:m.start()].count('\n')+1)
secs={'TestGapStandsWhereGapsMdSays':median('gappush'),'TestBigintGapStandsWhereGapsMdSays':median('gapbigint'),B:median('products'),A:median('agreement-only'),'TestScannerShardCoverage':median('coverage'),'TestPerformance':median('performance'),'TestProfileArtifacts':median('artifacts'),'TestProfileSnapshotsAgree':median('snapshots')}
oracles={'TestGapStandsWhereGapsMdSays':('Node runs the fixture and returns ab; lower.NotYet and exact What label are self-written. Location and execution of lowered output are not checked.',['external-run','self']),'TestBigintGapStandsWhereGapsMdSays':('Node returns 1237940039285380274899124223; lower.NotYet and exact What label are self-written. Location is not checked.',['external-run','self']),B:('Self: build recipes must complete without t.Fatal. Return paths are discarded, with no product existence or behavior assertion.','self'),A:('Pinned TypeScript-Go scanner, compared byte for byte to native and Node-executed port. Includes separate built-in mismatch witnesses, verified by W1.','external-run'),'TestScannerShardCoverage':('Self: corpus union, exactly-once ownership, mutant ownership and one planted mismatch. Git supplies tracked inputs, not expected scanner semantics.','self'),'TestPerformance':('TypeScript-Go token count compared with native and Node counts. Only total count is asserted; token kinds, values, offsets and speed have no assertion. M1 and M3 survive this row.','external-run'),'TestProfileArtifacts':('Self: file writes and compiler command success. S2 proves failed profile construction is detected; artifact behavior is checked in another row.','self'),'TestProfileSnapshotsAgree':('Pinned TypeScript-Go full bytes compared to release, profiled and Node port across all 18791 cases. Artifact directories freshly rebuilt for each mutant.','external-run')}
construct={B:'S1','TestProfileArtifacts':'S2','TestScannerShardCoverage':'W1'}
probe={A:'P1','TestPerformance':'P1','TestProfileSnapshotsAgree':'P1','TestGapStandsWhereGapsMdSays':'P2','TestBigintGapStandsWhereGapsMdSays':'P2',B:'P3','TestScannerShardCoverage':'P4'}
subs={A:'TestProfileSnapshotsAgree','TestProfileSnapshotsAgree':A,'TestGapStandsWhereGapsMdSays':'TestBigintGapStandsWhereGapsMdSays','TestBigintGapStandsWhereGapsMdSays':'TestGapStandsWhereGapsMdSays'}
def line(id,members):
 filt=lambda e:e.get('Test') in members and e['Action']=='output'
 outs=[e['Output'].strip() for e in events(id) if filt(e)]
 for text in outs:
  if any(k in text for k in ['gap changed:','native: line','release: line','native count','preparation failed','planted failure caught by','union 0','unknown argument:']):return text
 return next((x for x in outs if '--- FAIL:' in x),'<no failure>')
runs=json.loads((E/'runs.json').read_text());result=[]
for g,members in groups.items():
 kills=[id for id,m in matrix.items() if m[g]=='fail'];unique=[id for id in kills if sum(x=='fail' for x in matrix[id].values())==1]
 last=construct.get(g,kills[-1] if kills else None);v='setup-check' if g in construct else 'sacred' if unique else 'subsumed'
 idrun=next(x for x in reversed(runs) if x['id']==last)
 p=probe.get(g);pk=p and any(e['Action']=='fail' and e.get('Test') in members for e in events(p))
 obj=dict(test=g,package=P,file=files[members[0]],seconds=secs[g],oracle=oracles[g][0],oracle_kind=oracles[g][1],kills=kills,unique_kills=unique,last_proven_fail=last+': '+line(last,members),verdict=v,subsumed_by=[subs[g]] if g in subs else [],mutants_in_matrix=4,probe_kills=[p] if pk else [],subsumer_seconds=secs[subs[g]] if g in subs else None,vacuous=(not pk) if p else None,bounded=False,matrix_rows=[],evidence=idrun['command']+' > '+last+'.log 2>&1; '+line(last,members))
 if len(members)>1:obj['members']=members
 if g in construct:obj['construction_kills']=[construct[g]]
 if g==A:obj['witness_evidence']='W1: scanner_products_test.go:356: source scanned as Identifier Node: mutant survives comparison; weakening difference also fails the other three built-in witness owners.'
 result.append(obj)
(E/'rows.json').write_text(json.dumps(result,indent=2))
menu=json.loads((E/'menu.json').read_text());table=[]
for id,path,old,new,kind in menu:
 if not id.startswith('M'):continue
 orig=subprocess.check_output(['git','show',base+':'+path],cwd=R,text=True);lineno=orig[:orig.index(old)].count('\n')+1
 table.append('| '+id+' | '+path+':'+str(lineno)+' | `'+old.strip()+'` → `'+new.strip()+'` | '+', '.join(g for g in groups if matrix[id][g]=='fail')+' |')
tot=sum(r['wall'] for r in runs);primary=[next(r for r in reversed(runs) if r['id']==id) for id in matrix]
summary='Unit u159: 29 top-level tests, grouped into 8 rows at '+base[:12]+'.\nEnabled baseline passed in 34.372 seconds; no skips; nproc 5.\nFour primary mutants were caught: 1 sacred row, 4 subsumed hints and 3 setup-check rows.\nProduct family passed its empty-path probe; semantic rows failed their own probes.\nAll source edits restored; standalone diffs, raw logs and matrices accompany this report.\n'
friction='''The brief requires a clean baseline and scratch source mutations, but the repository corpus collector rejects dirty tracked TypeScript files. M1/M2/M3 first failed this guard. Those failures establish no semantic kill. Temporary local commits solved it without changing the collector; their hashes are saved. Each run was reset back to the starting commit before the next edit. Central replay needs a clean committed variant too, or the same guard will mask semantic results.

An unconditional early return in run made the remaining TypeScript body unreachable and disabled its flow narrowing, producing TS2339 rather than an empty answer. Removing the whole body fixed typing but removed a built-in regex mutation site, causing an unrelated preparation failure. Both attempts are retained as invalid. The final P1 uses an always-satisfied path.length >= 0 guard and preserves all mutation sites. Its product builds pass and semantic comparisons fail.

Family rules overlap with setup rules. All sixteen agreement wrappers call the same scannerShard checker, so they are one family. Coverage has its own body and distinct assertions, so it is a separate setup-check. Seven product wrappers share scannerProductTime with different build recipes, so they form one construction family. The initial agreement timings included Coverage; corrected family-only timings are saved and used.

The instruction to enable installable opt-ins required supplying a pinned TypeScript checkout, enabling the benchmark, creating profile artifacts and giving the snapshot row their directory. The deliberate ADAMIC_SCANNER_PLANT_FAILURE option is intentionally failing behavior, not a skipped row, and was not enabled for the baseline. Artifact directories are rebuilt for each mutation before snapshot execution; a clean snapshot directory reused under mutations would silently test the wrong executable.

The benchmark has a useful unique count guard but no performance threshold. Its name alone does not establish a speed regression gate. It passes both M1 and M3 even though full-answer rows show changed semantics. Conversely, the gap rows assert the same diagnostic constructor option and share only one measured kill. Their mutual subsumption rests on one mutant, not an argument to delete either. Agreement/snapshot subsumption rests on two mutants.

The product family detects a dropped build operation, yet every wrapper passes when scannerFetch returns an empty path. Its construction guard therefore does not prove that a usable product was returned. The profile-artifact row was checked by invalidating its profile compiler option; its empty-answer status is null because no single construction-entry empty probe was run for that row. This is an explicit coverage limit, not an inferred non-vacuity claim.

The desired three mutants per row conflicts with the four-rebuild limit for ports. Four primary changes were fixed before their results and spread over advance, run, contains and notYet. Construction edits and comparison weakening have separate S/W identifiers; entry probes have P identifiers. None contributes to primary uniqueness or subsumption. The single-switch suggestion was not used because each port mutation required a separate fresh native build; the explicit four-mutant rebuild allowance was followed.

Timing compiler preparation, semantic execution and external process work separately is not directly supplied by the brief's Go package line. runs.json records command wall time, and the requested medians use the test binary's own elapsed line. Buildcache miss timings and profile-build elapsed events remain in raw JSON. Their parallel durations must not be summed as a package wall time. Native products and the fetched corpus were not pushed; reproducible source diffs, commands and logs were.
'''
validation='''Every primary matrix ran all 29 top-level tests with the benchmark and profile options enabled; all rows have observed pass/fail results. No run exceeded 90 seconds, no Go panic truncated the matrix, and no row skipped. M1/M2/M3 compiled through the actual sanitized/release native recipes in their corrected whole-package runs; M4 passed go vet ./internal/lower/ and native builds. P/S/W Go changes passed go vet for their mutated package. Diffs apply to the starting origin/main commit. No other packages were audited, and repo-wide uniqueness remains for central replay. The port functions and scanner-package helpers are listed in function-inventory.txt; lower.cover/lower-functions.log list actual gap reaches. The entire transitive compiler graph used as port build preparation was not separately inventoried or mutated.
'''
report=summary+'\n```json\n'+json.dumps(result,indent=2)+'\n```\n\n| ID | Starting file:line | Change | Rows failed |\n|---|---|---|---|\n'+'\n'.join(table)+'\n\nSurvivors: none in the whole-package primary matrix. M1 and M3 survive the count-only benchmark but are caught by full-answer checks. No equivalent candidates.\n\n'+friction+'\nWarm tool setup was skipped (0 setup-script seconds); npm ci reports 460 ms. Timed commands total '+str(round(tot,3))+' wall seconds, including invalid and repeated runs but excluding initial fetch/npm/baseline and separately unmeasured vet commands. The baseline adds 34.372 binary seconds. Corrected primary run wall times: '+', '.join(r['id']+' '+str(round(r['wall'],3))+' s' for r in primary)+'. Their profile construction test elapsed times: '+', '.join(id+' '+str(next(e['Elapsed'] for e in events(id) if e['Action']=='pass' and e.get('Test')=='TestProfileArtifacts'))+' s' for id in matrix)+'. Total session about 18 minutes.\n\n'+validation
(E/'REPORT.md').write_text(report)
print(summary);print('seconds',secs);print('runtime wall',tot)
