from pathlib import Path
import subprocess,json,gzip,shutil
root=Path('/workspace/adamic');p=Path('/tmp/defend-conditions');base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip();d=root/'review/test-defend/internal-oracle-conditions/sessions'/base[:8];d.mkdir(parents=True,exist_ok=True)
def events(f):return [json.loads(s) for s in (p/f).read_text().splitlines() if s.startswith('{')]
assert any(e['Action']=='pass' and not e.get('Test') for e in events('restored.log'))
for f in p.iterdir():
 if not f.is_file():continue
 if f.suffix=='.log':(d/(f.name+'.gz')).write_bytes(gzip.compress(f.read_bytes(),mtime=0))
 else:shutil.copy2(f,d/f.name)
plan=json.loads((p/'plan.json').read_text());rows=[];matrix={}
core_regex=plan['regexes'][0][1]
for mid,file,before,after,line,change,target in plan['mutants']:
 es=sum([events(mid+'.'+name+'.log') for name in ['core','general','review','counts']],[])
 results={e['Test']:e['Action'] for e in es if e['Action'] in ['pass','fail','skip'] and e.get('Test') and '/' not in e['Test']}
 failed=[n for n,a in results.items() if a=='fail'];assert failed==[target],(mid,failed)
 proof=next(e['Output'].strip() for e in es if e.get('Test','').split('/')[0]==target and '.go:' in e.get('Output','') and ('stderr differs' in e['Output'] or 'exit codes differ' in e['Output']))
 matrix[mid]={'rows_failed':failed,'rows_passed':[n for n,a in results.items() if a=='pass'],'rows_skipped':[n for n,a in results.items() if a=='skip'],'top_level_results':results,'passed_subtests':[e['Test'] for e in es if e['Action']=='pass' and '/' in e.get('Test','')],'skipped_subtests':[{'test':e['Test'],'action':e['Action']} for e in es if e['Action']=='skip' and '/' in e.get('Test','')],'failing_line':proof}
 prior='TestEntriesRuntimeReadiness' if target=='TestEntriesProvenance' else 'TestEntriesProvenance'
 row={'test':target,'package':'internal/oracle','prior_verdict':'subsumed','subsumed_by':[prior],'defense':'defended','unique_mutant':mid+' '+file+':'+str(line),'attempts':[{'mutant':mid,'file_line':file+':'+str(line),'change':change,'rows_failed':failed}],'evidence':'ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-conditions/cache/'+mid+' timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '+core_regex+'; '+proof,'bounded':True,'matrix_rows':list(results),'rows_passed':matrix[mid]['rows_passed'],'base':base}
 rows.append(row)
