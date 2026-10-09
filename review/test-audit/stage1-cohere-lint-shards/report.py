import pathlib,json,re,statistics
p=pathlib.Path('review/test-audit/stage1-cohere-lint-shards'); menu=json.loads((p/'menu.json').read_text()); commands=json.loads((p/'commands.json').read_text()); rows=['TestMergePutsCasesBackInOrder','TestMergeKeepsLinesThatOnlyLookLikeCases','TestMergeRefusesAMissingOrRepeatedCase']; ids=['C'+str(i)for i in range(1,10)]
def events(mid):return [json.loads(x)for x in (p/(mid+'.log')).read_text().splitlines()if x.startswith('{')]
matrix={mid:[x['Test']for x in events(mid)if x['Action']=='fail'and'Test'in x and'/'not in x['Test']]for mid in ids+['P1']}; results=[]
oracles=['Handwritten complete expected byte string for interleaved case blocks in ascending order.','Handwritten complete expected byte string retaining indented case-like text and case 1x.','Handwritten diagnostic substrings for missing, repeated, non-case-start, empty-shard and extra-case inputs. Error required; complete diagnostic not compared.']
for row,line,oracle in zip(rows,[8,23,34],oracles):
 samples=[]
 for i in range(1,4):
  text=''.join(x.get('Output','')for x in events(row+'-'+str(i)));samples.append(float(re.search(r'\bok\s+\S+\s+(\d+\.\d+)s',text).group(1)))
 caught=[mid for mid in ids if row in matrix[mid]];last=caught[-1];failure=next(x['Output'].strip()for x in events(last)if x.get('Test')==row and 'shards_test.go:'in x.get('Output',''));command=next(x['command']for x in commands if x['id']==last)
 results.append(dict(test=row,package='stage1/cohere/lint/shards',file='stage1/cohere/lint/shards/shards_test.go:'+str(line),seconds=statistics.median(samples),oracle=oracle,oracle_kind='self',kills=caught,unique_kills=[],last_proven_fail=last+' '+failure,verdict='setup-check',subsumed_by=[],mutants_in_matrix=9,probe_kills=['P1'],subsumer_seconds=None,vacuous=False,bounded=False,matrix_rows=[],evidence=command+'; '+failure,timing_samples=samples,mutation_kind='construction',construction_unique_kills=[mid for mid in caught if len(matrix[mid])==1]))
