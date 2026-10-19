from pathlib import Path
import collections, json, subprocess
root=Path('/workspace/adamic')
evidence=Path('/tmp/adamic-void-evidence/evidence/compiler/lower-void-widening')
evidence.mkdir(parents=True,exist_ok=True)
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
base='141fec7702d4aed21429724b2afd7efcd980f277'
other=json.loads(Path('/tmp/void-evidence/admission-other-movement.json').read_text())
summary=json.loads(Path('/tmp/void-evidence/admission-final-summary.json').read_text())
impact=json.loads(Path('/tmp/void-evidence/typescript-impact.json').read_text())
probes=json.loads(Path('/tmp/void-evidence/diagnostics.json').read_text())
stat=subprocess.check_output(['git','show','--stat','--oneline',sha],cwd=root,text=True)
report=f'''Task #0p2kkth. Implementation `{sha}` on `compiler/lower-void-widening`, based on newest fetched origin/main `{base}`. Evidence is on `evidence/compiler/lower-void-widening`. The first green was pushed at 07:00 UTC on October 19, 2026.

Lower refuses exactly-void callable results viewed through any callable result other than exactly void, including each target union member and resolved generic parameter. It also refuses void-valued returns at non-void return sites. The fix says to declare the function's real return type. Both suffixes use the same pass.

The visitor calls the new pass at `internal/lower/refusals.go:258`, immediately before optional widening. `internal/lower/void_widening.go:10` owns the refusal; its call-signature helpers begin at lines 93 and 106. `internal/lower/functions.go:334` also recognizes void union members at implicit returns: w02 otherwise panicked when falling off the end. The original failing run is `logs/focused.log`. `internal/oracle/void_widening_test.go:23` walks and pins the 17 refused programs; line 13 registers the legal fixture.

Three existing fixtures needed migration, without exemptions: `async_plain.a:6` and `async_coverage_discard.a:11` now explicitly discard awaited void results; `regexp_replace/callback.a:32` declares the callback's actual undefined result. Their uncached oracle run passed. Their counts did not move.

Every refusal below has its complete repository-relative diagnostic in diagnostics.json and the adjacent checked-in .refused file. Every source ran successfully on Node and typescript-go accepted it before Lower refused it. All have the exact What in their first-line a-check header.

| Probe | Pinned diagnostic |
| --- | --- |
'''
for probe in probes:
 report+='| '+Path(probe['path']).name.removesuffix('.refused')+' | `'+probe['diagnostic'].replace('|','\\|')+'` |\n'
report+='''
The accepted fixture covers v09's exactly-void forwarding return, v08's declared void/number result, w02's implicit void union exit, number-valued functions viewed as void, and a string-valued arrow passed as void. Source Node, backend JavaScript, sanitized native, and the oracle's release native all agree, with empty stderr and exit 0:

```text
done
number
4
none
discarded
```

The nil-refusal mutant failed all 17 refused-program subtests, exit 1; several otherwise stop at a later NotYet, and the test still requires the intended Refused diagnostic. The union-only admission mutant compiled v04 without warnings, then sanitized native and backend JavaScript printed `undefined`, while source Node printed `number`. Each run exited 0 with empty stderr. The mutant files and Go overlays are retained; repository source was never mutated for these two experiments.

Linux 6.18.55, clang 20.1.8, Go 1.27.1, Node 24.19.0. Formatting and vet were clean. These final commands passed, with full output in logs:

```sh
gofmt -l cmd internal
go vet ./internal/lower ./internal/oracle
go test -count=1 -timeout 30m ./internal/lower ./internal/fresh ./internal/flow
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle
```

Only one counts row was added: `void_widening_legal.a`, allocations/frees/retains/releases/peak/regions = `8/8/7/18/6/0`. No other row or recorded count changed. The first counts attempt saw the three old fixture forms and failed; the final run after migration passed. Both logs are retained.

'''
report+=f'''The final comparison built adamic from the implementation commit and clean main, then ran `adamic c` on all {summary['completed']} tracked .a/.ts programs, excluding .d.ts, cohere/, and any node_modules directory. Timeouts: {len(summary['timeouts'])}. Every program whose first stop became this refusal is listed below; raw exit status, stderr, and timing for both compilers are in admission-final.jsonl. The exploratory scan also identified the three migrated fixture forms above.

'''
for row in summary['newly_refused']:
 report+='- `'+row['path']+'`\n'
report+=f'''
There were also nine existing compiler panics in tracked review/agree fixtures. Both compilers exited 2 at the same native source frames with identical panic messages; addresses and build-root paths differ. All nine paths and both full stacks are in admission-other-movement.json. No new panic or other exit-status movement occurred outside the 17 refused probes. The base binary used -buildvcs=false in a detached main worktree with the pinned cohere dependency linked; the head binary records the delivered SHA and vcs.modified=false. Build provenance is retained.

TypeScript compiler impact: a stock TypeScript {impact['version']} checker-API pattern inventory at upstream commit `050880ce59e30b356b686bd3144efe24f875ebc8` found {len(impact['hits'])} sites across {len(set(h['file'] for h in impact['hits']))} files. It included 77 original compiler files plus the generated diagnostic source ({impact['files']} inputs total). This is a pattern count, not a claim that pristine tsc passes Adamic's gate. Two checker diagnostics remain: {"; ".join(impact['diagnosticDetails'])} The rule was not narrowed. All sites and the scanner are retained.

| Compiler file | Sites |
| --- | ---: |
'''
for file,count in collections.Counter(h['file'] for h in impact['hits']).items(): report+=f'| `{file}` | {count} |\n'
report+='''
A separate existing ownership defect remains: discarding runtime-built strings through a void view leaks 223 bytes in three allocations on both base and branch. Both print `discarded`, while LeakSanitizer exits 1; source Node exits 0. `discard-dynamic.ts.txt` and both sanitizer logs preserve the witness. The final legal fixture uses the supplied number/literal-string probe shapes. This patch does not claim that dynamically allocated discarded results are released correctly, and does not add the later value-preserving ABI fix. The early stronger-control failure is preserved in focused-final.log.

Direct `() => void` to `() => {}` is already rejected by typescript-go with TS2322 under this checkout's options. The Lower fixture therefore uses the accepted union `(() => void) | (() => {})`; the direct checker rejection is retained.

Not run: the whole-repository ./... gate, a native build of TypeScript itself, or separate tsc/cohere CLI acceptance runs. The requested package gates and complete uncached oracle were run. No main push, force push, or review/ write was performed.

Implementation files and lines:

```text
'''+stat+'```\n'
(evidence/'REPORT.md').write_text(report)
