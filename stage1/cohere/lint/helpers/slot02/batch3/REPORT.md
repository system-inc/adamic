Built: react.IsEs5ComponentCall, react.isCreateClassName and collapse.isMathFunctionName in separate .a files.
Commits: claims 972f103 and c5f5f9e pushed before code; implementation d556337; the prior six helpers were already tested and pushed through 1c5e728.
Commands and outputs: setup 26s, nproc 5; isolated replacement batch PASS 24.828s; full helper package PASS 188.649s; filtered uncached oracle PASS 4.497s; go vet ./... PASS.
Mutants: drop createReactClass, drop bare factory calls, skip the React predicate, use math substring matching and change calc to Calc; compiled mutants are caught by the actual Go comparator.
Not covered: full repository gate, integrated rule findings/fixes/suggestions, arbitrary malformed AST graphs and the eight unavailable Tailwind live-engine/corpus gates.

# Selection and claim race

Previous helpers were finished and pushed before claiming this batch. Fetched all six origin codex/lint-helpers* branches and read every claim. All concrete helpers with more than six remaining consumers were reserved. The seven-consumer label strict option decoding and schema validation describes remaining per-rule custom configuration beyond the existing four shared option helpers, not an unclaimed Go helper symbol; those seven rules are named in the frozen ledger and the original report describes the limits.

Initially reserved StringAttributeValue and the two React helpers in 972f103 at 01:20:12 UTC. A later all-branch refresh exposed slot 01's earlier StringAttributeValue claim 6ed891f at 01:19:38. This slot withdrew StringAttributeValue even though its local comparison passed. No JSX duplicate file is delivered or counted. Slot 04's 2053731 at 01:20:22 was later than ours, and slot 04 explicitly withdrew both React helpers to this slot. The complete snapshots are in evidence/claims.log and final-claims.log.

Refetched every branch and read every claim before replacing the JSX helper. collapse.isMathFunctionName was unclaimed and tied the highest available named-helper count, six. Replacement claim c5f5f9e was pushed before implementation. Subsequent all-branch inspection found that symbol only in slot 02's claim. No additional helper is claimed.

Changes stay in slot-owned helper files, fixture data and standalone tests. No shared registration generator, shared rule harness, tracked cohere source or protected compiler file is edited. The oracles use temporary Go overlays. New Adamic files are .a; inherited options_json.ts is consumed unchanged and renamed .a only in temporary mutant copies. No PR opened.

# Observations

The capture contains 583 deduplicated rule/file/source inputs across all twelve consumers, including dynamically assembled fixture strings. The oracle deduplicates parser inputs by source and extension, canonicalizes file names to absolute /fixture plus extension and adds controls. There are 580 parsed sources, 11,627 projected nodes, 371 factory names and 1,247 math names. Nil adds one factory-call query, giving 11,628 call verdict/predicate observations.

isCreateClassName: the actual private Go helper decides every identifier/literal name and the boundary controls. Exactly createClass and createReactClass accept. Case differences, prefixes, suffixes, whitespace and embedded NUL decline. Escaped source identifiers are read through the real parser's decoded Text(), not source spelling.

IsEs5ComponentCall: the actual public Go helper is called on every parser node plus nil. A temporary overlay wraps only its isIdentifierNamed call site to record the dependency invocation count, exact receiver node identity and wanted name React, then calls the unchanged real private predicate. The projected prerequisite answers also come from that private predicate. The Adamic helper injects this externally owned dependency, preserving short-circuit order. Bare calls and React-namespaced calls to either factory accept. Nested callee parentheses are skipped; receiver parentheses are delegated to isIdentifierNamed. Computed access, unrelated namespaces, assertions, wrong member kinds and non-call nodes decline. Optional static access follows Go's parser kind. Controls include all of these shapes, missing ordinary matches, case differences, escaped identifiers and Unicode names. Valid call projections contain a callee; Go SkipParentheses panics for a missing expression, and the port explicitly panics for missing/cyclic arena edges. Malformed factory graphs and exact Go internal panic prose are not normal parity coverage.

isMathFunctionName: an oracle-only exporter calls the actual private Go helper in collapse. Go's own mathFunctions list supplies positive controls and prefix/suffix/case variants. Its 19 entries are calc, min, max, clamp, mod, rem, sin, cos, tan, asin, acos, atan, atan2, pow, sqrt, hypot, log, exp and round. Every captured consumer source and every decoded string/no-substitution-template text is probed directly and through actual Go ParseValue function-name extraction. The port preserves exact membership, distinct from hasMathFunction's wider substring probe. Table drift introduces a failing positive control rather than silently widening the claim.

