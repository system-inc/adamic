# Historical one-var unified certification

This records the earlier codex/lint-port-one-var proof. Current wave2 landing
validation and remaining wave-21 blockers are in [UNPARK_REPORT.md](UNPARK_REPORT.md).

The port is rebased onto requested harness 59451e23e, including multi-edit
finding support fca1616e6. It reports every upstream automatic edit through
RuleContext.reportRange's edits array. The former explicit refusal and manual
single-edit Finding construction are removed. No shared harness file changed.
The former blocked input is now an owned positive witness:

```typescript
function f() { var a; var b; }
```

The oracle's SourceFile listener gathers one-var findings before ordinary
node listeners. The initial all-rule comparison exposed a tie-order difference
against no-var. The native rule now preserves that collection order in its
own reports while still receiving only VariableDeclarationList nodes from the
kind-indexed registry. The full TestOneVar prefix remains unchanged and captures
all sixteen upstream test functions. Messages are unchanged from batch 3.

## Final evidence

TestRulesAgree PASS 52.21s: 2357 captured source/rule/options combinations,
13,209,539 byte-identical Go, original Node, emitted JavaScript and sanitized
native outputs. This includes findings, all automatic edit ranges/text,
suggestions and fixed-source results. The inherited malformed method-signature
controls retain the harness's explicit native/Node recovery-refusal checks.
No corpus entry or option guard was removed or relaxed.

TestOwnedWitnesses PASS 22.88s: 120,291 byte-identical outputs across those four
sides, selected and all-rule runs, including the two-edit join witness and the
original nonadjacent declaration witness. Combined parity command exits 0 in
75.104s.

TestMutants PASS 774.555s: all 41 registered mutants, 123 successful
side comparisons. The owned one-var-combine-message mutant changes only the
combine message, compiles and exits normally. Independent Go comparison catches
it on original Node, emitted JavaScript and sanitized native at case30 line507:
`Mutant combine 'var' statement.` versus `Combine this with the previous 'var'
statement.` It is not a compiler error or sanitizer crash.

Registry generation, gofmt and full go vet exit 0; formatting and vet logs are
empty. The first failed ordering run was superseded after its TestRulesAgree
failure; its remaining mutant work was interrupted, and its log is retained.
The final full mutant run began after the source correction and is the mutation
certification. Source hashes and the successful logs are under validation/multifix.
The earlier multi-edit refusal evidence remains historical in BLOCKER.md.

Commands, with /workspace/adamic-tools/env.sh sourced:

```sh
go run ./cmd/lint-registry
gofmt -l stage1/cohere/lint/rules/one-var
go vet ./...
go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses)$' -count=1 -v -timeout 30m > /workspace/one-var-multifix-parity2.log 2>&1
go test ./stage1/cohere/lint -run '^TestMutants$' -count=1 -v -timeout 30m > /workspace/one-var-multifix-mutants-final.log 2>&1
```

Uncovered: full repository gate, the seventeen-check required-input gate,
separate frozen compiler/repository populations, isolated native/Go throughput
and suppression application. This unit certifies the unified harness's complete
captured corpus, generated controls and owned witnesses; it makes no broader
performance claim. No checker handles are used by this syntax rule.
