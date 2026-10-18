import json,re
from pathlib import Path
out=Path(__file__).parent
load=lambda p:json.loads((out/p).read_text())
def rows(p):return [json.loads(s) for s in (out/p).read_text().splitlines()] if (out/p).exists() else []
checks=rows('checks.jsonl'); consumers=rows('consumers.jsonl'); mutants=rows('mutants.jsonl')
text='''Built: source e1d03f9700a51458833f12db1320d3d95374a025 applied as one commit on pinned main for compiler/loom-r-slice-2.
Commits: base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA is in the final response and git log.
Commands and outputs: focused source tests, Node and sanitizers, flow corpus checks, guards, counts and changed-package vet; exact results below.
Mutants: original semantic checks plus extended guard probes and structural inline probe; every observed catcher is below.
Not covered: full package suites/full gate, ahra/Loom, macOS and WASI; cycle-ledger original-source corpus skipped for missing configured input.

## Base and conflict resolutions

Fetched origin, verified the full base SHA above, checked out compiler/loom-r-slice-2 from that exact commit, then used `git cherry-pick --no-commit e1d03f9700a51458833f12db1320d3d95374a025`. This applies only the source commit's own parent-to-commit diff. No diff from main to the source tip was applied. The initial conflict hunks are retained in conflicts.diff.

Every conflicted hunk:

1. internal/ir/call_targets_guard_test.go, callTargetExports: source's custom archive product conflicted with main's centralized buildcache.GoExports. Kept main's GoExports and its explicit leakcheck dependency, called separately for native/lower/fresh/flow. Kept all four reader leaves, four product leaves, and four serial preparation leaves. The serial leaves prepare the actual shared product instead of calculating unused legacy inputs. Removed the superseded custom archive extraction implementation. Retained source's external test import alias handling and per-package stale-reader checks.
2. internal/native/runtime/adamic.h, adamic_thrown declaration: main made the pending exception thread-local. Kept `extern _Thread_local adamic_object *adamic_thrown`, added source's omitted-message comment and field-name helper declaration.
3. internal/native/runtime/exceptions.c, after adamic_error_new: main added OrdinaryToPrimitive, including lazy method lookup, ownership, pending-exception propagation and TypeError creation. Kept that whole function and source's separate adamic_error_field_name function. Kept thread-local state and source's absent-message shape, empty fallback and representation bytes.
4. internal/native/runtime/object.c, adamic_object_optional_find: main publishes shape and slot together in an atomic packed cache. Kept its atomic load, search, cache-store helper and absence behavior; inserted only source's field-name remapping before lookup. The ordinary find remapping auto-merged while keeping main's packed cache store.

Auto-merged lowering/emission changes were retained together with main's newer infrastructure; auto-merges.patch records those diffs. No production lowering change beyond the source's own diff was necessary. Main's newer jobserver support in units.go and newer lowering infrastructure remain in the auto-merged result.

Other necessary adaptations: the native Error C harness now initializes main's one-word caches with `{0}` instead of `{NULL, 0}`. The source self-compare test uses main's shared `leakcheck.Unbalanced(leakRun(counted))` instead of removed local `unbalanced`; the first guard compile attempt exposed this before running tests (initial-build-failure.txt).

Sixteen refusal fixtures moved into miscompile_2a_refused/ and weak_miscompiles_refused/, with corresponding path/fixture registrations updated. refused-fixtures.json lists every move and verifies unchanged source contents. Accepted fixtures remain in their original directories. Conservative assumption: all rejected fixtures belong under *_refused even though this pinned base's flow collector globs only top-level oracle files. Eight separate TestSlice2 flow leaves explicitly exercise the accepted Error and Weak programs; the source's 35 callback flow leaves remain. No other package's tests were edited to conceal a refusal.

## Runtime clearance

Every runtime +/- line is IDENTICAL to the source commit's own runtime +/- lines. There are no differing runtime lines to list. runtime-comparison.json contains both complete line lists; runtime.patch is the re-cut patch against the pinned base. adamic.h +2/-0; exceptions.c +15/-2; object.c +2/-0. Newer main context (thread-local pending errors, OrdinaryToPrimitive, atomic cache publication) remains untouched in the own diff.

## Counts

Took base counts.md, then regenerated exactly once on Linux with house clang 20.1.8, after focused compilation/tests passed: `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout=600s -args -update-counts`. Exit 0. Added 46 owned rows: 35 callback, three miscompile controls/Error, five Weak, three self-compare. No existing row changed or disappeared. All 46 added rows also match the source commit numerically (counts-source-comparison.json). counts-attribution.json and counts-rows.md list every row and number. This exact base has counts-only counts.md: main already moved run digests into keyed measurements; its header directs the command used. No old schema or Mac table was introduced.

## Guards and execution conditions

SortedMap, exact runtime-set witnesses, literal native-agreement registration, every-test parallel/explanation, runtime field-layout and placeless split-build guards pass. No compiler-consuming package or runtime feature set was added, so existing compiler-dependencies and runtime-set declarations suffice. All new parallel tests call t.Parallel first. Four serial export preparation leaves carry direct Not parallel explanations. TestCallTargetReaders used -timeout=90s. Vet covered all seven changed Go packages.

GOPROXY=https://proxy.golang.org|direct; source /workspace/adamic-tools/env.sh. nproc=5; cgroup cpu.max=400000 100000. Runners use four-CPU affinity and GOMAXPROCS=4. Setup passed, no workaround needed; its complete timing lines follow:

```text
'''
text+='\n'.join(s for s in (out/'setup.log').read_text().splitlines() if s.startswith('setup:'))+'\n```\n\n'
text+='## Focused commands\n\nAll test output was written to logs, never piped. Wall seconds include command compilation; individual leaf seconds are in *-seconds.json. Every long command had a hard timeout. No test over its 60-second share was rerun.\n\n| Check | Exit | Wall seconds | Command |\n|---|---:|---:|---|\n'
for r in checks:text+=f"| {r['name']} | {r['exit']} | {r['seconds']} | `{' '.join(r['command'])}` |\n"
text+='\n## Direct compiler consumers\n\nreader-packages.json records all 50 direct consumers discovered from Go imports, including test imports. Seven are covered by the focused checks above. The other 43 use these explicit selectors, avoiding whole suites. Packages without tests are compiled. TestOriginalCycleLedger skips without ADAMIC_CYCLE_LEDGER_ROOT; that corpus is not claimed as run. Setup also compiled the entire repository. No cohere code was copied.\n\n| Package | Selector | Exit | Wall seconds | Observations |\n|---|---|---:|---:|---|\n'
for r in consumers:
 obs='compile only, no tests' if r['compile_only'] else ', '.join(f"{e['Test']}: {e['Action']} {e.get('Elapsed',0)}s" for e in r['leaves'])
 text+=f"| {r['package']} | `{r['pattern']}` | {r['exit']} | {r['seconds']} | {obs} |\n"
