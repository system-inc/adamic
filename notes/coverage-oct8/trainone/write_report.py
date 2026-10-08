#!/usr/bin/env python3
"""Render captured observations without abbreviating disagreement output."""
import json
from pathlib import Path
import subprocess
unit = Path(__file__).resolve().parent
root = unit.parents[2]
evidence = unit / 'evidence'
results = json.loads((evidence / 'results.json').read_text())
base = subprocess.check_output(['git', 'rev-parse', 'origin/main'], cwd=root, text=True).strip()
lines = [
 'Built 13 .a coverage programs and reproducible four-way and mutation runners; 6 agree and 7 stop at compile time.',
 f'Commits: base {base}; reviewed range 54cbc125..9f16421c; programs/evidence commit 5ea8459c131a463688c16f14f83b15f6492dff96; report/cleanup follow-up is on codex/coverage-oct8-trainone.',
 'Commands and outputs: run.py records stdout, stderr and exit codes; setup took 210.148s on nproc=5; package logs are below.',
 'Mutants: numeric-index and structural-map guards caught by lowering assertions; missing substr operand caught by compiler panic; deleted-key filter caught by ASan; two-index sort and allocation-size guards survived focused native tests.',
 'Not covered: admitted source deletion/presence-slot enumeration or optional Record access, combined null/undefined receivers, allocation failure injection, 32-bit targets, and the full repository gate.',
 '',
 '## Observations',
 '',
 'This is coverage-only work based on current origin/main. No compiler or runtime change is committed, no main merge and no PR. The full named files and slice history were read, plus CLAUDE.md, README.md, docs/0.1.md and docs/memory.md. The emit_statements.go and emit_slots.go slice hunks were read in history and current context. [slice-history.log](evidence/slice-history.log) preserves the requested git log -p output.',
 '',
 'The landing commits are cfa29460 (optional indexing), c4b29169 (substr), 8c962b67 (record aliases) and be356537 (checked enumeration). The range also contains the optional ArrayIndex slot hunk in 6eef75d9, debugger in 6c38c305, namespace readiness in 70ae18c3, and the Uint16Array addition/withdrawal pair 277e3937/f381d094. There is no new union.c comment in this slice.',
 '',
 'Source truth uses Node 24.19.0 via oracle/node.mjs, which strips types with Node itself. Native uses the compiler CLI with --sanitize (the oracle flags: -O1, ASan, UBSan, no recovery) and without it (-O2). Emitted JavaScript runs through the same Node runner. Comparison checks the raw stdout/stderr bytes and exit status. Each successful sanitized program runs again with detect_leaks=1. Every supported case exits 0 with empty stderr, including the separate leak runs. No baseline runtime abort, invalid C, wrong runtime output or leak was observed.',
 '',
 '| Program | What was tried | Four-way result |',
 '|---|---|---|',
 '| optional-nested.a | a?.[i]?.[j], absent base/inner row, empty rows, runtime strings, index replacing base binding | Agree |',
 '| optional-map.a | optional Map get followed by optional indexing; missing map/key/value; effects k then i | Agree |',
 '| optional-null.a | separate null array and Map receivers, skipped index/key, runtime values | Agree |',
 '| substr.a | negative/fractional/NaN/finite huge arguments, optional length, both surrogate halves, receiver changed by start operand | Agree |',
 '| entries-order.a | checked record entries, saved snapshot before overwrite, integer boundaries, empty/huge/ordinary keys | Agree |',
 '| entries-two-indices.a | exactly two numeric keys inserted out of numeric order | Agree |',
 '| entries-delete.a | delete then reinsert numeric key and delete ordinary key | Compiler refusal |',
 '| entries-absent.a | omitted optional number field with contextual homogeneous entries type | Compiler refusal |',
 '| entries-presence.a | omitted optional field versus explicitly present undefined | Compiler refusal |',
 '| optional-record.a | optional string-key Record lookup with side effect | NotYet |',
 '| optional-map-nullish.a / optional-nested-nullish.a | receiver includes null and undefined together | NotYet at representation |',
 '| substr-omitted-start.a | omitted/undefined start accepted by Node but outside Adamic library declaration | Type-check errors |',
 '',
 'The existing syntax_substr.a already covers a numeric cross product including infinities and split UTF-16 boundaries. These additions exercise finite magnitude 1e300, negative zero, receiver mutation during argument evaluation and additional half-surrogate boundaries. Existing optional_indexing_map.a already replaces its map binding in the key; the added map program composes the optional get with a separately effectful optional index. Existing record runtime harnesses cover deletion below the source-language refusal.',
 '',
 '## Compiler-stop disagreements',
 '',
 'For the three compiler columns below, these are compile-command outputs, not outputs from nonexistent binaries. That phase is recorded in results.json. An empty output block means zero bytes. These stops are observations, not claims of silent miscompilation.',
 ]
