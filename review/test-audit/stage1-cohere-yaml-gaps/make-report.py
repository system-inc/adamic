import pathlib,json,subprocess,statistics,re,shutil
p=pathlib.Path('/tmp/u152');root=pathlib.Path('/workspace/adamic');dest=root/'review/test-audit/stage1-cohere-yaml-gaps';dest.mkdir(parents=True,exist_ok=True)
base=(p/'base.txt').read_text().strip();pkg='stage1/cohere/yaml';prod=['TestLexerGaps','TestStructuralPositionRefusal','TestClosed family','TestLexerMatchesGo','TestPropsMatchGo','TestSharedSliceAppendMatchesNode']
closed=['TestClosedStringPresenceGap','TestClosedLexerGaps','TestClosedValuePresenceGap'];rows=prod[:4]+['TestLexerMutants',prod[4],'TestPropsMutants',prod[5],'TestScalarMutants'];timings=json.loads((p/'timings.json').read_text());med={r:round(statistics.median(x['seconds'] for x in timings if x['test']==r),3) for r in dict.fromkeys(x['test'] for x in timings)};med['TestClosed family']=med.pop('TestClosed gaps family')
runs={x['id']:x for x in json.loads((p/'runs.json').read_text())}
def events(id):
 es=[]
 for s in (p/(id+'.log')).read_text().splitlines():
  try:es.append(json.loads(s))
  except ValueError:pass
 return es
def grouped(name):return 'TestClosed family' if name.split('/')[0] in closed else name.split('/')[0]
matrix={}
for id in ['M1','M2','M3']:
 es=events(id);matrix[id]={r:'unknown' for r in prod}
 for e in es:
  if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail']:
   row=grouped(e['Test']);status=e['Action']
   if row in prod and matrix[id][row]!='fail':matrix[id][row]=status
