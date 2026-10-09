import json,subprocess
from pathlib import Path
p=Path(__file__).resolve().parent
names=[s for s in (p/'discovery.log').read_text().splitlines()if s.startswith('Test')]
def events(file):
 out=[]
 for s in file.read_text().splitlines():
  try:out.append(json.loads(s))
  except ValueError:pass
 return out
matrix={}
for f in sorted(p.glob('*.log')):
 ev=events(f)
 if not ev:continue
 states={e['Test']:e['Action']for e in ev if 'Test'in e and '/'not in e['Test'] and e['Action']in ('pass','fail','skip')}
 ran=sorted({e['Test']for e in ev if 'Test'in e and '/'not in e['Test']})
 matrix[f.name]={'passed':sorted(n for n,s in states.items()if s=='pass'),'failed':sorted(n for n,s in states.items()if s=='fail'),'skipped':sorted(n for n,s in states.items()if s=='skip'),'unknown':sorted(set(ran)-set(states)),'binary_seconds':next((e['Elapsed']for e in reversed(ev)if 'Test'not in e and e.get('Action')in ('pass','fail')),None),'cooked':any('test timed out' in e.get('Output','')for e in ev),'failure_lines':[e['Output'].strip()for e in ev if 'operand evaluated more than once:'in e.get('Output','')or 'lost its inserted non-null check'in e.get('Output','')]}
metadata={'D1':{'file_line':'internal/javascript/view_fields.go:22','change':'change format-string constant to evaluate numeric receiver twice'},'D2':{'file_line':'internal/lower/non_null.go:109','change':'return early from non-null recording for static property declarations'}}
rows=[]
for test,ident,log in [('TestRequiredViewFieldOperandOnce','D1','D1.log'),('TestViewFieldReadiness family','D2','D2-narrowed.log')]:
 m=matrix[log];assert len(m['failed'])==1
 failed=m['failed'][0];assert failed==('TestRequiredViewFieldOperandOnce'if ident=='D1'else'TestViewFieldInheritedStaticReadiness')
 command=f"selector=$(cat review/test-defend/internal-oracle-typeof_dispatch/matrix-selector.txt); ADAMIC_BUILD_CACHE_DIR=/tmp/defend-oracle-typeof/cache/{ident} timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run \"$selector\" > {log} 2>&1"
 r={'test':test,'package':'internal/oracle','prior_verdict':'subsumed','subsumed_by':['TestDefaultTaggedSourceViews'],'defense':'defended','unique_mutant':ident+' '+metadata[ident]['file_line'],'attempts':[{'mutant':ident,**metadata[ident],'rows_failed':[test]}],'evidence':command+'; '+m['failure_lines'][0],'bounded':True,'unique_scope':'observed reached-test matrix; whole-package uniqueness unproved','matrix_rows':json.loads((p/'matrix-rows.json').read_text()),'passed_tests':m['passed'],'unknown_package_rows':sorted(set(names)-set(m['passed'])-set(m['failed']))}
 if ident=='D2':r['members']=['TestNarrowedFieldUsesSharedReadiness','TestViewFieldInheritedStaticReadiness'];r['defended_behavior']='static non-null assertion metadata, not a later field readiness read'
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n');(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
report='''Bounded unique catches found for both requested rows; retain both pending package-wide replay.
D1 demonstrates repeated operand evaluation; D2 demonstrates missing static non-null metadata.
Whole-package and wider replays cooked; unknown results are preserved rather than inferred.

'''
report+='Base: '+(p/'base.txt').read_text().strip()+'. See code-and-oracle.md for CUT, independent oracles, coverage and owner findings.\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n'
report+='| Run | Binary seconds | Passed top-level tests | Failed | Skipped | Unfinished |\n| --- | --- | --- | --- | --- | --- |\n'
for file,m in matrix.items():report+=f"| {file} | {m['binary_seconds']} | {len(m['passed'])} | {', '.join(m['failed'])} | {len(m['skipped'])} | {len(m['unknown'])} |\n"
report+='''
Only two production mutants were needed to find unique catches in the completed 50-test reached matrix. The full passed list is in each row and matrix.json. Fifty top-level functions become 49 rows when the two readiness members are grouped. No test, fixture, oracle or harness was changed. Production edits are restored before committing evidence.

D1's field value remains 7 and exit remains 0; the counter becomes 2 rather than Node's 1. This proves its name's single-evaluation promise. DefaultTaggedSourceViews passes. D1-expanded completed 189 of 195 top-level tests; only OperandOnce failed. Its six unfinished tests were rerun with the target functions and the bulk NativeAgreesWithNode family. That bulk run also cooked at 90 seconds, so incomplete native fixtures and any unfinished top-level results remain unknown. No full-package uniqueness is claimed.

D2's static assertion recording is absent, but the initializer's eager runtime check remains. The inherited-static readiness member reports lost inserted-check metadata; the other member and DefaultTaggedSourceViews pass in the 50-test replay. D2-expanded also cooked and was narrowed; each requested member was rerun alone. The family has a unique bounded catch. It does not follow that later field readiness is still tested: these migrated .ts fixtures stop at eager assertions before field reads.

Limitations and costs: the family name is not a listed Go Test and had to be resolved to its two actual members. Current origin/main differs from the audit base; base.txt pins replay locations. Full oracle fixture unions exceed the 90-second budget, even after excluding the three bulk top-level suites from the expanded matrix. SDK/corpus opt-in skips remain outside the uniqueness claim, listed per run in matrix.json. Go coverage does not instrument generated code or the C runtime. /tmp's 15 GB target is impossible on its 8.8 GB filesystem; only the completed prior unit's scratch cache was removed, leaving 6.4 GB free. No baseline assertion failure was observed; the 50-test and 195-test clean baselines pass. Broader replays consumed more than the suggested 20 minutes, bringing this session to roughly 26 minutes. Further global replay is left explicit rather than silently treating timeouts as passes.

Owner findings: OperandOnce genuinely compares the counter and full output with Node. The readiness family's historical names now describe later field behavior that its .ts assertions do not execute; its useful unique catch here guards recorded eager static checks. This is a rename/ownership finding, not permission to delete or weaken a test. Neither row promises a speed threshold without asserting one.

Validation: each Go production mutant passed go vet for its changed package. D1's emitted JavaScript ran successfully with wrong counter output, not a syntax error. D2 compiled and its direct assertion caught the metadata break. Both standalone diffs apply to the recorded base after restoration. Log files and full coverage profiles are pushed with the evidence.

Timing: setup skipped, nproc 5. npm.log records stage3/api npm ci. The run table separates each binary's own elapsed time, including native builds performed inside tests. Tool/build time outside the binary is not independently isolated. Whole-package completion, all compiler call graph exclusivity and repository-wide uniqueness are not covered.
'''
(p/'REPORT.md').write_text(report)
print([(r['test'],r['unique_mutant'],len(r['passed_tests']))for r in rows])
