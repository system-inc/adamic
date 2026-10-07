SHAs: claim b89a7d8; implementation 2edf8b8; evidence is this report's commit.
Built: proven scalar data descriptors, ordered defineProperties, descriptor-aware reflection/integrity, and Map/Set integrity.
Commands/results: validity runner 58 to 139 pass, zero disagreements; all 81 new passes match Node in three backends.
Mutants: all 12 caught by Node output comparisons, with exact changes and logs below.
Uncovered: 196 library first blockers and 321 language first blockers remain explicitly refused; full gate has inherited flow discovery failures.

## Measurements and classification

| built-ins/Object | Pass | Disagreement | Refused | Not TypeScript | Crashed | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|---:|
| Before | 58 | 0 | 598 | 1679 | 0 | 1076 | 3411 |
| After | 139 | 0 | 517 | 1679 | 0 | 1076 | 3411 |

These are observations from the unmodified runner at `origin/codex/test262-ts-validity` (`af128990588aab7ee52ea3ab8527d638009c4979`), archived outside the repository. Test262 is pinned to `5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81`. The baseline had 598 valid refusals rather than the older 638. Every baseline refusal was independently checked with real TypeScript 6.0.3 using the runner's source, strictness, lib, and prelude: 598 accepted, zero rejected. [Exact tsc evidence](exact-tsc.json) includes source hashes; [compressed exact sources](refused-sources.jsonl.gz) make the classification reproducible.

The claim was pushed before code: 359 first library blockers and 239 first language blockers, ordered defineProperty (186), seal (37), defineProperties (25), then smaller families. Initial counts identify the first blocker rather than promise a complete program will pass. Implementing descriptors exposed language dependencies including 47 accessor descriptors and 32 descriptor-aware for-in loops. The final split is 196 library and 321 language first blockers. [Remaining features and one-line language reproducers for @system_adamic](remaining.md) cover every named first blocker; [per-program final classification](after-classification.json) preserves full diagnostics.

Both original and capture runners produced identical before and after tables. Source hashes for all 517 remaining refusals match stock-tsc-accepted baseline sources. The 81 newly passing paths exactly match [the durable Node comparison corpus](new-passes.txt); no baseline passing program regressed.

## Implemented behavior and limits

Object.defineProperty handles scalar data descriptors on proven complete plain objects, including new reflective properties, omitted descriptor defaults, writable/enumerable/configurable flags, SameValue redefinition (NaN and signed zero), nonextensible targets, and catchable TypeError. Existing unboxed fields retain their checked types. Object.defineProperties proves its descriptor map and descriptors before application, sorts integer keys correctly, and preserves partial application on failure. Reflection and integrity account for descriptor flags and new keys. Numeric keys on empty proven plain objects use the existing JavaScript-compatible number formatter after argument evaluation.

Object.seal/preventExtensions and isExtensible/isSealed/isFrozen now support typed Map/Set receivers without modifying collection entries. Their empty own-property sets make nonextensible collections sealed and frozen. Weak integrity records are released at collection teardown. Object.keys/getOwnPropertyNames additionally handle proven dense arrays and nullish receiver TypeErrors. Unchanged plain-object let bindings reuse the Object slice's existing exact-shape proof for prototype reads. propertyIsEnumerable reads the enumerable descriptor instead of treating every own property as enumerable.

Descriptor metadata owns retained keys and scalar values but does not retain the receiver, and is destroyed for heap and region objects. There is no garbage collector or new ownership cycle. JavaScript lowering uses real Object.defineProperty and matching TypeErrors.

Remaining library blockers are explicit refusals, led by defineProperty (46), seal (35), constructors (15), getOwnPropertyDescriptor (11), and defineProperties (10). Array descriptor mutation needs descriptor-aware element presence and length; accessor descriptors need getter/setter execution. Host objects, boxed receivers, dynamic prototypes, reference-valued descriptors, unproven descriptor maps, and static-shape consumers after mutation remain outside the proven support. Object.values/entries/assign, JSON serialization, object spread, and for-in after descriptor mutation refuse rather than silently omit or expose keys. This unit does not implement the language dependencies in the handoff.

The algorithm port is the data-descriptor portion of V8 13.6.233 `src/objects/js-objects.cc`, `ValidateAndApplyPropertyDescriptor`, named in THIRD_PARTY_NOTICES.md. Shared hooks: `internal/lower/prototype.go` reuses shape proof and dispatches enumerable reads; `internal/javascript/javascript.go` adds Object helper dispatch; runtime `adamic.h` includes the descriptor header; `class_inheritance.c` destroys descriptor metadata; `heap.c` forgets collection integrity records. No runner, core lowering/emission, Map runtime, or Map/Set lowering files were edited.

