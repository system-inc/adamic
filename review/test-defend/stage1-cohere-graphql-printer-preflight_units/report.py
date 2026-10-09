import pathlib,json,re,subprocess
p=pathlib.Path('/tmp/def-graphql/evidence')
def events(f):
 es=[]
 for l in f.read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return es
base=events(p/'reached-baseline-parallel2.log');expected=sorted({e['Test'] for e in base if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']})
plans=json.loads((p/'plan.json').read_text());matrix=[];attempts=[]
def group(n):
 if re.fullmatch('TestPrinterWhitespaceGap_[0-9]+',n):return 'TestPrinterWhitespaceGap family'
 if re.fullmatch('TestPrinterAsGoCohere_[0-9]+',n):return 'TestPrinterAsGoCohere family'
 return n
for plan in plans:
 es=events(p/(plan['mutant']+'-matrix.log'));repair=p/(plan['mutant']+'-repair.log')
 if repair.exists():es+=events(repair)
 done={e['Test']:e['Action'] for e in es if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']}
 unknown=sorted(set(expected)-set(done));assert not unknown,(plan,unknown)
 failed=sorted({group(n) for n,v in done.items() if v=='fail'})
 errors=[e.get('Output','').strip() for e in es if e.get('Test','').startswith('TestPrinterWhitespaceGap_') and 'native:' in e.get('Output','')]
 matrix.append(dict(mutant=plan['mutant'],rows_failed=failed,top_level_results=done,unknown_rows=unknown,first_native_failure=errors[0] if errors else None,cooked_initial_run=any('panic: test timed out' in e.get('Output','') for e in events(p/(plan['mutant']+'-matrix.log')))))
 attempts.append(dict(mutant=plan['mutant'],file_line=f"{plan['file']}:{plan['line']}",change=plan['change'],rows_failed=failed))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
row=dict(test='TestPrinterWhitespaceGap family',package='stage1/cohere/graphql/printer',prior_verdict='subsumed',subsumed_by='TestPrinterFileDriver',defense='not defended',unique_mutant=None,attempts=attempts,evidence="ADAMIC_GRAPHQL_PRETTIER=/tmp/def-graphql/prettier ADAMIC_BUILD_CACHE_DIR=/tmp/def-graphql/cache/<id> timeout 120 go test -json -count=1 -timeout 90s -parallel 2 ./stage1/cohere/graphql/printer/ -run '^(TestPrinter(Whitespace.*|AsGoCohere.*|FileDriver|Shard.*|ConstructorGap))$'; D01: gaps_test.go:133: native: Expected <EOF>. (1:2), Go cohere Unexpected <EOF>. (1:2). D02: gaps_test.go:133: native: Unexpected <EOF>. (0:2), Go cohere (1:2). Exact commands and cooked-run repairs accompany the logs.",bounded=True,members=[f'TestPrinterWhitespaceGap_{i:03d}' for i in range(5)],matrix_rows=expected)
(p/'rows.json').write_text(json.dumps([row],indent=2))
report='''The whitespace-gap family was not defended after three aimed attempts.
Both diagnostic faults also fail the AsGo family; the work-growth attempt passes.
Results are bounded to the recorded current reached-row matrix, with cooked rows recovered by smaller reruns.

'''+json.dumps([row],indent=2)+'''

Starting origin/main: 92d196011e78208b45b3768dcf2c8c88e7ec132d. All five current numbered family members exist; Setup is separate. scope.log records all 51 top-level names in the current package. The family is one row, not five independent kills. It already contains both Node and native execution, so these are not two separate executor twins.

CODE UNDER TEST: the TypeScript GraphQL parser/printer port, specifically Parser.unexpected in parser.ts, getLocation and syntaxError in lexer.ts, and format's parse/refusal path in printer.ts. The source is executed through Node, native ASan/UBSan, and, in AsGo, Adamic's JavaScript backend and leak check. Every production diff contains only port-source changes; test bodies, fixtures, harness and external oracle implementations are untouched.

ORACLE: Go cohere decides the exact port refusal text. npm Prettier 3.9.6 and the embedded fork must accept these five whitespace corpus entries with empty formatted text. The repository's fixed gap list and five-case count are self-written controls. The family checks exact diagnostics, stdout/stderr/exit and the recorded cross-oracle gap. The file driver checks one valid query against a self-written output string. This row's name promises a whitespace-gap check, and the assertions do check that promise. It does not claim a performance comparison, but it has a 30-second own-work assertion, which motivated D03.

Coverage: family.cover and subsumer.cover are independently measured with -coverpkg on the printer package and internal/lower, internal/native. No family-exclusive Go block was measured; the subsumer has 4164 exclusive Go blocks due to a compiler cache miss. Cached products make these profiles unsuitable as TypeScript reach comparisons. The profiles measure Go machinery, not port statements in child processes. coverage-difference.json preserves every measured block. The semantic lead is empty/whitespace EOF refusals and coordinates, absent from the valid query tested by TestPrinterFileDriver. All five whitespace inputs are also present in the AsGo corpus, so input containment is a genuine possible subsumer.

D01 changes the unexpected-token diagnostic word. All five gap members fail with Expected <EOF> instead of Unexpected <EOF>; all four AsGo members also fail. FileDriver passes.
D02 changes the initial error line from 1 to 0. Gap members fail on exact coordinates, and all four AsGo members also fail. FileDriver passes. Its first matrix cooked; a smaller run repeats the whole gap family plus the unfinished AsGo member, recording terminal results rather than treating timeout as a kill.
D03 repeats the error-location scan eight times, resetting counters each pass and returning the same coordinates. This is explicitly allowed by the cost-row repeat-loop rule, rather than the ordinary insertion menu. Its initial matrix cooked without an assertion failure; a smaller run repeats the target family and unfinished rows. All recorded bounded rows pass. The production scan executes eight passes instead of one, while every compared answer remains correct. This is a work-growth survivor; it is not an equivalent candidate. A mild eightfold increase in this small operation does not establish what larger growth the 30-second guard would catch.

The two semantic faults refute general cover by the prior named file-driver row, but do not provide uniqueness: the AsGo family catches both. Three attempts cannot prove no unique fault exists. No test deletion or rewrite is recommended.

Frictions and limits:
- The full package cooked at 90.143 s with no earlier recorded failed test. The first broader narrowed run cooked at 90.234 s; its serialized version at 90.049 s; the serialized reached-row subset at 90.079 s. Protected-oracle-only upstream preflights were excluded because they never execute the mutated port. Two parallel slots produced a green reached-row baseline at 74.768 s, with no skips.
- D02 and D03 initially cooked at the binary budget. Their incomplete row observations were recovered by the smaller commands in repair-command.txt files. Timed-out package failure is not a production kill. matrix.json records grouped failures and every observed top-level terminal result.
- Compiler cache hits obscure Go per-test coverage of the port; direct TypeScript/native dynamic line coverage was not collected. Semantic input differences are stated rather than inferred from cache-dependent coverage.
- The port's lowering build key hashes repository contents. Evidence stayed under /tmp/def-graphql/evidence during every test run, avoiding log-driven cache invalidation. Logs and profiles are compressed when copied to review; original recorded commands name the uncompressed temporary logs.
- The work-budget cleanup assertion made a cost attempt appropriate even though this is a gap row rather than a throughput row. The added cost instruction permits repeating a loop, which the ordinary menu alone would not permit.
- Warm tools were available, but pinned npm dependencies were absent; npm install and npm ci were required in the scratch oracle directory, plus npm ci in stage3/api. The full cooked run observed TestPrinterThroughput skipped. No opt-in benchmark or Darwin-only suite is included in the bounded matrix.
- Git fetch initially worked, then shell Git requests began failing with could not read Username. Current cloud configuration showed no credential binding. The connected GitHub app remained able to read this repository, so publication uses its Git object and branch APIs without a force update or PR.

Timing: nproc=5; setup.sh skipped. Coverage runs passed in 17.712 and 9.123 binary seconds. Build and matrix command wall times are runs.json; native validation runs rebuild each mutant separately, and build logs preserve binary elapsed times. Repairs are repairs.json. Cold baseline timeouts and publication/authentication diagnosis consumed much of the session budget. Not covered: full-package/repository uniqueness, rows outside the bounded selector, exhaustive dynamic port coverage, controlled cost measurement, or compiler-only mutants. Production sources are restored.
'''
(p/'REPORT.md').write_text(report)
print([(x['mutant'],x['rows_failed'],x['unknown_rows']) for x in matrix])
