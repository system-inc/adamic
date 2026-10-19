import json,re,subprocess
from pathlib import Path
p=Path(__file__).parent
changes=json.loads((p/'counts-changes.json').read_text())
mutants=json.loads((p/'mutants-results.json').read_text())
checks=[json.loads(l) for l in (p/'checks.jsonl').read_text().splitlines()]
assert all(c['exit']==0 for c in checks) and len(mutants)==14 and all(m['caught'] for m in mutants)
seconds=[]
for f in sorted(p.glob('*-seconds.json')):
 if f.name=='test-seconds.json': continue
 for row in json.loads(f.read_text()): seconds.append(dict(check=f.name.removesuffix('-seconds.json'),**row))
for name,elapsed in re.findall(r'--- PASS: (\S+) \(([\d.]+)s\)',(p/'counts.log').read_text()):
 seconds.append(dict(check='counts',test=name,seconds=float(elapsed),action='pass'))
for line in (p/'counts-counterfactual.log').read_text().splitlines():
 e=json.loads(line)
 if e.get('Test') and e['Action']=='pass': seconds.append(dict(check='counts-attribution',test=e['Test'],seconds=e.get('Elapsed',0),action='pass'))
assert all(s['action']=='pass' and s['seconds']<60 for s in seconds)
(p/'test-seconds.json').write_text(json.dumps(seconds,indent=2)+'\n')
def rows(s):
 result={}
 for line in s.splitlines():
  a=[x.strip() for x in line.strip('|').split('|')]
  if len(a) in (5,7) and all(x.isdecimal() for x in a[1:]): result[a[0]]=line
 return result
