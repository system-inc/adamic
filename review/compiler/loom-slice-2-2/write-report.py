from pathlib import Path
import json, collections, subprocess
out=Path('review/compiler/loom-slice-2-2')
meta=json.loads((out/'recut.json').read_text());counts=json.loads((out/'counts-attribution.json').read_text());times=json.loads((out/'test-seconds.json').read_text())
checks=[json.loads(line) for line in (out/'checks.jsonl').read_text().splitlines()]
mutants=[]
for name in ['mutants-overlay-results.json','callback-mutants/results.json','mutants-self-compare-results.json','weak-isolated/result.json','registry-mutant/result.json']:
 p=out/name
 if p.exists():
  value=json.loads(p.read_text());mutants.extend(value if isinstance(value,list) else [value])
passed=[x for x in mutants if x.get('caught')]
lines=[
 'Built: slice 2 re-cut on main, with its two Weak counters adapted to main\'s keyed registry. Validation remains incomplete.',
 'Commit SHAs: base '+meta['base']+'; source '+meta['source']+'. Delivery commit and push are withheld while mandatory checks remain blocked.',
 'Commands and outputs: 62 oracle leaves and 74 support leaves pass; vet and registry ownership pass; one counts regeneration passes with 46 attributed additions.',
 'Mutants: 16 of 17 original semantic mutants verified, plus the new counter-registration mutant. The nullable-token mutant remains blocked; the inline-definition mutation is structural only.',
 'Not covered: four overdue oracle leaves, the unfinished reader guard, and the over-budget stage3 wrapper; no green delivery commit or push, full gate, or WASI run.',
 '',
 'The squash merge had no textual conflicts. Main\'s counts.md was restored before regeneration. The clean merge exposed one API compatibility issue: slice 2 appended a function returning two rows, while main now expects one keyed countCase per row. Only slice 2\'s weak_differential_test.go was adapted. Both liveness witnesses remain registered. Main also changed counts to a test family, so regeneration used the unanchored TestCountsAreRecorded selector.',
 '',
 'The conservative interpretation of the user\'s no-rerun rule applies to every active leaf exceeding its share. No overdue leaf was rerun. Compilation-only attempts and interrupted flow/fresh leaves that had not exceeded their share were continued; already completed leaves were excluded. The first green state within 30 minutes was not achieved.',
 '',
 '## Counts attribution',
 '',
 f"{counts['base_rows']} to {counts['after_rows']} table rows, including predicate rows: {counts['added']} additions, {counts['changed_existing']} existing-row changes, {counts['removed']} removals. All additions are owned by slice 2. No existing predicate or runtime row moved.",
 '',
 '| Owner | Added rows |', '|---|---:|'
]
for owner,number in sorted(collections.Counter(x['owner'] for x in counts['rows']).items()):lines.append(f'| {owner} | {number} |')
lines+=['','Every changed row, with its measured values, is in counts-attribution.json.','', '## Runtime changes against main','']
for path in subprocess.check_output(['git','diff','--cached','--name-only','--','internal/native/runtime'],text=True).splitlines():lines.append('- '+path)
lines+=['','These are the inherited Error message storage and lookup repairs. Native emission changes also include fields.go, emit_expressions.go, self_compare.go and units.go. There is no additional runtime repair in this re-cut. No code was copied from cohere.','', '## Observed checks','', '| Phase | Exit | Wall seconds |', '|---|---:|---:|']
for row in checks:lines.append(f"| {row['name']} | {row['exit']} | {row['seconds']} |")
lines+=['','Exact commands and outputs are in checks.jsonl and the named logs. All go test invocations used -count=1 and a hard test timeout, with output written directly to logs. No whole package suite or full gate was run. Source Node, JavaScript output, native release, ASan, UBSan and Linux leak checks were exercised by the unaffected oracle fixtures according to each fixture\'s contract. Freed Weak targets retain their explicit native/Node lifetime difference.','', '## Budget failures and setup','', 'The four first active oracle leaves ran into the package\'s 90-second timeout while cold runtime products were being built. They are TestCallbackWideningNumberEvery, TestCallbackWideningNullableObjectMap, TestWeakDifferentialHeldUnion and TestWeakDifferentialBooleanFilter. The reader guard was still active beyond its share when the support supervisor ended. TestFixturesSelfCompare passed semantically in 78.53 seconds, exceeding the 60-second limit. None of these six leaves was rerun.','', 'Setup exited 124 at its 600-second hard limit during go list -deps -export -json ./.... The generated environment was usable and sourced in subsequent shells. nproc=5; cpu.max=400000 100000, a four-CPU quota. Exact printed timing lines follow.','', '```text']
lines.extend(line for line in (out/'setup.log').read_text().splitlines() if line.startswith('setup:'))
lines+=['```','', 'Two focused compiler bootstrap attempts ended before any test leaf started, at 240.379 and 600.575 seconds. A later build found and prompted the keyed-counter compatibility fix. The oracle runtime products were subsequently prepared using native.RuntimeRequestFor, RuntimeRequest.Build with four workers, and main\'s existing buildcache inputs. Standalone native caches were populated under native\'s exact RuntimeRequest.Parts key. Product preparation and CLI warm-up logs are retained.','', '## Mutants','', '| Mutant | Observed catcher | Exit |', '|---|---|---|']
for row in mutants:
 catcher=', '.join(row.get('markers',[])) or ', '.join(row.get('catchers',[])) or row.get('catcher','')
 if row.get('structural_only'):catcher+='; structural only, not a semantic kill'
 lines.append('| '+row['mutant']+' | '+catcher.replace('|','\\|')+' | '+str(row.get('exit'))+' |')
lines+=['| nullable-token-omitted | Not run: its sole oracle catcher is overdue | none |','', 'The callback runner inherited an inaccurate final summary saying all five were caught, despite its explicit nullable-token skip. Its four-entry JSON ledger is authoritative; the source summary has been corrected for future use.\n\nThe constant self-equality mutant is built and checked inside TestSelfCompareConstantMutant: it compiled and ran successfully, and the NaN stdout disagreement caught it. Adaptation omission excludes the overdue NumberEvery leaf and requires the other eight original numeric witnesses to fail. The Weak callback guard is checked by exact diagnostic pins on the combined slice and separately isolated from the adapter using current main\'s object lowering. All fresh mutant sources are .go.txt overlays. Timeouts and build failures are not credited as kills.','', 'The union-liveness child completed after the collection supervisor ended. Its complete log records the intended FAIL in 9.14 seconds; its process exit code was not captured and is reported as unknown. The collection resumes from recorded kills, so that leaf was not rerun.','', '## Every recorded leaf\'s seconds','', '| Package | Test | Seconds |', '|---|---|---:|']
for row in times:lines.append(f"| {row['package'].rsplit('/',1)[-1]} | {row['test']} | {row['seconds']} |")
lines+=['','The same data is in test-seconds.json. It includes the over-budget stage3 wrapper; all other completed affected leaves are under 60 seconds. The unfinished leaves have no successful terminal timing event and remain explicitly blocked.','']
(out/'REPORT.md').write_text('\n'.join(lines))
