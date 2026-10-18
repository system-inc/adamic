import json
from pathlib import Path
out=Path(__file__).resolve().parent;checks=json.loads((out/'checks.json').read_text());events=[];mutants=[]
for p in sorted(out.glob('*.log')):
 rows=[];errors=[]
 for line in p.read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if not isinstance(r,dict):continue
  if 'Test' in r and r.get('Action') in ['pass','fail']:
   e=dict(log=p.name,test=r['Test'],seconds=r.get('Elapsed',0),result=r['Action']);rows.append(e);events.append(e)
  if r.get('OutputType')=='error':errors.append(r.get('Output',''))
 if p.name.startswith('mutant-'):mutants.append(dict(topic=p.stem.removeprefix('mutant-'),catchers=[r for r in rows if r['result']=='fail'],errors=errors))
assert len(mutants)==5 and all(r['catchers'] for r in mutants)
(out/'test-times.json').write_text(json.dumps(events,indent=2));(out/'mutants/results.json').write_text(json.dumps(mutants,indent=2));(out/'test-times.tsv').write_text('log\ttest\tseconds\tresult\n'+''.join(f"{r['log']}\t{r['test']}\t{r['seconds']}\t{r['result']}\n" for r in events))
admission=json.loads((out/'admission.json').read_text())
report='''Built: structural phantom-string proof recut on current main, with no additional language or lowering changes.
Commits: source ff53652e64d178017e71f9f4328937e41e877d9b; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; delivery SHA in final response.
Commands and outputs: exact bounded test argv, exits and wall seconds in checks.json; all positive checks passed.
Mutants: all five recorded source mutants caught by their named checks, without build errors or timeouts.
Not covered: full packages/full gate, WASI, historical full-mode 44-site closure census, or Workshop runner-cache shipping parity.

No squash conflicts. counts.md automatically merged, then was replaced byte-for-byte with main's table before exactly one Linux -update-counts run. The entire file diff is exactly six added source-fixture rows, with no deleted or changed existing row. Main 1f07ce2d removed the digest column again, placing digests in measurement-product keys; this newer infrastructure is retained. counts.patch and counts-attribution.json contain every row. No Mac table was retained.

Production, docs and tests retain the source's +/- lines: no manual lowering or test fix was necessary. Runtime +/- lines under internal/native/runtime are IDENTICAL to the source, both empty; runtime-comparison.json records both arrays. No new runtime set, IR map, package or product declaration was added. No build-directory or machine-path literal was added to products or tests.

All requested guards passed: TestEveryMapAProgramHoldsIsSorted, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy, and TestCallTargetReaders with -timeout 90s. The gob stability/nil SortedMap tests passed too. The proof uses existing IR String and Defined nodes. Existing packages need no new compiler-dependencies.json declaration. Every test already calls t.Parallel first; no test was deleted or split.

The focused lower tests hold representation exclusions, primitive object-proof isolation, literal constraints and missing-value checks. All six admitted fixtures agree with source Node, JavaScript backend, release native and native ASan/UBSan, with leak checks. Refusal tests retain runtime-layout/member-read and unknown-cast stops. Flow/freshness remainders, native consumers, guards and vet on lower/oracle passed. All test commands use -count=1 -timeout 90s -parallel 4, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1, with bounded drivers and subprocesses. Output goes directly to logs. test-times.tsv/json lists every observed leaf and parent; all are below 60 seconds.

Fresh admission comparison: built current-main and recut CLIs, restoring main's four existing production files through a baseline overlay, not older snapshots. The new proof file has no side effects and no main hook calls it in the baseline. Compared all 1,366 current oracle .a/.ts sources (excluding .d.ts), with four workers and a 20-second per-CLI hard limit: exactly six new admissions, all string_brand_ fixtures, and zero losses. Both first-error reasons, value of type __String | undefined and function returning __String | undefined, go from 1 to 0. admission.json records every observation and both binary hashes; admission-builds.json records exact successful build commands and seconds. Temporary CLI binaries were not committed. This is a first-error scan, not the historical full-mode closure census.

The unit's shared setup used GOPROXY=https://proxy.golang.org|direct and the house clang 20.1.8; Go 1.27.1, Node v24.19.0. /workspace/adamic-tools/env.sh was sourced. nproc=5, cgroup cpu.max=400000 100000 (four CPUs). setup.log records Node 0.019s, Go 0.021s, Markdown 0.058s, clang 0.129s, submodules 33.572s, shared cache 38.397s, Go build 324.920s, tests deferred 325.035s, cache warm 325.036s and done 325.089s. The initial TypeScript history fetch was interrupted; exact-pin shallow fetch followed by normal blob materialization recovered setup, as the first branch report records.

Assumption: preserve current main's counts measurement-product format instead of restoring its obsolete digest column. No cohere code was copied. Go formatting and Go whitespace checks passed. No PR or Loom submission was made. One commit is pushed only to compiler/loom-string-brand-3, without force.

Commands and observed outcomes:

'''
report+='\n'.join(f"- {r['label']}: exit {r['exit']}, {r['seconds']}s." for r in checks)
report+='\n\nRecorded source mutants and catchers:\n\n'+'\n'.join('- '+r['topic']+': '+', '.join(c['test']+f" ({c['seconds']}s)" for c in r['catchers'])+'.' for r in mutants)
report+='\n\nCounts additions:\n\n| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |\n|---|---:|---:|---:|---:|---:|---:|\n'+'\n'.join(json.loads((out/'counts-attribution.json').read_text())['added'])+'\n';(out/'REPORT.md').write_text(report)
print('Report generated; maximum observed test seconds:',max(e['seconds'] for e in events))
