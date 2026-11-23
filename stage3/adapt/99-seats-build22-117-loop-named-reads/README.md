# Family 117 diagnostic related-information loop

At the Nov 22 meter site `program.ts:804:18` in unit `program.ts:785:1`,
`formatDiagnosticsWithColorAndContext` destructures each item of
`diagnostic.relatedInformation`. The adapter changes that loop to bind the
item as `info` and read `file`, `start`, `length`, and `messageText` in
pattern order. The four `const` bindings are fresh per iteration and all four
reads precede the original body. The syntax-tree check verifies that `info`
does not occur in the enclosing function before the edit.

Adamic's Nov 22 ruling on #bbssdcy approved the named-read adaptation with
one fresh binding per iteration. On Nov 23 Adamic confirmed the getter audit
covers tsc's own producers. Diagnostic records use data properties. The
utility factories construct the records, `addRelatedInfo` appends them,
`attachFileToDiagnostic` attaches locations, and the checker and builder
collect or reconstruct them.

No ruling row governs an object binding on an array item. Rows 205 and 206
concern the iteration protocol. This edit uses ordinary named reads in the
existing checked read path. Row 395 states the applicable meaning:

> - **Every Adamic program prints exactly what it prints when JavaScript runs it; the listed divergences are the exceptions (Weak, panics inside `try`, typed-array writes, the enum initialization ReferenceError, WASI).** Meaning. Undated, docs; source: README.md:62, docs/memory.md:316-319,450, docs/typed-arrays.md:33-39, docs/verification/enum-init-reach.md:25-30. Spec: Semantics that differ from JavaScript.
