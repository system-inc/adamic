import json,pathlib,re
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-native-radix');prior=json.loads((p/'prior-rows.json').read_text());ms=json.loads((p/'mutants.json').read_text());runs=json.loads((p/'runs.json').read_text());by={r['id']:r for r in runs};mut={m['id']:m for m in ms}
def group(n):return 'TestRegExpBytecodeRandomNode family' if n.startswith('TestRegExpBytecodeRandomNodeUnit') else 'TestRecordMutants family' if n.startswith('TestRecordMutantsUnit') else n
def failures(id):return sorted(set(group(n) for n in by[id]['failed']))
def fail_line(id,row):
 members=[n for n in by[id]['failed'] if group(n)==row]
 for line in (p/(id+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Test','').split('/')[0] in members and re.search(r'\.go:\d+:',e.get('Output','')) and not re.search(r': compiling |: mode=|: cache |: build .* (hit|miss) ',e.get('Output','')):return e['Output'].strip()
 return None
assign={'TestToStringWithARadixOutOfRangePanics':['D1'],'TestRecordBenchmark':['D2'],'TestRegExpSearchNode':['D7'],'TestRegExpLintPatternsNode':['D7'],'TestRegExpBytecodeTest262':['D6','D7'],'TestRegExpNativeStepLimit':['D3','D4','D5'],'TestRegExpBytecodePatternUnits':['D6'],'TestRuntimeReleasePaths':[],'TestRuntimeStringEquality':['D2'],'TestRegExpBytecodeRandomNode family':['D6','D7']}
rows=[]
for r in prior:
 n=r['test'];ids=[id for id in assign[n] if id in by];unique=[id for id in ids if failures(id)==[n] and not by[id]['cooked']];defense='defended' if unique else 'not defended' if n=='TestRegExpNativeStepLimit' and len(ids)==3 else 'cannot-judge';id=unique[0] if unique else next((id for id in ids if n in failures(id)),None)
 evidence=(by[id]['command']+'; '+str(fail_line(id,n))) if id else 'No isolated targeted kill established; see coverage-differences.json and bounded matrix logs.'
 attempts=[dict(mutant=id,file_line=mut[id]['file']+':'+str(mut[id]['line']),change=mut[id]['before']+' -> '+mut[id]['after'],rows_failed=failures(id),intent=mut[id]['lead']) for id in ids]
 reason='Unique within completed direct-consumer matrix; other package suites are unknown.' if unique else 'Three attempts: changed diagnostic and removed limit configuration were also caught by the new boundary row; off-by-one limit was caught only by that boundary row.' if defense=='not defended' else 'Seven-mutant budget exhausted before three dedicated attempts. Exploratory shared-path runs do not establish redundancy.'
 rows.append(dict(test=n,package='internal/native',prior_verdict=r['verdict'],subsumed_by=r['subsumed_by'],defense=defense,unique_mutant=unique[0]+' '+mut[unique[0]]['file']+':'+str(mut[unique[0]]['line']) if unique else None,attempts=attempts,evidence=evidence,bounded=True,reason=reason,rows_passed=sorted(set(group(x) for x in by[id]['passed'])) if unique else [],code_under_test=r['file']+' exercises production runtime and regexp compilation; see code-and-oracle.md',oracle=r['oracle']))
(p/'rows.json').write_text(json.dumps(rows,indent=2));(p/'matrix.json').write_text(json.dumps({id:{'failed_rows':failures(id),'passed_rows':sorted(set(group(n) for n in by[id]['passed'])),'tests':json.loads((p/(id+'-scope.json')).read_text()),'cooked':by[id]['cooked']} for id in mut if id in by},indent=2))
notes='''Friction and limits:
- The audit labels were bounded, not package-wide. The current package has 282 Test functions; all were enumerated. The current exact step-limit boundary row was added to regexp matrices.
- The full clean run timed out at90.168 seconds with no observed individual failure. A broader direct-runtime baseline passed; D1's first broad run cooked at90.060 seconds. Its failures do not prove uniqueness and are retained separately. The subsequent per-function matrices completed or are explicitly flagged.
- Go -coverprofile with -coverpkg=./internal/native,./internal/regexp measures compiler and build code, not embedded C run in child binaries. All requested row/subsumer profiles are saved, with exact exclusive Go lines. No C coverage claim is made. Runtime leads use semantic inputs and assertions instead.
- Seven mutation slots cannot supply three dedicated attempts for ten rows. They were spent on invalid-radix diagnostics, undefined equality, three step-limit contracts, raw UTF16 pattern identity and low-surrogate lastIndex. Rows with fewer than three dedicated attempts are cannot-judge, not not defended. They remain candidates to keep.
- Some rows differ in input despite sharing a checker. The random wrappers are one family. A family member is never treated as its subsumer.
- The supplied radix audit excerpt truncates commands. Complete prior report, row data, scopes, plans and limits were fetched and saved.
- Every complete matrix lists its current test names and passed rows. Decoder corpora, unrelated compiler products, WASI and other emitted-program consumers were not rerun under mutants. No package-wide unique kill is asserted beyond the bounded direct-consumer scope.
Name/assertion findings for rows not defended:
- TestRecordBenchmark measures and logs timing, validates five nonnegative non-NaN durations, and compares Node workload results. It asserts no performance threshold and accepts Infinity. Its name does not establish protection against a slowdown.
- TestRegExpNativeStepLimit checks catastrophic-backtracking interruption, exit70 and a diagnostic substring. It does not check exact instruction count. The newer boundary row guards that separate promise.
- Search, Lint, Test262 and Random compare results for their own supplied inputs, including captures/groups/lastIndex. No missing name promise was established; insufficient mutants cannot justify weakening them.
- RuntimeReleasePaths checks shared ownership, live counts and a100000-object chain with two sanitizer settings. Its assertions do address the named release behavior. No exclusive production defect was tested within the budget.
No test or oracle was edited. All source edits are restored before committing evidence.
'''
(p/'friction-and-limits.md').write_text(notes)
report='Defended '+str(sum(r['defense']=='defended' for r in rows))+' rows in bounded direct-consumer matrices.\nAll other rows remain kept; incomplete defenses are cannot-judge.\nEvidence includes standalone diffs, compile/apply logs, coverage pairs and exact passed-row lists.\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n'+notes
(p/'REPORT.md').write_text(report)
print([(r['test'],r['defense'],r['unique_mutant']) for r in rows])
