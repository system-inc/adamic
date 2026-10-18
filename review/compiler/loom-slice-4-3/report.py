from pathlib import Path
import json
out=Path(__file__).resolve().parent;old=out.parent/'loom-slice-4'
for file,count in [('generic-mutants.log',3),('generator-mutants.log',6),('loader-mutants.log',14),('iteration-mutants.log',4),('extra-mutants.log',3)]:
 text=(out/file).read_text();assert text.count('caught')==count,(file,text)
assert len(json.loads((out/'reconciliation-mutants.json').read_text()))==3
scouts=json.loads((out/'scout-results.json').read_text());assert len(scouts)==10 and all(r['exit']==r['expected'] for r in scouts)
focused=json.loads((out/'focused-results.json').read_text());assert all(r['exit']==0 for r in focused)
fixtures=json.loads((out/'fixture-results.json').read_text());assert all(r['exit']==0 for r in fixtures) and sum(r['leaves'] for r in fixtures)==57
counts=json.loads((out/'counts-attribution.json').read_text());assert counts['added']==55 and counts['changed']==4 and not counts['removed'] and all(r['matches_source'] for r in counts['rows'])
rows=[]
for p in out.rglob('*.jsonl'):
 for line in p.read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if e.get('Test') and e.get('Action') in ['pass','fail','skip']:rows.append(dict(log=str(p.relative_to(out)),**e))
