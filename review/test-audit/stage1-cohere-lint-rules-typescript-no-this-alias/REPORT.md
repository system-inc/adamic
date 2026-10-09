u124: one live test, TestCompileProfiles, at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.
Clean baseline passed in 29.534s; nproc=5; no skips.
Verdict: untrue for the three semantic mutants tried; empty-entry probe passed.
All three survivors have changed-output native witnesses.
Evidence: test-audit/stage1-cohere-lint-rules-typescript-no-this-alias, review/test-audit/stage1-cohere-lint-rules-typescript-no-this-alias/.

```json
[
  {
    "test": "TestCompileProfiles",
    "package": "stage1/cohere/lint/rules/typescript-no-this-alias",
    "file": "stage1/cohere/lint/rules/typescript-no-this-alias/build_test.go",
    "seconds": 29.799,
    "oracle": "Self: Adamic load/lower, sanitized and release native Build, and JavaScript rendering must succeed. No product executes and no findings, exit behavior or diagnostic bytes are compared. Native survivor witnesses below are audit observations, not an external oracle invoked by this test.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "entry_probes": {
      "profile top-level driver": "P2",
      "Rule.visit (additional semantic entry)": "P1"
    },
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u124/cache/P2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/typescript-no-this-alias/ -run . > P2.log 2>&1; build_test.go:39: standalone profile: sanitized native, release native and emitted JavaScript compiled; P2 passed despite native stdout being empty"
  }
]
```

| ID | Origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/lint/rules/typescript-no-this-alias/rule.a:27 | `if(context.node(right).kind !== 'ThisKeyword')` becomes `if(context.node(right).kind === 'ThisKeyword')` | none |
| M2 | stage1/cohere/lint/rules/typescript-no-this-alias/rule.a:36 | `while(wrappers.includes(context.node(target).kind))` becomes `while(false)` | none |
| M3 | stage1/cohere/lint/rules/typescript-no-this-alias/rule.a:41 | `'thisAssignment', assignment` becomes `'thisAssignmentAudit', assignment` | none |
| P1 | stage1/cohere/lint/rules/typescript-no-this-alias/rule.a:11 | `visit(node: ParseNode, index: number): void {` becomes `visit(node: ParseNode, index: number): void { return;` | none |
| P2 | stage1/cohere/lint/rules/typescript-no-this-alias/profile.a:45 | `top-level manifest driver from const args through the end of the module` becomes `empty top-level driver (implicit void answer)` | none |

Survivors, from rebuilt native profiles over input.ts (const self = this; followed by (self) = this;):

- M1: clean native finding count 2, mutant count 0; ADAMIC_BUILD_CACHE_DIR=/tmp/u124/cache/M1 timeout 120 /tmp/u124/adamic build stage1/cohere/lint/rules/typescript-no-this-alias/profile.a -o /tmp/u124/M1-native > M1-build.log 2>&1;  timeout 120 /tmp/u124/M1-native /tmp/u124/evidence/manifest.txt > M1-output.log 2>&1. TestCompileProfiles passed.
- M2: clean native finding count 2, mutant count 1; ADAMIC_BUILD_CACHE_DIR=/tmp/u124/cache/M2 timeout 120 /tmp/u124/adamic build stage1/cohere/lint/rules/typescript-no-this-alias/profile.a -o /tmp/u124/M2-native > M2-build.log 2>&1;  timeout 120 /tmp/u124/M2-native /tmp/u124/evidence/manifest.txt > M2-output.log 2>&1. TestCompileProfiles passed.
- M3: clean native finding count 2, mutant count 2; both diagnostic IDs changed from thisAssignment to thisAssignmentAudit. Full native stdout differs despite equal counts. ADAMIC_BUILD_CACHE_DIR=/tmp/u124/cache/M3 timeout 120 /tmp/u124/adamic build stage1/cohere/lint/rules/typescript-no-this-alias/profile.a -o /tmp/u124/M3-native > M3-build.log 2>&1;  timeout 120 /tmp/u124/M3-native /tmp/u124/evidence/manifest.txt > M3-output.log 2>&1. TestCompileProfiles passed.

P1: native findings become zero while the test still passes. P2 removes the profile top-level manifest driver; its native executable prints no bytes while the test still passes. P2 is the primary executable-entry probe and establishes vacuous=true; P1 additionally probes Rule.visit.

The brief caused these ambiguities, costs and limits:

- The package has only a compilation test. The Go test does not invoke any runtime semantic function; the function inventory describes the code it compiles, not executed runtime coverage. Its native binaries are never executed, and emitted JavaScript is only written to disk. Compilation can fail for compiler/build regressions, but none of the three compilable semantic breaks was caught. The untrue verdict rests on those three mutants, not a claim that the test can never fail for any reason.
- Validation tests exist only as validation_test.go.txt. Go does not discover TestNextSupported, TestNextCorpus, TestNextOptions or TestNextThroughput. They are absent, not skipped opt-in rows. The separate Python validator and stored historical parity evidence are outside the requested live Go test scope and were not used for verdicts.
- This is an Adamic .a port. Go vet alone cannot validate a mutant in rule.a, so every standalone diff received the real sanitized/release native and JavaScript build exercised by the package test, plus a separate successful native profile build for changed-output witnesses. No oracle.go, Go cohere implementation, harness or test was edited.
- A runtime selector switch would require an extra file read on every visited AST node. With only three mutants, the brief permits separate rebuilds; each standalone mutation was applied directly and rebuilt. P1 and P2 were built separately as empty-answer probes and excluded from mutation counts.
- The package tests a whole profile. The owned functions are Rule constructor, Rule.visit, create, profile ancestry, profile visit, profile run and the top-level manifest loop. An empty executable entry is a driver with no top-level statements, because this profile has no named main function. Shared parser/scanner/context/settings dependencies compile too, but they were not mutated. Full transitive runtime coverage was not measured.
- Successful compilation is a self oracle, even though the audit separately executes native products. Counting those audit runs as an external-run oracle for the test would misstate what the test does.
- The native witness uses two aliases and default settings. M1 removes both reports, M2 loses the parenthesized target, M3 changes report IDs. This demonstrates changed behavior, not complete parity with upstream Go cohere. Option decoding, destructuring, filename variants, compound operators, Unicode ranges and large corpora were not audited independently.
- Warm tools still required npm ci in stage3/api; it reported 419ms. This test loads no other node_modules directory.
- Baseline and every whole-package mutant/probe matrix fit 90s. No narrowing, bounded result, timeout, panic, skip or unknown outcome was needed. Repository-wide uniqueness replay and other packages remain outside this audit.

Timing: warm env.sh worked; cloud setup was skipped. Baseline 29.534s. Alone samples [30.22, 29.799, 29.551], median 29.799s.

| Product | Package test (two native builds plus lowering/rendering), seconds | Separate native witness build wall seconds |
|---|---:|---:|
| clean | 29.534 | 11.220 |
| M1 | 29.580 | 11.258 |
| M2 | 29.792 | 11.333 |
| M3 | 30.055 | 11.345 |
| P1 | 29.495 | 11.175 |
| P2 | 17.363 | 8.995 |

Compiler executable build wall time 4.649s. Recorded audit command wall total 309.134s, excluding initial baseline/setup. Per-command timings and exact cache paths are in runs.json. Production and test sources were restored; only evidence is committed.

Total session time: about nine minutes.
