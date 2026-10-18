import json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent;checks=json.loads((out/'checks.json').read_text());events=[];mutants=[]
for p in sorted(out.glob('*.log')):
 rows=[];errors=[]
 for line in p.read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if 'Test' in r and r.get('Action') in ['pass','fail']:
   e=dict(log=p.name,test=r['Test'],seconds=r.get('Elapsed',0),result=r['Action']);rows.append(e);events.append(e)
  if r.get('OutputType')=='error':errors.append(r.get('Output',''))
 if p.name.startswith('mutant-'):mutants.append(dict(topic=p.stem.removeprefix('mutant-'),catchers=[r for r in rows if r['result']=='fail'],errors=errors))
(out/'test-times.json').write_text(json.dumps(events,indent=2));(out/'mutants/results.json').write_text(json.dumps(mutants,indent=2));(out/'test-times.tsv').write_text('log\ttest\tseconds\tresult\n'+''.join(f"{r['log']}\t{r['test']}\t{r['seconds']}\t{r['result']}\n" for r in events))
report='''Built: nullable string/object unions, checked stored scalar tags and optional boolean class fields recut on current main.
Commits: source bb8b389a3c3c567c29afe465e246c3ca9466f835; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; delivery SHA in final response.
Commands and outputs: exact bounded argv, exits and command seconds in checks.json; final positive checks passed.
Mutants: all 22 recorded source mutants and the recorded IR field-check mutant caught.
Not covered: full packages/full gate, WASI, general callable/object-kind unions, optional writes/delete and hidden never-array lowering.

No squash conflicts. Automatic merges in expression.go preserve main's newer infrastructure beside the source's lowering. counts.md was taken from main before exactly one Linux -update-counts run using house clang 20.1.8. Main 1f07ce2d has already removed the digest column again; digests key measurement products, and that current format is preserved.

Counts: sixteen new admitted fixture rows and one changed existing row, taste_stage3_representations.a: retains/releases 11/34 to 13/36; allocations, frees, peak and regions unchanged. A current-tree counterfactual overlay disables only the new checked string-field route and measures 11/34 again (counts-cause.log). The same overlay returns scout_boxed_scalar_fields.a to 95/150 from 114/169. Thus each changed row belongs to this lowering. counts-attribution.json and counts.patch hold every row. Only the admitted source fixtures and the known affected existing taste fixture were regenerated; no Mac table was retained.

The source placed six deliberate NotYet storage probes among runnable root fixtures. Main's strict flow corpus correctly failed to lower them. They are now under internal/oracle/testdata/scout_nullable_storage/, with unchanged bytes and all six existing oracle tests pointing to them. All six still run against Node and require the explicit lowering stop. No test was deleted and no flow/freshness test was edited. The touched taste subtests now call t.Parallel first. These are the only changes beyond the squash and regenerated counts.

Runtime +/- lines under internal/native/runtime are IDENTICAL to the source: both empty. The native union.go conversion change outside that directory is retained as written. runtime-comparison.json records the empty arrays.

All requested guards passed: TestEveryMapAProgramHoldsIsSorted (no new IR maps), TestRuntimeSetsAreEachUsed (no new runtime set), TestNativeAgreementRowsNameRegisteredPrograms (no new agreement call), TestEveryTestIsParallelOrSaysWhy (plus parallel nested taste leaves), and TestCallTargetReaders with -timeout 90s. No new compiler-consuming package or build product was added, so compiler-dependencies.json needs no entry. Products/tests add no build-directory or machine-path literals.

Focused checks cover admitted fixtures against source Node and both backends, native ASan/UBSan/leaks, the six storage refusals, checked field IR mutant, RegExp callback adapter, the two changed stage1 gap leaves, flow/freshness remainders, native consumer checks, guards and vet. Vet covers lower, native, oracle, stage1/cohere/selector and stage1/cohere/values. Every test uses -count=1 -timeout 90s -parallel 4 with GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1; subprocesses and drivers have hard limits. Output goes directly to logs. Individual leaf and parent seconds are in test-times.tsv/json; all are under 60 seconds. No split was needed.

The unit's shared setup completed with clang 20.1.8, Go 1.27.1, Node v24.19.0; /workspace/adamic-tools/env.sh was sourced for this branch. nproc=5, cgroup cpu.max=400000 100000 (four CPUs). setup.log preserves timing lines: Node 0.019s, Go 0.021s, Markdown 0.058s, clang 0.129s, submodules 33.572s, shared cache 38.397s, Go build 324.920s, tests deferred 325.035s, cache warm 325.036s, done 325.089s. The initial TypeScript history fetch was interrupted; a depth-1 exact-pin fetch and normal blob materialization recovered it, as the first branch report records.

Assumption: deliberate NotYet probes belong in their own fixture directory, retaining their Node/refusal tests rather than weakening another package's runnable corpus. No cohere code was copied. No PR or Loom submission was made. One commit is pushed only to compiler/loom-unions-3, without force.

Commands and observed outcomes:

'''
report+='\n'.join(f"- {r['label']}: exit {r['exit']}, {r['seconds']}s." for r in checks)
report+='\n\nRecorded source mutants and observed catchers:\n\n'+'\n'.join('- '+r['topic']+': '+', '.join(c['test']+f" ({c['seconds']}s)" for c in r['catchers'])+'.' for r in mutants)
report+='\n\nThe recorded IR field-check removal is held by TestNullableStringFieldCheckMutant: the checked baseline panics with exit 70; the unchecked JavaScript mutant prints undefined and exits 0. Source mutant errors are in mutants/results.json and the logs; Go compilation, clang diagnostics, panics and timeouts are not counted as semantic catchers.\n\nCounts additions:\n\n';counts=json.loads((out/'counts-attribution.json').read_text());report+='\n'.join(counts['added'][k] for k in sorted(counts['added']))+'\n';report+='\nExisting changed row:\n'+json.dumps(counts['changed_existing'],indent=2)+'\n';(out/'REPORT.md').write_text(report)
print('Report generated; maximum test seconds',max(e['seconds'] for e in events))
