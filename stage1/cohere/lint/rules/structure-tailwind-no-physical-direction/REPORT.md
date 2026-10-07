Built: all four claimed rule implementations now exist; strict description options are complete. Shared JSX and suppression gaps remain explicit.
Commits: f60c905a adds definite assignment; 7e888bd0 adds strict description options; fresh validation evidence is committed next.
Commands and outputs: 197 upstream cases and 77 compiler plus 140 stage1 sources match Go on Node, emitted JavaScript and sanitized native; 165 raw-source corpus observations all match.
Mutants: four rule mutants and one unknown-field options mutant build and run with zero exits and empty stderr, then fail only byte comparisons on all three backends.
Not covered: independent JSX parsing for 18 upstream cases, shared TSX witnesses, global directive suppression and exact invalid-option error prose; full repository gate not run.

## Fresh validation after Ahra's continuation

The original definite-assignment claim is now ported in its own directory. Its 18 captured upstream cases join the previously claimed three rules. All 197 projected-AST cases match in 57,743 canonical bytes. The independent parser matches all 179 cases whose Go tree contains no JSX. The 18 JSX cases remain tested through a Go AST projection only.

The complete compiler and stage1 raw-source pass has 55 batches and 165 backend observations, all with zero exits, empty stderr and byte equality. Compiler pin is unchanged. No compiler or stage1 source is excluded on diagnostics. Fresh results and input/output hashes are under evidence/finish-claimed. Historical evidence below records earlier snapshots and must not override these fresh observations.

The strict description decoder matches Go acceptance and decoded values on 27 option cases, including malformed shapes, tagged-key casing, duplicate keys, nulls and enum constraints. Invalid-option error prose is not compared; the rule refuses invalid data explicitly. Removing the unknown-field shape check is a separate successful-running byte-only mutant on all three backends.

Shared registry tests PASS 0.034s. The shared owned-witness gate FAILS 6.200s before candidate execution because it writes the positive Google JSX witness as .ts and parses it with ScriptKindTS. No shared file was edited. Global directive suppression belongs to the shared lint pipeline and is absent there; the standalone comparisons invoke the unmodified Go rule API without that pipeline. The shared .a harness branch was absent at the last fetch. Sources use Ahra's explicitly authorized .ts fallback.

Latest startup-inclusive findings per second, median of three 200-positive-case samples:

| Rule | Sanitized native | Node | Go | Pipeline |
| --- | ---: | ---: | ---: | --- |
| adamic/no-definite-assignment | 2,866 | 1,460 | 12,206 | Independent raw-source parsing |
| structure/tailwind-no-physical-direction | 2,823 | 1,348 | 20,813 | Independent raw-source parsing |
| @eslint-community/eslint-comments/require-description | 3,411 | 1,647 | 18,805 | Independent raw-source parsing |
| @next/next/google-font-display | 452 | 999 | 18,128 | Adamic loads projected AST; Go parses source |

Rule mutants: definite_assignment_subscription_lost, physical_direction_exemption_lost, description_ignored and fallback_display_ignored. Twelve healthy executions produced twelve byte-only catches. The options mutant adds three more. Native includes ASan and UBSan. These timings are observations on this machine, not a comparative end-to-end speed claim for the projected pipeline.

## Historical snapshots

Built: the same three candidates now use the authorized .ts fallback; no further rules claimed and no shared files edited.
Commits: prior rule commits 25c8936f, 593b3356, 8d113a23; previous pushed head 87be831e; fallback and blocker evidence committed next.
Commands and outputs: shared registry discovers all eight rules; registry tests PASS 0.027s; TestOwnedWitnesses FAIL 5.648s when its Go oracle parses the Google font JSX witness as TypeScript.
Mutants: prior physical_direction_exemption_lost, description_ignored and fallback_display_ignored had nine byte-only catches; not rerun after the extension fallback because the shared witness gate blocks first.
Not covered: complete ports and post-rename corpus parity; shared TSX parsing, native JSX parsing, strict options and suppression remain blocked or incomplete.

## Selection and ownership

The next claim was pushed on codex/lint-wave1-11 before any candidate source was written.
The first unclaimed helper-ready entry was position 46, structure/tailwind-no-physical-direction.
All earlier entries were named in claim files on fetched origin branches. The next two
unclaimed syntax-only inventory entries were @eslint-community/eslint-comments/require-description
and @next/next/google-font-display. The inventory came from origin/codex/lint-inventory.
None of the three was ported on fetched origin/main. The previous slot's assignment and
skip evidence remain in claims/wave1-11.md and claims/wave1-11-REPORT.md.

All new Adamic source is .a. Each candidate owns its descriptor, implementation,
messages, Go adapter, mutant and witness. The tailwind directory also owns the
cross-rule validation tools. No shared production dispatch, compiler, copied-file list
or oracle was edited. The helper update was merged from origin/codex/lint-helpers.
The requested docs/parallel-work.md remains absent, including on fetched origin/main.

## What the comparison establishes

