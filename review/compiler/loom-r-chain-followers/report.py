import json,re,subprocess
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-r-chain-followers';base=(out/'base.txt').read_text().strip();source='f5a13b10596bcd419badbd39f4fb229dbbdff203';assert base=='dbdf977f5b8b0979d414ebe6f20e7aadd312e309'
checks=json.loads((out/'checks.json').read_text());assert len(checks)==32,len(checks);assert all(c['exit']==(1 if c['label'].startswith('mutant-') else 0) for c in checks)
def events(label):
 return [json.loads(l) for l in (out/(label+'.log')).read_text().splitlines() if l.startswith('{')]
focus=json.loads((out/'focus.json').read_text())
for label in ['fixtures','counts-regeneration']:
 actual={r['Test'].split('/',1)[1] for r in events(label) if r.get('Action')=='pass' and '/internal/oracle/testdata/' in r.get('Test','')};assert actual==set(focus['fixtures']),(label,actual)
for pkg,names in focus['tests'].items():
 passed={r['Test'] for r in events(pkg.split('/')[-1]+'-focus') if r.get('Action')=='pass' and 'Test' in r};assert set(names)<=passed
# The corpus globs own each added top-level fixture; nested scanner fixtures are oracle-owned.
top={p.replace('internal/','../',1) for p in focus['fixtures'] if p.count('/')==3};flow={r['Test'].split('/',1)[1] for r in events('flow-consumers') if r.get('Action')=='pass' and r.get('Test','').startswith('TestFlowCorpusRemainder/')};assert top<=flow,(top-flow)
fresh={r['Test'].split('/',1)[1] for r in events('fresh-consumers') if r.get('Action')=='pass' and r.get('Test','').startswith('TestFreshCorpusRemainder/')};assert top<=fresh,(top-fresh)
mutants=[]
for case in json.loads((out/'mutants/cases.json').read_text()):
 label='mutant-'+case['topic'];rows=events(label);failed=[r for r in rows if r.get('Action')=='fail' and 'Test' in r];assert failed and '[build failed]' not in (out/(label+'.log')).read_text()
 mutants.append(dict(topic=case['topic'],kind='source overlay',caught=True,catchers=[dict(test=r['Test'],seconds=r.get('Elapsed',0)) for r in failed],errors=[r['Output'].strip() for r in rows if r.get('OutputType')=='error']))
ir_mutants={'empty-array-falsiness':'TestEmptyArrayRuntimeMutants/empty_array_is_falsy','fresh-array-sharing':'TestEmptyArrayRuntimeMutants/fresh_number_array_is_shared','generic-wrong-result':'TestGenericFunctionValueWrongResultMutant','numeric-concatenation-spelling':'TestScannerConcatenationMutant','boolean-concatenation-spelling':'TestScannerPrimitiveConcatenationMutants/boolean','undefined-concatenation-spelling':'TestScannerPrimitiveConcatenationMutants/undefined','null-concatenation-spelling':'TestScannerPrimitiveConcatenationMutants/null'}
passed={r['Test']:r for r in events('oracle-focus') if r.get('Action')=='pass' and 'Test' in r}
for topic,test in ir_mutants.items():
 assert test in passed;mutants.append(dict(topic=topic,kind='IR output mutant',caught=True,catchers=[dict(test=test,seconds=passed[test].get('Elapsed',0))],evidence='Passing test requires stdout-only disagreement against source Node in native and JS backends, with build, sanitizer, exit and leak failures rejected.'))
assert len(mutants)==22;(out/'mutants/results.json').write_text(json.dumps(mutants,indent=2))
times=[]
for log in sorted(out.glob('*.log')):
 for line in log.read_text().splitlines():
  if not line.startswith('{'):continue
  try:r=json.loads(line)
  except json.JSONDecodeError:continue
  if r.get('Action') in ['pass','fail','skip'] and 'Test' in r:times.append(dict(log=log.name,test=r['Test'],seconds=r.get('Elapsed',0),outcome=r['Action']))
