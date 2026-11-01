#!/usr/bin/env python3
"""Package the completed submit-head pair and report without changing prior evidence."""
import gzip,hashlib,json,pathlib,shutil,subprocess
repo=pathlib.Path('/workspace/adamic');base=repo/'stage3/census/changeb-road2-meter';out=base/'810ce7d9';raw=pathlib.Path('/workspace/scratch/paired-810ce7d9')
A='54412bc063f205c452ef6839c19ec01662327f7c';B='810ce7d9ced9105f188a6171cc384a61f62c70ac';old='141ad1d244d6a2d9c5403858086d505dfdef5fd5';previous='d299479dc9a530314191221b8b112b5315964b70'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def cmd(*args):return subprocess.check_output(args,cwd=repo,text=True).strip()
def read(p):return [json.loads(x) for x in p.read_text().splitlines()]
assert cmd('git','rev-parse','HEAD')==previous
assert cmd('git','rev-parse','54412bc0^{commit}')==A
assert cmd('git','merge-base',A,B)==A
assert cmd('git','rev-list','--count',f'{A}..{B}')=='3'
assert not cmd('git','diff','--name-only',old,A,'--','stage3/adapt')
assert (out/'base-diff-stat.txt').read_text()==subprocess.check_output(['git','diff','--stat',old,A,'--','internal','stage3/adapt','stage3/census'],cwd=repo,text=True)
assert not (out/'adapt-diff.txt').read_text()
manifest=json.loads((out/'input-sha256.json').read_text());root=raw.parent/'tree/src/compiler'
assert len(manifest)==79
assert {str(p.relative_to(root)) for p in root.rglob('*.ts')}==set(manifest)
assert all(sha(root/p)==h for p,h in manifest.items())
assert json.loads((raw/'box.json').read_text())['hostname']==json.loads((base/'box.json').read_text())['hostname']
records={s:read(raw/f'{s}.jsonl') for s in 'AB'}
assert records['A'][0]==records['B'][0]
assert [(r['file'],r['units']) for r in records['A'][1:]]==[(r['file'],r['units']) for r in records['B'][1:]]
assert sha(raw/'A.jsonl')==sha(raw.parent/'paired/A.jsonl')
logs=out/'logs';logs.mkdir(exist_ok=True)
for p in raw.glob('*.log'):shutil.copyfile(p,logs/p.name)
shutil.copyfile(raw.parent/'fetch-submit.log',logs/'fetch-submit.log')
for side in 'AB':
    assert json.loads((raw/f'{side}-resources.json').read_text())['exit_code']==0
    for suffix in ('.jsonl','-times.jsonl'):
        p=raw/(side+suffix)
        (out/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
        assert gzip.decompress((out/(p.name+'.gz')).read_bytes())==p.read_bytes()
    shutil.copyfile(raw/f'{side}-resources.json',out/f'{side}-resources.json')
shutil.copyfile(raw/'box.json',out/'box.json')
prov={'previous_evidence_commit':previous,'A':A,'B':B,'old_base':old,'submit_commits_above_base':3,'adaptations_changed':False,'adapted_tree_rebuilt':False,'adapted_tree_original_compiler':old,'TypeScript_source':cmd('git','-C',str(raw.parent/'upstream'),'rev-parse','HEAD'),'cohere':cmd('git','-C',str(repo/'cohere'),'rev-parse','HEAD'),'cohere_TypeScript':cmd('git','-C',str(repo/'cohere/TypeScript'),'rev-parse','HEAD'),'same_box':True,'control_ledger_identical_to_prior_controls':True,'checker_headers_identical':True,'unit_eligibility_identical':True,'input_hashes_unchanged':True,'recorder_backport':False,'binary_sha256':{s:sha(raw.parent/f'census-{p}') for s,p in [('A','54412bc0'),('B','810ce7d9')]},'raw_sha256':{p.name:sha(p) for p in sorted(raw.glob('*.jsonl'))},'overlay_sha256':{s:{p.name:sha(p) for p in sorted((raw.parent/f'overlay-{pin}').glob('*')) if p.is_file()} for s,pin in [('A','54412bc0'),('B','810ce7d9')]},'tool_versions':{'go':cmd('go','version'),'node':cmd('/workspace/adamic-tools/bin/node','--version'),'python':cmd('python3','--version')},'compiler_worktree_tracked_changes':{s:cmd('git','-C',str(raw.parent/f'compiler-{pin}'),'diff','--ignore-submodules','--name-only') for s,pin in [('A','54412bc0'),('B','810ce7d9')]}}
assert all(not x for x in prov['compiler_worktree_tracked_changes'].values())
assert prov['binary_sha256']['A']==json.loads((base/'provenance.json').read_text())['binary_sha256']['A']
assert cmd('git','diff','--name-only',A,B,'--','stage3/census/latent','stage3/meter')==''
(out/'provenance.json').write_text(json.dumps(prov,indent=2)+'\n')
h=json.loads((out/'headline.json').read_text());tim=h['timing_totals'];noise=h['noise'];coverage=json.loads((out/'unit-coverage.json').read_text());overloads=json.loads((out/'overload-target-rows.json').read_text());generic=json.loads((out/'generic-claim-reasons.json').read_text());resources={s:json.loads((out/f'{s}-resources.json').read_text()) for s in 'AB'}
assert h['lowering_regressions']==0
assert len(generic)==6 and all(r['A']==0 and r['B']==1 for r in generic)
assert all(r['A']==1 and r['B']==0 for r in overloads)
assert all(x['where']=='src/compiler/core.ts:1779:35' for r in generic for x in r['B_sites'])
rows=['# Submit head `810ce7d9` against its moved base','',
'**Zero control-to-submit lowering regressions. Distinct NotYet and Refused totals are unchanged.** Six already-stopped units move from overload-target stops to six target-specific **“generic claim TOut”** stops. Both first-error censuses completed all 79 files concurrently on the same box. This is **measured on a checker-rejected program**: a no-stop attempt is not a claim of ordinary accepted-program compilation or native output.','',
'| Metric | A: base `54412bc0` | B: submit `810ce7d9` |','| --- | ---: | ---: |','| Distinct NotYet | 652 | 652 |','| Distinct Refused | 2,115 | 2,115 |','| Errors / panics | 0 / 0 | 0 / 0 |','| Attempted units without lowering finding | 2,930 | 2,930 |','| Timed attempted units | 4,017 | 4,017 |',
f"| Unit lowering wall seconds (sum) | {tim['A']['wall_seconds']:.6f} | {tim['B']['wall_seconds']:.6f} |",
f"| Unit lowering CPU seconds (sum) | {tim['A']['cpu_seconds']:.6f} | {tim['B']['cpu_seconds']:.6f} |",
f"| Whole-run wall seconds | {resources['A']['wall_seconds']:.3f} | {resources['B']['wall_seconds']:.3f} |",
f"| Peak RSS (KiB) | {resources['A']['peak_rss_kib']} | {resources['B']['peak_rss_kib']} |",'',
f"The new base is **`{A}`**; the submit head is **`{B}`**, exactly three commits above it. The requested [base diff](base-diff-stat.txt) from `141ad1d244d6a2d9c5403858086d505dfdef5fd5` shows **six conformance fixture files, 37 insertions**; there are **no changes under `stage3/adapt` or `stage3/census`**. Therefore this pair **reuses `/workspace/scratch/tree`** rather than rebuilding it. The adaptation driver, source pin, API dependencies and dependency pins also match. All 79 source hashes were checked before and after the pair ([manifest](input-sha256.json)). Both census binaries were built at their own requested SHAs using the meter overlay commands; the new control binary is byte-identical to the earlier control.",'',
'No unit that lowers on A is absent or stopped on B: the [regression list is empty](lowering-regressions.json). All 2,930 attempted units without lowering-phase findings are retained. The earlier two regressions, `computeLineOfPosition` (`scanner.ts:520:1`) and `indexOfNode` (`utilities.ts:1351:1`), remain no-stop attempts on both sides ([explicit check](prior-regressions-check.json)). The [six changed first stops](unit-outcome-changes.md) are all in already-stopped units; no stopped unit becomes a no-stop attempt.','',
'The six requested overload-target rows each go **1 → 0**. For each target below, the exact replacement reason is **`a checked predicate target <Target> that the generic claim TOut may be instantiated at`**, each **0 → 1**, always at **`src/compiler/core.ts:1779:35`** (`test(value)` in `tryCast`). These are **six distinct sites under `(kind, where, reason, text)`, at one source location**, because the reasons name different targets. There are no “generic claim” reasons on A.','',
'| Target | Overload A | Overload B | Generic claim A | Generic claim B | Old overload site |','| --- | ---: | ---: | ---: | ---: | --- |']
for r in overloads:
    rows.append(f"| {r['target']} | {r['A']} | {r['B']} | 0 | 1 | {r['A_sites'][0]['where']} |")
rows += ['',
'All **12 changed reason rows** mention **predicates**: these six removals and six additions. The [changed-reason table](changed-reasons.md) gives every row, counts and up to five example locations per side. [Overload rows and sites](overload-target-rows.json), [all generic-claim reasons and all their sites/attempting units](generic-claim-reasons.json), and the [full generic-claim site table](generic-claim-sites.md) retain the complete details without truncation. The [view/predicate table](view-predicate-reasons.md) includes unchanged rows; no read/write-through-view row changes and no observed reason literally says “branded supertype”.','',
f"Observed unit lowering wall time is **{tim['B']['wall_seconds']-tim['A']['wall_seconds']:+.6f} seconds ({(tim['B']['wall_seconds']/tim['A']['wall_seconds']-1)*100:+.3f}%, {tim['B']['wall_seconds']/tim['A']['wall_seconds']:.5f}×)**; CPU is **{tim['B']['cpu_seconds']-tim['A']['cpu_seconds']:+.6f} seconds ({(tim['B']['cpu_seconds']/tim['A']['cpu_seconds']-1)*100:+.3f}%)**. Whole-run wall time is {resources['B']['wall_seconds']-resources['A']['wall_seconds']:+.3f} seconds. The [20 largest B−A unit differences](timing-top20.md) give absolute seconds and B/A ratios; [all units](timing-all.md) and [JSON](unit-times.json) preserve every matched unit. Unit timing includes registration, lowering and recovery, excluding loading/checking and refusal scanning.",'',
f"Noise floor: {noise['n']} identical-outcome import/interface/type-alias units whose lowering does not execute a runtime branded-predicate call. Median signed B−A is {noise['median_signed_delta_seconds']*1000:.3f} ms; median absolute difference **{noise['median_absolute_delta_seconds']*1000:.3f} ms**, p90 absolute **{noise['p90_absolute_delta_seconds']*1000:.3f} ms**. Median B/A is **{noise['median_ratio']:.4f}**, p10–p90 **{noise['p10_ratio']:.4f}–{noise['p90_ratio']:.4f}** ([baseline](noise-baseline.json)). The unaffected-unit median itself shifts toward B by about {(1-noise['median_ratio'])*100:.2f}%; this single pair does not isolate a causal speedup from GC, cache and scheduling variation.",'',
f"Both runs started at {resources['A']['started_utc']} (A) and {resources['B']['started_utc']} (B), within 0.3 ms. B ended at {resources['B']['ended_utc']}; A at {resources['A']['ended_utc']}. Both exited 0 without hitting the 5400-second bound. Same box `6923d492b4a0`: 5 CPUs (`nproc`), cgroup quota 400000/100000 (4 CPU equivalents), 18,882,662,400 bytes RAM, no swap ([box](box.json), [A resources](A-resources.json), [B resources](B-resources.json)).",'',
'Pins and reproduction:','',
f'- Control: `{A}`; submit: `{B}`.',
'- Adapted source: TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`, originally adapted once using `141ad1d2`, reused because adaptations are unchanged.',
f"- A binary SHA-256: `{prov['binary_sha256']['A']}`.",
f"- B binary SHA-256: `{prov['binary_sha256']['B']}`.",
'- Both compiler worktrees have identical census/recorder code, no recorder backport and no tracked edits. Nothing under `internal/` was changed. The initialized `cohere` pin is shared through the same dependency symlink arrangement. [Exact commands](commands.md), [provenance](provenance.json), [runner](run-censuses.py), [analysis](analyze.py), [packager](package.py) and [logs](logs/).',
'- [A ledger](A.jsonl.gz), [B ledger](B.jsonl.gz), [A unit times](A-times.jsonl.gz) and [B unit times](B-times.jsonl.gz) round-trip exactly to the raw streams. Existing meter [lowering_summary validation](validation.json) and the requested [census delta](census-delta.json) agree: **2,767 → 2,767** distinct NotYet/Refused sites over **79 files**. Three SkippedDependency sites on each side are excluded. Checker headers and all 4,543 unit eligibility records (4,017 attempted, 498 deferred, 28 skipped) are identical; the control raw ledger is byte-identical to the previous controls. `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1` throughout.','']
(out/'REPORT.md').write_text('\n'.join(rows))
print('Packaged '+str(out))