The unmodified pinned Go rules decide findings. A separate Go parser projection contains
only node kinds, source spans, children and literal text. No rule verdict is projected.
The .a rule functions run against this tree independently on source Node, the Adamic
JavaScript backend and sanitized native. This first pass tests rule semantics and both backends. The separate raw-source pass
below exercises Adamic parsing. Neither pass compares entire AST geometry.

The canonical byte stream includes the rule name, message ID, UTF-8 start/end offsets,
full message text and fixed source. All three upstream rules have no fixes or suggestions;
the Go harness rejects a newly introduced repair and the Adamic harness rejects one too.
The fixed source is therefore the original source, compared byte for byte. The stream
is an owned representation, not Go cohere's human-facing report renderer.

The Go fixture harness was overlaid in scratch to record each Run result. Original
upstream test assertions still ran. Captured unique source/rule/options combinations:
54 tailwind, 109 require-description, 16 google-font-display. Capture includes clean
and reporting cases. The Go pin, exact input hashes, output digest and corpus census
are in evidence/input-manifest.json. Generated registry files are excluded from the
stage1 source census; gaps are included. No diagnostic-based exclusions are applied.

## Remaining integration work

The production registry hardcodes rule.ts, generates rule.ts imports and rejects .a
mutants. Its ordinary tests and cloud/setup.sh warm gate fail when these candidate
directories are present. registration-compatibility.patch is a reviewable, scratch-tested
proposal confined to this owned directory. It was not applied to production files.
The isolated overlay discovers all eight descriptors, including the three candidates.
The shared lint harness also needs .a source copying and corpus collection before its
normal integration tests can run these directories.

The current stage1 parser does not have a validated JSX adapter. The owned driver
explicitly refuses .tsx in independent-parser mode. Google font behavior is consequently
tested through Go's JSX AST, not through native source-to-AST parsing. The filename-gated initial independent probe covered only one .js case. A broader
scratch probe removes that filename gate only for fixtures whose Go tree contains no
JSX nodes: 53 tailwind and 108 require-description cases, all 161 matching Go on every
backend. It does not pass projected nodes to the Adamic parser. The 18 remaining
upstream cases contain JSX and remain excluded from independent parsing.

Require-description currently receives valid decoded option data through Settings.
Unknown fields and malformed options do not yet have an equivalent owned Adamic decoder;
the Go adapter uses the strict real upstream decoder. Production suppression behavior,
including the directive-subject exemption, is outside this standalone harness. These
are candidate implementations, not completed ports meeting the full requested bar.

## Reproduce

Source the toolchain environment printed by cloud/setup.sh, then run from the repo root:

```
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py --scratch /tmp/wave11-next > /tmp/wave11-next-run.log 2>&1
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/post_validate.py --scratch /tmp/wave11-next --compiler /tmp/wave11-typescript > /tmp/wave11-next-post.log 2>&1
```

The compiler checkout must be microsoft/TypeScript commit
050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3). The runner checks this pin.
All subprocess stdout and stderr go to files. Post-validation requires a zero exit,
empty stderr and byte equality for each backend. Mutants must build and run successfully
with empty stderr and fail only the byte comparison.

Findings throughput uses three startup-inclusive repetitions of 200 copies of a positive
upstream case per rule. Go includes source parsing; Adamic loads the projected AST.
Native is sanitized. These timings are observations of different input pipelines and
cannot establish end-to-end native speed relative to Go. Measurements are recorded in evidence/summary.json. Additional source-to-findings
timings for the two non-JSX rules are in evidence/independent-benchmarks.json.

## Toolchain and scoped gate

The follow-up setup printed Go ready (0s), clang ready (0s), Node ready (0s), submodules
ready (0s), then failed its warm test command because rule.ts was missing in the new
require-description directory. It printed no successful warm/done timing. The existing
installed environment worked; it was sourced at /workspace/adamic-tools/env.sh.
Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc is 5, cgroup quota is 4 processors.
The original successful setup and timing lines are preserved in claims/wave1-11-REPORT.md.

Raw setup and scoped gate outputs are retained under evidence. The production registry
failure and overlay result are separate. The complete repository gate and repository-wide
vet were not run. No pull request was opened.

## Completed corpus comparisons

| Corpus | Inputs | Source bytes | Rule/source combinations | Projected AST | Adamic parses raw source |
| --- | ---: | ---: | ---: | --- | --- |
| TypeScript src/compiler, pinned v6.0.3 | 77 | 9,400,075 | 231 | All three backends match Go | All three backends match Go |
| stage1 .ts and .a, including gaps | 136 | 752,660 | 408 | All three backends match Go | All three backends match Go |
| Captured upstream fixtures | 179 | See input artifacts | 179 | All three backends match Go | 161 non-JSX cases match on all three; 18 JSX cases excluded |

The full raw-source corpus pass consists of 54 batches and 162 backend comparisons,
all successful with empty stderr. No unsupported input was excluded from either corpus.
The stage1 sources were fixed at the hashes in evidence/input-manifest.json.
Upstream findings are 33 tailwind, 55 require-description and 7 google-font-display.
Canonical upstream output is 47,098 bytes with SHA-256
f82aa3df68d0ce4b9c851c2d584139b0458c3b279314637c9fa294280882192a.
Per-batch byte counts, findings counts, hashes and comparison assertions are in
evidence/corpus-output-digests.json. Raw corpus logs remain in /tmp/wave11-next;
compact hashes are committed rather than hundreds of megabytes of identical output.

