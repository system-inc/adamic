Built: added rule.json name/numeric kinds manifests for all nine existing owned rules under listeners/<slug>; no new claims, runtime rule changes or shared-file edits.
Commits: parent b3d94d10 remains based on current main e8ba3d5d; publication targets only codex/typeaware-wave-13, never main or area/.
Commands and outputs: fetched explicit main/own-branch refs; go test ./stage1/cohere/typeaware/wave_13_more -count=1 -v PASS 0.036s; touched-package vet and diff check pass; one positive constructor control is 603 identical bytes in Go, native and ASan.
Mutants: nine wrong-kind JSON mutations are rejected by the independent production-listener comparison; nine existing source-declaration mutations are also rejected. Prior compiling semantic rule mutants and corpus agreement remain recorded on the unchanged runtime source.
Not covered: automatic descriptor loading, numeric ParseNode/supplied-node dispatch, three shared-parser inputs, full Go gate and emitted-JavaScript parity. No whole-branch green or speed improvement is claimed.

The existing class listenerKinds declarations remain unchanged. Each manifest
contains only name and kinds, with numeric IDs matching the pinned Go parser
and production cohere subscriptions. The JSON guard reuses the independent
production Go AST listener extraction and actual ast.Kind enum constants,
rejects string-valued kinds, wrong names, missing/duplicate/wrong kinds,
unknown fields and trailing data. A wrong-kind mutation in each manifest is
caught by that comparison. These are metadata mutations, not new compiled
semantic variants. The earlier nine semantic native mutants are unchanged.

Current main has no type-aware rule.json discovery/loader. The separate syntax
registration branch has a descriptor contract requiring factories, visits and
oracle adapters against a syntax-only RuleContext; it is not this bridge-backed
type-aware suite. These manifests declare the requested metadata for integration,
but are not advertised as complete registrations in that separate driver.
No fictional factory/visit hook or shared generator edit was added.

The shared ParseNode still exposes kind: string and the existing suite hands a
file/context to run(). The numeric node and supplied-node callback interface
have not landed. Legacy string relevance checks/refetches remain in the old
path, as recorded in SPEED_LISTENERS_REPORT.md. Ahra prohibits shared parser,
harness and registration edits, so this unit cannot complete that migration.
The latest message does not name a batch-8 Diagnostic SHA; no diagnostic or
finding-model rebase was attempted. Main is still e8ba3d5d.

Replayed the positive Object() case 0741b4d4821fe4db against the unchanged final
native binaries: all three modes exit 0, emit 603 equal finding/suggestion bytes
and have empty native stderr. Replayed the three blocked cases as well:

| Case | Go bytes / exit | Native and ASan bytes / exit | Refusal |
| --- | --- | --- | --- |
| 53d1c0a9ffc2fefa | 699 / 0 | 0 / 70 | JSX .ts: GreaterThanToken/SlashToken at 5 |
| aff6ced2ca61fe89 | 699 / 0 | 0 / 70 | JSX .ts: GreaterThanToken/Identifier at 7 |
| defcb4c4ce6a2921 | 611 / 0 | 0 / 70 | yield label: expected semicolon at 37 |

No sanitizer report appears. Full native gates were not repeated for JSON-only
metadata ignored by the current driver: runtime .a files, compiler, bridge,
submodules and current main are unchanged from the prior supported-input gate.
That gate is original PASS 242.664s, Nexus 129/129, core 430/433, normal/ASan
agreement on both frozen corpora and all nine compiling semantic mutants caught
by bytes. The claim remains partial; no additional rules were selected.

Previous isolated compiler medians remain 6.186187s native / 1.915524s Go for
original, 2.969367s / 0.441660s for Nexus and 3.382823s / 0.462135s for core.
No new timing claim is made. Toolchain reused; prior setup reports Go, clang,
Node and submodules 0s, cache warm/total 200s, nproc 5. All process output went
directly to log files. validation/json-listeners retains the exact control
streams, refusal messages, test/metadata-mutant logs, fetch log and hashes.
