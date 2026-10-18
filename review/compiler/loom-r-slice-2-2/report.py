import json,hashlib,re,subprocess
from pathlib import Path
out=Path(__file__).parent
load=lambda name:json.loads((out/name).read_text())
rows=lambda name:[json.loads(s) for s in (out/name).read_text().splitlines()] if (out/name).exists() else []
text='''Built: slice 2 plus naming, static-inventory and a-check fixes as one commit on compiler/loom-r-slice-2-2.
Commits: base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 4ca997c7a13f67e01591ab1722e9cc02ba90a4ee; delivery SHA in final response/git log.
Commands and outputs: both reported reds reproduced, corrected guards pass; focused tests, compiler consumers, counts, reader guards and vet rerun below.
Mutants: original semantic/guard probes and structural inline probe rerun; three new fix probes and six live header probes recorded below.
Not covered: full package suites/full gate, ahra or Loom submission, macOS/WASI/TSan; cycle-ledger original corpus skipped without its configured source tree.

## Reproduction and fixes

Fetched origin (main advanced to 99c369e0), but used exactly dbdf977f5b as explicitly requested. Created compiler/loom-r-slice-2-2 from that commit and applied only 4ca997c7's own diff with cherry-pick --no-commit. No application conflicts. The earlier slice's lowering, runtime and infrastructure resolutions are retained; the historical report is review/compiler/loom-r-slice-2/REPORT.md.

The exact reported tests were run before editing. TestEveryTestHasOneDescriptiveName failed with 35 colliding callback identifier families across internal/flow, internal/fresh and internal/oracle. TestRuntimeStaticsAreListed failed on exceptions.c:18, empty_message. names-before.log and statics-before.log retain the raw observations.

Renamed 35 slice-owned flow tests to TestFlowCallbackWidening... and 35 fresh tests to TestFreshCallbackWidening... . Kept oracle names, all checks and the naming rule/pinned exceptions unchanged. renames.json lists all 70 exact old/new identifiers. No tests were deleted, and no other package's tests were edited to make room.

Added only exceptions.c:empty_message:1 to internal/native/testdata/runtime-statics/storage.txt, next to main's existing error_name entry, with the protection audit. Its C type is not const, so the scanner requires review, but it has no mutable shared state: ADAMIC_STRING initializes references=0, slab=0, units=0, index=ADAMIC_LITERAL_INDEX. heap.c retain/drop paths leave a zero count unchanged; share.c skips zero-count non-region identities; string_index.c excludes immortal literals from units/index writes; string_append.c requires reference_count==1 for in-place writes. This follows main's existing immortal-literal convention and requires neither a lock nor thread-local storage. The unsafe-shared-state stop condition does not apply.

Strengthened the slice's existing TestErrorUndefinedMessageHasNoOwnProperty C harness to call string-unit lookup and share, then verify its literal's references, slab, units and index remain unchanged. The immutable-units mutant changes the real runtime cache-write guard and is caught by this assertion, not a compiler warning. An initial negative run returned through the assertion without freeing the temporary Error, so LeakSanitizer masked its expected exit 4; that attempt is retained and NOT credited. The harness now releases the Error before its intentional failure exit, and the final probe observes exit 4 with no sanitizer masking (fix-probes-final.log).

## Runtime clearance

NO runtime source line changes relative to 4ca997c7. internal/native/runtime/{adamic.h,exceptions.c,object.c} are byte-identical; all other runtime files are untouched. runtime-clearance.json records this. The only new runtime-related production-tree edit is the inventory key and its audit comment; the other edit is test-only C embedded in the slice's existing Go test. The original slice's runtime +/- lines remain exactly the original +19/-2 over three files, unchanged from the previous report.

## A-check

Built this branch's cmd/adamic and invoked `adamic c <path>` on every one of the 63 added/modified .a programs. Final results: 46 clean, 14 NotYet, three Refused, zero checker errors. The 46 clean programs also have counted native builds in the full counts refresh; source fixtures are held to Node and sanitizers by the focused tests. Clean and NotYet programs carry no a-check error header. All source program bodies are byte-identical to 4ca997c7 after removing an old or new first-line a-check header.

Exactly five header changes, fully recorded with unchanged-body SHA256 in header-changes.json:
- Add `// a-check: refused this escaping a constructor through a closure before every field is set` to miscompile_2a_refused/d7054e9_this_arrow.a and d7054e9_this_arrow_number.a.
- Remove obsolete `// a-check: compile` from stage3/fixtures/self-compare/array-fill.a, closure.a and sweep.a.

The existing array_structural.a refusal header already matches its measured reason and remains unchanged. No type-error headers were needed. acheck-final.json holds every command outcome, reason/code, body hash and header; per-file compiler stderr is retained. No a-check classification relies on a Mac table or a source-file exemption.

For each of the three refusal programs, physically removed and independently corrupted its header, compiled it again with the real branch compiler, applied README's documented reason/code predicate to the fresh diagnostic, and restored the bytes in finally blocks. All six fail the header predicate as expected (header-mutants.json). These are documented-header-policy checks using fresh compiler diagnostics, not a claim of running ahra/Loom or an unavailable historical Gate.aCheck implementation. The historical README report's pinned run.py path is absent from that object in this checkout; no gate implementation was modified.

## Counts, guards and execution conditions

Regenerated counts once after the header corrections with the Linux house clang 20.1.8. counts.md is byte-identical to 4ca997c7: no rows added, removed or changed in this fix. The complete slice still adds its original 46 owned rows versus dbdf977f5b. Counts leaf maximum is 3.53 seconds; counts-leaf-seconds.json records each row.

SortedMap, exact runtime witnesses, literal native-agreement registration, parallel-test explanations, descriptive unique names, static inventory, runtime field layouts and placeless split-build guards pass. No compiler-consuming package or runtime feature set was added. Reader guards use -timeout=90s; vet covers all seven changed Go packages. Four serial export preparation leaves keep direct Not parallel explanations; renamed tests call t.Parallel first.

All test output went to log files. Every long command had a hard timeout. Tests ran with GOMAXPROCS=4 and four-CPU affinity. nproc=5; cgroup cpu.max=400000 100000. Setup passed without workaround, using GOPROXY=https://proxy.golang.org|direct and the printed env.sh. Complete setup timing lines:

```text
'''
text+='\n'.join(s for s in (out/'setup.log').read_text().splitlines() if s.startswith('setup:'))+'\n```\n\n'
text+='## Focused commands and results\n\n| Check | Exit | Wall seconds | Command |\n|---|---:|---:|---|\n'
for r in rows('checks.jsonl'):text+=f"| {r['name']} | {r['exit']} | {r['seconds']} | `{' '.join(r['command'])}` |\n"
text+='\n## Compiler consumers\n\nThe seven compiler/fixture packages use the focused checks above; the other 43 direct consumers use the prior explicit focused plan again. Packages with no tests get compilation checks. The only skip is TestOriginalCycleLedger for absent ADAMIC_CYCLE_LEDGER_ROOT. This is not original compiler-corpus coverage. No cohere code was copied.\n\n| Package | Selector | Exit | Wall seconds | Leaf outcomes |\n|---|---|---:|---:|---|\n'
for r in rows('consumers.jsonl'):
 obs='compile only' if r['compile_only'] else ', '.join(f"{e['Test']}: {e['Action']} {e.get('Elapsed',0)}s" for e in r['leaves'])
 text+=f"| {r['package']} | `{r['pattern']}` | {r['exit']} | {r['seconds']} | {obs} |\n"
