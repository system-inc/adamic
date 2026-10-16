import json,subprocess
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-chain-followers-2'
base=(out/'main-base.txt').read_text().strip();assert subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip()==base
checks=json.loads((out/'checks.json').read_text());assert len(checks)==23,len(checks)
assert all(r['exit']==(1 if r['label'].startswith('mutant-') else 0) for r in checks)
def events(label):
 return [json.loads(line) for line in (out/(label+'.log')).read_text().splitlines() if line.startswith('{')]
focus=json.loads((out/'focus.json').read_text())
for label in ['counts-regeneration','fixtures']:
 actual={r['Test'].split('/',1)[1] for r in events(label) if r.get('Action')=='pass' and '/internal/oracle/testdata/' in r.get('Test','')}
 assert actual==set(focus['fixtures']+focus['main_witnesses']),(label,actual)
for package,names in focus['tests'].items():
 passed={r['Test'] for r in events(package.split('/')[-1]+'-focus') if r.get('Action')=='pass' and 'Test' in r};assert set(names)<=passed,(package,set(names)-passed)
mutants=[]
for c in json.loads((out/'mutants/cases.json').read_text()):
 label='mutant-'+c['topic'];result=events(label);failed=[r for r in result if r.get('Action')=='fail' and 'Test' in r];assert failed and '[build failed]' not in (out/(label+'.log')).read_text()
 mutants.append(dict(topic=c['topic'],kind='source overlay',file=c['file'],caught=True,catchers=[dict(test=r['Test'],seconds=r.get('Elapsed',0)) for r in failed],errors=[r['Output'].strip() for r in result if r.get('OutputType')=='error']))
ir_mutants={
 'empty-array-falsiness':'TestEmptyArrayRuntimeMutants/empty_array_is_falsy',
 'fresh-array-sharing':'TestEmptyArrayRuntimeMutants/fresh_number_array_is_shared',
 'generic-wrong-result':'TestGenericFunctionValueWrongResultMutant',
 'numeric-concatenation-spelling':'TestScannerConcatenationMutant',
 'boolean-concatenation-spelling':'TestScannerPrimitiveConcatenationMutants/boolean',
 'undefined-concatenation-spelling':'TestScannerPrimitiveConcatenationMutants/undefined',
 'null-concatenation-spelling':'TestScannerPrimitiveConcatenationMutants/null'}
passed={r['Test']:r for r in events('oracle-focus') if r.get('Action')=='pass' and 'Test' in r}
for topic,test in ir_mutants.items():
 assert test in passed,test
 mutants.append(dict(topic=topic,kind='IR output mutant',caught=True,catchers=[dict(test=test,seconds=passed[test].get('Elapsed',0))],evidence='The passing test requires clean native and JS execution and stdout-only disagreement with source Node.'))
assert len(mutants)==21
(out/'mutants/results.json').write_text(json.dumps(mutants,indent=2))
times=[]
for log in sorted(out.glob('*.log')):
 for line in log.read_text().splitlines():
  if not line.startswith('{'):continue
  r=json.loads(line)
  if r.get('Action') in ['pass','fail','skip'] and 'Test' in r:
   times.append(dict(log=log.name,test=r['Test'],seconds=r.get('Elapsed',0),outcome=r['Action']))
