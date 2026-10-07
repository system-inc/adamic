"""Render best-of-three paired measurements with per-sample provenance."""
import collections,json,pathlib
p=pathlib.Path(__file__).resolve().parent
rows=json.loads((p/'measurements.json').read_text())
assert len(rows)==72 and all(r['exit']==0 for r in rows)
metadata=json.loads((p/'build-flags.json').read_text())
best={}
for row in rows:
 key=(row['slug'],row['mode'],row['compiler'])
 if key not in best or row['seconds']<best[key]['seconds']:best[key]=row
lines=[
 'Built: selected-rule split builds with GOMAXPROCS jobs, concurrent correct/mutant builds, independent whole-file witnesses.\\',
 'Commits: merge a79f985bb18650f4224f19c02e2bf71b4bfeaaa9; implementation '+metadata['commit']+'.\\',
 'Commands and outputs: 72 interleaved author checks passed; parity, helper parity, cache guards and ordinary uncached all-rule checks passed.\\',
 'Mutants: seven lint cache omissions, eight native split/header mutants and four witness/concurrency guards caught; all owned mutants caught.\\',
 'Not covered: full repository gate or fifty-rule fleet; byte edit is one appended space; helper edits invalidate broadly.',
 '',
 'The branch is `devtools/lint-rule-check`. It merged `origin/area/developer-tools` at `a50bb48389bb470c1fcb7cc4754b988cbe3e2a11`, without rebasing. Only this branch is pushed.',
 '',
 'The fast path explicitly uses `native.Options{Sanitize: true, Split: true, Jobs: runtime.GOMAXPROCS(0)}` and starts the correct and mutant builds concurrently. Native port keys now include split mode and jobs. An inherited `ADAMIC_NATIVE_SPLIT` switch is cleared by the lint harness so existing full-registration builds and `ADAMIC_GATE_UNCACHED=1` retain whole-file compilation. The split compiler and its cache implementation were merged unchanged.',
 '',
 'Best of three, wall-clock seconds on the same implementation commit, with whole-file and split commands interleaved. Each rule/round/compiler has a separate initially empty user cache; subsequent edits reuse that cache. The Go build cache is primed and fixed. The ordinary compiler baseline builds correct and mutant ports sequentially, matching the preceding fast path. The clang wrapper only records and forwards compiler invocations; both paths use it.',
 '',
 '| loop | before whole-file | after split | instrument |',
 '|---|---:|---:|---|']
base='go test ./stage1/cohere/lint -run \'^TestRule$\' -count=1 -timeout 30m -v -args -rule '
for slug in ('no-var','no-empty','eqeqeq'):
 for mode,flag in [('cold',''),('warm-byte-edit',' -rule-byte-change'),('warm-unchanged',''),('helper-edit',' -rule-helper-change')]:
  a=best[slug,mode,'whole'];b=best[slug,mode,'split']
  lines.append(f'| {slug} {mode} | {a["seconds"]:.3f} | {b["seconds"]:.3f} | `{base+slug+flag}`; before adds `-rule-whole` |')
lines += ['', 'Build-flags line for every row: '+ '; '.join(f'{k}={metadata[k]}' for k in ('commit','nproc','cpu.max','go','clang','node','GOMAXPROCS','jobs','flags','GOCACHE'))+'. Cache mode and load before/after for each winning sample follow. [All 72 rows](measurements.json) also carry the full line, exact command, exit code, cache directory and trace.', '',
 '| winning sample | round | cache | load before | load after | log |', '|---|---:|---|---|---|---|']
for slug in ('no-var','no-empty','eqeqeq'):
 for mode in ('cold','warm-byte-edit','warm-unchanged','helper-edit'):
  for compiler in ('whole','split'):
   r=best[slug,mode,compiler]
   lines.append(f'| {slug} {mode} {compiler} | {r["round"]} | {r["cache_mode"]} | `{r["load_before"]}` | `{r["load_after"]}` | [{r["log"]}]({r["log"]}) |')