text+='\n## Mutants\n\nAll original semantic probes are rerun on this branch, including the embedded TestSelfCompareConstantMutant, which compiles/runs its NaN mutant and requires wrong stdout against Node. Static inline placement remains a structural-only clang catch and receives no semantic credit. No build failure or timeout is credited as a semantic catch. Full commands, overlays and logs are under this review directory; mutant Go sources have .go.txt extensions. The earlier narrow refusal-revert and isolated Weak probes retain main infrastructure instead of restoring obsolete compiler files.\n\n| Mutant | Caught | Wall seconds | Catcher |\n|---|---|---:|---|\n'
for r in rows('mutants.jsonl'):text+=f"| {r['mutant']} | {r['caught']} | {r['seconds']} | {'; '.join(r['markers'])} |\n"
for file in ['grain-mutants.json','fix-mutants.json']:
 if (out/file).exists():
  for r in load(file):text+=f"| {r['mutant']} | {r['caught']} | {r['seconds']:.3f} | {r['marker']} |\n"
text+='\nSix additional live header-policy catches are detailed in header-mutants.json. The first masked immutability probe is recorded separately in fix-mutants-initial.json and fix-mutants/immortal-units-write-initial/test.log; final proof is the exit-4 assertion after cleanup.\n'
text+='\n## Per-test timing evidence\n\nEvery added/touched leaf stays below 60 seconds; no over-budget leaf was rerun. *-seconds.json files hold individual measurements, consumers.jsonl holds consumer leaves, and mutant-test-seconds.json holds negative probes. Wall seconds above include Go compilation and may exceed a leaf\'s measured duration.\n\n| Test | Seconds | Check |\n|---|---:|---|\n'
for p in sorted(out.glob('*-seconds.json')):
 if p.name in ['counts-leaf-seconds.json','mutant-test-seconds.json']:continue
 for r in load(p.name):
  if '/' not in r['test']:text+=f"| {r['test']} | {r['seconds']} | {p.stem} |\n"
(out/'REPORT.md').write_text(text)