## Verification

Linux is the gate of record. `bash cloud/setup.sh` succeeded: go 0s, clang 0s, node 0s, submodules 0s, build cache warm 57s, total 57s, 5 processors; `nproc` returned 5. Go 1.27.1, clang 20.1.8, Node 24.19.0. Stock tsc initially was missing from PATH; installing TypeScript 6.0.3 into an isolated scratch prefix resolved it. See [setup output](setup.txt).

All commands sourced `/workspace/adamic-tools/env.sh`; validity commands additionally prepended `/workspace/scratch/object3-typescript/node_modules/.bin` to PATH. Test outputs were written directly to log files.

- From `/workspace/scratch/object3-runner`: `go run ./cmd/adamic-test262 -test262 /workspace/scratch/test262 -root /workspace/adamic -work /workspace/scratch/object3-before -adapt -json built-ins/Object`, then the same command with work directory `/workspace/scratch/object3-done`. [Before](before.json), [after](after.json), and [runner log](after.txt).
- `go test -count=1 -timeout 20m ./internal/lower ./internal/native ./internal/javascript`: passed; [package results](packages.txt).
- `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 20m ./internal/oracle -run 'Object|TestNativeAgreesWithNode/internal/oracle/testdata/library_object_'`: passed after restoring every mutant. Every new pass compared source Node with sanitized native, release -O2 and JavaScript; successful exits checked for leaks. [Final oracle log](oracle-restored.txt).
- Oracle count refresh passed; gofmt, go vet for affected packages, and `git diff --check` passed.
- `go test -count=1 -timeout 30m ./...` failed in internal/flow: its positive-fixture glob includes six intentional refusal probes inherited from the prototype branch. An early run also found two new refusal probes, which were moved into this unit's docs before shipping. The later unmodified full run has only the six inherited discovery failures. [Later full gate](full-gate.txt).
- `go test -overlay /workspace/scratch/object3-flow/overlay.json -count=1 -timeout 30m ./internal/flow` passed all positive-fixture SSA, trace, liveness and range checks. The scratch-only overlay excludes exactly the six inherited negative probes: try_assign, replaced, public_hash, iterator_view, groups_proto, groups_proto2. Repository flow files were not changed. [Positive flow result](flow.txt).

## Mutants

`python3 docs/library-object/unit3/run-mutants.py` ran all twelve on the final implementation, restoring exact original bytes after each. Each test reached a Node comparison and failed for different stdout or exit status. Compile failures, lower refusals, ASan errors and runtime diagnostics do not qualify. [Summary](mutants.txt). The script preserves each exact mutation and test filter. Individual output is recorded below.

| Mutant | Incorrect behavior | What caught it |
|---|---|---|
| [plain-object-view-proof](mutants/plain-object-view-proof.txt) | Accept all prototype receiver views | Prototype representation Node comparison |
| [numeric-key-sign](mutants/numeric-key-sign.txt) | Format abs(key) | numeric_keys negative-key Node comparison |
| [numeric-key-typed-slot-proof](mutants/numeric-key-typed-slot-proof.txt) | Permit numeric keys over unboxed typed slots | Refusal guard fallback: Node 9 versus native pointer-valued number |
| [library-typeerror-name](mutants/library-typeerror-name.txt) | Throw Error rather than TypeError | descriptor_errors Node comparison |
| [own-keys-array-boundary](mutants/own-keys-array-boundary.txt) | Omit last array index | own_keys_receivers Node comparison |
| [descriptor-same-value-zero](mutants/descriptor-same-value-zero.txt) | Treat positive and negative zero as SameValue | descriptor_same_value Node comparison |
| [descriptor-new-enumerable](mutants/descriptor-new-enumerable.txt) | Default new properties to enumerable | descriptor_order Node comparison |
| [define-properties-index-order](mutants/define-properties-index-order.txt) | Apply integer keys in descending order | define_properties partial-application Node comparison |
| [collection-extensibility](mutants/collection-extensibility.txt) | Reverse Map extensibility | collection_integrity Node comparison |
| [has-own-hidden-key](mutants/has-own-hidden-key.txt) | Ignore added descriptor keys | descriptor_errors Node comparison |
| [descriptor-consumer-proof](mutants/descriptor-consumer-proof.txt) | Accept static-shape values consumer after mutation | Refusal guard fallback: Node empty values versus native value 1 |
| [descriptor-map-proof](mutants/descriptor-map-proof.txt) | Accept mutated descriptor maps | Refusal guard fallback: Node conversion throws before defining x, native defines x |