locations = {
 'entries-absent': 'internal/lower/library_object.go:140-141 rejects optional field layouts even with a homogeneous result contract.',
 'entries-delete': 'internal/native/runtime/record.c:108 implements runtime deletion, but internal/lower/refusals.go:27 rejects source delete before that path.',
 'entries-presence': 'internal/lower/library_object.go:124-125 rejects a result representation containing undefined.',
 'optional-record': 'internal/lower/collections.go:367 stops optional non-tuple indexing before optional_indexing.go can run.',
 'optional-map-nullish': 'internal/lower/expression.go:40 reports the unrepresentable receiver type before optional_indexing_map.go:44 can lower it.',
 'optional-nested-nullish': 'internal/lower/expression.go:40 reports the unrepresentable receiver type before optional_indexing.go:10 can lower it.',
 'substr-omitted-start': 'internal/lower/string_substr.go:33-34 has an absent-start fallback, but the embedded substr declaration requires a number start; checker errors happen first.',
}
for name, value in results.items():
 if value['agree']:
  continue
 lines += ['', '### ' + name, '', '**Kind: compiler error.** ' + locations[name], '', 'Program:', '```typescript', (unit / (name + '.a')).read_text().rstrip('\n'), '```']
 for mode in ['node', 'sanitize', 'O2', 'javascript']:
  run = value[mode]
  lines += ['', f"{mode}: exit {run['exit']}, phase {run.get('phase', 'run')}; stdout verbatim:", '```text', run['stdout'].rstrip('\n'), '```', 'stderr verbatim:', '```text', run['stderr'].rstrip('\n'), '```']
lines += ['', '## Comment and documentation audit', '',
 'Every comment block in the named whole files was inspected. Matching here means consistent with the implementation, not an independently proven specification claim. There are no comments or doc blocks in entries_records.go.', '',
 '| File and comment | Observation |', '|---|---|',
 '| optional_indexing.go:8-9 | Effects saves the receiver, Conditional guards null/undefined, and only the true arm contains ArrayIndex/StringIndex. The new nested probe observes the claimed effects. |',
 '| optional_indexing_chain.go:8-14 | Only a direct optional property receiver with required array/string/typed-array field is admitted; parentheses are not skipped. Missing element representation can flow into a following optional read. |',
 '| optional_indexing_map.go:8-9,67-69 | Conditional guards receiver before key evaluation; structural object/new hazards are scanned even behind a library annotation. Structural-guard removal is caught. |',
 '| string_substr.go:8-9 | Helper call receives operands once; stringInteger normalizes, negative start is relative, count determines slice end. Receiver mutation probe prints the original string and effects sl. |',
 '| library_object.c:1 | Comment says static Object methods over proven, fixed shapes. Actual keys and checked values dispatch to mutable record storage. **Kind: comment that does not match the code**, if fixed shapes is intended as the scope of this translation unit. |',
 '| library_object.c:32,52-53,78,98,157 | Private # fields bypass freeze check; canonical uint32 index recognition excludes the listed strings; insertion sort is stable for ordinary names; class keys delegate to class descriptors; checked reflection uses ordered keys and one actual-value read per key. No mismatch found in these bodies. |',
 '| record.c:1,8-11 | Record stores keys in a counted Map behind wrapper slots; observable keys come from table, not wrapper shape. Iterator wraps record and snapshot with scalar cursor. Offsets are 0/1/key as documented. No layout proof changes were made. |',
 '| record.c:42-43,75 | Prototype member guard runs only after an own miss; byte comparisons are dispatched by lengths; diagnostic fits the stated maximum without allocation. Existing record tests compare the member list with Node. |',
 '| record.c:109,118-119 | delete always returns true; index recognizer bounds length to ten bytes before uint64 arithmetic, rejects uint32 max, -0 and 01. Source deletion is nevertheless refused. |',
 '| union.c:1,35,40,57-58,88,164 | Union box equality follows numeric/string value equality and identity otherwise; conversion defaults panic; shape metadata is a static linked registry; scalar tags match ir.Type values. Comments describe these code paths. |',
 '| emit_slots.go:9-10,14-15,51-52 | Optional slot path snapshots before aside(index), emits key and cleanup inside present branch; TypeOf retains lookup presence to distinguish missing from null. |',
 '| emit_statements.go slice comments at 67,205-206,261 | Debugger emits no native instruction; ordinary property write evaluates value before undefined check; kept reference is taken before releasing old slot. No mismatch found. Namespace readiness and withdrawn Uint16 hunk add no conflicting comment. |',
 '', 'The library_object.c header mismatch is a description-scope inference grounded in the record branches at lines 45,99,176; changing the header is outside this coverage-only delivery.',
 '', '## Mutations and package checks', '',
 'Each production mutation was independent and restored with finally. The mutant runners record exact replacement text and commands. No compiler-warning failure is counted as a successful detector. Focused checks:', '',
 '| Mutant | Check and observation |', '|---|---|',
 '| numeric-index-guard, optional_indexing.go:35 | TestOptionalIndexingKeepsUnsupportedStorageNotYet/string_numeric_key fails its diagnostic assertion, exit 1. |',
 '| structural-map-guard, optional_indexing_map.go:27 | TestOptionalIndexingMapShapeRefused fails its structural-receiver diagnostic assertion, exit 1. |',
 '| substr-missing-argument-guard, string_substr.go:33 | Existing syntax_substr.a reaches missing length; compiler panics index out of range [2] with length 2, exit 1. This proves the guard is exercised, not a runtime-output detector. |',
 '| record-deleted-key-filter, record.c:188 | TestRecordsAgainstNode/semantics fails with ASan heap-use-after-free in array_index through record_keys, exit 1. |',
 '| record-two-index-sort, record.c:178 | count > 1 changed to count > 2; TestRecordsAgainstNode exits 0. See full-package and new-witness results below. |',
 '| record-allocation-size-guard, record.c:162 | Removed count > SIZE_MAX / sizeof *indices panic branch; TestRecordsAgainstNode exits 0. See full-package result below. |',
 ]