old=rows(subprocess.check_output(['git','show','e3b5c979:internal/oracle/counts.md'],text=True))
comparison=[dict(key=c['key'],matches_source=old.get(c['key'])==c['after']) for c in changes]
assert all(c['matches_source'] for c in comparison)
(p/'counts-source-comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
lines=['Built: rebased compiler slice 3 with optional reads after calls, spread calls, JSON shorthand, refusal rulings, assignment witnesses and TSGo caching.',
'Commits: base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source e3b5c979fffa8faafa2e179b2b071b489b8d0db9; delivery SHA is in the final response and git log.',
'Commands and outputs: all focused checks, reader checks, changed-package vet, one counts regeneration and counts attribution passed.',
'Mutants: all 14 production mutants failed their intended tests; all eight embedded mutants were caught by their comparisons.',
'Not covered: whole package suites, full gate, the pre-generated flow corpus outside the remainder, WASI, macOS execution and the real checker ABI.',
'', '## Application and conflicts', '',
'Created compiler/loom-r-slice-3 from exactly dbdf977f5b, after git fetch origin. Fetched the source SHA explicitly and applied git cherry-pick --no-commit e3b5c979fffa8faafa2e179b2b071b489b8d0db9. Only its own parent-to-commit change was applied. The base stays fixed even if origin/main advances.',
'', 'Exactly two files conflicted, each with one marked hunk:',
'', '- internal/lower/refusals.go: main inserted parallelPreflight at the same preflight point where the slice inserted mergedFieldRulings and functionAnnotations. Kept parallelPreflight first, then mergedFieldRulings, then functionAnnotations, each returning its error. The auto-merged visitor retains main checks and the slice functionCallRefusal.',
'- internal/oracle/counts.md: main had five instantiation rows where the source added six TypeScript-mode ruling rows. Restored the entire base-main table, without hand merging any row, and regenerated it once after compilation and focused tests passed. The five instantiation rows remain unchanged; all six ruling rows are generated.',
'', 'Conservative interpretation of the worker scope: run focused selectors in every changed package and affected IR/flow/freshness/JavaScript readers, rather than whole package suites. No new compiler-consuming package or runtime feature set was added. The dependency registry needs no addition. TestEveryMapAProgramHoldsIsSorted and TestRuntimeSetsAreEachUsed pass. The pre-existing SortedMap types in the IR are retained.',
'', 'All new top-level oracle fixtures lower. FlowCorpusRemainder checked all 22 paths absent from the generated corpus, including this slice\'s seven new top-level oracle fixtures, with all four flow analyses. The refusal witnesses were already nested under refusal_rulings/, outside flow\'s top-level glob. No new top-level refused fixture remains to move. The nested computed optional-index witness now compiles and passes its dedicated oracle test.',
'', '## Counts', '',
'Exactly one -update-counts run on Linux with house clang 20.1.8: 22 additions, 11 existing numeric changes, zero removals. The command took 63.623 seconds wall time (62.094 seconds in the package); this is aggregate time, not a test leaf over 60 seconds. Its slowest test event was 5.07 seconds. The predicate-direction table is unchanged. Every changed numeric row matches the source commit\'s recorded value (counts-source-comparison.json).',
'', 'An isolated Go overlay restores only internal/lower/narrowed.go, internal/lower/exceptions.go, internal/native/reuse.go and internal/flow/build.go to the fixed base. Direct measurements of the 11 changed existing rows all reproduce the base values exactly: TestSlice3CountsAttribution passes, package 2.912 seconds. This did not regenerate or write counts.md. The optional-read/throw change therefore owns those 11 deltas. Ten rows now construct a catchable TypeError rather than stopping at the old invariant panic; 09_tree has conservative ownership/reuse changes from the new exceptional edges. New rows belong to the registered witnesses.',
'', 'Every moved row follows. Values are allocations/frees/retains/releases/peak/regions.', '', '| Fixture | Base | Rebased |', '|---|---|---|']
for c in changes:
 def values(s): return '/'.join(x.strip() for x in s.strip('|').split('|')[1:]) if s else 'new'
 lines.append('| '+c['key']+' | '+values(c['before'])+' | '+values(c['after'])+' |')
lines+=['', '## Runtime lines', '',
'The runtime delta against the fixed base adds precisely the same five lines as the source commit\'s own runtime delta, with no removals. There are no additional or altered runtime patch lines for clearance. Main\'s intervening runtime work stays intact. runtime-comparison.json asserts equality of every added/removed line.',
'', '```c', '// internal/native/runtime/adamic.h', '\tsize_t source_length;', '// internal/native/runtime/closure.c', '\tclosure->source_length = 0;', '// internal/native/runtime/union.c', '\tif (value != NULL && value->kind == adamic_kind_closure && strcmp(name, "length") == 0) {', '\t\treturn adamic_box_number((double)((adamic_closure *)value)->source_length);', '\t}', '```',
'', '## Validation', '',
'All Go test invocations use -count=1 and a hard -timeout; ordinary focused runs use -timeout=90s and -parallel=4. The runners set GOMAXPROCS=4 and four-CPU affinity. Oracle runs set ADAMIC_GATE_UNCACHED=1. Tests run source Node, JavaScript backend, release native, ASan/UBSan and Linux leak checks. The ruled missing-merged-field TypeScript cases pin their specified exit-70 checks independently of Node\'s undefined read. TSGo cache tests link inert ABI stubs: they test real cold/warm linking, Node output, placeless binary bytes and archive reuse, without claiming real checker ABI execution.',
'', 'Exact commands, exits and wall seconds are in checks.jsonl; raw outputs are the matching .log files. No test output was piped. Every recorded normal test event is below 60 seconds; no test was deleted, split or rerun for exceeding its budget. Every test event, including subtests and all counts rows, is listed in test-seconds.json. gofmt on every changed Go file and git diff --check outside inherited review evidence both print nothing. Historical patches/logs carried by the source are retained byte for byte and have their original patch whitespace.',
'', '| Check | Exit | Wall seconds | Command |', '|---|---:|---:|---|']
for c in checks:
 lines.append('| '+c['name']+' | '+str(c['exit'])+' | '+str(c['seconds'])+' | `'+ ' '.join(c['command']).replace('|','\\|')+'` |')
lines+=['', 'Counterfactual command: `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 timeout 180 go test ./internal/oracle -overlay=review/compiler/loom-r-slice-3/counts-counterfactual/overlay.json -run ^TestSlice3CountsAttribution$ -count=1 -json -timeout=90s -parallel=4`.',
'', 'Each selected top-level test\'s seconds follow; subtest seconds are in test-seconds.json.', '', '| Check | Test | Seconds |', '|---|---|---:|']
for s in seconds:
 if '/' not in s['test']: lines.append('| '+s['check']+' | '+s['test']+' | '+str(s['seconds'])+' |')
lines+=['', '## Mutants and catchers', '',
'Each production mutant ran independently through a Go overlay of the final source, with -timeout=80s and an outer timeout 90. All exited 1 at their intended test. Build failures, clang failures and timeouts were excluded from credit. Full mutated sources are .go.txt files under mutants/; commands and wall seconds are in mutants-results.json.',
'', '| Mutant | Catching test | Test seconds | Failure |', '|---|---|---:|---|']
for m in mutants:
 text=(p/'mutants'/m['name']/'test.txt').read_text()
 failures=re.findall(r'^--- FAIL: (Test\S+) \(([\d.]+)s\)',text,re.M)
 why='behavior assertion'
 if 'LeakSanitizer' in text: why+=' and LeakSanitizer (76 bytes in one allocation)'
 for n,s in failures: lines.append('| '+m['name']+' | '+n+' | '+s+' | '+why+' |')
lines+=['', 'The eight embedded mutants also pass their catch requirements: TestAssignmentProofScannerKeywordMutant, TestAssignmentProofIfNarrowingMutant, TestAssignmentProofReturnDefinedMutant, TestAssignmentProofChainedMutant, TestAssignmentProofCompoundMutant, TestAssignmentProofWiderTargetMutant, TestCallSpreadEvaluatedTwiceMutant and TestJSONStringifyShorthandWrongBindingMutant. Assignment witnesses and mutants pass stock TypeScript 6.0.3 with zero diagnostics. The six source mutants change Node stdout from its independent golden; spread double evaluation and wrong shorthand binding disagree with source Node in native and JavaScript. Each mutant\'s normal test seconds is in the preceding table.',
'', '## Setup and limits', '',
'The first setup returned 1 because the checkout left by the earlier interrupted task had an empty cohere worktree: cohere/TypeScript/tsc/go.mod was absent, despite setup reporting submodules ready. Restored the recorded cohere commit and initialized its pinned nested TypeScript submodule; no code was copied from it. A second setup stalled in the shared Go cache helper and was terminated (143). Used the supported ADAMIC_GOCACHE_OFF=1 fallback, and sourced /workspace/adamic-tools/env.sh for builds/tests. The formatting command initially lacked the sourced tool PATH; it was then run successfully with that environment. All source tests passed without code changes beyond the conflict resolution.',
'', 'Successful setup timing lines:', '', '```text']
lines += [l for l in (p/'setup-local.log').read_text().splitlines() if l.startswith('setup:')]
lines+=['```', '', 'nproc=5; cgroup cpu.max=400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0. The pinned TypeScript dependency install (timeout 120 npm ci --prefix stage3/api --ignore-scripts) exited 0.',
'', 'No ahra command, PR, Loom submission, main push or force push was performed. This unit does not claim a full gate, full package suites, execution of the older generated flow corpus, macOS/WASI validation, or real TSGo checker ABI behavior. The earlier unrelated review/compiler/loom-slice-3/BLOCKED.md remains untracked and is excluded from this commit. The delivery branch was absent remotely before work (remote-before.txt).']
(p/'REPORT.md').write_text('\n'.join(lines)+'\n')
print('Report written; test events',len(seconds),'max',max(s['seconds'] for s in seconds),'counts changes',len(changes))