lines += ['', 'Split object commands additionally use `-Wno-gnu-line-marker`, `-fdebug-prefix-map=<snapshot>=/adamic-units`, `-x cpp-output -c`; every exact compiler invocation is retained in its trace. Whole-file commands retain the ordinary build flags.', '',
 'The byte edit appends exactly one space to the private rule module. Its native artifact key misses, but generated C remains identical: all three rounds compile zero objects for both correct and mutant ports. This measures source-byte invalidation and object reuse, not a behavior-changing edit. The unchanged check reuses completed native artifacts and still executes both sanitized binaries.', '',
 'The helper edit adds a used numeric identity helper and routes `visit` through it. It changes declarations and program numbering while preserving findings. Counts below are actual clang `-c -x cpp-output` calls across the concurrent correct/mutant pair, not cache file counts. All 34 unit names (32 function groups, state and main) compile again. Shared unchanged objects need only one physical compilation across the pair; additional calls represent differing correct/mutant objects.', '',
 '| rule | cold object builds, rounds 1/2/3 | byte edit | unchanged | helper object builds, rounds 1/2/3 |', '|---|---|---|---|---|']
for slug in ('no-var','no-empty','eqeqeq'):
 def counts(mode):return '/'.join(str(r['compiled_objects']) for r in rows if r['slug']==slug and r['mode']==mode and r['compiler']=='split')
 lines.append(f'| {slug} | {counts("cold")} | {counts("warm-byte-edit")} | {counts("warm-unchanged")} | {counts("helper-edit")} |')
lines += ['', 'Every `.clang.json` retains compiler arguments, working directory and start time. The two preprocessing windows overlap, confirming concurrent builds. Each split build uses four jobs; together they can request eight compiler workers against the four-core quota. No compiler scheduling or emitter changes were made.', '',
 'Validation commands (all test output is saved, not piped):', '', '```bash', 'source /workspace/adamic-tools/env.sh', 'python3 stage1/cohere/lint/split-evidence/measure.py', 'python3 stage1/cohere/lint/split-evidence/checks.py', '```', '',
 '[checks.py](checks.py) records the exact filtered Go commands. [parity.log](parity.log) covers selected-versus-full registration, split-versus-fresh-whole native, cached-versus-uncached answers, and compiler selection even with `ADAMIC_NATIVE_SPLIT=1` inherited. [helper-parity.log](helper-parity.log) repeats the complete four-runtime and fresh whole-file comparison after the helper edit. [ordinary-uncached-all-rule.log](ordinary-uncached-all-rule.log) runs `ADAMIC_GATE_UNCACHED=1 ADAMIC_NATIVE_SPLIT=1 go test ./stage1/cohere/lint -count=1 -timeout 30m -v -run \'^TestRulesAgree$\'`, including every upstream package and inherited corpus row; the build log must say `split=false jobs=0`.', '',
 'Focused ordinary-build oracle: `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 ADAMIC_NATIVE_SPLIT=0 go test ./internal/oracle -run \'^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(functions|closures|generic_functions)\\.a$\' -count=1 -timeout 30m -v` passed all three intended fixture subtests: [filtered-oracle.log](filtered-oracle.log). `go vet ./stage1/cohere/lint ./cmd/adamic-lint-check` passed: [vet.log](vet.log).', '',
 'Observed byte-identical corpus sizes:', '', '| rule | unedited | helper edit |', '|---|---:|---:|']
for slug in ('no-var','no-empty','eqeqeq'):
 import re
 def corpus(log):
  text=(p/log).read_text();part=text.split('=== RUN   TestSplitWholeParity/'+slug,1)[1].split('=== RUN',1)[0]
  return re.search(r'outputs byte-identical: (\d+) bytes',part).group(1)
 lines.append(f'| {slug} | {corpus("parity.log")} | {corpus("helper-parity.log")} |')