if (evidence / 'survivors.json').exists():
 for result in json.loads((evidence / 'survivors.json').read_text()):
  lines += ['', f"Full native package under {result['name']}: `{' '.join(result['command'])}`, exit {result['exit']}.", '```text', (evidence / ('full-' + result['name'] + '.log')).read_text().rstrip('\n'), '```']
if (evidence / 'sort-witness.json').exists():
 lines += ['', '**Kind: branch no native-package test covers for exactly two indices.** internal/native/runtime/record.c:178-179 sorts two keys in main; threshold mutant skips it. New program is entries-two-indices.a. Four outputs under that mutant follow. This is a deliberate mutant, not a main defect.']
 witness = json.loads((evidence / 'sort-witness.json').read_text())
 for mode, run in [('node', results['entries-two-indices']['node']), *witness.items()]:
  lines += ['', f"{mode}: exit {run['exit']}; stdout verbatim:", '```text', run['stdout'].rstrip('\n'), '```', 'stderr verbatim:', '```text', run['stderr'].rstrip('\n'), '```']
if (evidence / 'sort-existing-oracle.log').exists():
 lines += ['', 'The existing entries_record_alias source oracle under the same sort mutant:', '```text', (evidence / 'sort-existing-oracle.log').read_text().rstrip('\n'), '```', 'Observed: the existing source oracle catches this mutant with stdout differs, exit 1. Thus exactly-two-key ordering is already covered at repository level; only the native-package test set missed it.']
