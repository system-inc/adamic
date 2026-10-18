from pathlib import Path
import json
out=Path(__file__).resolve().parent
base='dbdf977f5b8b0979d414ebe6f20e7aadd312e309';source='05eb9043dc8b01b32402ce33b91f97aadf524305'
all_events=[]
for p in (out/'logs').glob('*.jsonl'):
 if any(x in p.name for x in ['before-counts','pre-owner','initial']):continue
 for line in p.read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if r.get('Test') and r.get('Action') in ['pass','fail','skip']:all_events.append(dict(log=p.name,**r))
latest={}
for r in all_events:
 key=(r['Package'],r['Test'])
 if key not in latest or r['Time']>latest[key]['Time']:latest[key]=r
leaves=[r for r in latest.values() if not any(s['Package']==r['Package'] and s['Test'].startswith(r['Test']+'/') for s in latest.values())]
(out/'test-seconds.json').write_text(json.dumps(sorted(latest.values(),key=lambda r:(r['Package'],r['Test'])),indent=2)+'\n')
extras=json.loads((out/'extra-consumer-results.json').read_text());last={r['package']:r for r in extras};failures=[r for r in last.values() if r['exit']];red=[r for r in leaves if r.get('Elapsed',0)>=60];counts=json.loads((out/'counts-attribution.json').read_text());mutants=json.loads((out/'all-mutants-results.json').read_text());focused=json.loads((out/'focused-results.json').read_text());fixtures=json.loads((out/'fixture-results.json').read_text())
lines=[
'Re-cut slice 5 placeholders, parser step 24 and the negative-witness ledger onto the pinned main, preserving concurrent runtime infrastructure.',
'Delivery is one commit on '+base+'; source is '+source+'; the handoff gives the delivery SHA.',
'Core results: 38 focused commands, 55 final uncached differential fixtures, guards, call-target readers and changed-package vet pass; counts add 41 rows and change 21.',
'All 27 recorded independent mutants and two added metadata-publication mutants are caught, along with the source embedded mutants and eight witness repairs.',
'Not cleared for a heavy block: CSS took 63.12s, typeaware hit its 90s timeout during archive setup, and other consumer smoke tests have product-owner prerequisite failures; the gate was not run.',
'',
'The cherry-pick was --no-commit of the source commit itself, never a main-to-source diff. HEAD remained the exact pinned base until the single delivery commit. Conservative assumptions: the explicit pinned base overrides the current-main boilerplate; main owns packed caches, static owner tracking, atomic publication and async readiness rules; source Program-region lifetime remains pending. No cohere code was copied. Source .ts compatibility fixtures retain their dialect; no new .ts was authored. No ahra command, PR or Loom submission was made.',
'',
'Conflicts and main changes',
'',
'All 11 marked hunks across nine files are listed below; conflict-hunks.json retains their exact main/source/resolved text (counts is restored then regenerated). Main f93a65fb introduced concurrent lowering, fcf5c2ad concurrent runtime infrastructure, and e3282288 traced admission compiler products and the ledger schema.',
'']
for r in json.loads((out/'conflict-hunks.json').read_text()):lines.append('- `'+r['file']+'`, hunk '+str(r['hunk'])+': '+r['reason'])
lines += ['',
'Compatible-merge adaptations also replace library_object.c publication cache.index with an index derived from the returned slot. Reserved object-view writes find storage through immutable shape names, avoiding a second shared-cache read. Placeholder views use the actual static owner: main\'s owner-aware helper is exposed as adamic_object_field_owned and metadata indices derive from its returned slot.',
'',
'Runtime clearance',
'',
'Runtime edit lines differ from the source in exactly five files: adamic.h, class_static.c, library_object.c, object.c and share.c. Every differing added/deleted line is in runtime-line-differences.md. runtime-comparison.json retains both edit sequences; source-runtime-patch.json and rebased-runtime-patch.json contain complete exact patches as JSON strings. The other nine runtime files retain identical edit lines. union.c retains main\'s acquire load of shape_types in its surrounding context.',
'',
'The additional share.c change visits reference children of NodeArray interior metadata while leaving its zero-count interior header unshared. Optional metadata admits strings; omitting this traversal would leave reachable child reference counts unprepared for workers. TestParserNodeArrayMetadataSharing passes ASan/UBSan and leak checks in 1.02s. Its valid-C skip-metadata-sharing mutant fails with metadata child not shared. A second valid-C retain-interior-metadata mutant fails with interior metadata retained, proving the zero-count-header assertion too. This is an internal graph-publication invariant, separate from the Node differential fixtures. The first harness attempt failed -Wnewline-eof in 7.91s; its corrected run and mutant are retained. No counts fixture was added by this test. Post-sharing NodeArray counted-row, Node differential construction, async/shared-reference and runtime-set regressions pass.',
'',
'Counts rows',
'',
'Main\'s counts.md was restored verbatim before exactly one successful -update-counts invocation on Linux with house clang 20.1.8. It completed in 144.102s with 1,099 measured leaves; the largest count leaf was 14.73s. There are 41 added rows, 21 changed rows, no removals, no unattributed rows and no row differing from the source\'s measurement. counts-attribution.json records every before/after value and source attribution. The table below lists every affected row.',
'',
'| Row | Before | After |','|---|---|---|']
for r in counts['rows']:lines.append('| `'+r['fixture']+'` | '+('/'.join(r['before']) if r['before'] else 'added')+' | '+ '/'.join(r['after'])+' |')
lines += ['',
'Validation',
'',
'Commands and selectors are preserved verbatim in focused-results.json, fixture-results.json, consumer-results.json, extra-consumer-results.json and all-mutants-results.json. Complete go-test JSON output is under logs/. Raw .log files remain local as repository policy requires. test-seconds.json lists each observed test\'s seconds, including every new observed test. GOMAXPROCS=4 and GOFLAGS appended -p=4 while preserving -trimpath and -buildvcs=false; execution used the printed env.sh. Every go test had a 90s timeout except the counts family (180s); every orchestration/child command had an outer hard limit.',
'',
'- Focused selection: all 38 groups pass. A three-test parser-count group initially failed against main\'s untouched table, then passed after regeneration; both logs remain. Program-region lifetime and optional stack guard-cost tests explicitly skip and are not coverage claims.',
'- All 55 final fixtures pass with ADAMIC_GATE_UNCACHED=1 after the final runtime adaptation. These hold source Node, backend JavaScript, ASan/UBSan native, release native and successful-program leak observations. Initial fixture results are retained separately.',
'- TestEveryMapAProgramHoldsIsSorted passes; placeholder/factory IR maps remain SortedMap. The plain-map mutant is caught.',
'- TestRuntimeSetsAreEachUsed passes, including after the sharing adaptation. No new feature set is introduced.',
'- TestNativeAgreementRowsNameRegisteredPrograms passes. No unregistered agreement row is added.',
'- TestEveryTestIsParallelOrSaysWhy passes after adding the sharing test; that test calls t.Parallel first. No exemption or foreign-test edit was added.',
'- All three displayed-path product tests pass and their independent physical-path mutants are caught. No Go package is added, so compiler-dependencies.json needs no declaration.',
'- TestCallTargetReaders passes with -timeout 90s (39s initially, then 8.10s and finally 2.07s after the test edits).',
'- Flow corpus remainder, the two changed non-null flow corpus units, freshness corpus remainder and JavaScript mixed-union consumer pass.',
'- Main\'s async/shared-reference Node/native tests and TestParallelSlotPairs pass. Eight TestSlice5Witness repairs and stack regressions pass.',
'- Vet on cmd/adamic, internal/flow, fresh, ir, javascript, lower, native and oracle exits 0 without diagnostics; native vet repeats successfully after the new sharing test.',
'- Thirteen deliberate .a refusal controls moved unchanged into non_null_refused/. moved-refusals.json lists them. Their refusal tests pass. Accepted default-placeholder controls remain runnable; no foreign tests were changed to exclude a failing input.',
'',
'Consumer limitations and observed red',
'',
'The additional direct-consumer inventory is direct-consumers.json; extra-consumer-selection.json records bounded smoke selectors, not whole suites. Selected CSS TestCompositionMatchesGo_000 passed its assertions but took 63.12s in this four-CPU run. It is unchanged by this commit and was never rerun. This is an observation, not proof of its cold-tree timing or of the cause. Splitting another package\'s test is outside this unit\'s territory. The branch therefore has no claim of complete green consumer validation.',
'',
'Typeaware TestShadowIndexMissingBinding_000 also hit the 90s test timeout while its goroutine waited in harness.archive/buildProduct/typeAwareRunBuildCommand. Its stage0 phase had completed in 4.535561s; this does not establish a compiler semantic failure or its cause. It was not rerun.\n\nThe following are the latest failed consumer attempts. They are not mutant kills or compiler semantic evidence. Running the dedicated MarkdownInline and CSSNumbers printer owner tests succeeded (0.758s and 1.63s), but did not make the untraced CSSNumbers reader succeed. No product-owner implementation or unrelated test was edited.',
'',
'| Package | Selected test(s) | Result |','|---|---|---|']
for r in failures:
 command=r['command'];selector=command[command.index('-run')+1];errors=[]
 for p in (out/'logs').glob('extra-consumer-*.jsonl'):
  if '.pre-owner' in p.name:continue
  for line in p.read_text().splitlines():
   try:e=json.loads(line)
   except ValueError:continue
   if e.get('Package','').endswith(r['package'][1:]) and e.get('OutputType')=='error':errors.append(e['Output'].strip())
 lines.append('| `'+r['package']+'` | `'+selector+'` | '+('<br>'.join(errors[:2]) if errors else 'exit '+str(r['exit'])+'; see JSON log') .replace('|','\\|')+' |')
