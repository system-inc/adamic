import collections,json,re,statistics
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-r-v4-3';p=e/'fixed'
def read(name):return json.loads((p/name).read_text())
def jsonlines(name):return [json.loads(x) for x in (p/name).read_text().splitlines()]
focused=[r for r in jsonlines('focused-results.jsonl') if 'test'in r]
readers=jsonlines('readers/reader-results.jsonl');reader_tests=[r for r in readers if 'test'in r]
assert len(focused)==315 and all(r['exit']==0 and r['grain'] for r in focused)
assert all(r['exit']==0 and r.get('grain',True) for r in readers)
mutants=jsonlines('mutants/source-mutants.jsonl');assert len(mutants)==19 and all(r['killed'] for r in mutants)
for m in ['object-lookup-omission.json','primitive-proof-omission.json']:assert read(m)['killed']
summary=read('package-summary.json');js=read('javascript-times.json');js={s:[r['seconds'] for r in js if r['side']==s and r['mode']=='corpus'] for s in ['base','v4']}
for side in summary:assert summary[side]['status']=='pass' and not summary[side]['problems']
audit=read('a-check/summary.json');commands=jsonlines('commands.jsonl')
lines=[
'Built: views V4 carrying V3, with checker-proven primitive equality operands bypassing JavaScript adapter-root lookups.',
'Commits: one commit on exact dbdf977f5b8b0979d414ebe6f20e7aadd312e309, carrying 82a5dffd7654f3eb2cb4652a12b99401210f442a; final SHA is in the push report.',
f'Commands and outputs: 315 focused entries plus three new unit checks, {len(reader_tests)} focused reader entries, guards, vet, a-check and both complete JSON packages pass; timings below.',
'Mutants: all 19 carried source mutants, object-lookup omission, primitive-proof omission, removed header and wrong reason are caught.',
'Not covered: a full repository gate, ahra/Loom submission, or reproduction of the remote loaded Lunatic environment.',
'',
'The branch starts at the requested exact base, despite origin/main advancing. The no-commit cherry-pick of 82a5dffd had no conflicts. This commit includes its entire change and the primitive optimization. No runtime C line or counts row differs from 82a5dffd. No new oracle fixture or runtime feature set was added.',
'',
'Equality and inequality, plus Object.is, carry a two-operand checker certificate in IR before union boxing. Number, string, boolean, bigint, undefined and null, and unions solely of these, qualify. Scalar IR representations also bypass normalization for generated comparisons. Unknown, any, object, function and unions admitting them retain lookup. Unresolved or otherwise unproved types conservatively retain it. This does not add bigint lowering support.',
'',
'The compiler deliberately refuses a callable erased to source object type. Therefore the object-typed regression exercises the real JavaScript emitter on object-typed IR operands and real adapter root metadata. Source-level callable identity and cached/no-stacking tests separately pass in native release, sanitizers and JavaScript. The object-lookup compiler mutant runs successfully on Node but produces false false instead of true true, so the identity assertion fails semantically; no compiler error kills it.',
'',
'Full-corpus JavaScript measurements use the same 3,919-case pinned protocol, fresh processes, GOMAXPROCS=4 and alternating revision order. Each output is 300,372,755 bytes with source Node SHA-256 8571e190419a2e27c5796b3f1ff6485cfc824ddf161fcf48e9b3a55739ac614f, exit 0 and empty stderr. Source/corpus provenance and current compiler hashes are retained. Three runs per revision:',
'',
'| Revision | Samples (s) | Median (s) |',
'| --- | --- | --- |']
for side,label in [('base','dbdf977f5b'),('v4','V4-3')]:lines.append(f'| {label} | '+', '.join(f'{v:.3f}' for v in js[side])+f' | {statistics.median(js[side]):.3f} |')
lines += ['',f'V4-3/base median ratio: {statistics.median(js["v4"])/statistics.median(js["base"]):.4f}. The prior V4 diagnostic was 25.465 s base versus 48.148 s V4. That earlier single comparison and the runtime-only causal ablation remain in previous-diagnostic-report.md; neither ablation is a compiler change here.',
'',
'Both complete stage1/cohere/json package runs use -count=1 -json -parallel=4 -timeout=30m, outer limit 1860 s, GOMAXPROCS=4, sequentially with no competing validation work. The pinned corpus, Go oracle owners and lowered/native port products are settled first on both trees; the latter use -parallel=1. Commands and all raw JSON are in fixed/commands.jsonl and fixed/{base,v4}-*.jsonl.',
'',
'| Revision | Package elapsed (s) | Pass / skip events | Failed or >=60 s tests |',
'| --- | --- | --- | --- |']
for side,label in [('base','dbdf977f5b'),('v4','V4-3')]:
 s=summary[side];lines.append(f'| {label} | {s["package_seconds"]:.3f} | {s["completed"].get("pass",0)} / {s["completed"].get("skip",0)} | {len(s["problems"])} |')