lines += ['', '**Kind: branch no test covers in the runs performed.** record.c:162 overflow panic branch survives removal. On this 64-bit target, at most 2^32-1 canonical numeric keys can exist, below SIZE_MAX / sizeof(indexed_key). Inference: this guard cannot fire for a valid record here; the result does not prove 32-bit safety. No enormous allocation or fabricated runtime state was attempted.', '',
 'The candidate guard audit is not an exhaustive branch census. Error propagation, malloc failure paths, synthetic IR-only storage states, all optional-chain rejection variants and every scalar box branch were not independently mutated. Existing substr IR mutants and record runtime mutants are additionally exercised by the validation commands below.', '',
 'Existing mutant checks run as part of the package/filtered validation:', '',
 '| Existing mutants | What catches them |', '|---|---|',
 '| substr relative-start, truncate, NaN, undefined-length, length-not-end, length-clamp, evaluation-order | TestSyntaxSubstrMutants requires native and emitted-JavaScript stdout to differ from source Node while finishing and passing leak checks. |',
 '| record indices-in-insertion-order, uint32-max-as-index, deleted-key-iterated | TestRecordMutants requires differing Node stdout with balanced counts and clean sanitizer exits. |',
 '| record overwrite-key-leaked | TestRecordMutants requires LeakSanitizer. |',
 '| record stored-key-freed / own-slot-null-read | TestRecordMutants requires ASan use-after-free / UBSan null-member access respectively. |',
 '| record prototype-membership-restored, missing-read-silent, own-read-checked-as-missing | TestRecordReadMutants requires failure of the exact own-only stop contract, or erroneous stop for an own hit. |',
 '| entries unproven-as-proven on checked_misfit, checked_literal and record_misfit | TestEntriesProvenance pins exit 70 and diagnostic; mutants instead match source Node completion in native and JavaScript. |',
 '| entries drop literal membership | TestEntriesProvenance/entries_checked_literal pins the literal-contract exit/message. |',
 '| entries unchecked readiness primitive | TestEntriesRuntimeReadiness pins undefined-field exit/message in both backends; altered IR instead matches Node completion. |', '',
 '## Commands, setup and limits', '',
 'All test output went directly to log files. Setup succeeded; its timing lines and environment are verbatim:', '```text', (evidence / 'setup.log').read_text().rstrip('\n'), '```',
 'Initial full lowering test exited 1 because the Node type dependency was absent; see lower-initial.log. Workaround: `npm ci --prefix stage3/api --ignore-scripts` installed the lockfile-pinned packages, with no tracked lockfile change. The rerun exited 0:', '```text', (evidence / 'lower-final.log').read_text().rstrip('\n'), '```',
 'Reproduction from repository root:', '```sh', 'source /workspace/adamic-tools/env.sh', 'go build -o /tmp/trainone-adamic ./cmd/adamic > /tmp/trainone-build.log 2>&1', 'python3 notes/coverage-oct8/trainone/run.py > /tmp/trainone-fourway.log 2>&1', 'python3 notes/coverage-oct8/trainone/mutants.py > /tmp/trainone-mutants.log 2>&1', 'python3 notes/coverage-oct8/trainone/survivors.py > /tmp/trainone-survivors.log 2>&1', 'python3 notes/coverage-oct8/trainone/oracle-sort-mutant.py > /tmp/trainone-sort-oracle.log 2>&1', '```',
 'Final validation uses `bash notes/coverage-oct8/trainone/validate.sh`; its exact commands are saved in that script and exit statuses in evidence/validation-status.log. The unmutated lowering command is `go test ./internal/lower -count=1 -timeout 30m`. Full native runs under each surviving mutant used `go test ./internal/native -count=1 -timeout 30m`.', '', 'Use the path printed by setup on a different machine. Mutant runners edit the shared tree temporarily; run them sequentially and do not overlap a baseline compiler rebuild with them. run.py uses a prebuilt compiler and stores compile output separately from execution output.',
 ]
for name in ['validation-status.log', 'validation-native.log', 'validation-oracle.log', 'vet.log', 'format.log', 'diff-check.log', 'fourway.log']:
 if (evidence / name).exists():
  lines += ['', name + ':', '```text', (evidence / name).read_text().rstrip('\n'), '```']
lines += ['', 'Successful raw outputs, every failed compile output, commands and exit codes are preserved in evidence/results.json and the individual .stdout/.stderr/.json files. Generated JavaScript is preserved as each successful JavaScript compile stdout. Logs are committed so temporary-directory cleanup does not remove the observations.', '',
 '## Inferences and remaining scope', '',
 'All four agreed on everything the current compiler admitted in these probes. Optional-record access, delete, optional-presence enumeration and the combined nullish type stop before runtime, so this unit cannot claim four-way execution coverage for them. The source Node outputs establish what those programs do, while existing C harness checks establish only runtime primitive behavior below the admission boundary. No absence of defects beyond these inputs is inferred.', '',
 'The full repository gate was not run. The selected package/oracle checks and their limits are recorded above. No counts table was changed because these programs live under notes and are not registered as ordinary oracle fixtures. No main branch update, PR, compiler fix or changed refusal contract is part of this work.', '']
(unit / 'REPORT.md').write_text('\n'.join(lines))
