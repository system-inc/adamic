Built: replaced BooleanPropNaming's handwritten default matcher with the exact Go default as one JS RegExp literal.
Commits: based on pushed wave-17 tip 034aeba7 and current main f8013f0b; no new claims or shared-file edits.
Commands and outputs: fifth-batch Go byte oracle PASS 161.775s; 401 matcher names / 2346 identical bytes; sanitizer stderr empty; vet and listener checks PASS; setup 25s, nproc 5.
Mutants: syntax, memo and boolean rule mutants plus released-handle registry mutants caught again; regex literal [A-Z] to [a-z] caught at byte 18, with exit zero and clean sanitizer stderr.
Not covered: configurable patterns, production JSX, full component/props integration, numeric node-handler conversion and full repository gate; all remain explicit.

The existing matcher encoded the default's prefix test by hand. The new
requirement forbids handwritten regex matchers, so the rule now initializes
/^(is|has)[A-Z]([A-Za-z0-9]?)+/u once at module scope and calls test on it.
The Go default is DefaultBooleanPropNamingRulePattern in the unchanged pinned
production rule. origin/codex/lint-regex was not present in the fetched refs, so
there is no shared table row to use yet. No translated pattern is silently
invented for configured options. Configurable matching remains parked because
new RegExp(pattern, 'u') with a runtime pattern is not yet lowered by stage 0.
The default matcher contains no string-kind dispatch or refetch changes.

The full existing fifth-batch test compares findings, fixes and suggestions
against production Go on 155 upstream/edge controls, 44 ordinary-parser controls,
option/message variants and both frozen corpora, normal and ASan/UBSan. All
comparisons pass, including the existing comparison-only mutants and stale-handle
controls. JSX native decisions use independent raw AST projection, and the
production parser refusal remains explicit. This does not complete the parked
BooleanPropNaming rule. Only its already implemented default-pattern scope
changes here.

matcher_probe.a imports the actual shipped matcher and checks 401 values:
ASCII boundaries under each prefix, wrong prefixes, empty names, Unicode,
newlines and prefix-only matches. matcher_oracle.go uses Go regexp independently.
The sanitized native output is 2346 identical bytes with empty stderr. A real
mutation of the shipped matcher literal from [A-Z] to [a-z] compiles and exits
zero under sanitizers but differs at byte 18. The extracted mutated function and
probe are preserved; a compile failure is not counted as a catch. The initial
probe passed a boolean to Adamic's string-only console.log and was refused;
converting the result to the strings true/false fixed the probe before the
comparisons above. No rule implementation repair was needed for that refusal.

After tests and setup finished, three alternating quiet whole-process runs
compared complete diagnostic hashes each round:

| Corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 2.409440 | 0.362830 | 6.64 |
| repository | 0.335271 | 0.153053 | 2.19 |

Native remains slower. These are whole fifth-batch ordinary-parser timings,
not isolated RegExp or projected-JSX timings. Individual rounds and phases are
preserved in timing/measurements.json.

Commands used, with test output sent to log files:

```
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE17_FIFTH_ARTIFACTS=/workspace/wave-17-regex-default ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript TMPDIR=/workspace/wave-17-artifacts go test ./stage1/cohere/typeaware -run '^TestWave17FifthAgreementAndMutants$' -count=1 -v -timeout=30m
bash cloud/setup.sh
go vet ./...
python3 stage1/cohere/typeaware/wave_17_listeners/verify.py
go run stage1/cohere/typeaware/validation-wave-17-regex/matcher_oracle.go
/workspace/wave-17-regex-default/adamic build stage1/cohere/typeaware/validation-wave-17-regex/matcher_probe.a -o /workspace/wave-17-regex-default/matcher-native --tsgo /workspace/wave-17-regex-default/checker-asan.a --sanitize
/workspace/wave-17-regex-default/adamic build /workspace/wave-17-regex-default/matcher_mutant_probe.a -o /workspace/wave-17-regex-default/matcher-mutant --sanitize
python3 stage1/cohere/typeaware/validation-wave-17-landing/measure.py /workspace/wave-17-regex-default /workspace/wave-17-regex-default/timing /workspace/wave-17-typescript
```

No new three-rule claim is made by this compatibility update. Existing parked
claims retain their status. The analysis-dependent require-atomic-updates is not
selected, and the next analysis-independent batch remains outstanding.