(d/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(d/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
ref='origin/test-audit/internal-oracle-conditions';old=gzip.decompress(subprocess.check_output(['git','show',ref+':review/test-audit/internal-oracle-conditions/discovery.log.gz'])).decode();(d/'audit-discovery.log.gz').write_bytes(gzip.compress(old.encode(),mtime=0));a={l for l in old.splitlines() if l.startswith('Test')};b={l for l in (p/'list.log').read_text().splitlines() if l.startswith('Test')}
(d/'scope.json').write_text(json.dumps({'base':base,'current_top_level_tests':sorted(b),'added_since_audit':sorted(b-a),'removed_since_audit':sorted(a-b),'matrix_rows':rows[0]['matrix_rows'],'general_agreement_subtests':[t for t in matrix['D1']['passed_subtests'] if t.startswith('TestNativeAgreesWithNode/')],'review_agreement_subtests':matrix['D1']['skipped_subtests']+[{'test':'TestReviewProgramsAgreeWithNode/smoke.a','action':'pass'}],'outside_selected_rows_and_inputs':'unknown'},indent=2)+'\n')
report='''# Entries defense

Both TestEntriesProvenance and TestEntriesRuntimeReadiness are defended in the bounded current-package matrix. Neither result claims package-wide or repository-wide uniqueness beyond the observed rows.

Starting origin/main: BASE. Audit evidence: origin/test-audit/internal-oracle-conditions, review/test-audit/internal-oracle-conditions/. Its mutual subsumption rested on one shared mutant, M3 disabling checked Object calls in the JavaScript backend. Its own notes call that a sample, not a deletion recommendation.

## Code and oracles

Code under test: Adamic enumeration lowering in internal/lower/library_object.go and entries_provenance.go, native C generation/runtime enumeration, and JavaScript enumeration generation in internal/javascript/entries_provenance.go. Oracle: original successful programs run on Node. Checked failures use repository-authored exact panic messages, stdout and exit 70; these contracts are self, not external authority. The constructed-IR readiness row compares the unchecked variant with Node and pins the checked failure independently. Tests, harness, Node sources and their expected values were untouched.

## Difference and aimed mutants

Per-test coverage commands used -coverpkg=./internal/lower,./internal/native,./internal/javascript. Both clean runs passed. Provenance covered 3490 blocks, with 3061 absent from readiness. Readiness covered 434, with five absent from provenance. Raw profiles and coverage-diff.json retain every block.

Provenance actually calls Lower on ten sources, follows annotated and imported origins, inspects checked Object calls, compares three products and pins detailed diagnostic labels. Readiness does not call Lower: it constructs an ObjectLiteral whose numeric z slot is Uninitialized and an explicitly Checked Object.entries call. Its exclusive blocks include the uninitialized-field paths in JavaScript and native object emission. Both execute the shared checked enumeration emitter, but readiness feeds it field-readiness metadata that none of the ten provenance sources has. This is a semantic defense on shared lines, not an inference from exclusive coverage alone.

D1 changes call.ElementName from tsc's TypeToString(element) to the constant unknown at library_object.go:145. It preserves the checks, their exit 70 and successful answers. Provenance fails three source subcases on stderr at entries_provenance_test.go:51. Readiness still receives its hand-built ElementName number and passes. The general Node row's checked-fixture comparisons pass with both backends carrying the same wrong label. The full count sweep passes, so the catch is not a changed allocation table.

D2 replaces the generated readiness condition adamicFieldReadiness.get(object)?.has(key) with false at internal/javascript/entries_provenance.go:21. This flip makes checked enumeration read a stored numeric zero instead of treating its uninitialized field as undefined. Readiness fails at entries_provenance_test.go:129: exit codes differ, checked JavaScript returns completed and exit 0 instead of the pinned exit 70. Native behavior remains correct. Provenance and all selected general fixtures pass because their enumerated fields are initialized.

Both standalone diffs apply to BASE without switches, and each passed go vet in its changed package. Their native products built successfully under the harness's clang flags before runtime assertions failed. Every mutant had ADAMIC_BUILD_CACHE_DIR=/tmp/defend-conditions/cache/<id>, with ADAMIC_GATE_UNCACHED=1. matrix.json lists the exact failing lines and every observed passed row/subtest. There are no survivors. One successful aimed attempt per row sufficed; no additional mutations were needed.

## Matrix limits and clean results

The whole clean package timed out after 90.082 binary seconds, with no individual failing test observed. We narrowed using Object.entries/Object.values source callers and ir.ObjectCall tests. The bounded matrix contains the two targets, EntriesAcceptance, all five new top-level rows since the audit, the general Node family on 25 selected enumeration/library Object fixtures, and the complete TestCountsAreRecorded row. Review agreement selected smoke.a and the tuple-from-entries review case. That tuple case is pending and skipped on baseline and mutants. Remaining review agreement programs, other general-family fixtures and other package tests are unknown. SelfTest is a witness and its production results were not treated as witness-strength evidence.

The package grew from 193 to 198 top-level tests; none vanished. The five additions are FractionalPowersReachRuntime, ReviewProgramsAgreeWithNode, ReviewProgramsRefuse, ReviewProgramsNoLooseFiles and ReviewProgramsSelfTest. Their scope and outcomes are recorded in scope.json. EntriesAcceptance still skips for the absent worker-authored expectations.json corpus; no replacement was manufactured.

Clean bounded baseline binary seconds: provenance coverage 1.411, readiness coverage 0.488, core 2.439, general 2.927, review 0.406, complete counts 46.422. D1 complete counts passed in 46.330; D2 in 45.322. The restored two-row run passes. commands.json records actual rebuild/run wall times, caches and commands. Toolchain setup was warm and skipped; npm ci ran before baseline; nproc=5.

## Brief issues and time costs

- The mutual audit subsumption was based on a single broad JavaScript mutant. It had not exercised the provenance diagnostic contract or the staged uninitialized-field input separately.
- A 90-second package limit makes a full current-package uniqueness claim unavailable here. The defended verdicts are explicitly bounded. Outside kills remain unknown for central replay.
- /tmp has only 8.8 GB total capacity. Disk-first cleanup deleted only the prior /tmp/defend-parser-namespaces scratch/cache; afterward /tmp had 8.8 GB free and /workspace had 20 GB. No disk failure occurred. The 15 GB /tmp threshold is impossible in this workspace.
- Some initial filename selectors were over-escaped. Their raw logs are preserved as partial-selector/prior logs; only corrected replays count. Corrected general runs verified 25 passed fixture subtests. Corrected review runs verified smoke passes and the tuple case skips. An initial runner assertion expected two passes and rejected the known baseline skip; this was an evidence-validation error, not a test failure.
- The acceptance corpus and pending tuple case could not execute, limiting coverage.
- Twins and cost exceptions do not apply to these two rows: their inputs and compiler layers differ, and neither is a performance threshold row.

Owner finding: both names match assertions that execute. Provenance checks origins/check annotations and exact contracts. RuntimeReadiness checks an uninitialized slot at backend level and expressly does not claim source admission. Both stay. No tests were deleted, rewritten or weakened; production sources were restored before saving evidence.
'''.replace('BASE',base)
(d/'report.md').write_text(report)
print(d)
