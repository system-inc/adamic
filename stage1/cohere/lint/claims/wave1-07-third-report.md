Built one owned .a candidate for no-this-alias; two assertion rules have concrete suggestion protocol blockers.
Claim effa075c was pushed before implementation; candidate and evidence commits are listed in the final response.
Four-way checks passed: 38 new upstream cases, 200 corpus rows, six options; final suite PASS 86.345s.
Parenthesized-alias mutant compiled and ran; Node, emitted JavaScript and sanitized native byte comparisons caught it.
Not covered: two assertion ports and mutants, default .a integration, arbitrary invalid options, full repository gate.

# Selection and scope

Pushed everything already on codex/lint-wave1-07 (up to date), then fetched all
320 origin refs. Main snapshot ef3d907ecdc4c771b016f7d9c52372def057a340 was used
for queue selection, without merging newer main into this checkout. All 46 names
in the original helper handoff are claimed. Inventory syntax-only filtering excludes
needs_type_information, binding_only and declared_checker: 287 rules.

The first remaining names were no-non-null-asserted-optional-chain,
no-non-null-assertion and no-this-alias, all under @typescript-eslint.
Selection evidence JSON enumerating a rule with claims: [] does not reserve it.
Main contains policy messages and frequency observations for no-non-null-assertion,
but no executable implementation. Claims remain reserved; no blocked rule is marked ported.

Only the owned no-this-alias rule directory and this unit's claim/evidence files
were changed. Shared infrastructure files and the cohere submodule were not edited.

# Successful candidate

The rule reproduces Go's case-sensitive TypeScript filename gate, direct `this`
initializer test, identifier-only exemptions, opt-in destructuring reporting,
assignment target unwrapping, and all compound assignment operators. Its Go adapter
returns the unmodified upstream NoThisAlias rule and accepts captured decoded options.
Suggestions/fixes are absent upstream; fixed source remains unchanged.

The owned runner reuses the prior scratch-only compatibility overlay. Default
integration is still blocked because registry discovery requires rule.ts rather than .a.
The shared proposal stays unapplied. Every successful execution exits zero without stderr;
native parity and mutant runs use ASan/UBSan and leak checks.

# Validation and raw evidence

Setup initially failed (exit 1) at cache warming: the default lint TestMain tried
opening rules/eslint-comments-require-description/rule.ts. Installed Go, clang,
Node and submodules were ready at 0s. Workaround: source /workspace/adamic-tools/env.sh,
then GOFLAGS='-overlay=/tmp/wave07-third-final2/overlay.json' bash cloud/setup.sh,
with output redirected to /tmp/wave07-third-setup-overlay.log. Retry succeeded:
Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 17s, total 17s. nproc: 5.
Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

Compiler pin: TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8,
at /tmp/wave07-typescript. Sources comprise src/compiler plus this checkout's
stage1 .ts/.a files. All test output went directly to log files.

```
ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript python3 stage1/cohere/lint/rules/typescript-no-this-alias/validate.py --scratch /tmp/wave07-third-final2 --run '^(TestNextSupported|TestNextCorpus|TestNextOptions|TestNextThroughput|TestOwnedWitnesses)$'
python3 stage1/cohere/lint/rules/typescript-no-this-alias/validate.py --scratch /tmp/wave07-third-mutant --run '^TestMutants$/^parenthesized_alias_lost$'
go test -overlay=/tmp/wave07-third-final2/overlay.json ./stage1/cohere/lint/registry -count=1 -v
go vet -overlay=/tmp/wave07-third-final2/overlay.json ./...
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=5m
```

Final suite PASS 86.345s. Owned witnesses matched 18044 bytes. Captured 417 cases:
38 for no-this-alias; 415 supported cases matched 131001 bytes across Go, source
Node, emitted JavaScript and sanitized native. Two old JSX cases remain explicitly
excluded; there are no JSX exclusions among the 38 new-rule cases. Corpus: 200
file/rule rows matched 12214489 bytes. Six decoded-option variants matched 10165 bytes.
Registry controls PASS 0.016s; vet exit 0; external one-byte oracle PASS 0.258s,
uncached (native and Node each zero hits, one miss).

Go-only full rule-family test filter, run in cohere:
`go test ./internal/lint/rules/typescript -run '^(TestNoThisAlias|TestNoNonNullAssertion|TestNoNonNullAssertedOptionalChain)' -count=1 -v`.
PASS 0.017s. This is upstream verification, not proof of either blocked Adamic port.
Raw logs live in ../rules/typescript-no-this-alias/evidence/*.txt.

# Mutant and throughput

The mutant replaces assignment-target unwrapping with while(false). It compiles
and runs but loses the second finding in `const self = this; (self) = this;`.
Independent Go byte comparison catches it on all three Adamic paths. PASS 16.045s;
compile failures or sanitizer failures receive no semantic-mutant credit.

1000 synthetic alias declarations, 1000 findings; best of five full subprocess
executions including startup. Native timing uses release mode, parity is sanitized.

| Native findings/s | Node findings/s | Go findings/s |
| ---: | ---: | ---: |
| 160329.12 | 9036.54 | 113642.50 |

# Two concrete blockers

`const r = foo?.bar!;` produces a Go finding at [10,19) with a removal suggestion
at [18,19). `const r = foo!.bar;` produces a finding at [10,14) with two suggestion
edits: removal [13,14), then replacement [14,15) by `?.`.

The current Finding data model has one replacement and no independent edit ranges.
The unchanged Go oracle insists each suggestion contain one fix with exactly the
finding range, and both fixtures panic with `unexpected suggestion shape`.
Dropping suggestions or widening them would fail the requested Go parity bar.
Shared finding/output/oracle support belongs to the integration worker under
docs/parallel-work.md. No misleading partial assertion implementations were registered.

Reproduce the independent geometry and actual unchanged-protocol rejection:
`python3 stage1/cohere/lint/claims/wave1-07-third-evidence/probe.py --scratch /tmp/suggestion-gaps`.
Both probes pass; raw outputs and shared-oracle panic traces are in that directory.
This is blocker evidence, not a mutant, native parity or throughput result for those two rules.