lines+=['','All 2,048 ordinary agreement shards pass on each revision. No shard crosses the 60 s budget on V4-3. fixed/shard-times.csv lists every test on both sides; fixed/{base,v4}-tests.json retain outputs and backend timings. There is no observed output mismatch to reduce. These local observations do not establish the exact remote timeout mechanism.','']
for side in ['base','v4']:
 ts=read(side+'-tests.json');shards=[t for t in ts.values() if t['shard']];maximum=max(shards,key=lambda t:t['seconds']);lines.append(f'{side} slowest ordinary shard: {maximum["test"]}, {maximum["seconds"]:.3f} s.')
lines+=['','Focused validation runs exact top-level tests from fixed/selected-focused.json, using compiled test binaries, -test.count=1 -test.timeout=90s -test.parallel=1, outer 95 s per entry. Products run serially, other entries with two workers. Every TestV4 name is selected (fixed/v4-coverage-audit.json). All lane 5 boxed parameter/result and producer negatives run; all adapter identity/cache/no-stacking tests run. Native fixtures use release and ASan/UBSan/LeakSanitizer through internal/native. Source Node determines their expected behavior.']
lines.append(f'{len(focused)} focused entries: {sum(not r["skip"] for r in focused)} pass, {sum(r["skip"] for r in focused)} explicit skips; maximum {max(r["seconds"] for r in focused):.3f} s. {len([r for r in readers if r.get("build")])} reader packages compile, {len(reader_tests)} selected reader tests finish ({sum(r["skip"] for r in reader_tests)} skips); maximum {max(r["seconds"] for r in reader_tests):.3f} s. Optional/WASI skips retain their exact reasons in logs.')
lines+=['','New test seconds are in fixed/new-test-times.json. The primitive-union oracle fixture checks source Node, native release, sanitizers and JavaScript, including nullish unions, NaN and signed zero. The proof classifier separately checks bigint and mixed object/callable unions. The two JavaScript tests prove representation fallback and object identity.','',
'Guards pass: TestCallTargetReaders (-timeout 90s), TestEveryMapAProgramHoldsIsSorted, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestHasOneDescriptiveName, TestEveryTestIsParallelOrSaysWhy and TestEveryCCompilerRunsThroughNative. No guard or baseline was loosened. New tests have unique descriptive names and call t.Parallel first. The IR addition is a fixed boolean array, with no new map or call target. Existing native-agreement registrations and exact runtime witnesses remain intact. No compiler-consuming package, direct C compilation or build-directory-bearing product was added. go vet ran on internal/{flow,fresh,ir,javascript,load,lower,native,native/runtimeproducts,oracle} and stage1/cohere/json, with a final JavaScript vet after its test edit. Exact outputs are in fixed/*.log.','',
'A-check invokes the candidate compiler on all 47 added/modified .a paths, including review witnesses, through unchanged pinned Gate.aCheck (fbac28c62493f27a788edc02a18bc8edb68de5da). Status counts: 40 checked, 1 NotYet, 5 refused, 1 TS7009 checker error. All six expected-error headers already match; no header or body bytes changed. Both removed-header and wrong-reason mutants fail the exact predicate. fixed/a-check/status-table.md is the per-file table; statuses.json and gate-baseline.json contain real current diagnostics and hashes.','',
'Carried source mutants and exact catchers:','', '| Mutant | Catcher | Seconds |','| --- | --- | --- |']
for row in mutants:lines.append(f'| {row["mutant"]} | {row["test"]} | {row["seconds"]:.3f} |')
lines+=['| object-typed lookup omission | TestJavaScriptObjectOperandAdapterIdentity | '+f'{read("object-lookup-omission.json")["seconds"]:.3f}'+' |','| primitive proof omitted | TestV4PrimitiveUnionIdentityEmission | '+f'{read("primitive-proof-omission.json")["seconds"]:.3f}'+' |','',
'The focused driver hit disk exhaustion while opening a new log after 258 completed entries. Generated binaries and bulk stdout artifacts were removed; their hashes and measurements remain. Only unfinished entries resumed, with complete pass logs recovered rather than rerun. No assertion, timeout or 60 s leaf was retried. fixed/disk-cleanup.json, focused-driver.log and focused-resume.log preserve the interruption and recovery. Generated test binaries and multi-GB bulk outputs are excluded from the commit.','',
'Toolchain setup ran with GOPROXY=https://proxy.golang.org|direct, then sourced /workspace/adamic-tools/env.sh. Linux, clang 20.1.8, Go 1.27.1, Node 24.19.0; nproc=5, cgroup CPU quota=4 and memory limit=16 GiB. Setup timing lines:','', '```']
lines += [line for line in (e/'fix-setup.log').read_text().splitlines() if line.startswith('setup:')]
lines+=['```','']
(e/'REPORT.md').write_text('\n'.join(lines))
print('Report written')
