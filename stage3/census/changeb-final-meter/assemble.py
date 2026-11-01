import pathlib,json,gzip,shutil,hashlib,collections,subprocess
repo=pathlib.Path('/workspace/adamic'); out=repo/'stage3/census/changeb-final-meter'; raw=pathlib.Path('/workspace/scratch/changeb-final-meter')
read=lambda n:json.loads((out/n).read_text())
a,b=read('A-validation.json'),read('B-validation.json')
# Validate unchanged shared source bytes after both runs.
manifest=read('source-sha256.json')
assert all(hashlib.sha256((pathlib.Path('/workspace/scratch/tree/src/compiler')/p).read_bytes()).hexdigest()==h for p,h in manifest.items())
records={s:[json.loads(l) for l in (raw/(s+'.jsonl')).read_text().splitlines()] for s in ['A','B']}
assert records['A'][0]==records['B'][0]
assert read('census-delta.json')['before_sites']==a['totals']['NotYet']+a['totals']['Refused']
assert read('census-delta.json')['after_sites']==b['totals']['NotYet']+b['totals']['Refused']
assert not read('lowering-regressions.json')
assert not read('unknown-instance-units.json')
for s in ['A','B']:
 for suffix in ['.jsonl','-times.jsonl','.log']:
  with gzip.GzipFile(filename=str(out/(s+suffix+'.gz')),mode='wb',mtime=0) as f:f.write((raw/(s+suffix)).read_bytes())
 for suffix in ['-resource.json','-exit.txt']:
  shutil.copyfile(raw/(s+suffix),out/(s+suffix))
for n in ['delta.log','analysis.log','claims.log','cast-callers.log','start-utc.txt','end-utc.txt','pids.txt']:
 shutil.copyfile(raw/n,out/n)
# Every changed unit first-stop record, including overload declarations.
lines=['All 13 changed unit records were stopped in both runs. No successful A unit is absent or stopped on B. An attempted unit with no lowering-phase finding counts as lowering in this isolated census.','','| Unit | A first stop | B first stop |','|---|---|---|']
for r in read('unit-ledger-delta.json'):
 name=r['A']['unit'].get('name') or {'src/compiler/emitter.ts:1147:1':'notImplementedResolver','src/compiler/factory/parenthesizerRules.ts:675:1':'nullParenthesizerRules'}.get(r['where'],'KindVariableStatement')
 cells=[]
 for s in ['A','B']:
  cells.append('; '.join(f['where']+' '+f['kind']+': '+f['reason'].replace('|','&#124;') for f in r[s]['findings']))
 lines.append(f"| {name} ({r['where']}) | {cells[0]} | {cells[1]} |")
(out/'first-stop-delta.md').write_text('\n'.join(lines)+'\n')
# Explicitly name every relevant row, including unchanged view/predicate rows.
lines=['| Kind / reason | A | B | A example locations (up to 5) | B example locations (up to 5) |','|---|---:|---:|---|---|']
for reason in sorted(a['per_reason'].keys()|b['per_reason'].keys()):
 if not any(w in reason.lower() for w in ['view','branded supertype','predicate']):continue
 examples={}
 for s in ['A','B']:
  examples[s]=sorted({f['where'].replace('/workspace/scratch/tree/','') for r in records[s][1:] for f in r['findings'] if f['kind']+': '+f['reason']==reason})[:5]
 lines.append(f"| {reason.replace('|','&#124;')} | {a['per_reason'].get(reason,0)} | {b['per_reason'].get(reason,0)} | {'; '.join(examples['A'])} | {'; '.join(examples['B'])} |")
