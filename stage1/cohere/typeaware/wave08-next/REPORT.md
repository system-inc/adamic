# Wave 08 continuation

Built: no-uncleared-race-timeout is complete; the two process rules have partial native decision cores and pending raw fact adapters, not complete ports.
Commits: original three rules pushed in d0d6310f; continuation claim 126c0358 pushed before implementation; the continuation code and this report are committed together.
Checks: independent production Go byte agreement on 22 controls (13 findings, 8,196 bytes), compiler 77 (0, 5,780 bytes) and repository 287 (0, 18,485 bytes), normal and ASan/UBSan/LSan; isolated pending tests, vet and formatting passed.
Mutants: race rule compiled and exited zero with empty stderr, byte oracle caught difference 47; two raw-fact mutants and two native event-core mutants were caught by their separate tests; released handle refused with exit 70.
Not covered: complete process-rule integration, production Go finding agreement or production-rule mutants for those two rules; no additional rules claimed.

## Selection and claims

Fetched every origin head with:

```sh
git fetch origin '+refs/heads/*:refs/remotes/origin/*' --no-recurse-submodules \
  > /workspace/wave08-next-fetch.log 2>&1
```

Checked 325 origin refs and their 33 claim documents against all 197 entries of
the combined compiler/repository by-volume ranking. Detected 25 ranked base ports
through actual native registrations on origin/main and origin/codex/tsgo-c-library
(existing filenames often abbreviate rule names). Ninety-six ranked entries were
mentioned in claims; 76 entries remained. Equal volumes use lexical rule-name
ordering. The audit with each rule's matching registrations and claim paths is
validation/selection.json. Names in existing claims were conservatively excluded,
including released claims still named there.

The first three available entries were:

1. nexus/correctness-no-process-exit-after-output, combined volume 0.
2. nexus/correctness-no-uncleared-race-timeout, combined volume 0.
3. nexus/correctness-require-blocking-standard-streams, combined volume 0.

The claim update was committed and pushed in 126c0358 before implementation.
Work continues on codex/typeaware-wave-08; no PR is opened. All continuation
changes are inside this directory except the expressly requested claim update.
No shared harness, generator, dispatcher or protected compiler file was edited.
New Adamic sources and stored lint controls are .a. Ambient TypeScript declarations
are stored as text and materialized as .d.ts only in scratch validation.

## Completed rule

no_uncleared_race_timeout.a implements the production default rule: library
PromiseConstructor.race, library Promise construction, const promise references,
executor-local global timers, discarded handles, void/expression bodies,
unread variable/assignment handles, parentheses, shorthand references and nested
function/class exclusions. It reuses registered node-symbol-details,
binding-declarations, wave08-symbol and syntax-metadata questions. declarations.a
is an owned decoder. No new question or shared registration is needed for this
rule. The pinned standard library's declaration scripts are identified by their
compiler default-library status; non-library declaration scripts are parsed to
exclude external modules. Module-block timers must belong to declare global.

suite.a is an isolated runner using one checker program per manifest and the
existing canonical diagnostic representation. testdata/oracle.go invokes only the
production cohere race rule unchanged through the independent loader and walk.
It imports no bridge implementation. Findings, complete message text and spans,
fixes and suggestions compare; this rule emits no fixes or suggestions, and their
empty serialized payloads compare too. Sorting preserves duplicate reports.

The 22 controls include discarded/retained handles, assignment and const forms,
const versus let promise references, shorthand reads, nested callbacks, shadowed
symbols, generic Promise constructors, globalThis/window and DOM overloads,
Unicode/CRLF and repeated references producing duplicate diagnostics. They are
lint input fixtures, not claims that Adamic compiles every fixture as a program.
The two frozen populations are exactly the prior 77 TypeScript v6.0.3 compiler
files (050880ce59e30b356b686bd3144efe24f875ebc8) and 287 repository roots.
Portable manifests and complete compressed outputs are preserved in validation.
Repository/compiler source hashes were already recorded in the original wave-08
validation and the input population did not change.

| Population | Findings | Bytes | Normal | Sanitizers |
| --- | ---: | ---: | --- | --- |
| controls 22 | 13 | 8,196 | identical | identical |
| compiler 77 | 0 | 5,780 | identical | identical |
| repository 287 | 0 | 18,485 | identical | identical |

The zero-volume corpora alone cannot prove this rule fires. The positive controls
and mutant supply that evidence. The mutant reverses the lost-handle predicate,
compiles, exits 0 with empty stderr, and only the production Go byte oracle catches
it at byte 47. Mutant binaries are removed. Querying a released checker handle
using node-symbol-details panics with exit 70 and the established invalid-handle
message. The bridge archive is unchanged from the original wave-08 gate, which
also proved the retaining-registry mutant, C ABI ownership and sanitizers.

## Partial process rules and exact blockers

Neither process rule is marked complete or included in the default runner.
--process and --blocking deliberately panic before any finding output. This
prevents a zero-volume corpus from disguising an incomplete implementation.

Prepared components:

* resolved_callee.go and resolved_callee.a: raw resolved-signature declaration,
  source/module status, body bounds, function flags and return flags. Native
  callee eligibility declines overload signatures without bodies, generators,
  unawaited async functions, foreign global scripts and never-returning calls.