matrix['M4']={r:('over-budget' if r in ['TestLexerMatchesGo','TestPropsMatchGo'] else 'pass') for r in prod}
assert all(v!='unknown' for id,m in matrix.items() for v in m.values())
oracles={
'TestLexerGaps':('Node output is pinned to 2\\n; NotYet type and diagnostic substring are self-written compiler-limit pins.',['external-run','self']),
'TestStructuralPositionRefusal':('Node output is pinned to 1\\n; Refused type and Span[] diagnostic substring are self-written refusal pins. M1 is rejected for returning a different diagnostic class.',['external-run','self']),
'TestClosed family':('Live Node stdout, checked against hand-written fixture pins, is compared byte-for-byte with sanitized native and emitted JavaScript stdout. M2 fails valueConjunction.ts.',['external-run','self']),
'TestLexerMatchesGo':('Live Go cohere lexer, Node port source, emitted JavaScript and yaml@2.9.0 exact-byte agreement. The completed M1 kill is a lowering precondition failure; the port mutant M4 times out before comparison.','external-run'),
'TestPropsMatchGo':('Live Go cohere private resolver through an export overlay, Node port source, emitted JavaScript and yaml@2.9.0 exact-byte agreement. M1 fails lowering; M3 triggers a sanitized native exit before byte comparison.','external-run'),
'TestSharedSliceAppendMatchesNode':('Live Node stdout pinned to a\\nx\\n, compared byte-for-byte against native and sanitized native at offsets 0 and 48. M3 changes native stdout to a\\n\\n.',['external-run','self']),
'TestLexerMutants':('Witness compares built-in port-mutant native and Node bytes against live Go cohere lexer output; W1 makes the comparison always agree and all five subcases fail.','external-run'),
'TestPropsMutants':('Witness compares built-in port-mutant native and Node bytes against live Go cohere resolver output; W2 makes the comparison always agree and all three subcases fail.','external-run'),
'TestScalarMutants':('Witness compares built-in port-mutant native and Node bytes against live Go cohere scalar output; W3 makes the comparison always agree and all three subcases fail.','external-run')}
files={'TestLexerGaps':'gaps_test.go','TestStructuralPositionRefusal':'gaps_test.go','TestClosed family':'gaps_test.go','TestLexerMatchesGo':'lexer_test.go','TestLexerMutants':'lexer_test.go','TestPropsMatchGo':'props_test.go','TestPropsMutants':'props_test.go','TestSharedSliceAppendMatchesNode':'scalar_runtime_gap_test.go','TestScalarMutants':'scalar_test.go'}
verdicts={'TestLexerGaps':('subsumed',['TestStructuralPositionRefusal']),'TestStructuralPositionRefusal':('subsumed',['TestLexerGaps']),'TestClosed family':('sacred',[]),'TestLexerMatchesGo':('subsumed',['TestPropsMatchGo']),'TestPropsMatchGo':('overlapping',['TestStructuralPositionRefusal','TestSharedSliceAppendMatchesNode']),'TestSharedSliceAppendMatchesNode':('subsumed',['TestPropsMatchGo'])}
lastids={'TestLexerGaps':'M1','TestStructuralPositionRefusal':'M1','TestClosed family':'M2','TestLexerMatchesGo':'M1','TestPropsMatchGo':'M3','TestSharedSliceAppendMatchesNode':'M3','TestLexerMutants':'W1','TestPropsMutants':'W2','TestScalarMutants':'W3'}
needles={'TestLexerGaps':'gap changed or closed:','TestStructuralPositionRefusal':'refusal changed or closed:','TestClosed family':'native ASan/UBSan/LSan: "TRUE','TestLexerMatchesGo':'lexer_test.go:203:','TestPropsMatchGo':'props_test.go:101:','TestSharedSliceAppendMatchesNode':'native "a\\n\\n"','TestLexerMutants':'native missed mutant','TestPropsMutants':'native missed mutant','TestScalarMutants':'native missed mutant'}
report=[]
for row in rows:
 id=lastids[row];file=files[row];source=subprocess.check_output(['git','show',base+':'+pkg+'/'+file],cwd=root,text=True);member=closed[0] if row=='TestClosed family' else row;line=next(i for i,s in enumerate(source.splitlines(),1) if s.startswith('func '+member+'('))
 fails=[e['Output'].strip() for e in events(id) if needles[row] in e.get('Output','')];assert fails,(row,needles[row]);failure=fails[0]
 if row=='TestScalarMutants':failure=failure.replace('scalar_test.go:153:','scalar_test.go:154 (origin; log line 153):')
 if row=='TestPropsMatchGo':failure+='; ERROR: AddressSanitizer: heap-buffer-overflow'
 witness=id.startswith('W');kills=[] if witness else [m for m in ['M1','M2','M3'] if matrix[m][row]=='fail'];unique=[] if witness else [m for m in kills if sum(v=='fail' for v in matrix[m].values())==1];verdict,subsumers=('witness',[]) if witness else verdicts[row]
 probes=[] if witness else ['P2' if row=='TestLexerMatchesGo' else 'P3' if row=='TestPropsMatchGo' else 'P1']
 obj={'test':row,'package':pkg,'file':f'{pkg}/{file}:{line}','seconds':med[row],'oracle':oracles[row][0],'oracle_kind':oracles[row][1],'kills':kills,'unique_kills':unique,'last_proven_fail':id+' '+failure,'verdict':verdict,'subsumed_by':subsumers,'mutants_in_matrix':1 if witness else 4,'completed_production_mutants':0 if witness else 3,'probe_kills':probes,'subsumer_seconds':med[subsumers[0]] if verdict=='subsumed' else None,'vacuous':None if witness else False,'bounded':True,'matrix_rows':[row] if witness else prod,'evidence':f'ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/{id} '+runs[id]['command']+' > '+id+'.log 2>&1; '+failure,'nproc':5,'unknown_mutants':['M4'] if row in ['TestLexerMatchesGo','TestPropsMatchGo'] else []}
 if witness:obj['witness_kills']=[id]
 if row=='TestClosed family':obj['members']=closed
 report.append(obj)