Actual Go expected output is removed before the corpus reaches Adamic. Baselines and credited mutants compare byte-for-byte on Node source, emitted JavaScript and ASan/UBSan native, then against Go. Every successful run exits zero without stderr or sanitizer findings. A compilation or runtime failure is not credited as a semantic mutant. The isolated replacement comparison passed in 24.828s with four initial mutants; the final package run adds the React-predicate mutant and passes the new batch in 23.20s. The complete helper package passes in 188.649s with all sixteen semantic mutants caught.

# Capture, regeneration and external limits

Capture workflow is adapted from slot 03's earlier workflow, using temporary overlays to record runtime fixture sources before external-engine checks. Pinned cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db. The final React package capture passed in 5.892s. Tailwind returned exit 1 in 10.223s for unavailable installations and empty external corpora. These eight known live checks failed:

- TestClassOrderFixturesActuallyRan
- TestUnknownClassFixturesActuallyRan
- TestConflictFixturesActuallyRan
- TestConflictingClassesPlacementIsAccountedFor
- TestCanonicalFixturesActuallyRan
- TestCanonicalClassesPlacementIsAccountedFor
- TestUnknownClassesPlacesEveryCorpusClass
- TestClassOrderLiveMatchesTheEngineOverTheCorpus

This is not a passing Tailwind rule gate. All six math-helper consumers are captured before those checks and are tested against actual Go without needing the external engine. The script permits only the bounded known failure set, refuses unexpected failures and requires every selected rule. Both generated artifacts reproduced byte for byte on recapture. Source SHA256: c6dc573505e8d54a3b6a33d0cb32fa24847857e316fe16044ad88246bb38512c. Coverage SHA256: c0ec72dfac07f909118346dc965bc69f91f72a9f5fabde58c4ff9184082d24c0.

first.log is the superseded local JSX/React comparison before the ownership race was discovered. Its JSX comparison and three JSX mutants are not delivery evidence. second.log records the replacement comparison using the final three helper implementations. final.log is the complete final package run.

# Compiled mutants

| Helper | Mutation | First Go witness |
|---|---|---|
| isCreateClassName | omit createReactClass | line 117: name:false versus name:true |
| IsEs5ComponentCall | drop the bare Identifier factory branch | line 431: call:false:0:-1: versus call:true:0:-1: |
| IsEs5ComponentCall | skip isIdentifierNamed namespace guard | line 385: call:false:0:-1: versus call:false:1:14:React, proving the omitted dependency call |
| isMathFunctionName | use name.includes(functionName) | line 12001: math:true versus math:false |
| isMathFunctionName | change calc table entry to Calc | line 12341: math:true versus math:false |

All five compile and execute on source Node, emitted JavaScript and sanitized native, agree with one another and differ from independent Go observations. The full package also reruns all eleven earlier/inherited semantic mutants: JSON raw controls, schema overlapping oneOf, strict unknown fields, policy interpolation, JSX attribute-name kind, computed property identifier, reader cache namespace, namespaced member base, partial class fields, Unicode class whitespace and literal slice copying. Each is caught by its existing Go comparator; prior reports specify the witnesses. No production file is mutated.

# Commands, timing and handoff

All test output goes directly to log files. Source /workspace/adamic-tools/env.sh in each shell before toolchain commands, from /workspace/adamic.

```
bash cloud/setup.sh > stage1/cohere/lint/helpers/slot02/batch3/evidence/setup.log 2>&1
nproc > stage1/cohere/lint/helpers/slot02/batch3/evidence/nproc.log
python3 stage1/cohere/lint/helpers/slot02/batch3/testdata/regenerate.py > stage1/cohere/lint/helpers/slot02/batch3/evidence/regenerate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch3$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot02/batch3/evidence/second.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot02/batch3/evidence/final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot02/batch3/evidence/oracle.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot02/batch3/evidence/vet.log 2>&1
git diff --check > stage1/cohere/lint/helpers/slot02/batch3/evidence/format.log 2>&1
```

Setup succeeded: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 26s, total 26s, nproc 5. Tools are Go 1.27.1, clang 20.1.8 and Node 24.19.0. Filtered uncached oracle passed all six fixtures in 4.497s, zero result cache hits and six probe misses. Vet and source whitespace checks exit zero with empty logs. The full repository test gate was not run; the complete touched helper package and filtered external oracle form this bounded worker gate.

RULES.md names every consumer. The local readiness.json records each residual dependency after these three helpers without rewriting the shared frozen ledger or claiming rule implementation. Observation: eighteen dependency entries removed across twelve distinct rules. Inference from the ledger: zero rules lose their final listed blocker from this batch alone. Other slots' implementation dependencies are not assumed delivered here. Adapter fidelity, the other slots' helpers, integrated diagnostics/fixes/suggestions and live design systems remain integration work.
