import pathlib,json,shutil,gzip,hashlib
p=pathlib.Path('review/test-audit/internal-oracle-stack')
rows=json.loads((p/'rows.json').read_text());matrix=json.loads((p/'matrix.json').read_text());probes=json.loads((p/'probe-results.json').read_text());plan=json.loads((p/'mutant-plan.json').read_text())
scope=[r['test'] for r in rows]+['TestNativeAgreesWithNode']
report=[]
for i,r in enumerate(rows):
 name=r['test'];es=matrix[-1]['events'];fail=next(e['Output'].strip() for e in es if e.get('Test','').startswith(name+'/') and 'stack_test.go:' in e.get('Output',''))
 report.append(dict(test=name,package='internal/oracle',file=r['file']+':'+str(r['line']),seconds=r['seconds'],oracle=('Original TypeScript source executed by Node; full stdout, stderr and exit status compared to sanitized and release native products.' if i==0 else 'Node validates the exact start output, panic text and exit 70 at 1024 KiB. At 512/256 KiB the expected same contract is self-written; Node is deliberately not run. M04 proves stderr is checked even when exit 70 and stdout remain correct.'),oracle_kind=('external-run' if i==0 else ['external-run','self']),kills=[m['id'] for m in matrix if name in m['kills']],unique_kills=([] if i==0 else ['M02']),last_proven_fail='M04: '+fail,verdict=('subsumed' if i==0 else 'sacred'),subsumed_by=([rows[1]['test']] if i==0 else []),mutants_in_matrix=4,probe_kills=[q['id'] for q in probes if q['test']==name and q['exit']!=0],subsumer_seconds=(rows[1]['seconds'] if i==0 else None),vacuous=False,bounded=True,matrix_rows=scope,evidence=matrix[-1]['command']+'; ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u069/cache/M04; '+fail))
(p/'report.json').write_text(json.dumps(report,indent=2)+'\n')
summary='Unit u069: two rows present, neither moved or vanished.\nBase: 6c60da091afddc9c2fe88b3a1067845b6dc79cb3; nproc 5; toolchain warm.\nWhole package timed out at 90.062 s; clean bounded runs passed.\nFour production mutants: all killed; both rows rejected both empty-answer probes.\nBounded verdicts: long arguments subsumed; small stacks sacred.\n'
table='| ID | origin/main location | Change | Failed rows |\n|---|---|---|---|\n'
for q,m in zip(plan,matrix):table+='| '+q['id']+' | '+q['file']+':'+str(q['line'])+' | `'+q['old'].strip()+'` to `'+q['new'].strip()+'` | '+', '.join(m['kills'])+' |\n'
notes='''Survivors: none in the bounded matrix. No equivalent candidates claimed.

Scope and limitations:
- The brief cites 8de93800f4; fetched origin/main was 6c60da091afddc9c2fe88b3a1067845b6dc79cb3. Both requested names exist there in stack_test.go at lines 17 and 89. No grouping applies: bodies assert distinct contracts.
- Whole-package baseline exceeded its 90-second binary budget without an observed assertion failure. Narrowed via stack fixture references and native emission/runtime call paths to both unit rows plus only TestNativeAgreesWithNode/internal/oracle/testdata/stack_overflow.a. Other members of that corpus row and all other package rows remain unknown. Unique kills mean unique only in this bounded set. No repo-wide testing performed.
- matrix_rows contains the top-level corpus row because counts are over rows. Only its named fixture ran; its full corpus family was not measured or judged. The long row is subsumed by the small row on three caught production mutants, a small-set hint, not deletion advice.
- CODE UNDER TEST: native.C and native functionBody stack-check emission; native runtime find_stack_limit, ADAMIC_CHECK_STACK, adamic_stack_overflow and adamic_panic. ORACLE: original source run by Node, plus the self-written small-stack contract below 1024 KiB. Coverage records 314 reached Go production functions, listed in reached-functions.txt; runtime-reach.txt records the C chain. Lowering and unused JavaScript generation prepare the fixture, so were not mutated or empty-probed.
- Ordinary three-run timing uses the suite's warm gate cache. Those binary seconds measure the row as normally invoked, not uncached native execution. Mutation and probe runs explicitly disable the gate cache and give each mutant its own build cache.
- The arguments-alone subcase clears its environment, so an environment mutation switch would silently stop selecting C mutations. The scratch switch instead reads /tmp/u069/mutant. Its clean uncached baseline passed. Switch scaffolding is absent from every standalone diff and supports no verdict by itself.
- Four mutants were fixed before outcome inspection. The brief's approximate three-per-row goal would suggest six, but its explicit maximum-four rule for native rebuilds limits this unit to four. They span reservation, margin, emission and panic text.
- Native.C empty probe fails during native build, which demonstrates rejection of no generated answer, not a production semantic kill. Runtime constructor empty probe leaves a zero limit and fails native execution. Both are separate probe_kills. No positive subcase passed either probe.
- Native runtime diffs compile with the builder's release and sanitizer clang flags. Go emitter/probe diffs pass go vet ./internal/native/. Exact commands and durations are in mutant-plan.json and probe-plan.json. Linux native behavior was exercised; Darwin/WASI alternate stack-limit branches were not separately covered by this slice. WASI SDK opt-in was enabled for the baseline. No skip event was observed before that baseline timed out; tests not reached are unknown.
- No outside authority is claimed for the below-1MiB panic expectation. The 1MiB Node execution checks the same expected value. Assertions compare complete stdout/stderr/exit, not just panic exit status: M04 still exits 70 with start output, and both rows fail on its changed stderr.
- The final restored-source bounded run passed; runtime/emitter production sources are restored. Replaying standalone diffs is necessary to settle package and repo uniqueness.

Costs:
Toolchain setup skipped (warm env); nproc=5. npm ci log records dependency setup. Whole package binary: 90.062 s timeout. Clean two-row uncached binary: 1.313 s (3.399 s command wall). Clean three-row uncached binary: 1.293 s. Scratch Go compilation: 8.118 s. Solo timings: long 0.133/0.107/0.103 s, median 0.107; small 0.100/0.110/0.103 s, median 0.103. Runtime validation and Go vet totals are recorded per diff.
Native rebuilding is included in each mutation command, with distinct caches. Separate native compiler wall time was not isolated, so the following report is an inclusive rebuild-and-run measurement, not a compiler-only claim.
'''
for m in matrix:
 elapsed=next(e['Elapsed'] for e in reversed(m['events']) if e['Action']=='fail' and not e.get('Test'))
 notes+=f"{m['id']}: command wall {m['wall_seconds']:.3f} s; binary including native build {elapsed:.3f} s.\n"
notes+='Probes command wall: '+', '.join(f"{q['id']}/{q['test']} {q['wall_seconds']:.3f} s" for q in probes)+'.\n'
(p/'REPORT.md').write_text(summary+'\n```json\n'+json.dumps(report,indent=2)+'\n```\n\n'+table+'\n'+notes)
for source in ['/tmp/u069-audit.py','/tmp/u069-measure.py','/tmp/u069-report.py']:shutil.copy(source,p/pathlib.Path(source).name)
# Keep exact output losslessly, reduce committed evidence size.
for log in (p/'logs').glob('*.log'):
 with log.open('rb') as inp,gzip.open(str(log)+'.gz','wb') as out:shutil.copyfileobj(inp,out)
 log.unlink()
(p/'SHA256SUMS').write_text(''.join(hashlib.sha256(f.read_bytes()).hexdigest()+'  '+str(f.relative_to(p))+'\n' for f in sorted(p.rglob('*')) if f.is_file() and f.name!='SHA256SUMS'))
