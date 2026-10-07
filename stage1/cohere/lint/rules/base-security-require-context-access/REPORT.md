Built: context-access listener, alias resolution, soft-use checks, policy messages and strict options; the two Tailwind claims have portable components under pending/.
Commits: context-access 42d6fa47; canonical components 3fed4da1; ordering components 2cc9128a; claim d316a491; earlier claimed work 2ec28ca6.
Commands and outputs: 54 Go fixture cases match in 15,596 bytes on Node, emitted JavaScript and sanitized native; full configured compiler/stage1 source comparisons pass. Registry PASS 0.054s.
Mutants: context_access_alias_lost and an unknown-field decoder mutant each build and run with zero exits and empty stderr, then fail byte comparison on all three backends.
Not covered: native parsing of positive decorated parameters, configured shared witnesses, exact invalid-option error prose, complete Tailwind rule entry points and the full repository gate.

## Context-access implementation

The listener runs once per SourceFile, gathers named-import aliases before checking any member, then visits methods, constructors and accessors. It preserves the last requirement for a repeated key, collapses repeated injections, accepts method or class access decorators, distinguishes called GraphQlFieldResolver from a bare decorator, and checks missing-value types syntactically. A constructor's finding covers the constructor; other findings cover the name. No fix or suggestion is proposed, matching Go.

The exact policy entry and strict target descriptor come from the pinned shared helper catalog. Strict decoding and normalized requirement values match Go on 26 cases, including tagged-key casing, unknown fields, null string elements, duplicate keys and required fields. Invalid-option error prose is not compared. The Go adapter invokes the real upstream decoder. An unselected factory does not decode another rule's options.

All 54 unique upstream source/options cases were captured while the original assertions ran. The Go AST projection supplies kinds, spans, child indexes and literal text, never findings. Source Node, Adamic-emitted JavaScript and ASan/UBSan native independently run the rule. Their canonical diagnostic IDs, UTF-8 spans, messages and unchanged source match Go in 15,596 bytes. This canonical format is owned, not cohere's report renderer.

The complete compiler and stage1 snapshots match Go both through projected ASTs and through independent raw-source parsing. The first raw-source run used no requirements and consequently had no enabled rule; it is not credited as a configured semantic check. The final raw-source run uses the live protected-key requirements on all 77 compiler files (9,400,075 bytes) and 149 stage1 files (783,146 bytes), with 174 healthy backend comparisons. Its census, hashes, byte counts and healthy comparison results are in evidence/configured-corpora.json. All inputs are included; generated registries alone are excluded. All real-corpus findings are zero, so the positive semantic controls are the upstream fixtures and alias mutant.

## Shared gaps

Positive raw-source validation refuses on all three backends with:

```
adamic: panic: parser slice unsupported primary AtToken at 10 in /repository/source/SecurityRequireContextAccess.ts
```

The shared parser cannot read decorated parameters. The 54 projected cases therefore do not establish native source-to-AST parity for this rule. No parser or compiler file was changed.

TestOwnedWitnesses still stops at the pre-existing Google-font JSX witness: it writes a .ts filename and the Go oracle uses ScriptKindTS. This run FAILS 5.856s before candidate execution. Context-access also needs configured witness metadata: registration_test.go makes each row only `<path>\t<rule>` and has no requirements column, so this rule's default configuration correctly enforces nothing. The positive owned witness has its required settings recorded in testdata/witness.options.json for integration; the current harness does not consume that sidecar. No silent default configuration was invented in the adapter.

The published origin/codex/lint-harness-dot-a is now available at 2650ad59. It adds .a loading, emitted-JavaScript comparison, profiling and single edit-range support. Its context still has no Program/CSS design-system input, and its Go oracle still rejects multiple fixes. It was inspected, not merged or edited. This rule uses Ahra's authorized .ts fallback until shared harness integration; pending Tailwind components are .a.

## Throughput and reproduction

Fresh median of three startup-inclusive repetitions of 200 positive cases: sanitized native 297 findings/s, source Node 1,059 findings/s and Go 23,311 findings/s. Adamic loads the Go AST projection; Go parses source. These are different input pipelines and do not establish comparative end-to-end speed.

Source /workspace/adamic-tools/env.sh, then run the owned scripts with stdout and stderr redirected to logs:

```
python3 stage1/cohere/lint/rules/base-security-require-context-access/validate.py --scratch /tmp/wave11-security > /tmp/context-upstream.log 2>&1
python3 stage1/cohere/lint/rules/base-security-require-context-access/post_validate.py --scratch /tmp/wave11-security --compiler /tmp/wave11-typescript > /tmp/context-post.log 2>&1
python3 stage1/cohere/lint/rules/base-security-require-context-access/raw_corpus_validate.py --baseline /tmp/wave11-security --scratch /tmp/wave11-security-configured --compiler /tmp/wave11-typescript > /tmp/context-raw.log 2>&1
python3 stage1/cohere/lint/rules/base-security-require-context-access/options_validate.py --scratch /tmp/wave11-security-options > /tmp/context-options.log 2>&1
```

The compiler pin is 050880ce59e30b356b686bd3144efe24f875ebc8. Raw comparisons require zero exits, empty stderr and byte equality. No compiler refusal or sanitizer failure earns mutant credit. Original cloud setup succeeded in 89s with Go/clang/Node timing 0s, submodules 1s, warm cache 89s; nproc is 5, quota 4. The exact timing lines remain in claims/wave1-11-REPORT.md. The full gate is blocked by the shared positive-witness/parser contracts and was not run. No PR was opened.

## Tailwind claims

Both claims remain reserved and incomplete. Their separate component directories are under pending/ because the shared registry requires every immediate rule directory to have a complete listener descriptor. Registering these fragments as finished rules would be incorrect. The portable components and their tests are retained for integration; their reports distinguish component oracle proof from full-rule findings/fixes parity. No further rules are claimed.
