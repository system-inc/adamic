import csv,json,pathlib,subprocess
root=pathlib.Path.cwd(); out=root/'review/compiler/loom-hidden-boundaries-2'
records=[]
for p in sorted(out.glob('*.json')):
 try:r=json.loads(p.read_text())
 except ValueError:continue
 if isinstance(r,dict) and 'command' in r and 'tests' in r:
  r['label']=p.stem;records.append(r)
(out/'commands-results.json').write_text(json.dumps(records,indent=2)+'\n')
with (out/'test-seconds.csv').open('w') as stream:
 writer=csv.writer(stream);writer.writerow(['run','test','outcome','seconds'])
 for r in records:
  for t in r['tests']:writer.writerow([r['label'],t['name'],t['action'],t['seconds']])
base=(out/'base-main-sha.txt').read_text().strip()
paths=subprocess.check_output(['git','diff',base,'--name-only'],text=True).splitlines()
production=[p for p in paths if p.endswith('.go') and not p.endswith('_test.go')]
(out/'runtime-files.txt').write_text('\n'.join(production)+'\n')
(out/'counts.diff').write_text(subprocess.check_output(['git','diff',base,'--','internal/oracle/counts.md'],text=True))
parts=[f'''Re-cut the rejected hidden-boundaries change with an exact receiver-runtime witness.
Source commit 688e1cd84e5fa449714a7837c41a47d2d987c969; main base {base}; the final response supplies the delivery SHA.
Source Node and all three compiled lanes agree; runtime witness/products, focused controls, counts, reader guard and vet passed.
The original 45 mutant detections were rerun, plus one new stale-witness mutant reproducing Loom's failure.
Full packages, full gate, full census, historical replay and performance checks were not run.

Loom change: chg_9qbdqc9eer68bywzjmnrs2q9be. Branch: compiler/loom-hidden-boundaries-2. No PR or Loom submission was made. The earlier focused verification omitted TestRuntimeSetsAreEachUsed, allowing the stale witness to pass worker verification.

Why the old witness stopped needing receivers

census_optional_values_system.a contains System.write, System.readFile and System.exit as literal methods, plus process.stdout.write and process.exit. None reads this or super, including defaults and nested arrows. receiverIndependentMethod proves that syntactically; objectMethod marks their IR functions Receiver=false and creates no this parameter. Their calls then use the ordinary closure ABI. This removes fifteen receiver retain/release pairs, moving the fixture from 67/91 to 52/76 retains/releases while allocations/frees stay 29/29. Native emission finds no receiver closure and emits no ADAMIC_CLOSURE_RECEIVERS define. native.Features is exactly []. This is the intended compiler change, not lost source behavior.

The replacement

host_void_method.a has receiver.write reading this.prefix. Its system.write method is receiver-independent, but receiver.write still needs the receiver ABI. Native emission turns on exactly [ADAMIC_CLOSURE_RECEIVERS], with no convention or canonical-closure feature. Replaced only the receiver-only runtimeSets witness and added its reason as a comment. Neither fixture was edited.

The set is retained. No test or product stops using it. All eleven feature sets remain, and all four receiver-only archives remain owned by TestProduct_NativeRuntimesSanitized0, TestProduct_NativeRuntimesRelease0, TestProduct_NativeRuntimesCounted0 and TestProduct_NativeRuntimesWASI0. All twelve runtime-product test groups passed, building the eleven sets in each of four option modes. The initial native product run skipped WASI before its SDK was available; the subsequent enabled WASI run passed all three groups.

Four-lane observations

| Fixture | Source Node | JavaScript backend Node | Release native | ASan/UBSan native | Features |
|---|---|---|---|---|---|
| census_optional_values_system.a | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | [] |
| host_void_method.a | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | identical output, empty stderr, exit 0 | [ADAMIC_CLOSURE_RECEIVERS] |

LeakSanitizer also passes for both. receiver-four-lanes.log records every lane's stdout/stderr/exit explicitly. Source Node stdout, which every compiled lane reproduces byte for byte:

census_optional_values_system.a:
```text
written
exit:undefined
exit:undefined
exit:7
file:utf8
file:ascii
missing
exit:undefined
exit:2
```

host_void_method.a:
```text
hello
receiver: called
argument
last
```

Integration and scope

Fetched current main and squash-merged the previous branch's delivery commit. The squash had no conflicts. Git automatically combined expression.go and counts.md; main's null/undefined template interpolation is preserved. The new authored source change is in internal/oracle/runtimes_test.go only. The delivery retains the previous compiler net change, TypeScript hatch controls and historical evidence. No cohere code was copied or submodule changed. New evidence lives under review/compiler/loom-hidden-boundaries-2; previous evidence is historical provenance, not this run's results.

Production compiler files changed against main:

''']
parts.append('\n'.join('- '+p for p in production)+'\n\nNo native runtime C or header source changed. No new package, compiler consumer or product was added, so compiler-dependencies.json needs no change. Products retain main\'s portable keys; no test/product hard-coded build directory was introduced. Mutant source files are .go.txt and overlays only.\n\nSetup\n\nGOPROXY=https://proxy.golang.org|direct; nproc=5, cpu.max=400000 100000 (four CPUs). All verification shells use GOMAXPROCS=4; ordinary focused tests use -parallel 4, runtime products and the two four-lane witnesses use -parallel 1. Go 1.27.1, Node 24.19.0, native clang 20.1.8 and WASI SDK 27 with clang 20.1.8. Both setup runs succeeded. Timing lines:\n\n```text\n')
for name in ['setup.log','setup-wasi.log']:
 parts.append(name+'\n'+'\n'.join(line for line in (out/name).read_text().splitlines() if line.startswith('setup:'))+'\n')