(out/'view-predicate-reasons.md').write_text('\n'.join(lines)+'\n')
source=subprocess.check_output(['git','show','84fccdc821a3155ae66c8137c2834eaff5c6cb69:internal/lower/predicates_branded.go'],cwd=repo,text=True).splitlines()
(out/'unknownInstance-source.txt').write_text('B 84fccdc821a3155ae66c8137c2834eaff5c6cb69 internal/lower/predicates_branded.go\n'+'\n'.join(f'{i+1}: {source[i]}'.rstrip() for i in range(164,230))+'\n')
(out/'REPORT.md').write_text('''A and B completed the same 79-file adapted tsc tree in concurrent first-error censuses. **Zero units that lower on A stop on B.** All 2,930 lowering units are retained, with no missing unit keys. This is the meter's isolated, checker-rejected-program measurement; its output guards remained enabled.

| Distinct sites | A | B |
|---|---:|---:|
| NotYet | 653 | 651 |
| Refused | 2,180 | 2,180 |
| error | 0 | 0 |
| panic | 0 | 0 |
| SkippedDependency | 3 | 3 |

Both ledgers pass `stage3/meter/report.py`'s existing `lowering_summary` coverage and distinct-count validator: 79/79 files, unique `(kind, where, reason, text)` sites. [A validation](A-validation.json), [B validation](B-validation.json). Headers/checker diagnostics are identical; all 79 source hashes were verified unchanged after both runs.

B changes **13 unit stop records**, all already stopped on A: [every unit, name, A/B reason and location](first-stop-delta.md), [machine-readable records](unit-ledger-delta.json). The successful-A sites absent or stopped in B list is **empty**, as is the list of A-lowering/B-stopped units: [lowering-regressions.json](lowering-regressions.json). Separately, eight A lowering-phase *finding* sites disappear from B: [complete list](A-lowering-findings-absent-in-B.json).

There are **13 changed reason rows**, five growing/new: [all counts and up to five example locations on each side](changed-reasons.md), [requested delta](census-delta.md). The branded-supertype array-target row rises 0→2; new predicate-target rows name DynamicNamedDeclaration, LeftHandSideExpression, NamedDeclaration and TypeNode (each 0→1). Predicate-overload rows ClassElement, JSDocPropertyLikeTag, LateBoundDeclaration, LeftHandSideExpression, LiteralExpression and TypeElement each fall 1→0. `a value of type HasJSDoc` and `reading Debug` each fall 1→0. **No view row changes**; all view/read/write/predicate/branded-supertype wording is explicitly named with A/B counts in [view-predicate-reasons.md](view-predicate-reasons.md).

The second question, **unknownInstance**: **zero units** stop for a body “lowered without knowing” a type; there are therefore no associated claim paths or branded declarations to enumerate. The exact B source constructs the wording as `"a checked predicate claim of "+l.checker.TypeToString(l.concrete(claim.Type()))+", over "+l.checker.TypeToString(parameter)+", which this body is lowered without knowing"` at `internal/lower/predicates_branded.go:182`; [pinned excerpt including unknownInstance](unknownInstance-source.txt), [empty unit list](unknown-instance-units.json). The program does declare branded supertypes, including `LeftHandSideExpression` with required `_leftHandSideExpressionBrand: undefined` at `types.ts:2448–2449`; [supplementary stock-TypeScript source inventory](branded-supertype-source-inventory.json).

`tryCast` / `cast`, separately: declarations at `core.ts:1778:1` / `1783:1` are **deferred generic instantiations, instantiated=true on both sides**, not standalone attempted units. The predicate calls in their bodies are `test(value)` at `1779` / `1784`. There is **no first stop at core.ts:1779** in either ledger. The observed `cast` stop belongs to `nullParenthesizerRules` (`factory/parenthesizerRules.ts:675:1`) at `core.ts:1784:32`: A says `checked predicate overload target LeftHandSideExpression`; B says `a checked predicate target LeftHandSideExpression over a generic argument TIn & ({} | null), read on the true side as Expression & LeftHandSideExpression`. Its predicate claim is `value is TOut`, instantiated to LeftHandSideExpression, which is a declared branded supertype on this path. It remains **NotYet**, outside unknownInstance.

[Every direct tryCast/cast call and containing-unit outcome](cast-callers.md) records 78 source calls in 22 units: 14 units NotYet, six skipped checker bodies, one Refused, and one lowering. `identifierToKeywordKind` (`utilitiesPublic.ts:865:1`, `tryCast(token, isKeyword)` at `867:20`) **lowers on both A and B**. `transformES2020` remains Refused. Thus a blanket “all callers stay NotYet” is not what this census observes. Static calls in skipped bodies, or beyond an earlier unit stop, do not establish a reached predicate call.

Per-unit lowering time across **4,018 matched attempted units** totals **794.908220 s A → 803.547514 s B**, **+8.639294 s (+1.0868%; ratio 1.010868)**; process CPU within unit timers is 842.766367→849.785866 s. [Top 20 increases, seconds and ratios](unit-times.md), [all units](unit-times.csv), [summary](timing-summary.json). Timers include unit registration/lowering/recovery and exclude checking, refusal scans and other untimed work. On 1,292 import/export/interface/type-alias units without lowered predicate bodies, median absolute paired delta is **6.280 ms**, p95 **63.790 ms**; median B/A **0.9912**, p05–p95 **0.6849–1.4107**. This single concurrent pair has substantial scheduling/cache/GC noise; the measured increase is not a repeat-run causal estimate.

| Whole run | A | B |
|---|---:|---:|
| Wall seconds | 1437.456142 | 1447.622914 |
| Peak RSS KiB | 6068172 | 5828264 |
| Exit code | 0 | 0 |

Box: `nproc=5`, CPU quota 400000/100000 (4 CPUs), host memory 18,882,662,400 bytes, cgroup memory limit 17,179,869,184 bytes (16 GiB), no swap. Resource capture uses Python `wait4` around each timeout child. An initial launch found `/usr/bin/time` missing and exited before either census ran; the corrected runner and failed-launch logs are preserved.

Exact compilers: **A `470e16f8315f7a83c5c7b7f6f2a77d8328db0a25`**, **B `84fccdc821a3155ae66c8137c2834eaff5c6cb69`**. Upstream TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Adapted once with A's stage3 files. Both contain identical timing tree `b12676c48d04c7930df05eb0984be1fbba0f7a13`; no recorder patch was needed. Both census binaries used their own SHA's meter build; compiler `internal/` files are unchanged.

Binary SHA-256s:

- A: `845025b39eb6c23a91fe681151202417e4c2f7e912630e79c3a2ca97d63f414b`
- B: `0d896e3261ff0bcb749f4e2754a06372935db41ed066a6e912e6cc5107b4e93b`

[Exact commands and launch correction](commands.md), [paired runner](run-censuses.py), [source hashes](source-sha256.json), [box/provenance](provenance.json). Gzipped A/B ledgers, timing streams, setup/apply/build/run logs and resource records are adjacent. Only this evidence directory is committed and pushed to `typescript/changeb-final-meter`.
''')
print('All final invariants pass; evidence assembled.')