assert all(r['seconds']<60 for r in times)
(out/'test-times.json').write_text(json.dumps(times,indent=2))
(out/'test-times.tsv').write_text('log\ttest\tseconds\toutcome\n'+''.join(f"{r['log']}\t{r['test']}\t{r['seconds']:.3f}\t{r['outcome']}\n" for r in times))
counts=json.loads((out/'counts-attribution.json').read_text());assert len(counts['added'])==15 and not counts['changed_existing']
paths=subprocess.check_output(['git','diff','--name-only',base],text=True).splitlines();runtime=[p for p in paths if p.startswith('internal/native/runtime/')];assert not runtime
implementation=[p for p in paths if p.startswith('internal/') and p.endswith('.go') and not p.endswith('_test.go')]
report=f'''Built: parked change chg_ws3ng9gpgkm7f18c9n12fygt1d recut onto current main as compiler/loom-chain-followers-2.
Commits: source 376febb44da1cc8018f1382f0bc488941375fcdc; base main {base}; final SHA in the delivery message.
Commands and outputs: 23 bounded focused commands completed, positive checks exit 0 and 14 source-mutant runs exit 1 at their named tests.
Mutants: all 21 caught, 14 source overlays plus seven native/JavaScript stdout mutants held against Node.
Not covered: full packages/full Loom gate, WASI, or dependency-blocked tasks already excluded from the authoritative source.

The branch begins at main {base}. Used git merge --squash 376febb44da1cc8018f1382f0bc488941375fcdc. Main was unchanged at the final fetch. One commit will be pushed to the new branch, with no force push, PR or Loom submission; the old branch is untouched.

Exactly two conflicts:

- internal/lower/exceptions.go: kept both adjacent switch cases in throwsOut. Main's c6e584809ceaf4822c96b9bf91d6ab15b6db6269, "Three lowering gaps, on main", landed #q9hz1bf's object-spread accessor throw edge. Its ir.ObjectLiteral case calls l.result.ObjectSpreadMayThrow(node), so throws from enumerable literal getters propagate through callers and cleanup. Kept that unchanged, and also kept the source's ir.PhantomMember case, found = found || !node.Optional, for a non-optional read that can throw a TypeError on undefined. Both types participate in throw propagation. The recorded landed diff is exceptions-landed.patch.
- internal/oracle/counts.md: took main with git checkout --ours, not the parked table. Regenerated once after the compiler resolution, measuring all fifteen source fixtures plus main's seven new witnesses. Main's {counts['main_rows_preserved']} existing rows retain identical values; exactly fifteen rows are added. TestCountsAreRecordedTable then passed. No prior failed regeneration or second regeneration occurred.

The automatic merges in internal/flow/build.go and internal/lower/expression.go retain main's object-spread flow edge, static constructor spread handling and nullish template work alongside the source's phantom brands, primitive concatenation and contextual generic function values. Existing tests, including main's, were not edited. No test was removed or ran 60 seconds or more; no split was needed.

Focused verification, all uncached with ADAMIC_GATE_UNCACHED=1, GOMAXPROCS=4, -parallel 4, -count=1 and -timeout 90s:

'''+''.join(f"- {r['label']}: exit {r['exit']}, {r['seconds']:.3f} command seconds.\n" for r in checks if not r['label'].startswith('mutant-'))+'''
checks.json gives every exact argv and command wall time; focus.json names all tests and fixtures. Tests write stdout and stderr directly to per-command logs, never a pipe. Each subprocess has a 240-second hard limit, the runner a 900-second limit. The counts command measures only these 22 fixture leaves and preserves unmeasured rows. The oracle run independently verified exactly those same 22 leaves, including Node truth, JS backend, native release/sanitizers and leak checks. Source fixtures are fifteen; main witnesses are accessor_spread_throw_first/middle/last, plain_data_spread, static_constructor_spread_gap/owned and template_nullish. The seven changed stage3 status rows also passed, preserving their recorded Node bytes. The lower and IR checks include TestClassFeaturesThrowingAccessorSpread and TestObjectSpreadAccessorEffects, proving the kept main side of the exceptions conflict.

Reader guard and vet: TestCallTargetReaders ran with -timeout 90s; go vet ran on ./internal/flow ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle. Both passed. Changed Go files are gofmt clean. Tests/products add no build-directory or machine-path literals; there is no new compiler-consuming package, so cloud/fast-gate/compiler-dependencies.json needs no change. No cohere code was copied; generic inference retains its pinned checker/shim reference.

Every test's seconds and outcome are recorded in test-times.tsv and test-times.json, including nested leaves, family wrappers and mutant failures. Maximum measured test seconds: '''+f"{max(r['seconds'] for r in times):.3f}"+'''. Compilation and process startup are included separately in the command seconds above. The inherited top-level arrow checks passed again; their source overlays fail before C emission, so invalid C or -Werror is not the catcher.

All 21 mutants and their catchers:

'''+''.join('- '+r['topic']+' ('+r['kind']+'): '+', '.join(c['test']+f" ({c['seconds']:.3f}s)" for c in r['catchers'])+'.\n' for r in mutants)+'''
The fourteen overlays are reconstructed by applying only their mutation delta, comparing each stored mutant to parked source 376febb4, then transplanting the delta onto the resolved current tree. No old full-file snapshot overwrites landed main behavior. Each source mutant exits 1 with a named failed test and no build failure. The seven in-test IR mutants require stdout-only disagreement with Node in both backends; sanitizer/build failure is not accepted. Full failed-test messages and the in-test assertions' evidence are in mutants/results.json and the logs.

Runtime files changed against main: none. Compiler implementation files changed:

'''+''.join('- '+p+'\n' for p in implementation)+'''
Counts additions, allocations/frees/retains/releases/peak/regions:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
'''+''.join(row+'\n' for row in counts['added'].values())+'''
counts-attribution.json records all added rows and an empty changed-existing set. counts-main.md.txt is the main table before regeneration. New evidence lives under review/compiler/loom-chain-followers-2; earlier evidence remains historical under the original branch directories.

Setup: exported GOPROXY=https://proxy.golang.org|direct, ran timeout 600 bash cloud/setup.sh, and sourced the printed /workspace/adamic-tools/env.sh. It succeeded, including the real sanitizer overflow probe. nproc=5, cpu.max=400000 100000 (four-CPU quota). Cumulative timing lines: Go ready 0.144s; submodules ready 0.156s; Node ready 0.230s; Markdown ready 0.375s (validation step 0.034s); clang ready 0.750s; shared cache ready 1.889s; Go build ready 51.627s; test binaries deferred 51.706s; build cache warm 51.708s; done 51.743s. Toolchain: Go 1.27.1, clang 20.1.8, Node v24.19.0. Full timing lines and flags are in setup.log.

Conservative scope assumption: use the parked SHA 376febb4 as authoritative, carrying its compiler support and independent arrow tests. Its #agwccbw, #p6a00tt and #k881crd exclusions remain unchanged. Loom owns the full gate; this worker did not run whole packages or submit the change. Existing historical evidence bytes were preserved.
'''
(out/'REPORT.md').write_text(report)
print('Validated all 21 catchers, all 22 fixture/count leaves, and',len(times),'test timing outcomes; maximum',max(r['seconds'] for r in times))