assert all(r['seconds']<60 for r in times);maximum=max(r['seconds'] for r in times)
(out/'test-times.json').write_text(json.dumps(times,indent=2));(out/'test-times.tsv').write_text('log\ttest\tseconds\toutcome\n'+''.join(f"{r['log']}\t{r['test']}\t{r['seconds']:.3f}\t{r['outcome']}\n" for r in times))
counts=json.loads((out/'counts-attribution.json').read_text());assert len(counts['added'])==15 and not counts['changed_existing'];runtime=json.loads((out/'runtime-comparison.json').read_text());assert not runtime['changed_lines'] and runtime['source_own_runtime_diff']==runtime['rebased_runtime_diff']==''
assert subprocess.check_output(['git','diff',base,'--','internal/native/runtime'],text=True)==''
setup=[line for line in (out/'setup.log').read_text().splitlines() if line.startswith('setup:')]
report=f'''Built: applied only f5a13b1's own compiler change to pinned main dbdf977f5b as compiler/loom-r-chain-followers.
Commits: source {source}; exact base {base}; delivery SHA in the final message.
Commands and outputs: all focused positive checks exit 0; exact argv, exits and command seconds in checks.json, all output in logs.
Mutants: all 22 from the source report caught: fifteen source overlays fail named tests, seven in-test output mutations disagree with Node in both backends.
Not covered: whole packages, the full Loom gate, WASI or ahra loom check; unchanged main count rows were preserved rather than remeasured.

Base and application: fetched origin and the exact source SHA, then created compiler/loom-r-chain-followers from {base}. Used git cherry-pick --no-commit {source}, so the applied change is source~1..source, never main..source. The base stays pinned to the requested SHA even if origin/main advances. The inherited source scope and its exclusions are unchanged. Existing consumer tests were not edited to fit this change. New evidence is under review/compiler/loom-r-chain-followers; older source reports are retained as historical evidence.

Conflict and resolution, one file and one hunk:

internal/oracle/counts.md, combined hunk @@ -427,19 -427,9 +427,25 @@ after array_literal_empty_union.a. Pinned main contributes thirteen async rows (async_coverage_discard/parameters/reject_eager/reject_empty/reject_nested/typeof/unions/values, async_nested/plain/three/throw/typeof). The source adds arrow_this_super/field/stored at the same insertion point. Saved conflicts.patch, then used git checkout --ours -- internal/oracle/counts.md, preserving pinned main's complete table without hand-merging any row. After the compiler and fixture checks passed, ran exactly one Linux -update-counts command with house clang 20.1.8, selecting the fifteen source fixture leaves as supported by counts_test.go's writeCounts. The canonical writer inserted the source rows in registration order, preserving the async rows and every other main row. No other file conflicted. The automatic merges retain pinned main's newer compiler behavior.

Counts: exactly fifteen added rows, no changed or removed existing row. All {counts['preserved_main_rows']} pinned-main rows keep their original values. TestCountsAreRecordedTable passed afterward. Counts refresh follows the command in counts.md's header and the partial-refresh preservation contract documented above writeCounts in internal/oracle/counts_test.go. The fifteen added rows are attributable one-for-one to this change's new registered oracle fixtures; the source and all fixtures compiled before the one refresh. counts-main.md.txt is the untouched base table, and counts-attribution.json independently compares every row.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
'''+''.join(v+'\n' for v in counts['added'].values())+'''
Runtime C lines: NONE. git diff source~1 source -- internal/native/runtime is empty, as is git diff base HEAD/worktree -- internal/native/runtime. Thus the runtime +/- lines are identical to the source's own runtime lines, with no line needing runtime clearance. runtime-comparison.json records both diffs and an empty changed_lines array. Code generation changes live in Go, as in the source; main's runtime C remains intact.

Verification and reader coverage:

'''+''.join(f"- {r['label']}: exit {r['exit']}, {r['seconds']:.3f} command seconds.\n" for r in checks if not r['label'].startswith('mutant-'))+f'''
All 15 exact source oracle fixture leaves passed uncached against source Node, JavaScript emission, native sanitizers and leak checks. All 12 added top-level fixtures also passed the flow and freshness corpus remainders; flow's 27-program remainder runs single assignment, mutation ranges, graph paths and liveness, retaining its independent Node observations, and finished in 4.500 test seconds. Freshness's remainder passed in 1.450 seconds. The three nested scanner fixtures are held by their oracle leaves. Every new top-level fixture lowers; no newly added refused top-level fixture needed relocation to a *_refused directory. No consumer or corpus list was changed. Source lowering tests, native memory-plan consumers, a JavaScript representation consumer, the seven changed stage3 status rows, and IR readers passed. Setup's authorized go build also compiled all repository packages, retaining main's product key behavior.

TestCallTargetReaders used -timeout 90s. go vet passed on internal/flow, fresh, ir, javascript, lower, native and oracle. Guards also passed for SortedMap-only IR program maps, runtime-set witnesses, native agreement registration, and test parallelism. The new node holds no map, no runtime feature set was added, no package was added, and no compiler-consumer declaration is needed. Products and tests introduce no build-directory or machine-path literals. Changed Go files are gofmt clean. No code was copied from cohere; the source retains its pinned checker/shim inference reference.

Each test's seconds, including parent tests, leaves and intentional mutant failures, are in test-times.tsv and test-times.json. Maximum measured test: {maximum:.3f} seconds. All runs use GOMAXPROCS=4, -parallel 4, -count=1, -timeout 90s and ADAMIC_GATE_UNCACHED=1 on one four-CPU quota instance. No test crossed 60 seconds, so none was deleted, split or rerun after crossing the limit. Compilation/startup time is recorded separately as command wall seconds. Each test command has its Go timeout and a 240-second subprocess hard limit; the runner has a 900-second limit and was monitored. Test stdout/stderr went directly to log files and was never piped.

Mutants and catchers:

'''+''.join('- '+m['topic']+' ('+m['kind']+'): '+', '.join(c['test']+f" ({c['seconds']:.3f}s)" for c in m['catchers'])+'.\n' for m in mutants)+'''
The fifteenth source mutant is phantom-fresh from the latest source REPORT.md: removing ir.PhantomMember from the existing conservative opaque-call freshness model is caught by TestFreshCorpusRemainder. Every stored source mutant is reconstructed as only its mutation delta against f5a13b1, then applied to the rebased file; an old whole-file snapshot never replaces main's newer code. Each source run exited 1 with named test failures, not a Go build failure. Arrow mutants are caught by IR frame assertions before C emission. The seven in-test IR mutations require stdout-only disagreement with Node in both backends; build/sanitizer/exit/leak failures do not qualify. Exact failure messages and catchers are retained in mutants/results.json and per-mutant logs.

Toolchain and setup: exported GOPROXY=https://proxy.golang.org|direct; ran timeout 600 bash cloud/setup.sh and sourced its printed /workspace/adamic-tools/env.sh. Setup succeeded and proved the sanitizer with a real heap-buffer-overflow probe. nproc=5, cpu.max=400000 100000, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node v24.19.0. Exact timing lines:

'''+''.join('- '+line+'\n' for line in setup)+'''
Scope assumptions: the named source commit's own diff is authoritative, including its conservative opaque-call freshness treatment and parallel source tests. No freshness precision is inferred beyond that model. The pinned base overrides the generic current-main instruction. The full gate belongs to Loom, so checks are focused rather than whole-package runs. No ahra CLI, PR creation or Loom submission was run. The old branches remain untouched and the new branch is pushed without force. Historical evidence bytes and inherited fixture whitespace are preserved.
'''
(out/'REPORT.md').write_text(report);print('Validated 22 mutants, 15 fixture/count leaves, 12 new flow/fresh top-level programs and',len(times),'test outcomes; max',maximum)