lines += ['', "The 11 additional compiler consumers declared by main's dependency registry were also attempted; declared-consumer-results.json holds commands and logs. Ten pass. stage3/lane TestApplyPython fails because its apply checkout has neither the product manifest nor ADAMIC_TYPESCRIPT_SOURCE pointing to TypeScript v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8. This is a missing gate-input prerequisite; no foreign test or source pin was changed.", '', 'Independent mutants and catchers', '', '| Mutant | Catcher |', '|---|---|']
for r in mutants:
 name=r.get('mutant',r.get('name'));cmd=r.get('command',[]);catcher=r.get('catcher',r.get('test',''))
 if cmd and '-run' in cmd:catcher=cmd[cmd.index('-run')+1]+(' '+catcher if catcher else '')
 lines.append('| `'+name+'` | `'+catcher.replace('|',' / ')+'` |')
lines += ['',
'All 29 independent mutant catchers exited nonzero through the intended assertion or behavior, not a compiler/build failure. omit-refusal-state specifically checks nil-result initialization via a Go panic. Embedded source mutants include publication/key/JSON, wrong factory payload, rewind, stack guard, flow-check removal, null-tag, spread readiness, wrong-zero/erasure/capture/exception/missing-value/message changes and eight negative-witness repairs. Their focused assertions require the intended disagreement.',
'',
'Toolchain and recovery', '',
'The first conflict resolver\'s closing-marker regex was greedy and truncated file tails. Inspection caught this; all exact three-way inputs were restored from git objects, the regex was corrected, and every hunk was reapplied. The first setup consequently failed on unexpected EOF in JavaScript/native Go files and is not credited. The repaired setup succeeded with ADAMIC_GOCACHE_OFF=1. setup-observations.json records the complete timing lines.',
'']
lines += ['- '+x for x in json.loads((out/'setup-observations.json').read_text())['timing_lines']]
lines += ['', 'nproc=5; cpu.max=400000 100000 (four CPU quota). Go 1.27.1, Node 24.19.0, clang 20.1.8. GOPROXY was exported as https://proxy.golang.org|direct before setup. Subsequent tests unset GOCACHEPROG for the local Go cache.', '', 'Not covered: whole-package suites, full gate, ahra/Loom submission, admission census, pending Program-region lifetime, opt-in guard-cost measurement, WASI or non-Linux systems. Consumer prerequisites and the observed CSS budget red remain explicit handoff limitations. Runtime line changes require the user\'s runtime-clearance route; no implicit clearance is claimed.']
(out/'REPORT.md').write_text('\n'.join(lines)+'\n')
(out/'validation-summary.json').write_text(json.dumps(dict(base=base,source=source,focused_successful_commands=sum(r['exit']==0 for r in focused),fixture_programs=sum(len(r['paths']) for r in fixtures),independent_mutants=len(mutants),additional_consumers_attempted=len(last),additional_consumers_total=len(json.loads((out/'extra-consumer-selection.json').read_text())),failed_consumers=[r['package'] for r in failures],over_budget=[dict(test=r['Test'],seconds=r['Elapsed']) for r in red],timed_out_consumers=['./stage1/cohere/typeaware'],complete_green=False),indent=2)+'\n')
print('report updated; consumers',len(last),'failed',len(failures),'budget red',len(red))
