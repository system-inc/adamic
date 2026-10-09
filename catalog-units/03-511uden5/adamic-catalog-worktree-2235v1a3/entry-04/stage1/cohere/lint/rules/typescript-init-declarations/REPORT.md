@typescript-eslint/init-declarations now runs through the unified kind-indexed harness. Its descriptor subscribes to VariableDeclarationList, declares node: true and takes the handed ParseNode. The typed Go oracle adapter decodes field 5 to InitDeclarationsOptions, retaining the upstream inert default when no mode is configured. All new Adamic modules are .a.

The port starts from the ledger's batch4-typescript copy and compares the differing batch4 copy with the pinned Go implementation. Its initializer scan stays in initializer.a because only this rule uses it. It scans between declaration children, so spaces, comments, definite-assignment tokens and type annotations do not hide an initializer. Ambient declarations are detected through retained declare modifiers using existing RuleContext.has. Loop heads count as initialized, loop bodies retain their own behavior, and const/using/await using are exempt under never. The two message builders preserve upstream text verbatim. Always reports the name alone; never reports the complete declarator. The rule has no fixes or suggestions.

The supplied witness-options commit 29c41e102 is included unchanged as a prerequisite, cherry-picked as 7880e534b because area had not merged it at branch creation. All authored changes live in this rule directory; the shared dependency's five files were not edited. Integration may drop the duplicate prerequisite when it lands upstream. Source branch hashes and all seven actual upstream Test names are recorded in evidence/provenance.json. The prefix TestInitDeclarations includes both imported pass/fail tests, destructuring, constant bindings, both decoder tests and inert-default tests. The captured source corpus covers 51 passing inputs, 26 failing inputs, seven destructuring cases, four constant-binding cases and four default/explicit-mode controls. The independent upstream decoder assertions also run.

Three firing witnesses carry the decoded upstream Go options in sidecars: always, never and never with IgnoreForLoopInit. They exercise annotations, whitespace/comments, a Unicode name, loop heads versus bodies, ambient namespaces and constant/resource bindings. Each own-rule row must fire in Go; all-rule rows keep default options. No shared registry list or oracle guard is changed.

Validation commands, with output written directly to the committed compressed logs:

```sh
bash cloud/setup.sh > /tmp/init-declarations-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry > /tmp/init-declarations-registry-final.log 2>&1
gofmt -l stage1/cohere/lint/rules/typescript-init-declarations/oracle.go > /tmp/init-declarations-gofmt.log
go vet ./stage1/cohere/lint/... > /tmp/init-declarations-vet-final.log 2>&1
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$' -count=1 -timeout=30m -v > /tmp/init-declarations-final.log 2>&1
# Run inside cohere:
go test ./internal/lint/rules/typescript -run '^TestInitDeclarations' -count=1 -timeout=10m -v > /tmp/init-declarations-upstream.log 2>&1
```

Setup reports Go, clang, Node and submodules ready in 0s each; cache warm 49s; total 49s on nproc 5, cpu.max 400000 100000. The independent production cohere formatter formatted all three owned Adamic modules; its checker then reported zero findings. gofmt and vet logs are empty. Registry generation discovers 41 rules. All requested tests pass, with no skipped selected tests, and all seven upstream functions pass independently in 0.015s.

--- PASS: TestRulesAgree (50.71s)
--- PASS: TestMutants (727.47s)
--- PASS: TestOwnedWitnesses (21.39s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint	799.592s

Go, source Node, emitted JavaScript and sanitized native match 13068145, 94587 serialized bytes for the upstream/corner-case corpus and owned witnesses respectively. Full descriptions, byte ranges, finding order and unchanged fixed-source output are compared. Existing explicit parser-recovery refusals in other rules retain the harness assertions; no new exclusion is introduced for this rule.

The owned mutant replaces the always report target nameIndex with declarationIndex. It compiles and runs successfully on all three ports; only the byte comparison catches the range widening. For pending: number, all three ports report range 4 19 instead of Go's 4 11. The focused mutant proof and the complete 41-mutant matrix both pass; every matrix name and exact mismatches are retained in evidence.

No known checker, dynamic-RegExp, parser or Tailwind blocker was hit by this rule. The full repository gate, full compiler-source corpus and performance benchmarks were not run for this isolated port.