(p/'report.json').write_text(json.dumps(report,indent=2)+'\n');(p/'matrix.json').write_text(json.dumps({'base':base,'rows':prod,'statuses':matrix,'outside_slice':'unknown','timeouts_are_kills':False},indent=2)+'\n')
with (p/'matrix.csv').open('w') as out:
 out.write('row,M1,M2,M3,M4\n')
 for r in prod:out.write(r+','+','.join(matrix[m][r] for m in ['M1','M2','M3','M4'])+'\n')
metadata=[('M1','internal/lower/object.go',1265,'flip len(arguments) != 1 to == 1 in the push arity guard'),('M2','internal/native/runtime/string_build_impl.h',131,'change true text constant from true to TRUE'),('M3','internal/native/runtime/string_append.c',86,'drop result->length = written'),('M4',pkg+'/lexer.ts',21,'change flow indicator comma code 44 to hyphen code 45')]
notes='''Observed limits and brief friction:

1. Current origin/main is 3bf0a5d9e74d197a38f61982e43b2d791f17b563, newer than the brief's 8de93800f4. All 11 requested functions remain in their listed files. Discovery is saved; no moved or vanished row.
2. The three TestClosed wrappers differ only in inputs to closedGap. They form one TestClosed family with nine fixture inputs, reducing 11 functions to nine rows. Its three timing runs select all three members together.
3. The first whole-package run used the specified outer 120-second backstop and exited 124 without any observed failed test. TestMain ran a setup child for 41.877 seconds before m.Run, so the outer clock cut off the suite before its own 90-second test clock finished. The initial run skipped two out-of-scope library rows, TestComposeMatchGo and TestCSTMatchesGo. The enabled slice baseline and restored baseline had no skips and no failures. Kills outside the slice remain unknown.
4. ADAMIC_YAML_LIBRARY is required. LexerMatchesGo skips after several substantive comparisons if it is absent; PropsMatchGo skips before comparing any of its gathered outputs. Both were enabled for every slice baseline, timing, mutation and probe using yaml@2.9.0 and prettier@3.9.6 in /tmp/u152/library. API npm ci and the library installation succeeded, but their duration was not instrumented.
5. M4 makes the flow lexer fail to advance on '-x'. The initial matrix timed out at 90.016 seconds. The two reached rows were then run separately and each exhausted 90 seconds. All other production rows were replayed separately and passed. M4 has unknown kills, not zero observed kills, and is neither a survivor nor an equivalent candidate. A two-second Node behavior witness returns the token stream before M4 and exits 124 after M4. It supports changed behavior only, not a test kill.
6. The childguard defaults are 30 minutes to first output, two minutes of stall and 60 minutes total. The Go binary's timeout panics before these guards terminate the hanging native child, which runs in its own process group. Three audit-owned orphan groups required explicit cleanup. The first orphan also competed with the witness runs. The three-run Good timings happened before any mutation or orphan and are unaffected.
7. The completed M1 kills for the lexer and property agreement tests are lowering precondition failures. M3's property kill is an ASan heap-buffer-overflow before the equality assertion. These are real integration catches; they do not demonstrate a completed nonempty port-output disagreement. P2 and P3 separately demonstrate rejection of empty driver output, but probes are excluded from worthiness and subsumption.
8. The two negative gap guards pin self-written diagnostic classes and text while Node proves the inputs execute. M1 makes the structural refusal become a NotYet error and the structural row rejects it. These oracle pins will also fail if the compiler intentionally closes the limitations, as their messages explain.
9. Subsumption rests on three completed production matrices: one shared M1 kill for the gap guards and lexer row, and two shared kills for the property row. This is evidence for a defender's review, not a deletion recommendation. M4 cannot settle any additional lexer/property relation. Sacred means exclusive within these six production rows only; central replay must settle wider uniqueness.
10. The four-mutant rebuild limit takes precedence over aiming for three mutations per row. The mutations were fixed before outcomes and spread over an arity guard, a runtime constant, runtime string append and a port delimiter function. They were replayed individually with fresh per-id ADAMIC_BUILD_CACHE_DIR values, rather than a source switch, so each standalone diff is exactly the version run. Native Build also keys its runtime library on source contents; it rebuilt changed runtime versions.
11. Go and TypeScript function inventories were generated before mutations. Go coverage records 562 reached lower/native functions; the port signature list contains 125 functions. The two C implementation files were read in full before choosing their mutations. Coverage is at Go function/block level, not C or TypeScript dynamic coverage; the TypeScript inventory is a conservative imported call surface.
12. W1/W2/W3 are the allowed witness comparison edits. Each forces the equality check to report agreement and every built-in mutant subcase rejects that result. W3 also removes an otherwise-unused bytes import to compile. Its failure is log line 153, mapped to origin/main line 154. Production-mutant precondition failures in witnesses were deliberately excluded from the production matrix.
13. P1 empties the Lower entry used by the compiler integration rows. Closed gaps recover the ensuing C-emission nil panic as a test failure; the shared-slice row does not, so it was run alone and its panic was observed. P2/P3 empty the port driver mains, leaving the Go and YAML-library oracles unchanged. Every production row failed its own entry probe. Witness vacuity is null because production probes do not apply to their checks.
14. There were no completed-matrix survivors. No repo-wide tests, unrequested YAML rows, C dynamic coverage, exhaustive semantic mutants, or changes to production behavior were retained. Native build timings in builds.json are additional warm-cache validation rebuilds, not measurements of the first cold runtime compilation within each matrix.
'''
summary='\n'.join(['u152: 11 discovered functions, nine rows after family grouping; none moved or vanished.','Base: '+base+'; nproc 5; warm setup skipped.','Enabled slice baseline passed in 50.595s; restored slice passed in 52.186s; no slice skips.','Bounded verdicts: one sacred family, four subsumed rows, one overlapping row, three witnesses.','Evidence branch: test-audit/stage1-cohere-yaml-gaps; path review/test-audit/stage1-cohere-yaml-gaps/.'])
md=summary+'\n\n```json\n'+json.dumps(report,indent=2)+'\n```\n\n| ID | Origin/main file:line | Change | Observed failed rows |\n| --- | --- | --- | --- |\n'
for id,file,line,change in metadata:
 failed=', '.join(r for r in prod if matrix[id][r]=='fail') or ('unknown: LexerMatchesGo and PropsMatchGo over budget' if id=='M4' else '')
 md+=f'| {id} | {file}:{line} | {change} | {failed} |\n'