lines += ['', 'Mutants run and caught:', '', '| mutant | change | check and observation |', '|---|---|---|']
for kind,component,change in [('oracle','oracle/adapter/no-var','no-var adapter becomes silent'),('capture','capture/filter','fires filter becomes stays-silent'),('go-output','go-output/manifest','input var becomes let'),('port','port/modules','main prints an extra line'),('node','node/modules','main prints an extra line'),('javascript','javascript/modules','emitted module prints an extra line'),('javascript-build','javascript-build/modules','main prints an extra line')]:
 lines.append(f'| lint {kind} | drop `{component}`; {change} | `TestLintCacheInvalidation/{kind}` fails with `stale {kind} answer` and `cached_matches_before=true`; [cache-guards.log](cache-guards.log) |')
for name,change,test,message in [
 ('native-flags','drop ordered flags from object key','TestUnitCacheFlagsHoldSanitizer','sanitized rebuild reused uninstrumented object'),
 ('native-state','duplicate state across units','TestUnitsPreserveSharedState','shared state (uncached=0)'),
 ('native-literal','rewrite identifiers inside string tokens','TestSplitTokensDoNotRewriteLiterals','literal changed'),
 ('native-determinism','swap unit order on repeated split','TestUnitsPreserveSharedState','split changed'),
 ('native-header-flatten','remove system-header provenance markers','TestUnitSystemHeaderProvenance','system-header provenance lost'),
 ('native-header-header-key','drop header bytes from object key; change header macro 1 to 2','TestUnitSystemHeaderProvenance','output="1\\n" want="2\\n"'),
 ('native-header-flags','drop flags and reuse release object','TestUnitCacheFlagsHoldSanitizer','sanitized rebuild reused uninstrumented object'),
 ('native-header-blanket-warning','disable user-code diagnostics globally','TestUnitSystemHeaderProvenance','user-header extension must be rejected'),
 ('inherited-split-switch','remove clearing of inherited opt-in switch','TestWholeBuildSelection','inherited split switch can replace the whole-file witness'),
 ('uncached-selection','remove uncached guard from selected build options','TestWholeBuildSelection','uncached build options:'),
 ('serial-build','remove goroutine launches; findings still agree and owned mutants are caught','compilation_overlap','correct and mutant compilation windows do not overlap'),
 ('parity-output','add puts to only the actual whole-file C witness, compiled and run cleanly','TestSplitWholeParity/no-var','split and whole-file outputs differ')]:
 lines.append(f'| {name} | {change} | `{test}`: `{message}`; [{name}-mutant.log]({name}-mutant.log) |')
lines += ['', 'Every measured author check also runs its owned rule mutant on Node, emitted JavaScript and sanitized native. The mutants are `var declaration suppressed`, `empty function body reported`, and `suggestion applied as fix`; none survived. Native split guards pass without mutations in [native-split-guards.log](native-split-guards.log).', '',
 'Setup completed successfully: [setup.log](setup.log), including tool timings and its complete build-flags/load line. It reports done at 44.164 seconds, nproc 5, cpu.max `400000 100000`, Go 1.27.1, clang 20.1.8, Node 24.19.0; test binaries were deferred by the merged setup script.', '',
 'An initial pre-commit check was refused because harness bytes changed during its oracle build: [input-change-refusal.log](input-change-refusal.log). After edits stopped, the checks passed. This refusal was not a benchmark sample.', '',
 'Limits: this checkout has five registered baseline rules, not a completed fifty-rule fleet. No full repository `go test ./...`, compiler corpus, throughput or profile gate was rerun. The focused lint and native checks above were run. The broad helper invalidation is observed; stable module-qualified numbering and smaller declaration dependency sets remain compiler work, as described in CLANG_UNITS.md. No rule directory, submodule, registry validation or native implementation was edited by this fast-path change.']
(p/'REPORT.md').write_text('\n'.join(lines)+'\n')
print('report written')
