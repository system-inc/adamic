import json
from pathlib import Path
out=Path(__file__).resolve().parent
checks=json.loads((out/'checks.json').read_text());events=[];mutants=[]
for p in sorted(out.glob('*.log')):
 rows=[]
 for line in p.read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if 'Test' in r and r.get('Action') in ['pass','fail']:
   e=dict(log=p.name,test=r['Test'],seconds=r.get('Elapsed',0),result=r['Action']);rows.append(e);events.append(e)
 if p.name.startswith('mutant-'):mutants.append(dict(topic=p.stem.removeprefix('mutant-'),catchers=[r for r in rows if r['result']=='fail']))
(out/'test-times.json').write_text(json.dumps(events,indent=2));(out/'mutants/results.json').write_text(json.dumps(mutants,indent=2))
(out/'test-times.tsv').write_text('log\ttest\tseconds\tresult\n'+''.join(f"{r['log']}\t{r['test']}\t{r['seconds']}\t{r['result']}\n" for r in events))
report='''Built: recut chain followers on current main, preserving lowering and adding conservative freshness compatibility.
Commits: source d4c561e317fd66aaf029fdd1ca12084c12e3af9e; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; delivery SHA in final response.
Commands and outputs: exact bounded commands, exit codes and wall seconds in checks.json; final positive checks passed.
Mutants: all 14 recorded source mutants, seven recorded output mutants, and one added freshness mutant caught.
Not covered: whole packages, full Loom gate, WASI; the explicit no-whole-packages instruction was followed with focused consumer checks.

No squash conflicts. Git automatically merged counts.md, which was then replaced with main's table before exactly one Linux -update-counts run. Main 1f07ce2d removed the digest column again and moved digests into measurement-product keys. This newer infrastructure is retained. Only the fifteen source fixtures were measured; the complete file diff has exactly fifteen added rows and no other changed line. counts.patch and counts-attribution.json list every row.

The additional production fix is internal/fresh/library_language.go: ir.PhantomMember now uses the existing conservative opaque-call model, retaining every operand effect, escaping operands, clobbering parameters and treating the result as outside. Assumption: make no new freshness claim, even though a successful phantom read produces undefined. TestFreshCorpusRemainder exposed the missing node in four phantom fixtures; removing the case again is the added mutant, caught there. No consumer test was changed. Two nested source tests were also made parallel in generic_function_value_test.go and phantom_brands_test.go.

Runtime +/- lines are IDENTICAL to the source: both diffs under internal/native/runtime are empty. runtime-comparison.json contains the arrays. No runtime line differs.

Guards satisfied:

- TestEveryMapAProgramHoldsIsSorted passed. The new node holds no map.
- TestRuntimeSetsAreEachUsed passed. No runtime set was added.
- TestNativeAgreementRowsNameRegisteredPrograms passed. No new native-agreement call was added.
- TestEveryTestIsParallelOrSaysWhy passed. All top-level source tests comply; two nested tests were also repaired.
- TestCallTargetReaders passed with -timeout 90s, again after the last test edits.
- No new package consumes the compiler, so compiler-dependencies.json requires no new entry. No product or test adds a build-directory or machine-path literal.

Focused verification includes source Node, JavaScript backend, sanitized native and leak checks for the fifteen fixtures, seven output mutants, changed stage3 status rows, flow and freshness remainders, native and JavaScript consumers, guards and vet. Vet covers internal/flow, ir, javascript, lower, native, oracle and fresh. All test commands use -count=1 -timeout 90s -parallel 4, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1. Output goes directly to files. Each test's seconds are in test-times.tsv and test-times.json, including parents and intended failures. No test reached 60 seconds; none was deleted or split. The cold counts command includes compilation separately from test time.

Setup: the initial attempt was interrupted with exit 143 during an oversized TypeScript promisor history fetch. Fetching the exact pinned d92d9bfee114c80be2c375d72edae966176e3a4f with --depth=1 succeeded. The next attempt exited 1 because GIT_NO_LAZY_FETCH=1 also disabled a blob needed for checkout. Normal blob fetching then completed setup, exit 0. The printed /workspace/adamic-tools/env.sh was sourced. nproc=5, cpu.max=400000 100000 (four CPUs). clang 20.1.8, Go 1.27.1, Node v24.19.0. Final cumulative timings: Node 0.019s; Go 0.021s; Markdown 0.058s; clang 0.129s; submodules 33.572s; shared cache 38.397s; Go build 324.920s; tests deferred 325.035s; cache warm 325.036s; done 325.089s. All attempts and recovery are logged. An early guard invocation failed before tests because the submodule go.mod was absent; its actual later run passed.

No cohere code was copied. Generic inference refers to the pinned shim. Historical evidence remains byte for byte; git diff --check reports inherited whitespace in historical patch/log files and a fixture EOF, while Go source formatting is clean. No PR or Loom submission was made. The new branch is pushed without force.

Commands and observed outcomes:

'''
report+='\n'.join(f"- {r['label']}: exit {r['exit']}, {r['seconds']}s." for r in checks)
report+='\n\nSource mutants and their catchers:\n\n'+'\n'.join('- '+r['topic']+': '+', '.join(c['test']+f" ({c['seconds']}s)" for c in r['catchers'])+'.' for r in mutants)
report+='''

Recorded output mutants: empty-array-falsiness; fresh-array-sharing; generic-wrong-result; numeric-concatenation-spelling; boolean-concatenation-spelling; undefined-concatenation-spelling; null-concatenation-spelling. Each is caught only by stdout disagreement against Node in both backends; its test rejects build, sanitizer, exit or leak failure as a catcher. Source mutants fail their named assertion or lowering check, rather than invalid C or a Go build failure.

Counts additions:

'''
counts=json.loads((out/'counts-attribution.json').read_text());report+='\n'.join(counts['added'][k] for k in sorted(counts['added']))+'\n';(out/'REPORT.md').write_text(report)
print('Report generated; maximum observed test seconds:',max(r['seconds'] for r in events))