assert not any(r['Action']=='fail' or r.get('Elapsed',0)>=60 for r in rows)
(out/'test-seconds.json').write_text(json.dumps(rows,indent=2)+'\n')
summary=json.loads((old/'mutants-summary.json').read_text())
summary['observed_in_this_recut']=True
(out/'mutants-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
text='''Re-cut all five slice 4 members: generic body relations, generics scout, iteration, project references and generators, retaining conservative protocol accessor effects.
Delivery is one commit on main d79302e50c022c997c00f7c68ddb7475f9ee65fc; source bad75ff43f66d8a89b17ffb7a48af039f170defe; the handoff supplies the delivery SHA.
Focused named tests, 57 differential fixture leaves plus user_iterators, stock-TypeScript project controls, counts, structural guards, call-target readers and changed-package vet passed.
All 49 recorded mutants were caught again, including the two accessor-effect mutants and the reader ownership mutant; mutants-summary.json lists every name and catcher.
Not covered: whole-package suites, all transitive consumer suites, the full gate, opt-in generic-body census, WASI, admission census or benchmarks.

The branch starts from freshly fetched origin/main. git merge --squash origin/compiler/loom-slice-4-2 has no conflicts. Main did not move between the two re-cuts. No implementation or test adaptation was necessary beyond the requested squash: main's later infrastructure merged cleanly with this lowering. One recorded mutant needed a harness-only adaptation: main now makes nodeTypesIndex return a rooted path, so standalone-node-prelude changes append(checkRoots, index) to append(roots, index), instead of the former ResolveFile expression. The missing-site assertion is retained in the log and is not credited as a mutant kill. The first 12 loader mutants were not repeated; loader-mutants-resume.py ran the two remaining cases. The new evidence supersedes historical run results retained in the source's review trees.

Main infrastructure retained: LoadDisplayed and sourceFS display callbacks; placeless GoExports and runtime products; the literal nativeAgreementPrograms registry; main's existing compiler-dependency declarations; and the current prepare-and-measure counts API. The source adds stage3/project-loader/probe depending on internal/load to cloud/fast-gate/compiler-dependencies.json while preserving all current entries. The typed call-target audit includes oracle test readers with explicit ownership and reasons. The unowned-reader mutant proves this extension rejects an unapproved reader.

The task described a digest column, but this fetched main keeps numeric rows in counts.md and digests in keyed count-measurement products. Main's exact counts.md was restored before one successful Linux -update-counts invocation. It completed in 111.726 seconds. Exactly 55 new rows and four existing rows move; no rows are removed and no unrelated row moves. All 59 measured values match the source slice exactly. counts-attribution.json lists every row, before/after values, source equality and cause.

The four existing rows, allocations/frees/retains/releases/peak/regions:

- regexp.a: 437/437/382/417/58/0 to 437/437/383/418/58/0, generator result boxed-field ownership.
- user_iterators.a: 663/663/455/904/64/0 to 1050/1050/677/994/97/0, cached protocol dispatch and close ownership.
- user_iterators_rest_tdz.a: 5/0/5/3/5/0 to 11/3/7/3/8/0, cached protocol dispatch and close ownership.
- internal/fresh/testdata/regexp_tree.ts: 91/91/81/85/35/0 to 91/91/85/89/35/0, generator result boxed-field ownership.

Neither the source nor this re-cut changes internal/native/runtime C or header lines. Their +/- sequences are identical and empty; runtime-comparison.json records this. No runtime difference was introduced.

Every requested guard passes:

- TestEveryMapAProgramHoldsIsSorted: all Program-held maps remain SortedMap.
- TestRuntimeSetsAreEachUsed: every existing feature set still has an exact witness; no new feature set or registry entry was necessary. Counts and native fixtures require no undeclared set.
- TestNativeAgreementRowsNameRegisteredPrograms: every native agreement call still names its registered program by literal; no bypass or exemption was added.
- TestEveryTestIsParallelOrSaysWhy: source tests already parallelize first or have the required immediately preceding reason.
- TestCallTargetReaders, -timeout 90s: 17.49 seconds including GoExports setup; source ownership declarations remain enforced.
- New compiler consumer: stage3/project-loader/probe is declared in the manifest, as above.
- Products retain main's displayed names, prefix maps and portable keys. No build-directory name was introduced in a product or new test.
- No observed test leaf reaches 60 seconds. The largest observed leaf is 21.68 seconds, user_iterators. Every test's seconds, including new test leaves, are in test-seconds.json. No test was deleted, weakened, or required a further split.

Exact validation:

- focused.py selects named source tests in flow, fresh, load, lower and oracle, 12 per bounded command, requiring each selected name to emit a terminal event. All pass. It includes the source's embedded protocol, numeric/optional result, borrowed-across-suspension and enum mutants. The generic-body census retains its documented opt-in skip.
- fixtures.py: all 57 selected leaves pass uncached source Node, backend Node, native ASan/UBSan, release native and completed-program leak checks. The separate TestNativeAgreesWithNode_user_iterators passes in 21.689 seconds.
- verify.py and verify_types.py: stock TypeScript source-root, transitive/diamond, ambient-union, duplicate declaration, conflicting-option and diagnostic controls pass. Sanitized native and JavaScript outputs agree with stock/source Node. Temporary .ts files are stock-TypeScript adapters of existing .a fixtures, not newly authored Adamic programs or delivered source files. Scratch binaries and generated adapters are excluded from the commit; RESULT.json receipts are retained.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -timeout 180s -args -update-counts: pass, one rewrite.
- go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run the four structural guard names -count=1 -json -timeout 90s: pass; guards.jsonl holds exact names and results.
- go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s: pass, 17.49-second leaf.
- go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/oracle ./stage3/project-loader/probe: exit 0, no diagnostics.
- gofmt -l on changed Go files: no output. Product/source whitespace is clean. Historical raw logs, patch artifacts and the upstream TEST262 license retain their original bytes.

Every command used setup's env.sh, GOPROXY=https://proxy.golang.org|direct, GOMAXPROCS=4, -p=4 and main's -buildvcs=false -trimpath. Test output went directly to files. Setup completed successfully: Go ready 0.019s, Node ready 0.021s, submodules 0.066s, markdown ready 0.067s, clang ready 0.161s, shared cache ready 0.437s, done 69.458s. All setup timing lines are in setup-timings.json, with the original local setup.log. nproc=5, cpu.max=400000 100000, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node 24.19.0. Tests unset GOCACHEPROG to use the local cache.

All recorded mutant catchers follow. The independent runners require their intended behavior/assertion and reject a build failure as a kill. The source's two borrow mutants require ASan heap-use-after-free, and the cycle mutant requires LeakSanitizer. Four scout controls pass and all six scout mutants fail with their expected assertions.

| Member | Mutant | Catcher |
|---|---|---|
'''
for r in summary['mutants']:text+='| '+r['member']+' | '+r['mutant']+' | '+r['catcher']+' |\n'
text+='''
The exact selectors and exits are in focused-results.json, fixture-results.json, mutants.json, extra-mutants.json, scout-results.json, reconciliation-mutants.json, overlay descriptors and the bounded runner scripts. Individual .log files stay under this review directory locally, in accordance with main's rule that stray logs do not reach commits.

Conservative semantic assumption: protocol property lookup can run an arbitrary accessor and publish its receiver, even for a scalar result. IteratorField and IteratorMethod retain the source's conservative arbitrary-call effects; both pure-load mutants fail their receiver-escape assertions. No precision improvement is assumed. No cohere code was copied, no PR was opened, no main push or force push occurred, and nothing was submitted to Loom.
'''
(out/'REPORT.md').write_text(text)