parts.append('```\n\nCounts\n\nOne successful TestCountsAreRecorded regeneration ran on this base. It left the squashed counts table byte-identical. Against main the inherited change adds 33 fixture rows and changes nine existing rows; the witness fix moves no row. counts.diff contains every row. Current-ABI and main-ABI checks reproduce all nine sets of counts, and the wrong-ABI mutant fails all nine owning count assertions.\n\nCommands and outputs\n\nEvery test wrote directly to a log; none was piped. Every long command has an outer timeout and every correctness Go test has -timeout 90s. Full count/product families use longer bounded process limits to run their already split top-level leaves; every measured leaf stayed below 60 seconds. No test was deleted or split.\n\n| Run | Exit | Command seconds |\n|---|---:|---:|\n')
for r in records:
 if not r['label'].startswith('mutant-'):parts.append(f"| {r['label']} | {r['exit']} | {r['seconds']} |\n")
parts.append('\nExact command arrays are in commands-results.json. logs.json preserves every raw log in the commit, since the repository intentionally ignores .log files; the named local logs remain in this review directory. Every leaf\'s seconds, including products, counts and expected mutant failures, are in test-seconds.csv. Go compilation is included in command seconds; test seconds include in-test setup.\n\nMutants\n\nThe first mutant batch reached its 600-second outer limit after 33 catches. Only unfinished checks were resumed under another hard limit; completed catches were not repeated. Compiler overlays are regenerated from the current merged tree, with original patch positions mapped through main\'s edits rather than using stale full-source overlays. Each caught mutant must exit 1 with an owning assertion; no build-warning or timeout failure counts. The four flow mutants additionally require their specific failure marker.\n\n| Mutant run | Exit | Actual failing catchers |\n|---|---:|---|\n')
for r in records:
 if r['label'].startswith('mutant-') or r['label'] in ['counts-audit-wrong-abi-mutant','stale-runtime-witness-mutant']:
  parts.append(f"| {r['label']} | {r['exit']} | {', '.join(t['name'] for t in r['tests'] if t['action']=='fail')} |\n")
parts.append('''
The two remaining original semantic mutants are in the green oracle controls: TestHiddenBoundary04PresentArrayMutant catches undefined returned for present input through Node stdout (0|0 versus 7|0); TestHiddenBoundary04EagerIndexMutant catches eager index evaluation through Node side-effect counts. Together with the 38 compiler overlays, four flow overlays and one wrong-ABI overlay, these are the original 45 detections. The new stale-runtime-witness overlay restores census_optional_values_system.a as the witness and fails TestRuntimeSetsAreEachUsed with the same "turns on []" message as Loom. It tests the witness check itself while preserving correct compiler and fixture behavior.

Limits

No whole-package test run or full gate ran. Historical census/replay/performance evidence inherited by the squash was not remeasured. WASI runtime archives were built, but WASI fixture execution was not requested or run. The old ABI attribution's TDZ control agrees on empty stdout and an explicit stop; it does not assert equality of raw Node/Adamic error text or exit status. No PR was opened and no Loom submission was made.
''')
(out/'REPORT.md').write_text(''.join(parts))
