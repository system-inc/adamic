# SourceFile finding order

Area checked: `334509eea8a49b8085187e206c495cc6aa24c5c4`.
Upstream Go pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`.

`input.ts.txt` selects all registered rules. Both diagnostics start at byte zero.
Go emits max-classes-per-file before @typescript-eslint/no-import-type-side-effects.
The unmodified area emits them in reverse order on Node, emitted JavaScript,
and sanitized native; the complete outputs are `area-{go,node,emitted,native}.txt`.
Those three port outputs are identical to one another. The fix moves class counting
and reporting into the SourceFile listener, matching Go. `fixed-all-runtimes.txt`
is identical to `area-go.txt` on all three runtimes.

`fixed.log` proves all 29 captured upstream max-classes-per-file cases, its owned
witnesses alone and with all rules, the package's complete owned witness suite,
and mutant `one excess class is silently allowed` (Node, emitted JavaScript,
sanitized native built with ASan/UBSan).

The two saved Go test sources are temporary harness adapters, not package code.
To replay, copy `area-proof.go.txt` into the area's lint package as
`max_order_review_test.go`, then run
`go test ./stage1/cohere/lint -run TestMaxClassesAreaOrderProof -count=1 -v`.
On this fix branch, copy `fixed-proof.go.txt` into that same temporary path and run
`go test ./stage1/cohere/lint -run 'TestMaxClassesOwnParity|TestOwnedWitnesses|TestMutants/one_excess_class_is_silently_allowed' -count=1 -v`.
Use the configured cloud tool environment and GOPROXY. For an isolated worktree
whose cohere checkout is a symlink, set `GOFLAGS=-buildvcs=false`.
The area replay creates `/tmp/imports-max-order-proof/input.ts` and saves full
outputs there; create that directory before the replay. Remove the temporary Go
test after replay. The committed witness is `../../testdata/source-file-order.ts.txt`.

`organized-area-failure.txt` records the organized-imports branch's combined-witness
failure on the uncorrected area. Both new rule branches pass upstream parity, their
mutants, TestJsxInventoryDiscovery, and TestJsxLintTrees with the area's discovery.
Organized imports' own history and diff contain no max-classes change.
`integrated.log` proves TestOwnedWitnesses, JSX discovery/tree checks, and all three
relevant mutants after independently merging both rule branches and this correction
onto the area. The report-order discrepancy is fixed without bundling another rule.