md+='\nWitness diffs: W1 lexer_test.go:263, W2 props_test.go:177, W3 scalar_test.go:153. Each comparison changes bytes.Equal(...) to true; W3 removes its unused import.\n\nEmpty-entry diffs: P1 internal/lower/lower.go:20; P2 lex_main.ts:25; P3 props_main.ts:99. Probes do not count as mutants.\n\nSurvivors: none among completed matrices. M4 is over budget; its Node witness changes a terminating token stream into nontermination.\n\n'+notes
wall=sum(x['wall'] for x in timings);binary=sum(x['seconds'] for x in timings)
md+=f'\nTiming: warm toolchain setup 0 seconds, nproc 5. Twenty-seven timing runs total {binary:.3f} binary seconds and {wall:.3f} command wall seconds. Whole-package setup products reported Go driver 3.47s, lowering 20.91s, native build 17.19s. Matrix, witness, probe, narrowed replay and rebuild durations and commands are in runs.json, narrow-runs.json, m4-other-runs.json and builds.json. The initial npm installs were not timed. Separate warm native builds measured M2 0.192494s, M3 0.245134s, M4 0.904854s. Overall unit work began about 13:12 UTC and finished about 13:40 UTC.\n'
(p/'REPORT.md').write_text(md)
for f in p.iterdir():
 if f.is_file() and (f.suffix in ['.log','.json','.txt','.diff','.py','.out','.ts','.md']):shutil.copy2(f,dest/f.name)
(dest/'library-package.json').write_text((p/'library/package.json').read_text())
print('Evidence files:',len(list(dest.iterdir())),'bytes',sum(f.stat().st_size for f in dest.iterdir()))
print(summary)