(p/'rows.json').write_text(json.dumps(results,indent=2));(p/'matrix.json').write_text(json.dumps(matrix,indent=2));table='| ID | Origin shards.go line | Change | Failing rows |\n|---|---:|---|---|\n'
for m in menu:table+='| '+m['id']+' | '+str(m['line'])+' | `'+m['menu']+'` | '+', '.join(matrix[m['id']])+' |\n'
(p/'mutants.md').write_text(table);baseline=''.join(x.get('Output','')for x in events('baseline'));baseline_seconds=re.search(r'\bok\s+\S+\s+(\d+\.\d+)s',baseline).group(1);wall=sum(x['wall']for x in commands);vet=sum(x['seconds']for x in json.loads((p/'validation.json').read_text()));witness=sum(x['seconds']for x in json.loads((p/'witness-commands.json').read_text()))
report=f'''u125 audited all three current top-level tests, no families or skips.
Starting origin/main: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2; nproc 5; warm Go 1.27.1.
Clean baseline passed in {baseline_seconds}s own binary, 0.175s shell build/run wall.
Nine construction mutations: seven caught, two witnessed changed-behavior survivors.
All three rows are setup-check; empty Merge fails every row; source restored.

CODE UNDER TEST: Go shard-output construction Merge and isCaseLine. Run is not reached.
ORACLE: handwritten output bytes and diagnostic substrings, all self. No Node/cohere/native execution.

'''+json.dumps(results,indent=2)+'\n\n'+table+'''
Survivors:
C5: clean Merge preserves "case 0\\ncase \\nfixed\\ta\\n"; mutant returns malformed case line "case ". Missing minimum-digit boundary guard in current tests.
C6: clean Merge returns all ten cases 0..9; mutant reports case 9 missing. Current positive case numbers do not cover digit 9.
Witness command: go run /tmp/u125/witness.go with clean source, then each standalone survivor applied. Outputs witness-clean.log, witness-C5.log, witness-C6.log; exact driver and command timings retained.
No equivalent candidates or unexplained survivors.

Brief ambiguities, interpretation and time costs:
- This package constructs the test suite's merged shard output. Its rows check that construction, so the brief's setup-check rule applies. C IDs deliberately identify construction edits. Their observed failures appear in kills for replay, but unique_kills stays empty because construction uniqueness does not establish production sacredness. construction_unique_kills records observed singleton catches separately.
- The stage1 path does not mean these tests execute a stage1 port. All bodies call Go Merge; none calls Run. Mutating .a/.ts, cohere or native lowering would not test the asserted construction. Native rebuilding is inapplicable.
- Three functions share Merge but assert different properties: order, preservation of non-header lines, and rejection errors. They are not input-only wrappers around a shared checker, so no family grouping.
- The negative row feeds deliberately bad shard output directly to Merge. This tests construction validation; it does not witness a weakened external agreement oracle, and no test/harness assertions were edited.
- Whole-package scope came from the current list, not stale filenames. No missing or opt-in tests, no helpers, no external tools required by these bodies.
- The one-line header prefix can fail negative tests through a different error. Only observed failures are reported; a production-worthy verdict is not inferred from those precondition failures.
- Error substring checks verify narrower behavior than full diagnostic equality. A different error containing the same substring could pass. Positive rows compare complete bytes, not counts or exit codes.
- A literal empty-answer return with its old body retained triggers vet unreachable-code diagnostics. P1 replaces the complete Merge body and passes vet. This remains a probe, never a construction or production kill.
- All nine menu choices were saved before testing. They cover framing, numbering, duplicate rejection, header recognition, digit bounds, block bytes and total-case counts. No inserted-statement supplemental mutants.
- Sequential switch-free Go mutations were used because this tiny package builds and runs in fractions of a second. This also directly validates every saved diff without selector dependencies. Build/check wall is recorded per mutation. The four-mutant rebuild limit concerns native port/compiler rebuilds, which this unit performs none of.
- npm ci in stage3/api was mandatory in the brief even though no test loads node_modules. Installation was not separately timed; no other Node dependency directory is reached.
- An initially broad caller search included generated lint tests and produced excessive output. It did not run any other package. Scope and mutations were taken only from the three bodies read in full and shards.go.
- No run reached 90 seconds, so no bounded matrix, timeout narrowing, panic reruns or unknown rows. No repository-wide uniqueness claim.
- Seven caught construction mutations prove these checks can fail; the two survivor witnesses also show concrete construction boundaries the tests miss. Neither survivor is called equivalent or silently treated as covered.

Timing and uncovered work:
Setup skipped, 0s; nproc 5. Baseline binary '''+baseline_seconds+f'''s; shell build/run 0.175s.
Nine isolated timing runs, ten matrix/probe runs and restored package run: {wall:.3f}s combined shell build/run wall. Standalone vet compile/check wall: {vet:.3f}s. Three survivor-witness build/runs: {witness:.3f}s. Individual measurements and exact commands retained.
Each row's seconds is the median of three fresh -count=1 own-binary ok lines. No native builds; npm install wall not recorded independently. Completed within the 20-minute budget.
Not covered: Run subprocess launch/count-only aggregation/concurrency/manifest handling; additional malformed initial headers, unterminated headers, signed numbers, huge numbers and CRLF; other Adamic packages; external/native lint behavior. Source restored, final clean test passed. No PR or main push.
'''
(p/'REPORT.md').write_text(report);print(json.dumps(results,indent=2));print(table);print('build/run',wall,'vet',vet,'witness',witness,'baseline',baseline_seconds)
