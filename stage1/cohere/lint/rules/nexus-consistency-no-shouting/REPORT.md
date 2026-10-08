# Nexus consistency no shouting

Ported the ledger-assigned batch 3 rule into its own directory. The descriptor
subscribes to SourceFile, takes the supplied ParseNode, and has node: true.
All Adamic modules are .a. No shared context, finding, driver, generator,
oracle or comparison file was edited.

The branch started from area/stage1-lint at d3a37422c and was rebased onto the
user-supplied 4f18a05c9 harness fix before certification. Upstream is the real
cohere submodule at 715ba94f3608a6500086b1076ce5cb7e51b836db, not the batch
implementation. The ledger names origin/codex/stage1-lint-batch3:stage1/cohere/lint/rules/no_shouting.ts.
The batch 3 comment anchoring and literal exclusion helpers
are local because this new directory is their only added consumer. The static
acronym, currency and two-letter policy lists come from the pinned Go source.

The Go adapter decodes manifest field 5 into ConsistencyNoShoutingOptions,
including allow. The description is copied verbatim from upstream policy.
The first four distinct shouting tokens remain in their observed order.
Go regexes use JS RegExp literals from codex/lint-regex's translated table;
masking uses the real regex engine and matches Go's blanking order. Source
ranges are converted by the shared reporting model. There are no automatic
fixes or suggestions for this upstream rule.

upstreamTest is TestConsistencyNoShouting. The actual eight upstream functions
are Fires, StaysSilent, KnowsTheAcronymsTheReviewFound,
MasksASingleQuotedCapitalPhrase, MasksAnIndentedCodeBlock,
MasksADoubleQuoteThatWraps, RespectsTheAllowOption and NamesTheTokens with that
prefix. All eight are captured, including options; there is no narrowed prefix.

The banner witness reports five distinct shouting tokens; the masking witness
keeps code and string contents quiet while ALWAYS reports. The allow witness
uses a sidecar to suppress WIDGET while five other tokens still report.
The mutant shouting_fourth_token_omitted changes tokens.slice(0, 4) to
tokens.slice(0, 3) in messages.a. It must compile and exit successfully;
only the message-byte comparison against independent Go can reject it.

## Observed checks

- bash cloud/setup.sh initially failed while warming Go's build cache with
  no space left on device. Only obsolete reproducible Go cache entries were
  removed; no source, required input or assertion was removed. The retry passed:
  Go 0s, clang 1s, Node 1s, submodules 2s, cache warm 203s, total 203s.
  nproc is 5; cgroup cpu.max is 400000 100000; memory is 17.6 GB.
- go run ./cmd/lint-registry passed on the final base.
- go test ./stage1/cohere/lint -run '^Test(OwnedWitnesses|RulesAgree)$'
  -count=1 -timeout 30m -v passed. TestRulesAgree: 117.58s and 13,074,983
  identical bytes across Go, source Node, emitted JavaScript and ASan/UBSan
  native. TestOwnedWitnesses: 21.59s and 115,657 identical bytes.
- After adding the explicit allow sidecar, the unfiltered TestOwnedWitnesses
  rerun passed: 21.25s and 116,849 identical bytes on the same four backends.
- go test ./stage1/cohere/lint -run '^TestMutants$' -count=1 -timeout 30m -v
  passed in 800.377s. All 41 registered mutants were caught. The owned
  shouting_fourth_token_omitted mutant compiled, exited zero with clean stderr,
  and was caught on Node, emitted JavaScript and sanitized native by the
  missing fourth quoted token in the banner message. No compiler failure
  or runtime refusal counts as the kill.
- gofmt -l of the owned oracle returned no files. go vet ./... passed with
  empty output. Logs for all checks are committed under evidence/.

The pre-fix comparison and mutant runs were interrupted on user steering and
are not certified. The successful pre-fix sanitized manual build is historical
only; the final unified gates rebuild both output backends. Existing harness
malformed-syntax recovery refusals are reported explicitly by TestRulesAgree;
no new skip, weakened guard or exclusion was introduced for this rule.
No full repository gate, new throughput benchmark or live cohere formatting
run is claimed. The broader required stage 1 input checks were not invoked by
the focused lint gates.