Run the additional independent checks after validate.py and post_validate.py:

```
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/independent_validate.py --scratch /tmp/wave11-next > /tmp/wave11-independent-run.log 2>&1
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/corpus_independent.py --scratch /tmp/wave11-next > /tmp/wave11-independent-full-run.log 2>&1
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/benchmark_independent.py --scratch /tmp/wave11-next > /tmp/wave11-own-benchmark-run.log 2>&1
```

## Findings per second

Median of three repetitions, 200 positive cases and 200 findings per repetition.
These include process startup and use sanitized native. Measurements were made on this
cloud machine and are not a speed guarantee.

| Rule | Native findings/s | Node findings/s | Go findings/s | Input pipeline |
| --- | ---: | ---: | ---: | --- |
| structure/tailwind-no-physical-direction | 2,211 | 1,482 | 20,288 | All parse raw source independently |
| @eslint-community/eslint-comments/require-description | 2,216 | 1,252 | 17,585 | All parse raw source independently |
| @next/next/google-font-display | 499 | 1,415 | 19,089 | Adamic loads Go-projected JSX AST; Go parses source |

The full projected-AST timing series for all three rules is also retained in
evidence/summary.json, including 903 native findings/s for tailwind and 1,565 for
require-description. Those timings exercise a different input pipeline from the
independent-parser table. Emitted JavaScript parity was checked but its throughput
was not profiled.

## Scoped gate results

| Command | Result |
| --- | --- |
| Final adamic types validation.ts | PASS, empty output |
| Owned Go utility vet | PASS, empty output |
| Go upstream TestNoPhysicalDirection, TestRequireDescription, TestGoogleFontDisplay | PASS, 54/109/16 captured unique cases |
| Plain registry and comment-helper package tests | Registry FAIL because rule.ts is missing; comment helpers PASS |
| Registry and comment-helper package tests with scratch registry overlay | PASS, registry 0.067s, comment helpers 212.772s |
| Formatted overlay registry tests, verbose | PASS, including structural rejection controls |
| ADAMIC_GATE_UNCACHED=1 external TestTheOracleCatchesOneByte | PASS, 0.714s |
| git apply --check --unidiff-zero registration-compatibility.patch | PASS; no production file changed |

The patch uses zero-context hunks so the committed patch text does not introduce
whitespace warnings from diff context markers. Apply it with --unidiff-zero if the
shared owner accepts it. It addresses registry module discovery and .a mutants; it
does not supply the shared lint harness's .a copying or JSX parsing.

All three semantic mutants have nine successful executions and nine byte-comparison
catches. Build logs and first differing bytes with output hashes are preserved in
evidence/mutants.json and the named build logs. Compiler refusal, sanitizer errors and
Go test failures earned no mutant credit.

## Ahra correction and current stopping point

Ahra authorized .ts rule modules until the shared .a codemod lands and required workers
to stop at any other blocker without editing shared files. Only files inside these
three owned rule directories changed in this follow-up. Eight Adamic modules were
renamed from .a to .ts, their imports and mutant paths were updated, and the owned
validation scripts now use validation.ts. The historical evidence above is from the
pre-rename source snapshot; it must not be treated as a fresh post-rename corpus run.
Renaming source files changes the source filenames seen by filename-sensitive rules.

Fetching refs/heads/codex/lint-harness-dot-a failed with:

```
fatal: couldn't find remote ref refs/heads/codex/lint-harness-dot-a
```

The unmodified shared registration generator now discovers all eight rules and its
ordinary package tests pass without an overlay. The three candidate descriptors are
readable. The former .a proposal is retained only as historical material; it is not
needed by this .ts fallback and was never applied to shared production files.

The next shared gate stops before running the candidate on Node or native:

```
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout 10m
panic: invalid corpus .../next-google-font-display-0.ts: ['>' expected. ',' expected. Variable declaration expected. Expression expected. Expression expected.]
FAIL github.com/system-inc/adamic/stage1/cohere/lint 5.648s
```

Exact cause: registration_test.go's ownedWitnesses writes every raw witness to a
filename ending in .ts. testdata/oracle.go line 89 always calls ParseSourceFile with
core.ScriptKindTS, then rejects any parse diagnostics. GoogleFontDisplay requires a
JSX link element, so its positive fixture cannot pass that oracle input contract.
This needs a shared TSX witness/parser contract; changing the witness to a string
would remove the finding and defeat its purpose. Neither shared file was edited.

Independent native JSX parsing remains unvalidated for the 18 upstream JSX cases.
Require-description's strict Adamic option decoder and production suppression
integration also remain incomplete. The three candidates are not declared finished.
Work stopped at the shared witness blocker, as instructed, and no additional claim
was made. Correction logs are committed under evidence/ahra-correction-*.log.