* program_loads.go and program_loads.a: program membership, raw source text and
  static/import-equals/dynamic-import/require literal resolutions. They leave
  runtime edge selection and blocking judgments to Adamic.
* process_flow.a: native graph-event walk, minimal sets of write try-chains,
  catch rollback, loop fixpoint, exit termination, ordered blocking termination,
  and started/handed-on function decisions.
* no_process_exit_after_output.a: native callee eligibility, event decisions and
  exact production message/report construction.
* require_blocking_standard_streams.a: reached-function traversal, entry decline,
  source ordering and the exact reason/count message variants.

First blocker: registered signature/signature-shape and call-returns questions
return types/parameters/return types, but not the actual resolved signature's
Declaration(). Picking the first binding declaration changes overloaded callees'
behavior. wave08-resolved-callee supplies the missing raw identity. The existing
bridge dispatcher does not recognize it.

Second blocker: registered module-links enumerates only top-level import/export
specifiers. The blocking rule's program index also needs import-equals, dynamic
import and require literal resolution, plus raw program source for its transitive
ordering analysis. wave08-program-loads prepares those facts. The existing
bridge dispatcher does not recognize it either.

Direct tests prove both raw extraction functions on checker nodes, including
static and dynamic imports. Calls through the current native bridge explicitly
fail with `unsupported checker question`, exit 70, before any finding output.
Registering these adapters would require edits outside this directory, which
Ahra's correction forbids: "Keep your changes inside your own rule directories"
and "If anything else blocks you, say exactly what it is and stop, rather than
editing shared files." No inferred permission or approval request was added.

Native graph construction and production integration are also unfinished. The
prepared decision cores consume preclassified graph events; they do not build
cohere-equivalent CFGs, recognize all process symbols, build the blocking index
or perform the complete writer/blocker traversal. Those remaining parts are
explicit partial-work limits, not capabilities supplied by the pending questions.
A graph-event core test is not production-rule byte agreement. Registering the
questions alone therefore does not make these two rules complete. They remain
claimed for integration, and no further rules are taken.

Separate semantic mutants actually run:

| Mutation | Observation |
| --- | --- |
| Resolved callee return flags replaced by 0 | compiled; direct checker fixture assertion fails |
| Resolved module target emptied | compiled; static/dynamic resolution assertion fails |
| Native write state requires two held chains | compiled, exit 0, empty stderr; Go event expectations differ |
| Native blocking event ignored | compiled, exit 0, empty stderr; Go event expectations differ |

The event-core control also runs under sanitizers. testdata/flow_oracle.go contains
independent Go expectations for those event cases, not the complete production
lint rule. No production-rule mutant claim is made for either partial port.
Pinned flag masks (Number 64, Never 262144, Generator 1, Async 2) are checked in Go.

## Measured time

After builds and other test jobs finished, three alternating Go/native rounds
compared complete output each time. Process wall time includes loading, parsing,
judgments, rendering and teardown. Builds and sanitizers are excluded.

| Corpus | Go median | Native median | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 0.468330s | 2.407551s | 5.14071 |
| repository | 0.169830s | 0.336913s | 1.98382 |

Raw rounds and bridge phase counters are in validation/results.json. These are
observations for the completed race rule, not timing for all three claimed rules.
Setup and toolchain are unchanged: setup passed in 132s; nproc 5, four-core CPU
quota; Go 1.27.1, clang 20.1.8 and Node 24.19.0. Original setup timing lines are
in ../validation-wave08/setup.log. No submodule pins were changed.

## Commands and outputs

All test output went directly to log files, without pipes. Source the installed
environment before these commands. The existing archives were built and fully
ownership-tested for the original three rules; no linked Go code changed here.
They can be rebuilt independently with go build -buildmode=c-archive.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave08-next/validate.py \
  --artifacts /workspace/wave08-next-final \
  --stage0 /workspace/wave08-validation/adamic \
  --archive /workspace/wave08-validation/checker.a \
  --asan-archive /workspace/wave08-validation/checker-asan.a \
  --compiler-root /workspace/typescript-wave08-corpus \
  > /workspace/wave08-next-final-validation.log 2>&1
# PASS: complete race-rule bytes, sanitizers, byte mutant, timings and refusals.
python3 stage1/cohere/typeaware/wave08-next/validate_pending.py \
  --artifacts /workspace/wave08-next-extra \
  --stage0 /workspace/wave08-validation/adamic \
  > /workspace/wave08-next-extra.log 2>&1
# Same procedure as the executed scratch script; PASS, two fact and two core mutants.
go test ./stage1/cohere/typeaware/wave08-next -count=1 -v \
  > /workspace/wave08-next-pending-tests.log 2>&1
# PASS 0.077s, including pinned native masks.
go vet ./... > /workspace/wave08-next-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave08-next \
  > /workspace/wave08-next-gofmt.log
# Both empty; production cohere formatter via the .a wrapper is idempotent too.
```

The portable validate_pending.py argument parser differs from the executed
scratch script only in paths/options. Both procedures are saved in source and
results are preserved. The full repository gate and original 26-rule suites
were not rerun; the original wave-08 report records the complete bridge gate and
filtered Node oracle. This continuation adds one complete rule and two partial
ports. No claim of three completed new rules is made.
