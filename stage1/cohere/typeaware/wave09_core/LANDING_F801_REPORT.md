Built: rebased wave-09 onto current origin/main f8013f0b and re-greened all owned rule and component oracles.
Commits: pre-rebase pushed tip 3b4a0aa1e; rebased implementation tip 1c10a7478; evidence commit follows on codex/typeaware-wave-09 only.
Commands/output: original wave suite, checker, bridge, filtered Node oracle, targeted vet and all fourteen component verifiers passed.
Mutants: all existing rule/component semantic mutants, released-handle checks and bridge sanitizer mutants reran successfully; logs retain each catcher.
Not covered: full regex compiler, full constructor tracking, numeric handed-node execution, all upstream options or the full repository gate; no new claims.

The latest landing request made this rebase the unit of work. Fetch found
origin/main f8013f0baac41ddc340d76f83bddde38536a8f07. Rebase completed
without conflicts. A final fetch after all tests confirmed the same main SHA.
Only codex/typeaware-wave-09 is published, with an explicit lease on its old
remote tip 3b4a0aa1e8934204378c3e7cb80e9ba30d6b851a. No main or area
branch is pushed. There were no implementation changes or new claims.

The compiler was rebuilt with go build -o /workspace/wave-09-core/adamic
./cmd/adamic before component verification. The original suite used the same
77-file compiler and 287-file frozen repository manifests as prior landing,
with ADAMIC_WAVE09_ARTIFACTS=/workspace/wave-09-f801-original and the
ADAMIC_WAVE09_REPOSITORY_MANIFEST, ADAMIC_WAVE09_COMPILER_MANIFEST and
ADAMIC_TYPESCRIPT_SOURCE paths recorded in LANDING_REPORT.md.

Commands, each with stdout/stderr redirected to /workspace/wave-09-f801-*.log:

```
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v
go test ./bridge/tsgo/checker ./bridge/tsgo -count=1 -timeout=15m -v
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
```

Every Python verifier below was run via python3 under wave09_core, sequentially
after rebuilding the compiler. Tests write log files and were not piped.

* label_var/verify.py
* verify_helpers.py
* misleading_character_class/verify_patterns.py
* misleading_character_class/verify_literal_tokens.py
* misleading_character_class/verify_offsets.py
* invalid_regexp/verify_quotes.py
* invalid_regexp/verify_format.py
* invalid_regexp/verify_frontend.py
* misleading_character_class/verify_literals.py
* misleading_character_class/verify_constructors.py
* misleading_character_class/verify_constants.py
* misleading_character_class/verify_bindings.py
* listeners/verify.py
* invalid_regexp/verify_case.py

The original suite passed in 106.183 seconds, checker in 0.171 seconds,
bridge in 78.342 seconds and filtered Node oracle in 4.685 seconds.
Targeted vet passed with empty output. Original controls match 44 findings;
compiler matches 14 and repository 4, including normal and sanitized output.
Label controls match 18 findings; both corpora match zero, with the same one
explicit parser-rejected control. The literal slice includes full suggestions
and fixes in native, sanitized, source Node and emitted JavaScript runs.
Unicode canonicalization matches all 2,228,228 answers / 15,603,280 bytes.

Semantic mutants rerun: prefer-const eligibility (byte 52), radix judgment
(9897), type-parameter flag (12746), assertion fix end (12158), label scope
membership and value-mask, flags precedence, emoji-modifier bounds, surrogate
folding and malformed-pattern guard, rune quoting, error-prefix trimming,
literal suggestion range, cooked/raw offsets, constructor span, constant
coercion, scoped-write eligibility (240), alias-chain eligibility (883),
listener kind (166), ASCII canonicalization guard (2180), fold representative
(379). These build and run; the independent output comparisons catch them.
Released-registry and checker-retention mutations fail the required panic-70
expectation. Bridge memory/length/ownership mutations are checked by its
rerun sanitizer suite. Exact outputs and catchers appear in retained logs.

Original suite whole-process native/Go times: compiler 2.725083/0.590128
seconds; repository 0.454586/0.160121 seconds. Label medians:
compiler 1.568374/0.327472 seconds; repository 0.244421/0.160093 seconds.
These runs overlapped other validation; no isolated throughput claim is made.
The existing successful setup is reused: 88 seconds, nproc 5.

Shared integration remains blocked: ParseNode.kind is string-only and the
shared Rules.ask refetches by index. There is no numeric handed-node listener
contract on current main. The six numeric rule.json declarations are verified,
but legacy execution is not claimed compliant with the requested speed rule.
No batch-8 Diagnostic SHA was supplied. Exact native regex compilation and
constructor/reference tracking remain unfinished independent of that gap.
Shared harness, generator, parser and compiler sources were left untouched.
Fresh test logs live in validation-landing-f801. Previous reports retain the
full corpus streams and component details; this landing does not upgrade
partial components to complete rule parity.