text+='\n## Mutants and catchers\n\nExternal mutant commands, overlays and full logs are retained under mutants/. Go mutant sources have .go.txt names. A build failure, timeout or compiler warning does not count as a semantic kill. Original refusal-revert is expressed narrowly as the omitted adapter plus its recorded eight refusal assertions, rather than restoring five obsolete compiler files over main. The isolated Weak callback probe disables that stop and uses this pinned base\'s object.go to omit the adapter, retaining main\'s newer unrelated object lowering.\n\n| Mutant | Caught | Wall seconds | Observed catcher markers |\n|---|---|---:|---|\n'
for r in mutants:text+=f"| {r['mutant']} | {r['caught']} | {r['seconds']} | {'; '.join(r['markers'])} |\n"
if (out/'grain-mutants.json').exists():
 for r in load('grain-mutants.json'):text+=f"| {r['mutant']} | {r['caught']} | {r['seconds']:.3f} | {r['marker']} |\n"
text+='\nTestSelfCompareConstantMutant is the embedded original semantic probe: the mutated NaN comparison compiles and executes successfully, and the oracle observes wrong stdout; the test passes only when that mutant is killed. All 17 original semantic probes and four extended semantic/guard probes were caught. Negative probe leaves also stayed below 60 seconds (maximum 36.64 seconds); mutant-test-seconds.json records them. The static-inline-definition probe is structural only, caught by clang, and receives no semantic credit. Refusal boundary probes distinguish conservative rejection from accidental admission; sanitizer probes independently establish the unsafe behavior where applicable.\n'
text+='\n## Individual affected test timings\n\n| Test | Seconds | Check |\n|---|---:|---|\n'
for p in sorted(out.glob('*-seconds.json')):
 for r in load(p.name):
  if '/' not in r['test']:text+=f"| {r['test']} | {r['seconds']} | {p.stem} |\n"
(out/'REPORT.md').write_text(text)
