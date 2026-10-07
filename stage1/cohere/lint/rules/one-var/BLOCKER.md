# Historical one-var multi-edit blocker

This historical blocker is resolved by harness fca1616e6 and the rule's
reportRange migration on base 59451e23e. See REPORT.md for current validation.
The earlier failure evidence below is retained as history.

Source: origin/codex/stage1-lint-batch3, rules/one_var.ts, as named by
DEDUP_LEDGER.md. Messages were moved verbatim; helpers used only by this rule
are local. Existing RuleContext parent, scanner, has and functionLike methods
provide shared functionality without context edits. rule.json uses named
VariableDeclarationList listeners, node:true and the full TestOneVar prefix
(all sixteen real upstream test functions). Its adapter returns decoded
upstream OneVarSettings, preserving absent/null defaults rather than nil.

This is a partial migration, not a certified port. The area at d3a37422c has
one automatic edit per finding. Upstream OneVar returns multiple edits when
joining adjacent declarations. Reproducer, default options:

```typescript
function f() { var a; var b; }
```

Go's joinFixes returns an edit changing the previous terminator to a comma and
an edit removing the second var keyword. stage1/cohere/lint/testdata/oracle.go:75
panics `unexpected fix shape` when serializing these. Finding has no extraFixes
field, unlike batch 3. The migrated rule explicitly refuses multiple edits
instead of silently dropping them. No guard, serializer or test was modified.

The no-fix positive witness is `var first; consume(); var second;`.
The message mutant is declared but NOT proven caught: the unified TestMutants
fails in its independent Go oracle before compiling or executing the mutant.
The full TestOneVar prefix was not narrowed to hide repair cases.

Commands (output files are committed under validation):

```
go run ./cmd/lint-registry
gofmt -w stage1/cohere/lint/rules/one-var/oracle.go
go vet ./...
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants/one-var-combine-message)$' -count=1 -v -timeout 30m
go test ./stage1/cohere/lint -run '^TestMutants$/one-var-combine-message' -count=1 -v -timeout 30m
```

Registry and vet exit 0. TestRulesAgree FAIL 180.07s (2357 captured combinations),
TestOwnedWitnesses FAIL 21.85s, targeted TestMutants FAIL 8.165s: each oracle
exits 2 with the same fix-shape panic. The first combined filter does not select
the targeted mutant, so its separate command is the mutation evidence.
Setup succeeds in 123s (Go0, clang1, Node1, submodules1, cache123); nproc5,
cgroup quota4. No full repository gate, all-mutant gate, complete byte parity,
source/native/emitted-JavaScript witness agreement, sanitizer check or native/Go
timing is claimed. Work stops at the shared automatic-multi-edit model blocker.
Witness-options 29c41e102 is not on this area; the witness needs no options.
